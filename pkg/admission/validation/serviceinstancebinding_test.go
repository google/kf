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

package kfvalidation

import (
	"context"
	"errors"
	"testing"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	kffake "github.com/google/kf/v2/pkg/client/kf/clientset/versioned/fake"
	"github.com/google/kf/v2/pkg/client/kf/informers/externalversions"
	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func TestValidateAppExists(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := &v1alpha1.App{}
	app.Name = "valid-app"
	app.Namespace = "test-ns"
	app.APIVersion = "kf.dev/v1alpha1"
	app.Kind = "App"

	// Create a fake App lister. By default, this returns a "not found" error if the
	// requested App does not exist.
	kfClient := kffake.NewSimpleClientset(app)
	informers := externalversions.NewSharedInformerFactory(kfClient, 0)
	appInformer := informers.Kf().V1alpha1().Apps().Informer()
	appLister := informers.Kf().V1alpha1().Apps().Lister()

	informers.Start(ctx.Done())
	cache.WaitForCacheSync(ctx.Done(), appInformer.HasSynced)

	validBinding := &v1alpha1.ServiceInstanceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test-ns",
		},
		Spec: v1alpha1.ServiceInstanceBindingSpec{
			BindingType: v1alpha1.BindingType{
				App: &v1alpha1.AppRef{
					Name: "valid-app",
				},
			},
			InstanceRef: corev1.LocalObjectReference{
				Name: "valid-service",
			},
		},
	}

	invalidBinding := &v1alpha1.ServiceInstanceBinding{
		Spec: v1alpha1.ServiceInstanceBindingSpec{
			BindingType: v1alpha1.BindingType{
				App: &v1alpha1.AppRef{
					Name: "missing-app",
				},
			},
		},
	}

	cases := map[string]struct {
		serviceinstancebinding *v1alpha1.ServiceInstanceBinding
		want                   error
	}{
		"app exists": {
			serviceinstancebinding: validBinding,
		},
		"app doesn't exist": {
			serviceinstancebinding: invalidBinding,
			want:                   errors.New("App \"missing-app\" does not exist. The binding cannot be created"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := validateAppExists(appLister, tc.serviceinstancebinding)
			testutil.AssertErrorsEqual(t, tc.want, got)
		})
	}
}

func TestValidateServiceInstanceExists(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	si := &v1alpha1.ServiceInstance{}
	si.Name = "valid-service"
	si.Namespace = "test-ns"
	si.APIVersion = "kf.dev/v1alpha1"
	si.Kind = "ServiceInstance"

	// Create a fake ServiceInstance lister. By default, this returns a "not found" error if the
	// requested ServiceInstance does not exist.
	kfClient := kffake.NewSimpleClientset(si)
	informers := externalversions.NewSharedInformerFactory(kfClient, 0)
	serviceInstanceInformer := informers.Kf().V1alpha1().ServiceInstances().Informer()
	serviceInstanceLister := informers.Kf().V1alpha1().ServiceInstances().Lister()

	informers.Start(ctx.Done())
	cache.WaitForCacheSync(ctx.Done(), serviceInstanceInformer.HasSynced)

	validBinding := &v1alpha1.ServiceInstanceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test-ns",
		},
		Spec: v1alpha1.ServiceInstanceBindingSpec{
			BindingType: v1alpha1.BindingType{
				App: &v1alpha1.AppRef{
					Name: "valid-app",
				},
			},
			InstanceRef: corev1.LocalObjectReference{
				Name: "valid-service",
			},
		},
	}

	invalidBinding := &v1alpha1.ServiceInstanceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test-ns",
		},
		Spec: v1alpha1.ServiceInstanceBindingSpec{
			BindingType: v1alpha1.BindingType{
				App: &v1alpha1.AppRef{
					Name: "valid-app",
				},
			},
			InstanceRef: corev1.LocalObjectReference{
				Name: "missing-service",
			},
		},
	}

	cases := map[string]struct {
		serviceinstancebinding *v1alpha1.ServiceInstanceBinding
		want                   error
	}{
		"service instance exists": {
			serviceinstancebinding: validBinding,
		},
		"service instance doesn't exist": {
			serviceinstancebinding: invalidBinding,
			want:                   errors.New("ServiceInstance \"missing-service\" does not exist. The binding cannot be created"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := validateServiceInstanceExists(serviceInstanceLister, tc.serviceinstancebinding)
			testutil.AssertErrorsEqual(t, tc.want, got)
		})
	}
}
