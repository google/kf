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

package kf

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"sync"
	"time"

	"kf-operator/pkg/apis/kfsystem/v1alpha1"
	operandv1alpha1 "kf-operator/pkg/apis/operand/v1alpha1"
	opclient "kf-operator/pkg/client/clientset/versioned/typed/operand/v1alpha1"
	"kf-operator/pkg/client/injection/client"
	"kf-operator/pkg/operand"

	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/logging"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	pkgreconciler "knative.dev/pkg/reconciler"
)

var (
	// OwnerNamespace labels an Operand when it is owned by a non-clusterscoped resource.
	OwnerNamespace = "operand.kuberun.cloud.google.com/owner-ns"
	// OwnerName labels an Operand when it is owned by a non-clusterscoped resource.
	OwnerName       = "operand.kuberun.cloud.google.com/owner-name"
	reconcilePeriod time.Duration
)

func init() {
	flag.DurationVar(&reconcilePeriod, "kfsystem_reconcile_period", 1*time.Minute, "Time to wait between periodic reconciles. Defaults to 1 minute..")
}

// OperandReconciler is the general structure of the specific reconciler.
type operandReconciler struct {
	Name          string
	KubeClient    kubernetes.Interface
	OperandClient opclient.OperandV1alpha1Interface
	Ctor
}

// CreateOperandReconciler creates and returns an Interface for the given configuration.
func CreateOperandReconciler(ctx context.Context, name string, ctor Ctor) Interface {
	return &operandReconciler{
		Name:          name,
		KubeClient:    kubeclient.Get(ctx),
		OperandClient: client.Get(ctx).OperandV1alpha1(),
		Ctor:          ctor,
	}
}

// Interface exposes the reconciler fields expected to be used outside
// this package.
type Interface interface {
	// Reconcile is called from the Reconciler in ReconcileKind.
	Reconcile(ctx context.Context, kfs *v1alpha1.KfSystem) pkgreconciler.Event
	Finalize(ctx context.Context) pkgreconciler.Event
}

// Ctor defines the strongly typed interfaces to be implemented by a
// specific resource reconciler.
type Ctor interface {
	// MarkInstallSucceeded indicates that we've completed the install.
	MarkInstallSucceeded(kfs *v1alpha1.KfSystem)
	//MarkInstallNotReady indicates to the user that we're waiting for the install to complete.
	MarkInstallNotReady(kfs *v1alpha1.KfSystem)
	// MarkInstallFailed indicates to the user that we've ceased waiting on the install to complete.
	MarkInstallFailed(kfs *v1alpha1.KfSystem, error string)
	// CalculateOperand calculates the operand this reconciliation cycle.
	CalculateOperand(context.Context, *v1alpha1.KfSystem) (*operandv1alpha1.OperandSpec, error)
}

type safeCtor struct {
	*sync.Mutex
	Ctor
}

// MakeSafe creates a Ctor that blocks Mark* calls on the given mutex.
func MakeSafe(c Ctor, m *sync.Mutex) Ctor {
	return &safeCtor{m, c}
}

func (s safeCtor) MarkInstallSucceeded(kfs *v1alpha1.KfSystem) {
	s.Lock()
	defer s.Unlock()
	s.Ctor.MarkInstallSucceeded(kfs)
}

func (s safeCtor) MarkInstallNotReady(kfs *v1alpha1.KfSystem) {
	s.Lock()
	defer s.Unlock()
	s.Ctor.MarkInstallNotReady(kfs)
}

func (s safeCtor) MarkInstallFailed(kfs *v1alpha1.KfSystem, err string) {
	s.Lock()
	defer s.Unlock()
	s.Ctor.MarkInstallFailed(kfs, err)
}

// Reconcile handles installing the manifest
func (o operandReconciler) Reconcile(ctx context.Context, kfs *v1alpha1.KfSystem) pkgreconciler.Event {
	if kfs.GetStatus().ObservedGeneration != kfs.GetGeneration() {
		o.Ctor.MarkInstallNotReady(kfs)
	}
	opSpec, err := o.CalculateOperand(ctx, kfs)
	if err != nil {
		o.Ctor.MarkInstallFailed(kfs, fmt.Sprintf("Failed to create operand %v", err))
		return err
	}

	op, err := operand.EnsureOperand(ctx, o.Name, *opSpec, o.OperandClient, map[string]string{OwnerName: kfs.GetName()})
	if err != nil {
		o.Ctor.MarkInstallFailed(kfs, err.Error())
		return err
	}
	if op.Status.IsFalse() {
		// Mark install failed to avoid deadlock on webhook.
		err = errors.New("operand is not installed")
		o.Ctor.MarkInstallFailed(kfs, err.Error())
		return err
	}
	if !op.Status.IsReady() || op.GetStatus().ObservedGeneration != op.GetGeneration() {
		o.MarkInstallNotReady(kfs)
		return nil
	}
	o.MarkInstallSucceeded(kfs)
	return nil
}

func (o operandReconciler) Finalize(ctx context.Context) pkgreconciler.Event {
	logger := logging.FromContext(ctx)
	err := o.OperandClient.Operands().Delete(ctx, o.Name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		logger.Infof("The operand %s has been deleted.", o.Name)
		return nil
	}
	return err
}
