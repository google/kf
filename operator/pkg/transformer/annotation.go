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

package transformer

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes/scheme"
	"knative.dev/pkg/logging"
)

// requiredAnnotationsFromEnvVars is a map of annotations that must be present as environment variables. The key is the
// environment variable name. The value is the annotation's name. So for each entry in this map, we must find the
// corresponding envVar and whenever this operator creates a Pod (directly or indirectly [e.g. a Deployment]), it must
// ensure that the Pods have the annotation name (value in this map) with the env var's value.
var requiredAnnotationsFromEnvVars = map[string]string{
	"COMPONENT_NAME_ANNOTATION":    "components.gke.io/component-name",
	"COMPONENT_VERSION_ANNOTATION": "components.gke.io/component-version",
}

// TestOnlyChangeRequiredAnnotationsFromEnvVars should be called only in unit tests. It changes the required annotations
// to be whatever is specified. It returns a function to undo the changes. The expected usage is:
// defer transformer.TestOnlyChangesRequiredAnnotationsFromEnvVars(nil)()
func TestOnlyChangeRequiredAnnotationsFromEnvVars(newValue map[string]string) func() {
	original := requiredAnnotationsFromEnvVars
	requiredAnnotationsFromEnvVars = newValue
	return func() {
		requiredAnnotationsFromEnvVars = original
	}
}

// GetPodAnnotationsFromEnv gets all the annotations that should be injected into Pods created either directly (e.g. Pod) or
// indirectly (e.g. Deployment) by this operator. It is intended to be used with an AnnotationTransformer.
func GetPodAnnotationsFromEnv() (map[string]string, error) {
	podAnnotations := make(map[string]string, len(requiredAnnotationsFromEnvVars))
	for envVarName, annotationName := range requiredAnnotationsFromEnvVars {
		if value := os.Getenv(envVarName); value != "" {
			podAnnotations[annotationName] = value
		} else {
			return podAnnotations, fmt.Errorf("required envVar '%s' was blank", envVarName)
		}
	}
	return podAnnotations, nil
}

// Annotation is a transformer that injects annotations into Pods created both directly (e.g. Pod) and
// indirectly (e.g. Deployment).
type Annotation struct {
	PodAnnotations map[string]string
}

// Transform creates a func to inject annotattions into Pods.
func (a *Annotation) Transform(ctx context.Context) func(u *unstructured.Unstructured) error {
	logger := logging.FromContext(ctx).Desugar()
	return func(u *unstructured.Unstructured) error {
		switch u.GetKind() {
		// TODO need to use PodSpecable duck type in order to remove duplicates of deployment, daemonSet
		case "Pod":
			return a.updatePod(logger, u)
		case "Deployment":
			return a.updateDeployment(logger, u)
		case "DaemonSet":
			return a.updateDaemonSet(logger, u)
		case "StatefulSet":
			return a.updateStatefulSet(logger, u)
		case "Job":
			return a.updateJob(logger, u)
		default:
			return nil
		}
	}
}

func (a *Annotation) addAnnotations(podSpec *v1.PodTemplateSpec) {
	if podSpec.Annotations == nil {
		podSpec.Annotations = make(map[string]string)
	}
	for k, v := range a.PodAnnotations {
		podSpec.Annotations[k] = v
	}
}

func (a *Annotation) updatePod(logger *zap.Logger, u *unstructured.Unstructured) error {
	var pod = &v1.Pod{}
	if err := scheme.Scheme.Convert(u, pod, nil); err != nil {
		logger.Error("Error converting Unstructured to Pod", zap.Error(err), zap.Any("unstructured", u), zap.Any("pod", pod))
		return err
	}

	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string)
	}
	for k, v := range a.PodAnnotations {
		pod.Annotations[k] = v
	}

	if err := scheme.Scheme.Convert(pod, u, nil); err != nil {
		return err
	}
	// The zero-value timestamp defaulted by the conversion causes
	// superfluous updates
	u.SetCreationTimestamp(metav1.Time{})

	logger.Debug("Finished injecting annotations into the Pod", zap.String("name", u.GetName()), zap.Any("unstructured", u.Object))
	return nil
}

func (a *Annotation) updateDeployment(logger *zap.Logger, u *unstructured.Unstructured) error {
	var deployment = &appsv1.Deployment{}
	if err := scheme.Scheme.Convert(u, deployment, nil); err != nil {
		logger.Error("Error converting Unstructured to Deployment", zap.Error(err), zap.Any("unstructured", u), zap.Any("deployment", deployment))
		return err
	}

	a.addAnnotations(&deployment.Spec.Template)
	if err := scheme.Scheme.Convert(deployment, u, nil); err != nil {
		return err
	}
	// The zero-value timestamp defaulted by the conversion causes
	// superfluous updates
	u.SetCreationTimestamp(metav1.Time{})

	logger.Debug("Finished injecting annotations into the Deployment", zap.String("name", u.GetName()), zap.Any("unstructured", u.Object))
	return nil
}

func (a *Annotation) updateDaemonSet(logger *zap.Logger, u *unstructured.Unstructured) error {
	var daemonSet = &appsv1.DaemonSet{}
	if err := scheme.Scheme.Convert(u, daemonSet, nil); err != nil {
		logger.Error("Error converting Unstructured to daemonSet", zap.Error(err), zap.Any("unstructured", u), zap.Any("daemonSet", daemonSet))
		return err
	}

	a.addAnnotations(&daemonSet.Spec.Template)
	if err := scheme.Scheme.Convert(daemonSet, u, nil); err != nil {
		return err
	}
	// The zero-value timestamp defaulted by the conversion causes
	// superfluous updates
	u.SetCreationTimestamp(metav1.Time{})

	logger.Debug("Finished injecting annotations into the daemonSet", zap.String("name", u.GetName()), zap.Any("unstructured", u.Object))
	return nil
}

func (a *Annotation) updateStatefulSet(logger *zap.Logger, u *unstructured.Unstructured) error {
	var statefulSet = &appsv1.StatefulSet{}
	if err := scheme.Scheme.Convert(u, statefulSet, nil); err != nil {
		logger.Error("Error converting Unstructured to statefulSet", zap.Error(err), zap.Any("unstructured", u), zap.Any("statefulSet", statefulSet))
		return err
	}

	a.addAnnotations(&statefulSet.Spec.Template)
	if err := scheme.Scheme.Convert(statefulSet, u, nil); err != nil {
		return err
	}
	// The zero-value timestamp defaulted by the conversion causes
	// superfluous updates
	u.SetCreationTimestamp(metav1.Time{})

	logger.Debug("Finished injecting annotations into the StatefulSet", zap.String("name", u.GetName()), zap.Any("unstructured", u.Object))
	return nil
}

func (a *Annotation) updateJob(logger *zap.Logger, u *unstructured.Unstructured) error {
	var job = &batchv1.Job{}
	if err := scheme.Scheme.Convert(u, job, nil); err != nil {
		logger.Error("Error converting Unstructured to Job", zap.Error(err), zap.Any("unstructured", u), zap.Any("job", job))
		return err
	}

	a.addAnnotations(&job.Spec.Template)
	if err := scheme.Scheme.Convert(job, u, nil); err != nil {
		return err
	}
	// The zero-value timestamp defaulted by the conversion causes
	// superfluous updates
	u.SetCreationTimestamp(metav1.Time{})

	logger.Debug("Finished injecting annotations into the Job", zap.String("name", u.GetName()), zap.Any("unstructured", u.Object))
	return nil
}
