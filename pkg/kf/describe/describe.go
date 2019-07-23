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
	"fmt"
	"io"
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
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
				for k, v := range meta.Labels {
					fmt.Fprintf(w, "%s=%s\n", k, v)
				}
			})
		}
	})
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
