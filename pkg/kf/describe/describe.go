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

package describe

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	kfv1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/services"
	"github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	"sigs.k8s.io/yaml"
)

// EnvVars prints out environment variables.
func EnvVars(w io.Writer, vars []corev1.EnvVar) error {

	return SectionWriter(w, "Environment", func(w io.Writer) error {
		// TODO handle additional types of variables with different ValueFrom
		// fields like Secret or ConfigMap
		for _, e := range vars {
			if _, err := fmt.Fprintf(w, "%s:\t%s\n", e.Name, e.Value); err != nil {
				return err
			}
		}

		return nil
	})
}

// TypeMeta prints information about the type.
func TypeMeta(w io.Writer, meta metav1.TypeMeta) error {
	return TabbedWriter(w, func(w io.Writer) error {
		if _, err := fmt.Fprintf(w, "API Version:\t%s\n", meta.APIVersion); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "Kind:\t%s\n", meta.Kind); err != nil {
			return err
		}

		return nil
	})
}

// ObjectMeta prints the normal object metadata associated with standard K8s
// objects.
func ObjectMeta(w io.Writer, meta metav1.ObjectMeta) error {

	return SectionWriter(w, "Metadata", func(w io.Writer) error {
		if _, err := fmt.Fprintf(w, "Name:\t%s\n", meta.Name); err != nil {
			return err
		}

		if meta.Namespace != "" {
			if _, err := fmt.Fprintf(w, "Space:\t%s\n", meta.Namespace); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "Creation Timestamp:\t%s\n", meta.CreationTimestamp); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "Age:\t%s\n", translateTimestampSince(meta.CreationTimestamp)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "Generation:\t%d\n", meta.Generation); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "UID:\t%s\n", meta.UID); err != nil {
			return err
		}

		if meta.DeletionTimestamp != nil {
			if _, err := fmt.Fprintf(w, "Terminating Since:\t%s\n", translateTimestampSince(*meta.DeletionTimestamp)); err != nil {
				return err
			}

			if meta.DeletionGracePeriodSeconds != nil {
				if _, err := fmt.Fprintf(w, "Termination Grace Period:\t%ds\n", *meta.DeletionGracePeriodSeconds); err != nil {
					return err
				}
			}
		}

		if len(meta.Labels) == 0 {
			if _, err := fmt.Fprintf(w, "Labels:\t<none>\n"); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(w, "Labels:\t\n"); err != nil {
				return err
			}

			if err := IndentWriter(w, func(w io.Writer) error {
				return Labels(w, meta.Labels)
			}); err != nil {
				return err
			}
		}

		return nil
	})
}

// Labels prints a label map by the alphanumeric sorting of the keys.
func Labels(w io.Writer, labels map[string]string) error {

	type pair struct {
		key   string
		value string
	}
	var list []pair

	for k, v := range labels {
		list = append(list, pair{k, v})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].key < list[j].key
	})

	for _, pair := range list {
		if _, err := fmt.Fprintf(w, "%s=%s\n", pair.key, pair.value); err != nil {
			return err
		}
	}

	return nil
}

// DuckStatus prints a table of status info based on the duck status.
func DuckStatus(w io.Writer, duck duckv1beta1.Status) error {

	return SectionWriter(w, "Status", func(w io.Writer) error {
		// Print the overall status, this should be one of either ConditionReady
		// OR ConditionSucceeded. Ready is for tasks that will run continuously
		// and succeeded is for one time tasks.
		if cond := duck.GetCondition(apis.ConditionReady); cond != nil {
			if err := duckCondition(w, *cond); err != nil {
				return err
			}
		}

		if cond := duck.GetCondition(apis.ConditionSucceeded); cond != nil {
			if err := duckCondition(w, *cond); err != nil {
				return err
			}
		}

		// XXX: We may want to print out observedgeneration here. That might
		// confuse users so we should be careful about the wording we use to
		// present it.

		// Print the rest of the conditions in a table.
		return SectionWriter(w, "Conditions", func(w io.Writer) error {
			conds := duck.GetConditions()
			if len(conds) == 0 {
				return nil
			}

			sort.Slice(conds, func(i, j int) bool {
				return conds[i].Type < conds[j].Type
			})

			if _, err := fmt.Fprintln(w, "Type\tStatus\tUpdated\tMessage\tReason"); err != nil {
				return err
			}

			for _, c := range conds {
				if c.Type == apis.ConditionReady || c.Type == apis.ConditionSucceeded {
					continue
				}

				if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					c.Type,
					c.Status,
					translateTimestampSince(c.LastTransitionTime.Inner),
					c.Message,
					c.Reason,
				); err != nil {
					return err
				}
			}

			return nil
		})
	})
}

func duckCondition(w io.Writer, cond apis.Condition) error {

	return SectionWriter(w, string(cond.Type), func(w io.Writer) error {
		if _, err := fmt.Fprintf(w, "Ready:\t%s\n", cond.Status); err != nil {
			return err
		}

		if !cond.LastTransitionTime.Inner.IsZero() {
			if _, err := fmt.Fprintf(w, "Time:\t%s\n", cond.LastTransitionTime.Inner); err != nil {
				return err
			}
		}
		if cond.Message != "" {
			if _, err := fmt.Fprintf(w, "Message:\t%s\n", cond.Message); err != nil {
				return err
			}
		}
		if cond.Reason != "" {
			if _, err := fmt.Fprintf(w, "Reason:\t%s\n", cond.Reason); err != nil {
				return err
			}
		}

		return nil
	})
}

// SourceSpec describes the source of an application and the build process it
// will undergo.
func SourceSpec(w io.Writer, spec kfv1alpha1.SourceSpec) error {

	return SectionWriter(w, string("Source"), func(w io.Writer) error {
		switch {
		case spec.IsContainerBuild():
			if _, err := fmt.Fprintln(w, "Build Type:\tcontainer"); err != nil {
				return err
			}
		case spec.IsBuildpackBuild():
			if _, err := fmt.Fprintln(w, "Build Type:\tbuildpack"); err != nil {
				return err
			}
		case spec.IsDockerfileBuild():
			if _, err := fmt.Fprintln(w, "Build Type:\tdockerfile"); err != nil {
				return err
			}
		default:
			if _, err := fmt.Fprintln(w, "Build Type:\tunknown"); err != nil {
				return err
			}
		}

		if spec.ServiceAccount != "" {
			if _, err := fmt.Fprintf(w, "Service Account:\t%s\n", spec.ServiceAccount); err != nil {
				return err
			}
		}

		if spec.IsContainerBuild() {
			if err := SectionWriter(w, "Container Image", func(w io.Writer) error {
				containerImage := spec.ContainerImage

				if _, err := fmt.Fprintf(w, "Image:\t%s\n", containerImage.Image); err != nil {
					return err
				}

				return nil
			}); err != nil {
				return err
			}
		}

		if spec.IsBuildpackBuild() {
			if err := SectionWriter(w, "Buildpack Build", func(w io.Writer) error {
				buildpackBuild := spec.BuildpackBuild

				if _, err := fmt.Fprintf(w, "Source:\t%s\n", buildpackBuild.Source); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(w, "Stack:\t%s\n", buildpackBuild.Stack); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(w, "Bulider:\t%s\n", buildpackBuild.BuildpackBuilder); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(w, "Destination:\t%s\n", buildpackBuild.Image); err != nil {
					return err
				}
				return EnvVars(w, buildpackBuild.Env)
			}); err != nil {
				return err
			}
		}

		if spec.IsDockerfileBuild() {
			if err := SectionWriter(w, "Dockerfile Build", func(w io.Writer) error {
				build := spec.Dockerfile

				if _, err := fmt.Fprintf(w, "Source:\t%s\n", build.Source); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(w, "Dockerfile Path:\t%s\n", build.Path); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(w, "Destination:\t%s\n", build.Image); err != nil {
					return err
				}

				return nil
			}); err != nil {
				return err
			}
		}

		return nil
	})
}

// AppSpecInstances describes the scaling features of the app.
func AppSpecInstances(w io.Writer, instances kfv1alpha1.AppSpecInstances) error {

	return SectionWriter(w, string("Scale"), func(w io.Writer) error {
		hasExactly := instances.Exactly != nil
		hasMin := instances.Min != nil
		hasMax := instances.Max != nil

		if _, err := fmt.Fprintf(w, "Stopped?:\t%v\n", instances.Stopped); err != nil {
			return err
		}

		if hasExactly {
			if _, err := fmt.Fprintf(w, "Exactly:\t%d\n", *instances.Exactly); err != nil {
				return err
			}
		}

		if hasMin {
			if _, err := fmt.Fprintf(w, "Min:\t%d\n", *instances.Min); err != nil {
				return err
			}
		}

		if hasMax {
			if _, err := fmt.Fprintf(w, "Max:\t%d\n", *instances.Max); err != nil {
				return err
			}
		} else if !hasExactly {
			if _, err := fmt.Fprint(w, "Max:\t∞\n"); err != nil {
				return err
			}
		}

		return nil
	})
}

// AppSpecTemplate describes the runtime configurations of the app.
func AppSpecTemplate(w io.Writer, template kfv1alpha1.AppSpecTemplate) error {

	return SectionWriter(w, string("Resource requests"), func(w io.Writer) error {
		resourceRequests := template.Spec.Containers[0].Resources.Requests
		if resourceRequests != nil {
			memory, hasMemory := resourceRequests[corev1.ResourceMemory]
			storage, hasStorage := resourceRequests[corev1.ResourceEphemeralStorage]
			cpu, hasCPU := resourceRequests[corev1.ResourceCPU]

			if hasMemory {
				if _, err := fmt.Fprintf(w, "Memory:\t%s\n", memory.String()); err != nil {
					return err
				}
			}

			if hasStorage {
				if _, err := fmt.Fprintf(w, "Storage:\t%s\n", storage.String()); err != nil {
					return err
				}
			}

			if hasCPU {
				if _, err := fmt.Fprintf(w, "CPU:\t%s\n", cpu.String()); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// HealthCheck prints a Readiness Probe in a friendly manner
func HealthCheck(w io.Writer, healthCheck *corev1.Probe) error {
	return SectionWriter(w, "Health Check", func(w io.Writer) error {
		if healthCheck == nil {
			return nil
		}

		if healthCheck.TimeoutSeconds != 0 {
			if _, err := fmt.Fprintf(w, "Timeout:\t%ds\n", healthCheck.TimeoutSeconds); err != nil {
				return err
			}
		}

		if healthCheck.TCPSocket != nil {
			if _, err := fmt.Fprintln(w, "Type:\tport (tcp)"); err != nil {
				return err
			}
		}

		if healthCheck.HTTPGet != nil {
			if _, err := fmt.Fprintln(w, "Type:\thttp"); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, "Endpoint:\t%s\n", healthCheck.HTTPGet.Path); err != nil {
				return err
			}
		}

		return nil
	})
}

// ServiceInstance
func ServiceInstance(w io.Writer, service *v1beta1.ServiceInstance) error {

	return SectionWriter(w, "Service Instance", func(w io.Writer) error {
		if service == nil {
			return nil
		}

		if _, err := fmt.Fprintf(w, "Name:\t%s\n", service.Name); err != nil {
			return err
		}

		if service.Spec.PlanReference.ClusterServiceClassExternalName != "" {
			if _, err := fmt.Fprintf(w, "Service:\t%s\n", service.Spec.PlanReference.ClusterServiceClassExternalName); err != nil {
				return err
			}
		}

		if service.Spec.PlanReference.ServiceClassExternalName != "" {
			if _, err := fmt.Fprintf(w, "Service:\t%s\n", service.Spec.PlanReference.ServiceClassExternalName); err != nil {
				return err
			}
		}

		if service.Spec.PlanReference.ClusterServicePlanExternalName != "" {
			if _, err := fmt.Fprintf(w, "Plan:\t%s\n", service.Spec.PlanReference.ClusterServicePlanExternalName); err != nil {
				return err
			}
		}

		if service.Spec.PlanReference.ServicePlanExternalName != "" {
			if _, err := fmt.Fprintf(w, "Plan:\t%s\n", service.Spec.PlanReference.ServicePlanExternalName); err != nil {
				return err
			}
		}

		if service.Spec.Parameters != nil {
			var params map[string]interface{}
			err := json.Unmarshal(service.Spec.Parameters.Raw, &params)
			if err != nil {
				return err
			}
			prettyParams, err := yaml.Marshal(params)
			if err != nil {
				return err
			}
			if err := SectionWriter(w, "Parameters", func(w io.Writer) error {
				if _, err := fmt.Fprint(w, string(prettyParams)); err != nil {
					return err
				}

				return nil
			}); err != nil {
				return err
			}
		}

		cond := services.LastStatusCondition(*service)
		if _, err := fmt.Fprintf(w, "Status:\t%s\n", cond.Reason); err != nil {
			return err
		}

		return nil
	})
}

// RouteSpecFieldsList prints a list of routes
func RouteSpecFieldsList(w io.Writer, routes []kfv1alpha1.RouteSpecFields) error {
	return SectionWriter(w, "Routes", func(w io.Writer) error {
		if routes == nil {
			return nil
		}

		return TabbedWriter(w, func(w io.Writer) error {
			if _, err := fmt.Fprintln(w, "Hostname\tDomain\tPath\tURL"); err != nil {
				return err
			}

			for _, route := range routes {
				if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					route.Hostname,
					route.Domain,
					route.Path,
					route.String()); err != nil {
					return err
				}
			}

			return nil
		})
	})
}

// MetaV1Beta1Table can print Kubernetes server-side rendered tables.
func MetaV1Beta1Table(w io.Writer, table *metav1beta1.Table) error {
	return TabbedWriter(w, func(w io.Writer) error {
		cols := []string{}
		for _, col := range table.ColumnDefinitions {
			cols = append(cols, col.Name)
		}

		if _, err := fmt.Fprintln(w, strings.Join(cols, "\t")); err != nil {
			return err
		}

		for _, row := range table.Rows {
			cells := []string{}
			for _, cell := range row.Cells {
				cells = append(cells, fmt.Sprintf("%v", cell))
			}
			if _, err := fmt.Fprintln(w, strings.Join(cells, "\t")); err != nil {
				return err
			}
		}

		return nil
	})
}
