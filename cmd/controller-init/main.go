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

package main

import (
	"flag"
	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	cv1alpha3 "knative.dev/pkg/client/clientset/versioned/typed/istio/v1alpha3"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/logging/logkey"
	"log"
)

const (
	component = "controller-init"
)

var (
	masterURL  = flag.String("master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	kubeconfig = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
)

func main() {
	flag.Parse()
	cm, err := configmap.Load("/etc/config-logging")
	if err != nil {
		log.Fatal("Error loading logging configuration:", err)
	}

	config, err := logging.NewConfigFromMap(cm)
	if err != nil {
		log.Fatal("Error parsing logging configuration:", err)
	}

	logger, _ := logging.NewLoggerFromConfig(config, component)
	defer logger.Sync()
	logger = logger.With(zap.String(logkey.ControllerType, component))

	clusterConfig, err := clientcmd.BuildConfigFromFlags(*masterURL, *kubeconfig)
	if err != nil {
		logger.Fatalw("Failed to get cluster config", zap.Error(err))
	}

	istioClient, err := cv1alpha3.NewForConfig(clusterConfig)
	if err != nil {
		logger.Fatalw("Failed to get the istio client set", zap.Error(err))
	}

	// Clean out virtualservices in the kf namespace. Virtualservices should be in the namespace they were created in
	logger.Info("Deleting virtualservices in the `kf` namespace...")
	err = istioClient.VirtualServices(v1alpha1.KfNamespace).DeleteCollection(nil, v1.ListOptions{})
	if err != nil {
		logger.Fatalw("Failed to delete virtualservices in the `kf` namespace", zap.Error(err))
	}
}
