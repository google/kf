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
	"log"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	sourceconfig "github.com/google/kf/pkg/reconciler/source/config"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	cv1alpha3 "knative.dev/pkg/client/clientset/versioned/typed/istio/v1alpha3"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/logging/logkey"
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

	// Check into the config-secrets. Its *possible* that the ConfigMap
	// doesn't exist. This is because the ConfigMap has to point at a secret
	// that an operator HAS to create. Therefore, we need to create a
	// placeholder ConfigMap if one doesn't exist.
	//
	// NOTE: This is necessary for kf v0.3 and beyond.
	{
		corev1Client, err := clientcorev1.NewForConfig(clusterConfig)
		if err != nil {
			logger.Fatalw("Failed to get the corev1 client set", zap.Error(err))
		}
		_, err = corev1Client.
			ConfigMaps(v1alpha1.KfNamespace).
			Get(sourceconfig.SecretsConfigName, metav1.GetOptions{})

		switch {
		case apierrs.IsNotFound(err):
			// ConfigMap isn't found, so we should create one.
			logger.Warnw("ConfigMap %q not found. Creating a placeholder ConfigMap...", sourceconfig.SecretsConfigName)
			if err := createPlaceholderConfigMap(corev1Client); err != nil {
				logger.Fatalw("Failed to create ConfigMap %q: %v", sourceconfig.SecretsConfigName, err)
			}
		case err != nil:
			logger.Fatalw("Failed to fetch ConfigMap %q: %v", sourceconfig.SecretsConfigName, err)
		}
	}

	istioClient, err := cv1alpha3.NewForConfig(clusterConfig)
	if err != nil {
		logger.Fatalw("Failed to get the istio client set", zap.Error(err))
	}

	// Clean out VirtualServices in the kf namespace. VirtualServices should
	// be in the namespace as their associated RouteClaim.
	//
	// NOTE: Older versions of Kf stored VirtualServices in the Kf namespace
	// to emulate a non-existent ClusterVirtualService. Newer versions of Kf
	// have moved away from this pattern, but this could lead to some stale
	// VirtualServices in the Kf namespace. Therefore this block of code
	// should be revisted in later versions of Kf when the upgrade path from
	// Kf 0.2 is no longer supported.
	logger.Info("Deleting virtualservices in the `kf` namespace...")
	err = istioClient.VirtualServices(v1alpha1.KfNamespace).DeleteCollection(nil, v1.ListOptions{})
	if err != nil {
		logger.Fatalw("Failed to delete virtualservices in the `kf` namespace", zap.Error(err))
	}
}

func createPlaceholderConfigMap(corev1Client *clientcorev1.CoreV1Client) error {
	cm := &corev1.ConfigMap{}
	cm.Name = sourceconfig.SecretsConfigName

	_, err := corev1Client.
		ConfigMaps(v1alpha1.KfNamespace).
		Create(cm)
	return err
}
