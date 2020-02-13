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
func EnvVars(w io.Writer, vars []corev1.EnvVar) {

	SectionWriter(w, "Environment", func(w io.Writer) {
		// TODO handle additional types of variables with different ValueFrom
		// fields like Secret or ConfigMap
		for _, e := range vars {
			fmt.Fprintf(w, "%s:\t%s\n", e.Name, e.Value)
		}
	})
}

// TypeMeta prints information about the type.
func TypeMeta(w io.Writer, meta metav1.TypeMeta) {
	TabbedWriter(w, func(w io.Writer) {
		fmt.Fprintf(w, "API Version:\t%s\n", meta.APIVersion)
		fmt.Fprintf(w, "Kind:\t%s\n", meta.Kind)
	})
}

// ObjectMeta prints the normal object metadata associated with standard K8s
// objects.
func ObjectMeta(w io.Writer, meta metav1.ObjectMeta) {

	SectionWriter(w, "Metadata", func(w io.Writer) {
		fmt.Fprintf(w, "Name:\t%s\n", meta.Name)
		if meta.Namespace != "" {
			fmt.Fprintf(w, "Space:\t%s\n", meta.Namespace)
		}
		fmt.Fprintf(w, "Creation Timestamp:\t%s\n", meta.CreationTimestamp)
		fmt.Fprintf(w, "Age:\t%s\n", translateTimestampSince(meta.CreationTimestamp))
		fmt.Fprintf(w, "Generation:\t%d\n", meta.Generation)
		fmt.Fprintf(w, "UID:\t%s\n", meta.UID)

		if meta.DeletionTimestamp != nil {
			fmt.Fprintf(w, "Terminating Since:\t%s\n", translateTimestampSince(*meta.DeletionTimestamp))

			if meta.DeletionGracePeriodSeconds != nil {
				fmt.Fprintf(w, "Termination Grace Period:\t%ds\n", *meta.DeletionGracePeriodSeconds)
			}
		}

		if len(meta.Labels) == 0 {
			fmt.Fprintf(w, "Labels:\t<none>\n")
		} else {
			fmt.Fprintf(w, "Labels:\t\n")

			IndentWriter(w, func(w io.Writer) {
				Labels(w, meta.Labels)
			})
		}
	})
}

// Labels prints a label map by the alphanumeric sorting of the keys.
func Labels(w io.Writer, labels map[string]string) {

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
		fmt.Fprintf(w, "%s=%s\n", pair.key, pair.value)
	}

}

// DuckStatus prints a table of status info based on the duck status.
func DuckStatus(w io.Writer, duck duckv1beta1.Status) {

	SectionWriter(w, "Status", func(w io.Writer) {
		// Print the overall status, this should be one of either ConditionReady
		// OR ConditionSucceeded. Ready is for tasks that will run continuously
		// and succeeded is for one time tasks.
		if cond := duck.GetCondition(apis.ConditionReady); cond != nil {
			duckCondition(w, *cond)
		}

		if cond := duck.GetCondition(apis.ConditionSucceeded); cond != nil {
			duckCondition(w, *cond)
		}

		// XXX: We may want to print out observedgeneration here. That might
		// confuse users so we should be careful about the wording we use to
		// present it.

		// Print the rest of the conditions in a table.
		SectionWriter(w, "Conditions", func(w io.Writer) {
			conds := duck.GetConditions()
			if len(conds) == 0 {
				return
			}

			sort.Slice(conds, func(i, j int) bool {
				return conds[i].Type < conds[j].Type
			})

			fmt.Fprintln(w, "Type\tStatus\tUpdated\tMessage\tReason")
			for _, c := range conds {
				if c.Type == apis.ConditionReady || c.Type == apis.ConditionSucceeded {
					continue
				}

				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					c.Type,
					c.Status,
					translateTimestampSince(c.LastTransitionTime.Inner),
					c.Message,
					c.Reason,
				)
			}
		})
	})
}

func duckCondition(w io.Writer, cond apis.Condition) {

	SectionWriter(w, string(cond.Type), func(w io.Writer) {
		fmt.Fprintf(w, "Ready:\t%s\n", cond.Status)

		if !cond.LastTransitionTime.Inner.IsZero() {
			fmt.Fprintf(w, "Time:\t%s\n", cond.LastTransitionTime.Inner)
		}
		if cond.Message != "" {
			fmt.Fprintf(w, "Message:\t%s\n", cond.Message)
		}
		if cond.Reason != "" {
			fmt.Fprintf(w, "Reason:\t%s\n", cond.Reason)
		}
	})
}

// SourceSpec describes the source of an application and the build process it
// will undergo.
func SourceSpec(w io.Writer, spec kfv1alpha1.SourceSpec) {

	SectionWriter(w, string("Source"), func(w io.Writer) {
		switch {
		case spec.IsContainerBuild():
			fmt.Fprintln(w, "Build Type:\tcontainer")
		case spec.IsBuildpackBuild():
			fmt.Fprintln(w, "Build Type:\tbuildpack")
		case spec.IsDockerfileBuild():
			fmt.Fprintln(w, "Build Type:\tdockerfile")
		default:
			fmt.Fprintln(w, "Build Type:\tunknown")
		}

		if spec.ServiceAccount != "" {
			fmt.Fprintf(w, "Service Account:\t%s\n", spec.ServiceAccount)
		}

		if spec.IsContainerBuild() {
			SectionWriter(w, "Container Image", func(w io.Writer) {
				containerImage := spec.ContainerImage

				fmt.Fprintf(w, "Image:\t%s\n", containerImage.Image)
			})
		}

		if spec.IsBuildpackBuild() {
			SectionWriter(w, "Buildpack Build", func(w io.Writer) {
				buildpackBuild := spec.BuildpackBuild

				fmt.Fprintf(w, "Source:\t%s\n", buildpackBuild.Source)
				fmt.Fprintf(w, "Stack:\t%s\n", buildpackBuild.Stack)
				fmt.Fprintf(w, "Bulider:\t%s\n", buildpackBuild.BuildpackBuilder)
				fmt.Fprintf(w, "Destination:\t%s\n", buildpackBuild.Image)
				EnvVars(w, buildpackBuild.Env)
			})
		}

		if spec.IsDockerfileBuild() {
			SectionWriter(w, "Dockerfile Build", func(w io.Writer) {
				build := spec.Dockerfile

				fmt.Fprintf(w, "Source:\t%s\n", build.Source)
				fmt.Fprintf(w, "Dockerfile Path:\t%s\n", build.Path)
				fmt.Fprintf(w, "Destination:\t%s\n", build.Image)
			})
		}
	})
}

// AppSpecInstances describes the scaling features of the app.
func AppSpecInstances(w io.Writer, instances kfv1alpha1.AppSpecInstances) {

	SectionWriter(w, string("Scale"), func(w io.Writer) {
		hasExactly := instances.Exactly != nil
		hasMin := instances.Min != nil
		hasMax := instances.Max != nil

		fmt.Fprintf(w, "Stopped?:\t%v\n", instances.Stopped)

		if hasExactly {
			fmt.Fprintf(w, "Exactly:\t%d\n", *instances.Exactly)
		}

		if hasMin {
			fmt.Fprintf(w, "Min:\t%d\n", *instances.Min)
		}

		if hasMax {
			fmt.Fprintf(w, "Max:\t%d\n", *instances.Max)
		} else if !hasExactly {
			fmt.Fprint(w, "Max:\t∞\n")
		}
	})
}

// AppSpecTemplate describes the runtime configurations of the app.
func AppSpecTemplate(w io.Writer, template kfv1alpha1.AppSpecTemplate) {

	SectionWriter(w, string("Resource requests"), func(w io.Writer) {
		resourceRequests := template.Spec.Containers[0].Resources.Requests
		if resourceRequests != nil {
			memory, hasMemory := resourceRequests[corev1.ResourceMemory]
			storage, hasStorage := resourceRequests[corev1.ResourceEphemeralStorage]
			cpu, hasCPU := resourceRequests[corev1.ResourceCPU]

			if hasMemory {
				fmt.Fprintf(w, "Memory:\t%s\n", memory.String())
			}

			if hasStorage {
				fmt.Fprintf(w, "Storage:\t%s\n", storage.String())
			}

			if hasCPU {
				fmt.Fprintf(w, "CPU:\t%s\n", cpu.String())
			}
		}

	})
}

// HealthCheck prints a Readiness Probe in a friendly manner
func HealthCheck(w io.Writer, healthCheck *corev1.Probe) {
	SectionWriter(w, "Health Check", func(w io.Writer) {
		if healthCheck == nil {
			return
		}

		if healthCheck.TimeoutSeconds != 0 {
			fmt.Fprintf(w, "Timeout:\t%ds\n", healthCheck.TimeoutSeconds)
		}

		if healthCheck.TCPSocket != nil {
			fmt.Fprintln(w, "Type:\tport (tcp)")
		}

		if healthCheck.HTTPGet != nil {
			fmt.Fprintln(w, "Type:\thttp")
			fmt.Fprintf(w, "Endpoint:\t%s\n", healthCheck.HTTPGet.Path)
		}
	})
}

// ServiceInstance
func ServiceInstance(w io.Writer, service *v1beta1.ServiceInstance) {
	SectionWriter(w, "Service Instance", func(w io.Writer) {
		if service == nil {
			return
		}

		fmt.Fprintf(w, "Name:\t%s\n", service.Name)

		if service.Spec.PlanReference.ClusterServiceClassExternalName != "" {
			fmt.Fprintf(w, "Service:\t%s\n", service.Spec.PlanReference.ClusterServiceClassExternalName)
		}

		if service.Spec.PlanReference.ServiceClassExternalName != "" {
			fmt.Fprintf(w, "Service:\t%s\n", service.Spec.PlanReference.ServiceClassExternalName)
		}

		if service.Spec.PlanReference.ClusterServicePlanExternalName != "" {
			fmt.Fprintf(w, "Plan:\t%s\n", service.Spec.PlanReference.ClusterServicePlanExternalName)
		}

		if service.Spec.PlanReference.ServicePlanExternalName != "" {
			fmt.Fprintf(w, "Plan:\t%s\n", service.Spec.PlanReference.ServicePlanExternalName)
		}

		if service.Spec.Parameters != nil {
			var params map[string]interface{}
			err := json.Unmarshal(service.Spec.Parameters.Raw, &params)
			if err != nil {
				panic(err)
			}
			prettyParams, err := yaml.Marshal(params)
			if err != nil {
				panic(err)
			}
			SectionWriter(w, "Parameters", func(w io.Writer) {
				fmt.Fprintf(w, string(prettyParams))
			})
		}

		cond := services.LastStatusCondition(*service)
		fmt.Fprintf(w, "Status:\t%s\n", cond.Reason)
	})
}

// RouteSpecFieldsList prints a list of routes
func RouteSpecFieldsList(w io.Writer, routes []kfv1alpha1.RouteSpecFields) {
	SectionWriter(w, "Routes", func(w io.Writer) {
		if routes == nil {
			return
		}

		TabbedWriter(w, func(w io.Writer) {
			fmt.Fprintln(w, "Hostname\tDomain\tPath\tURL")

			for _, route := range routes {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					route.Hostname,
					route.Domain,
					route.Path,
					route.String())
			}
		})
	})
}

// MetaV1Beta1Table can print Kubernetes server-side rendered tables.
func MetaV1Beta1Table(w io.Writer, table *metav1beta1.Table) error {
	TabbedWriter(w, func(w io.Writer) {
		cols := []string{}
		for _, col := range table.ColumnDefinitions {
			cols = append(cols, col.Name)
		}

		fmt.Fprintln(w, strings.Join(cols, "\t"))

		for _, row := range table.Rows {
			cells := []string{}
			for _, cell := range row.Cells {
				cells = append(cells, fmt.Sprintf("%v", cell))
			}
			fmt.Fprintln(w, strings.Join(cells, "\t"))
		}
	})

	return nil
}
