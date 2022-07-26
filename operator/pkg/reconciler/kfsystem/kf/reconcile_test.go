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

package kf_test

import (
	"context"
	"sync"
	"testing"

	operatorclient "kf-operator/pkg/client/injection/client"
	kfsystemreconciler "kf-operator/pkg/client/injection/reconciler/kfsystem/v1alpha1/kfsystem"
	kfsystem "kf-operator/pkg/reconciler/kfsystem"
	kftesting "kf-operator/pkg/reconciler/testing/kf"
	optest "kf-operator/pkg/testing/operand/v1alpha1"

	"github.com/Masterminds/semver/v3"
	mf "github.com/manifestival/manifestival"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	v1listers "k8s.io/client-go/listers/core/v1"
	clientgotesting "k8s.io/client-go/testing"
	ktesting "k8s.io/client-go/testing"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/ptr"
	rtesting "knative.dev/pkg/reconciler/testing"

	. "kf-operator/pkg/reconciler/testing"
	. "kf-operator/pkg/testing/k8s"
	. "kf-operator/pkg/testing/kfsystem/v1alpha1"
	. "kf-operator/pkg/testing/manifestival"
)

var (
	availableVersions = []*semver.Version{
		semver.MustParse("2.3"),
		semver.MustParse("2.4"),
		semver.MustParse("2.5"),
	}
)

func TestReconcile(t *testing.T) {
	now := metav1.Now()

	table := rtesting.TableTest{
		{
			Name: "controller not ready",
			Key:  "test/test-obj",
			Objects: []runtime.Object{
				KfSystemWithDefaults("test-obj", WithKfEnabled("0.1")),
			},
			WantCreates: []runtime.Object{
				optest.Operand("kf", optest.WithClusterOwner("test-obj"), optest.WithCheckDeploymentHealth, optest.WithSteadyState(t,
					Deployment("controller"), Deployment("webhook"))),
			},
			WantPatches: []clientgotesting.PatchActionImpl{
				KfSystemPatchFinalizer("test-obj"),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: KfSystemWithDefaults("test-obj", WithKfEnabled("0.1"), WithTargetKfVersion("0.1"), WithKfInstallNotReady),
			}},
			WantEvents: []string{
				rtesting.Eventf(corev1.EventTypeNormal, "FinalizerUpdate", "Updated %q finalizers", "test-obj"),
			},
			CmpOpts:                 CommonOptions,
			SkipNamespaceValidation: true,
			Ctx:                     withManifestResources(Deployment("controller"), Deployment("webhook")),
		},
		{
			Name: "install kf",
			Key:  "test/test-obj",
			Objects: []runtime.Object{
				KfSystemWithDefaults("test-obj", WithKfEnabled("0.1"), WithControllerCACerts("cacerts")),
				immutableSecret("cacerts"),
				optest.Operand("kf", optest.WithClusterOwner("test-obj"), optest.WithCheckDeploymentHealth,
					optest.WithSteadyState(t, Deployment(
						"controller",
						WithDeploymentVolumes(Volume("cacerts", WithVolumeSecretSource(&corev1.SecretVolumeSource{SecretName: "cacerts"}))),
						WithDeploymentContainer(Container("some-container", "some-image")),
					),
						Deployment("test-deployment"),
						Deployment("webhook"),
					),
					optest.WithLatestActiveOperandCreated("sha256-4f53cda18c"),
					optest.WithLatestActiveOperandReady("sha256-4f53cda18c"),
					optest.WithOperandInstallSuccessful()),
			},
			WantPatches: []clientgotesting.PatchActionImpl{
				KfSystemPatchFinalizer("test-obj"),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: KfSystemWithDefaults("test-obj", WithKfEnabled("0.1"), WithTargetKfVersion("0.1"), WithKfInstallSucceeded("0.1"), WithControllerCACerts("cacerts")),
			}},
			WantEvents: []string{
				rtesting.Eventf(corev1.EventTypeNormal, "FinalizerUpdate", "Updated %q finalizers", "test-obj"),
			},
			SkipNamespaceValidation: true,
			Ctx: withManifestResources(
				Deployment("controller", WithDeploymentContainer(Container("some-container", "some-image"))),
				Deployment("test-deployment"),
				Deployment("webhook")),
		},
		{
			Name: "failed apply",
			Key:  "test/test-obj",
			Objects: []runtime.Object{
				KfSystemWithDefaults("test-obj", WithKfEnabled("0.1")),
			},
			WantPatches: []clientgotesting.PatchActionImpl{
				KfSystemPatchFinalizer("test-obj"),
			},
			WantCreates: []runtime.Object{
				optest.Operand("kf", optest.WithClusterOwner("test-obj"), optest.WithCheckDeploymentHealth, optest.WithSteadyState(t, Deployment("test-deployment"))),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: KfSystemWithDefaults("test-obj", WithKfEnabled("0.1"), WithTargetKfVersion("0.1"), WithKfInstallFailed("inducing failure for create operands")),
			}},
			WantEvents: []string{
				rtesting.Eventf(corev1.EventTypeNormal, "FinalizerUpdate", "Updated %q finalizers", "test-obj"),
				rtesting.Eventf(corev1.EventTypeWarning, "InternalError", "inducing failure for create operands"),
			},
			WithReactors: []ktesting.ReactionFunc{
				rtesting.InduceFailure("create", "operands"),
			},
			CmpOpts:                 CommonOptions,
			SkipNamespaceValidation: true,
			WantErr:                 true,
			Ctx:                     withManifestResources(Deployment("test-deployment")),
		},
		{
			Name: "finalize - delete Operand",
			Key:  "test/test-obj",
			Objects: []runtime.Object{
				Deployment("test-deployment"),
				KfSystemWithDefaults("test-obj", WithKfEnabled("0.1"), WithKfSystemFinalizer, WithDeletionTimestamp(&now)),
				optest.Operand("kf", optest.WithCheckDeploymentHealth, optest.WithSteadyState(t, Deployment("controller")), optest.WithLatestActiveOperandCreated("sha256-4f53cda18c"), optest.WithLatestActiveOperandReady("sha256-4f53cda18c"), optest.WithOperandInstallSuccessful()),
			},
			WantPatches: []clientgotesting.PatchActionImpl{
				RemoveFinalizerAction("test-obj"),
			},
			WantEvents: []string{
				rtesting.Eventf(corev1.EventTypeNormal, "FinalizerUpdate", "Updated %q finalizers", "test-obj"),
			},
			WantDeletes: []clientgotesting.DeleteActionImpl{
				{
					ActionImpl: clientgotesting.ActionImpl{
						Verb:     "delete",
						Resource: corev1.SchemeGroupVersion.WithResource("operands"),
					},
					Name: "kf",
				},
			},
			SkipNamespaceValidation: true,
			Ctx:                     withManifestResources(Deployment("test-deployment")),
		},
		{
			Name: "getting certs secret fails",
			Key:  "test/test-obj",
			Objects: []runtime.Object{
				KfSystemWithDefaults("test-obj", WithKfEnabled("0.1"), WithControllerCACerts("cacerts")),
			},
			CmpOpts:                 CommonOptions,
			SkipNamespaceValidation: true,
			WantErr:                 true,
			Ctx:                     withManifestResources(Deployment("test-deployment")),
			WantPatches: []clientgotesting.PatchActionImpl{
				KfSystemPatchFinalizer("test-obj"),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: KfSystemWithDefaults("test-obj", WithKfEnabled("0.1"), WithTargetKfVersion("0.1"), WithControllerCACerts("cacerts"), WithKfInstallFailed(`Failed to create operand secret "cacerts" not found`)),
			}},
			WantEvents: []string{
				rtesting.Eventf(corev1.EventTypeNormal, "FinalizerUpdate", "Updated %q finalizers", "test-obj"),
				rtesting.Eventf(corev1.EventTypeWarning, "InternalError", `secret "cacerts" not found`),
			},
		},
		{
			Name: "certs secret is mutable - nil",
			Key:  "test/test-obj",
			Objects: []runtime.Object{
				KfSystemWithDefaults("test-obj", WithKfEnabled("0.1"), WithControllerCACerts("cacerts")),
				nilMutableSecret("cacerts"),
			},
			CmpOpts:                 CommonOptions,
			SkipNamespaceValidation: true,
			WantErr:                 true,
			Ctx:                     withManifestResources(Deployment("test-deployment")),
			WantPatches: []clientgotesting.PatchActionImpl{
				KfSystemPatchFinalizer("test-obj"),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: KfSystemWithDefaults("test-obj", WithKfEnabled("0.1"), WithTargetKfVersion("0.1"), WithControllerCACerts("cacerts"), WithKfInstallFailed(`Failed to create operand secret kf/cacerts must be immutable`)),
			}},
			WantEvents: []string{
				rtesting.Eventf(corev1.EventTypeNormal, "FinalizerUpdate", "Updated %q finalizers", "test-obj"),
				rtesting.Eventf(corev1.EventTypeWarning, "InternalError", `secret kf/cacerts must be immutable`),
			},
		},
		{
			Name: "certs secret is mutable - nil",
			Key:  "test/test-obj",
			Objects: []runtime.Object{
				KfSystemWithDefaults("test-obj", WithKfEnabled("0.1"), WithControllerCACerts("cacerts")),
				mutableSecret("cacerts"),
			},
			CmpOpts:                 CommonOptions,
			SkipNamespaceValidation: true,
			WantErr:                 true,
			Ctx:                     withManifestResources(Deployment("test-deployment")),
			WantPatches: []clientgotesting.PatchActionImpl{
				KfSystemPatchFinalizer("test-obj"),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: KfSystemWithDefaults("test-obj", WithKfEnabled("0.1"), WithTargetKfVersion("0.1"), WithControllerCACerts("cacerts"), WithKfInstallFailed(`Failed to create operand secret kf/cacerts must be immutable`)),
			}},
			WantEvents: []string{
				rtesting.Eventf(corev1.EventTypeNormal, "FinalizerUpdate", "Updated %q finalizers", "test-obj"),
				rtesting.Eventf(corev1.EventTypeWarning, "InternalError", `secret kf/cacerts must be immutable`),
			},
		},
	}

	factory := MakeFactory(func(ctx context.Context, listers *Listers) controller.Reconciler {
		manifest := FakeManifestivalOrDie(ctx)
		secretLister := v1listers.NewSecretLister(listers.IndexerFor(&corev1.Secret{}))
		podLister := v1listers.NewPodLister(listers.IndexerFor(&corev1.Pod{}))

		reconciler := kfsystem.NewKfReconciler(ctx,
			optest.TestFactory,
			availableVersions,
			func(version string) (*mf.Manifest, error) {
				return manifest, nil
			},
			&sync.Mutex{},
			secretLister,
			podLister,
		)

		return kfsystemreconciler.NewReconciler(ctx,
			logging.FromContext(ctx),
			operatorclient.Get(ctx),
			listers.GetKfSystemLister(),
			controller.GetEventRecorder(ctx),
			kftesting.FakeTestReconciler(reconciler),
		)
	})

	table.Test(t, factory)
}

func withManifestResources(objs ...runtime.Object) context.Context {
	return WithManifestResources(context.Background(), objs...)
}

func immutableSecret(name string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "kf",
		},
		Immutable: ptr.Bool(true),
	}
}

func nilMutableSecret(name string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "kf",
		},
	}
}

func mutableSecret(name string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "kf",
		},
		Immutable: ptr.Bool(false),
	}
}
