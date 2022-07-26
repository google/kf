/*
Copyright 2019 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"flag"
	"log"

	"k8s.io/client-go/tools/clientcmd"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/signals"

	"kf-operator/pkg/reconciler/activeoperand"
	"kf-operator/pkg/reconciler/clusteractiveoperand"
	"kf-operator/pkg/reconciler/kfsystem"
	"kf-operator/pkg/reconciler/operand"
)

var (
	// MasterURL is the address of the Kubernetes API server.
	MasterURL = flag.String("master-url", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	// Kubeconfig is the path to a kubeconfig.
	Kubeconfig = flag.String("kubeconfig-url", "", "Path to a kubeconfig. Only required if out-of-cluster.")
)

func main() {
	flag.Parse()

	cfg, err := clientcmd.BuildConfigFromFlags(*MasterURL, *Kubeconfig)
	if err != nil {
		log.Fatal("Error building kubeconfig", err)
	}
	sharedmain.MainWithConfig(
		signals.NewContext(),
		"appdevexperience-operator",
		cfg,
		kfsystem.NewController,
		clusteractiveoperand.NewController,
		activeoperand.NewController,
		operand.NewController)
}
