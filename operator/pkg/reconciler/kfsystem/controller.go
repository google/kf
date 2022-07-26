// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kfsystem

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"sync"

	operatorclient "kf-operator/pkg/client/injection/client"
	kfsysteminformer "kf-operator/pkg/client/injection/informers/kfsystem/v1alpha1/kfsystem"
	operandinformer "kf-operator/pkg/client/injection/informers/operand/v1alpha1/operand"
	kfreconciler "kf-operator/pkg/client/injection/reconciler/kfsystem/v1alpha1/kfsystem"
	mfinject "kf-operator/pkg/manifestival/injection/client"
	"kf-operator/pkg/operand"
	kfoperand "kf-operator/pkg/operand/kf"
	"kf-operator/pkg/reconciler/kfsystem/kf"
	"kf-operator/pkg/transformer"

	semver "github.com/Masterminds/semver/v3"
	"github.com/go-logr/zapr"
	mf "github.com/manifestival/manifestival"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	podinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/pod"
	secretinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/secret"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
)

const (
	kfPath = "kf"
)

// NewController registers eventhandlers to enqueue events.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	mfClient := mfinject.Get(ctx)
	logger := logging.FromContext(ctx)
	logger.Info("Extracting informers and clients")
	kfSystemInformer := kfsysteminformer.Get(ctx)
	operandInformer := operandinformer.Get(ctx)
	kubeClient := kubeclient.Get(ctx)
	secretInformer := secretinformer.Get(ctx)
	podInformer := podinformer.Get(ctx)
	client := operatorclient.Get(ctx)
	a, err := transformer.GetPodAnnotationsFromEnv()
	if err != nil {
		panic(fmt.Sprintf("Unable to create operand factory from environment %+v", err))
	}
	of := operand.CreateFactory(transformer.Annotation{a})

	var lock sync.Mutex
	c := &Reconciler{
		kubeClient: kubeClient,
		client:     client,
		reconcilers: []kfoperand.Interface{
			NewKfReconciler(
				ctx,
				of,
				FindAvailableVersions(kfPath),
				KfLookup(ctx, mfClient),
				&lock,
				secretInformer.Lister(),
				podInformer.Lister(),
			),
		},
	}

	impl := kfreconciler.NewImpl(ctx, c)

	logger.Info("Setting up event handlers")
	kfSystemInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	secretInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			if object, ok := obj.(metav1.Object); ok {
				return object.GetNamespace() == kfPath
			}
			return false
		},
		Handler: controller.HandleAll(func(interface{}) { impl.GlobalResync(kfSystemInformer.Informer()) }),
	})

	// When the KubeRun CRD is used, and is an Owner of Operands, we should rely on filtering.
	operandInformer.Informer().AddEventHandler(controller.HandleAll(impl.EnqueueLabelOfClusterScopedResource(kfoperand.OwnerName)))

	return impl
}

// AvailableVersionsOrDie finds all semantic version direcctories under the given root.
func AvailableVersionsOrDie(dir string) []*semver.Version {
	ds, err := ioutil.ReadDir(filepath.Join(os.Getenv("KO_DATA_PATH"), dir))
	if err != nil {
		panic(err)
	}
	vs := make([]*semver.Version, len(ds))
	for i, d := range ds {
		vs[i], err = semver.NewVersion(d.Name())
		if err != nil {
			panic(err)
		}
	}
	sort.Sort(semver.Collection(vs))
	return vs
}

// NewKfReconciler creates a new reconciler for the Kf component.
func NewKfReconciler(
	ctx context.Context,
	of operand.Factory,
	versions []*semver.Version,
	kfManifest func(string) (*mf.Manifest, error),
	lock *sync.Mutex,
	secretLister v1listers.SecretLister,
	podLister v1listers.PodLister,
) kfoperand.Interface {
	return kfoperand.CreateOperandReconciler(
		ctx,
		"kf",
		kfoperand.MakeSafe(
			&kf.Reconciler{
				Factory:      of,
				Versions:     versions,
				Lookup:       kfManifest,
				SecretLister: secretLister,
				PodLister:    podLister,
			},
			lock,
		),
	)
}

func getManifestString(ctx context.Context, path string) string {
	content, err := ioutil.ReadFile(manifestPath(path))
	if err != nil {
		panic(fmt.Errorf("getManifestString for path: %v, error: %v", path, err))
	}
	return string(content)
}

func getManifestOrExit(ctx context.Context, path string, mfClient mf.Client) *mf.Manifest {
	manifest, err := getManifest(ctx, path, mfClient)
	if err != nil {
		panic(fmt.Errorf("getManifestOrExit for path: %v, error: %v", path, err))
	}
	return manifest
}

func getManifest(ctx context.Context, path string, mfClient mf.Client) (*mf.Manifest, error) {
	logger := logging.FromContext(ctx)

	manifest, err := mf.NewManifest(manifestPath(path), manifestOptions(mfClient, logger)...)
	if err != nil {
		logger.Errorf("error creating the Manifest for %v: %v", path, err)
		return nil, err
	}

	return &manifest, nil
}

func manifestOptions(mfClient mf.Client, logger *zap.SugaredLogger) []mf.Option {
	return []mf.Option{
		mf.UseClient(mfClient),
		mf.UseLogger(zapr.NewLogger(logger.Desugar()).WithName("manifestival")),
	}
}

func manifestPath(path string) string {
	koDataDir := os.Getenv("KO_DATA_PATH")
	return filepath.Join(koDataDir, path+"/")
}

// FindAvailableVersions finds all semantic version directories under the given root.
func FindAvailableVersions(dir string) []*semver.Version {
	ds, err := ioutil.ReadDir(filepath.Join(os.Getenv("KO_DATA_PATH"), dir))
	if err != nil {
		panic(err)
	}
	vs := make([]*semver.Version, len(ds))
	for i, d := range ds {
		vs[i], err = semver.NewVersion(d.Name())
		if err != nil {
			panic(err)
		}
	}
	sort.Sort(semver.Collection(vs))
	return vs
}

// KfLookup looks up a particular minor version of kf.
func KfLookup(ctx context.Context, mfClient mf.Client) func(string) (*mf.Manifest, error) {
	return func(version string) (*mf.Manifest, error) {
		return getManifest(ctx, filepath.Join(kfPath, version), mfClient)
	}
}
