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
	"strings"

	kfv1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

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

// AppSpecInstances describes the scaling features of the app.
func AppSpecInstances(w io.Writer, instances kfv1alpha1.AppSpecInstances) {

	SectionWriter(w, string("Scale"), func(w io.Writer) {

		hasReplicas := instances.Replicas != nil

		fmt.Fprintf(w, "Stopped?:\t%v\n", instances.Stopped)

		if hasReplicas {
			fmt.Fprintf(w, "Replicas:\t%d\n", *instances.Replicas)
		}
	})
}

// AppSpecAutoscaling describes the autoscaling features of the app.
func AppSpecAutoscaling(w io.Writer, autoscalingSpec *kfv1alpha1.AppSpecAutoscaling) {
	if autoscalingSpec == nil {
		return
	}

	SectionWriter(w, string("Autoscaling"), func(w io.Writer) {

		fmt.Fprintf(w, "Enabled?:\t%v\n", autoscalingSpec.Enabled)

		if autoscalingSpec.MinReplicas != nil {
			fmt.Fprintf(w, "MaxReplicas:\t%d\n", *autoscalingSpec.MaxReplicas)
		}

		if autoscalingSpec.MinReplicas != nil {
			fmt.Fprintf(w, "MinReplicas:\t%d\n", *autoscalingSpec.MinReplicas)
		}

		SectionWriter(w, "Rules", func(w io.Writer) {
			fmt.Fprintln(w, "RuleType\tTarget")
			for _, c := range autoscalingSpec.Rules {
				fmt.Fprintf(w, "%s\t%d\n",
					c.RuleType,
					*c.Target,
				)
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
