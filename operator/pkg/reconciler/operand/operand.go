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

package operand

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"sort"
	"sync"
	"time"

	"kf-operator/pkg/apis/operand/v1alpha1"
	opclient "kf-operator/pkg/client/clientset/versioned/typed/operand/v1alpha1"
	operandreconciler "kf-operator/pkg/client/injection/reconciler/operand/v1alpha1/operand"
	operandlisters "kf-operator/pkg/client/listers/operand/v1alpha1"

	poperand "kf-operator/pkg/operand"

	"knative.dev/pkg/ptr"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/logging"
	pkgreconciler "knative.dev/pkg/reconciler"

	"github.com/hashicorp/go-multierror"
)

// reconciler implements controller.reconciler for Operand resources.
type reconciler struct {
	lock                       *sync.Mutex
	resourceReconciler         poperand.ResourceReconciler
	operandClient              opclient.OperandV1alpha1Interface
	clusterActiveOperandLister operandlisters.ClusterActiveOperandLister
	enqueueAfter               func(interface{}, time.Duration)
}

var (
	algorithm       = "sha256"
	reconcilePeriod time.Duration

	// OwnerName labels the owner of ClusterActiveOperand.
	OwnerName = "clusteractiveoperand.kuberun.cloud.google.com/owner-name"
)

func init() {
	flag.DurationVar(&reconcilePeriod, "operand_reconcile_period", time.Minute, "Period with which to reconcile the operand CR when there are no changes to it. Defaults to 1 minute")
}

// Check that our reconciler implements operandreconciler.Interface
var _ operandreconciler.Interface = (*reconciler)(nil)

func deleteOptions() metav1.DeleteOptions {
	do := &metav1.DeleteOptions{GracePeriodSeconds: ptr.Int64(30)}
	policy := metav1.DeletePropagationForeground
	do.PropagationPolicy = &policy
	return *do
}

type sortLiveRefs []v1alpha1.LiveRef

func (a sortLiveRefs) Len() int {
	return len(a)
}

func (a sortLiveRefs) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a sortLiveRefs) Less(i, j int) bool {
	if a[i].Group != a[j].Group {
		return a[i].Group < a[j].Group
	}
	if a[i].Kind != a[j].Kind {
		return a[i].Kind < a[j].Kind
	}
	if a[i].Namespace != a[j].Namespace {
		return a[i].Namespace < a[j].Namespace
	}
	return a[i].Name < a[j].Name
}

func (r *reconciler) ReconcileKind(ctx context.Context, o *v1alpha1.Operand) pkgreconciler.Event {
	defer r.enqueueAfter(o, reconcilePeriod)

	reconcileErr := r.doReconcileKind(ctx, o)
	if _, err := r.operandClient.Operands().UpdateStatus(ctx, o, metav1.UpdateOptions{}); err != nil {
		logging.FromContext(ctx).Errorw("error updating operand status", "error", err)
	}
	return reconcileErr
}

func (r *reconciler) doReconcileKind(ctx context.Context, o *v1alpha1.Operand) pkgreconciler.Event {
	logger := logging.FromContext(ctx)
	o.Status.InitializeConditions()
	o.Status.ObservedGeneration = o.GetGeneration()

	logger.Info("Calculating GVRs")
	refMap, err := r.deriveLiveRefs(o)
	if err != nil {
		return err
	}

	refs := make([]v1alpha1.LiveRef, 0, len(refMap))
	for ref := range refMap {
		refs = append(refs, ref)
	}
	sort.Sort(sortLiveRefs(refs))

	j, err := json.Marshal(refs)
	if err != nil {
		return err
	}
	hash := sha256.Sum256(j)
	name := fmt.Sprintf("%s-%s", algorithm, hex.EncodeToString(hash[:])[:10])

	logger.Infof("Configured for CAO %s", name)

	// TODO: Remove this for future releases. http://b/201351749
	if err := r.InjectOwnerLabels(ctx); err != nil {
		return err
	}

	if err := r.createClusterActiveOperandIfNecessary(ctx, o, name, refs...); err != nil {
		return err
	}

	// Perform SteadyState install
	if o.Generation != o.Status.InstalledSteadyStateGeneration {
		o.Status.InstalledSteadyStateGeneration = 0
		if result, err := r.applyAndGetState(ctx, o.Spec.SteadyState, o.Spec.CheckDeploymentHealth); err != nil {
			o.Status.MarkOperandInstallFailed(err)
			return err
		} else if !result.IsReady() {
			o.Status.MarkOperandInstallNotReady(result.String())
			return nil
		}
		o.Status.InstalledSteadyStateGeneration = o.Generation
	}

	// Perform PostInstall
	allResources := make([]unstructured.Unstructured, 0, len(refMap))
	for _, u := range refMap {
		allResources = append(allResources, u)
	}
	if result, err := r.applyAndGetState(ctx, allResources, o.Spec.CheckDeploymentHealth); err != nil {
		o.Status.MarkOperandPostInstallFailed(err)
		return err
	} else if !result.IsReady() {
		o.Status.MarkOperandPostInstallNotReady(result.String())
		return nil
	}
	o.Status.MarkOperandInstallSuccessful()

	r.lock.Lock()
	defer r.lock.Unlock()

	ao, err := r.operandClient.ClusterActiveOperands().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to find CAO %s [%v] \n Will be creating a new one", name, err)
	}

	if !ao.Status.IsReady() {
		return nil
	}

	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(OwnerName, selection.Equals, []string{o.Name})
	if err != nil {
		return fmt.Errorf("failed to create requirement for operand %s labels [%v]", o.Name, err)
	}
	selector = selector.Add(*requirement)

	// We're ready, clean up previous CAOs.
	clusterActiveOperands, err := r.clusterActiveOperandLister.List(selector)
	if err != nil {
		return fmt.Errorf("failed to list clusteractiveoperands for operand %s [%v] ", o.Name, err)
	}

	for _, cao := range clusterActiveOperands {
		if cao.Name != name {
			logger.Infof("Deleting CAO %s in favor of newer CAO %s", cao.Name, name)
			if err := r.operandClient.ClusterActiveOperands().Delete(ctx, cao.Name, deleteOptions()); err != nil && !apierrors.IsNotFound(err) {
				return fmt.Errorf("deleting the previous CAO %s caused error %w", cao.Name, err)
			}
		}
	}

	o.Status.MarkLatestActiveOperandReady(name)
	return nil
}

type state interface {
	fmt.Stringer
	IsReady() bool
}

type healthcheckState string

func (s healthcheckState) IsReady() bool {
	return s == poperand.Installed
}

func (s healthcheckState) String() string {
	return (string)(s)
}

// This function handles the transition to the version of operand controller that injects Owner label.
// Remove this function after next version
func (r reconciler) InjectOwnerLabels(ctx context.Context) error {
	logger := logging.FromContext(ctx)

	caos, err := r.clusterActiveOperandLister.List(labels.Everything())
	if err != nil {
		return fmt.Errorf("error fetching caos when trying to inject owner labels, %w", err)
	}

	for _, cao := range caos {
		// existing caos have no labels.
		if cao.Labels != nil {
			continue
		}
		if len(cao.OwnerReferences) != 1 {
			logger.Warnf("cao %s has zero or more than one ownerreferences, skipping injecting owner label")
			continue
		}

		// Do not modify cached objects.
		copy := cao.DeepCopy()
		// Inject owner label to ClusterActiveOperand if it doesn't exist.
		copy.Labels = make(map[string]string)
		copy.Labels[OwnerName] = copy.OwnerReferences[0].Name

		_, err = r.operandClient.ClusterActiveOperands().Update(ctx, copy, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to inject labels to cao %s, %v", copy.Name, err)
		}

		logger.Infof("injected owner label to cao %s", copy.Name)
	}

	return nil
}

type noHealthcheckState string

func (s noHealthcheckState) IsReady() bool {
	return s == poperand.Installed || s == HealthCheckDeploymentsNotReady
}

func (s noHealthcheckState) String() string {
	return (string)(s)
}

// applyAndGetState applies the given resources and returns their state.
func (r reconciler) applyAndGetState(ctx context.Context, resources []unstructured.Unstructured, healthcheckDeployemnts bool) (state, error) {
	if err := r.resourceReconciler.Apply(ctx, resources); err != nil {
		return nil, err
	}

	if result, err := r.resourceReconciler.GetState(ctx, resources); err != nil {
		return nil, err
	} else if healthcheckDeployemnts {
		return healthcheckState(result), nil
	} else {
		return noHealthcheckState(result), nil
	}
}

func (r reconciler) deriveLiveRefs(operand *v1alpha1.Operand) (map[v1alpha1.LiveRef]unstructured.Unstructured, error) {
	refs := make(map[v1alpha1.LiveRef]unstructured.Unstructured)
	// First calculate our latest ActiveOperand
	var result *multierror.Error
	for _, unstr := range append(operand.Spec.SteadyState, operand.Spec.PostInstall...) {
		if ref, err := r.makeRef(unstr); err != nil {
			result = multierror.Append(result, err)
		} else {
			refs[*ref] = unstr
		}
	}
	return refs, result.ErrorOrNil()
}

func (r reconciler) makeRef(u unstructured.Unstructured) (*v1alpha1.LiveRef, error) {
	gvk := u.GroupVersionKind()
	return &v1alpha1.LiveRef{
		Group:     gvk.Group,
		Kind:      gvk.Kind,
		Name:      u.GetName(),
		Namespace: u.GetNamespace(),
	}, nil
}

func (r reconciler) createClusterActiveOperandIfNecessary(ctx context.Context, o *v1alpha1.Operand, name string, refs ...v1alpha1.LiveRef) error {
	// Check that CAO exists.
	logger := logging.FromContext(ctx)
	_, err := r.operandClient.ClusterActiveOperands().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to fetch ClusterActiveOperand %s, %v", name, err)
		}

		o.Status.ResetLatestCreatedActiveOperand(name)

		logger.Infof("Creating latest CAO %s", name)
		cao := &v1alpha1.ClusterActiveOperand{
			ObjectMeta: metav1.ObjectMeta{
				Name:   name,
				Labels: map[string]string{OwnerName: o.Name},
			},
			Spec: v1alpha1.ClusterActiveOperandSpec{
				Live: refs,
			},
		}
		ref := kmeta.NewControllerRef(o)
		cao.SetOwnerReferences([]metav1.OwnerReference{*ref})
		_, err = r.operandClient.ClusterActiveOperands().Create(ctx, cao, metav1.CreateOptions{})
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("latest created failed %w", err)
		}
		logger.Infof("Created CAO %s", name)
		return nil
	}

	return nil
}
