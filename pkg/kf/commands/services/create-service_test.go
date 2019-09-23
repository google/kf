// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package services_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	servicecatalogclientfake "github.com/google/kf/pkg/client/servicecatalog/clientset/versioned/fake"
	"github.com/google/kf/pkg/kf/commands/config"
	servicescmd "github.com/google/kf/pkg/kf/commands/services"
	utils "github.com/google/kf/pkg/kf/internal/utils/cli"
	"github.com/google/kf/pkg/kf/testutil"
	servicecatalogv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clienttesting "k8s.io/client-go/testing"
)

func TestNewCreateServiceCommand(t *testing.T) {

	plan := servicecatalogv1beta1.ServicePlan{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db-service-free",
			Namespace: "custom-ns",
		},
		Spec: servicecatalogv1beta1.ServicePlanSpec{
			ServiceBrokerName: "broker-a",
			CommonServicePlanSpec: servicecatalogv1beta1.CommonServicePlanSpec{
				ExternalName: "free",
				Free:         false,
			},
			ServiceClassRef: servicecatalogv1beta1.LocalObjectReference{
				Name: "db-service-id",
			},
		},
	}

	class := &servicecatalogv1beta1.ServiceClass{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db-service-id",
			Namespace: "custom-ns",
		},
		Spec: servicecatalogv1beta1.ServiceClassSpec{
			ServiceBrokerName: "broker-a",
			CommonServiceClassSpec: servicecatalogv1beta1.CommonServiceClassSpec{
				ExternalName: "db-service",
			},
		},
	}

	planList := &servicecatalogv1beta1.ServicePlanList{
		Items: []servicecatalogv1beta1.ServicePlan{
			plan,
		},
	}

	clusterPlan := servicecatalogv1beta1.ClusterServicePlan{
		ObjectMeta: metav1.ObjectMeta{
			Name: "db-service-free",
		},
		Spec: servicecatalogv1beta1.ClusterServicePlanSpec{
			ClusterServiceBrokerName: "broker-a",
			CommonServicePlanSpec: servicecatalogv1beta1.CommonServicePlanSpec{
				ExternalName: "free",
				Free:         true,
			},
			ClusterServiceClassRef: servicecatalogv1beta1.ClusterObjectReference{
				Name: "db-service-id",
			},
		},
	}

	clusterClass := &servicecatalogv1beta1.ClusterServiceClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "db-service-id",
		},
		Spec: servicecatalogv1beta1.ClusterServiceClassSpec{
			ClusterServiceBrokerName: "broker-a",
			CommonServiceClassSpec: servicecatalogv1beta1.CommonServiceClassSpec{
				ExternalName: "db-service",
			},
		},
	}

	clusterPlanList := &servicecatalogv1beta1.ClusterServicePlanList{
		Items: []servicecatalogv1beta1.ClusterServicePlan{
			clusterPlan,
		},
	}

	cases := map[string]struct {
		Args      []string
		Setup     func(t *testing.T) *servicecatalogclientfake.Clientset
		Namespace string

		ExpectedErr     error
		ExpectedStrings []string
	}{
		"too few params": {
			Args:        []string{},
			ExpectedErr: errors.New("accepts 3 arg(s), received 0"),
		},
		"command params get passed correctly": {
			Args:      []string{"db-service", "free", "mydb", `--config={"ram_gb":4}`},
			Namespace: "custom-ns",
			Setup: func(t *testing.T) *servicecatalogclientfake.Clientset {
				return servicecatalogclientfake.NewSimpleClientset(planList, class)
			},
			ExpectedStrings: []string{
				"db-service",
				"mydb",
				"free",
				"ram_gb",
			},
		},
		"service from cluster broker": {
			Args:      []string{"db-service", "free", "mydb", `--config={"ram_gb":4}`},
			Namespace: "custom-ns",
			Setup: func(t *testing.T) *servicecatalogclientfake.Clientset {
				return servicecatalogclientfake.NewSimpleClientset(clusterPlanList, clusterClass)
			},
			ExpectedStrings: []string{
				"db-service",
				"mydb",
				"free",
				"ram_gb",
			},
		},
		"none for broker": {
			Args:      []string{"db-service", "free", "mydb", "--broker=does-not-exist"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T) *servicecatalogclientfake.Clientset {
				client := servicecatalogclientfake.NewSimpleClientset(planList, class)
				return client
			},
			ExpectedErr: errors.New("no plan free found for class db-service for the service-broker does-not-exist"),
		},
		"none for cluster broker": {
			Args:      []string{"db-service", "free", "mydb", "--broker=does-not-exist"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T) *servicecatalogclientfake.Clientset {
				client := servicecatalogclientfake.NewSimpleClientset(clusterPlanList, clusterClass)
				return client
			},
			ExpectedErr: errors.New("no plan free found for class db-service for the service-broker does-not-exist"),
		},
		"cluster over namespaced": {
			Args:      []string{"db-service", "free", "mydb", `--config={"ram_gb":4}`},
			Namespace: "custom-ns",
			Setup: func(t *testing.T) *servicecatalogclientfake.Clientset {
				client := servicecatalogclientfake.NewSimpleClientset(planList, clusterPlanList, class, clusterClass)
				client.PrependReactor("*", "serviceplans", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("dont ask for namespaced plans")
				})
				return client
			},
			ExpectedStrings: []string{
				"db-service",
				"mydb",
				"free",
				"ram_gb",
			},
		},
		"empty namespace": {
			Args:        []string{"db-service", "free", "mydb", `--config={"ram_gb":4}`},
			ExpectedErr: errors.New(utils.EmptyNamespaceError),
		},
		"defaults config": {
			Args:      []string{"db-service", "free", "mydb"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T) *servicecatalogclientfake.Clientset {
				return servicecatalogclientfake.NewSimpleClientset(planList, class)
			},
		},
		"bad path": {
			Args:        []string{"db-service", "free", "mydb", `--config=/some/bad/path`},
			Namespace:   "custom-ns",
			ExpectedErr: errors.New("couldn't read file: open /some/bad/path: no such file or directory"),
		},
		"bad server call": {
			Args:      []string{"db-service", "free", "mydb"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T) *servicecatalogclientfake.Clientset {
				client := servicecatalogclientfake.NewSimpleClientset()
				client.PrependReactor("*", "*", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("server-call-error")
				})
				return client
			},
			ExpectedErr: errors.New("server-call-error"),
		},
		"bad server call listing cluster plans": {
			Args:      []string{"db-service", "free", "mydb"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T) *servicecatalogclientfake.Clientset {
				client := servicecatalogclientfake.NewSimpleClientset(clusterPlanList, planList, clusterClass, class)
				client.PrependReactor("*", "clusterserviceplans", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("server-call-error")
				})
				return client
			},
			ExpectedErr: errors.New("server-call-error"),
		},
		"bad server call listing plans": {
			Args:      []string{"db-service", "free", "mydb"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T) *servicecatalogclientfake.Clientset {
				client := servicecatalogclientfake.NewSimpleClientset(planList, class)
				client.PrependReactor("list", "serviceplans", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("server-call-error")
				})
				return client
			},
			ExpectedErr: errors.New("server-call-error"),
		},
		"bad server call creating from cluster": {
			Args:      []string{"db-service", "free", "mydb"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T) *servicecatalogclientfake.Clientset {
				client := servicecatalogclientfake.NewSimpleClientset(clusterPlanList, clusterClass)
				client.PrependReactor("create", "serviceinstances", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("server-call-error")
				})
				return client
			},
			ExpectedErr: errors.New("server-call-error"),
		},
		"bad server call creating": {
			Args:      []string{"db-service", "free", "mydb"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T) *servicecatalogclientfake.Clientset {
				client := servicecatalogclientfake.NewSimpleClientset(planList, class)
				client.PrependReactor("create", "serviceinstances", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("server-call-error")
				})
				return client
			},
			ExpectedErr: errors.New("server-call-error"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {

			buf := new(bytes.Buffer)
			p := &config.KfParams{
				Namespace: tc.Namespace,
			}

			var client *servicecatalogclientfake.Clientset
			if tc.Setup != nil {
				client = tc.Setup(t)
			} else {
				client = servicecatalogclientfake.NewSimpleClientset()
			}

			cmd := servicescmd.NewCreateServiceCommand(p, client)
			fmt.Fprintf(os.Stderr, "%s", buf.String())
			cmd.SetOutput(buf)
			cmd.SetArgs(tc.Args)
			_, actualErr := cmd.ExecuteC()
			if tc.ExpectedErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectedErr, actualErr)
			}

			testutil.AssertContainsAll(t, buf.String(), tc.ExpectedStrings)
		})
	}
}
