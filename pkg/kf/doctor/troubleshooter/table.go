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

package troubleshooter

import (
	"context"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/doctor"
	"github.com/google/kf/v2/pkg/kf/internal/genericcli"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"knative.dev/pkg/injection/clients/dynamicclient"
)

// TroubleshootingTest returns true if the object might be affected
// by some condition or false otherwise.
type TroubleshootingTest func(obj *unstructured.Unstructured) bool

// Component contains troubleshooting information for a single Kubernetes
// component.
type Component struct {
	// Type contains the Kubernetes type represented by this component.
	Type genericcli.Type

	// Problems contains a list of possible problems that can occur with the
	// object.
	Problems []Problem
}

// DiagnoseSingle diagnoses a single Kubernetes object.
func (c *Component) DiagnoseSingle(ctx context.Context, d *doctor.Diagnostic, obj *unstructured.Unstructured) {
	d.Run(ctx, fmt.Sprintf("namespace=%q,name=%q", obj.GetNamespace(), obj.GetName()), func(ctx context.Context, d *doctor.Diagnostic) {
		for _, problem := range c.Problems {
			problem.diagnoseSingle(ctx, d, obj)
		}
	})
}

// Problem represents a high-level observable issue with the object that can be
// caused by one or more issues.
type Problem struct {
	// Description is a markdown description of the problem. This should be
	// a single sentence.
	Description string

	// Filter returns True if the problem is happening.
	Filter TroubleshootingTest

	// Causes contains a list of possible causes for the problem that can be
	// tested.
	Causes []Cause
}

// diagnoseSingle checks for the observable problem on the object and
// if it exists, looks for causes.
func (p *Problem) diagnoseSingle(ctx context.Context, d *doctor.Diagnostic, obj *unstructured.Unstructured) {
	d.Run(ctx, fmt.Sprintf("Symptom: %s", p.Description), func(ctx context.Context, d *doctor.Diagnostic) {
		if !p.Filter(obj) {
			return
		}

		for _, cause := range p.Causes {
			cause.diagnoseSingle(ctx, d, obj)
		}
	})
}

// Cause represents a single technical issue that may be causing an undesired
// observed behavior.
type Cause struct {
	// Description is a markdown description of the cause.
	Description string

	// Filter returns True if the problem is caused by this cause.
	// If nil, it will match all objects and the recommendation will be at the
	// WARN level.
	Filter TroubleshootingTest

	// Recommendation is a markdown description of what the user should do to
	// resolve the problem.
	Recommendation string
}

// diagnoseSingle checks for causes on the object and if they exist reports
// recommendations.
func (c *Cause) diagnoseSingle(ctx context.Context, d *doctor.Diagnostic, obj *unstructured.Unstructured) {
	d.Run(ctx, fmt.Sprintf("Cause: %s", c.Description), func(ctx context.Context, d *doctor.Diagnostic) {
		if c.Filter != nil && !c.Filter(obj) {
			return
		}

		message := fmt.Sprintf(
			"Cause: %s\nRecommendation: %s",
			c.Description,
			heredoc.Doc(c.Recommendation),
		)

		if c.Filter == nil {
			d.Warn(message)
		} else {
			d.Error(message)
		}
	})
}

// TroubleshootingCloser creates closures binding Components to Kubernetes
// clusters.
type TroubleshootingCloser struct {
	p *config.KfParams
}

// NewTroubleshootingCloser creates a new object that can bind Components to
// a Kubernetes cluster so troubleshooters can be executed against it.
func NewTroubleshootingCloser(p *config.KfParams) *TroubleshootingCloser {
	return &TroubleshootingCloser{
		p: p,
	}
}

// Close creates reifies the component with the informaiton necessary to
// actuate its test against the currently targeted cluster.
func (tc *TroubleshootingCloser) Close(component Component, resourceName string) doctor.Diagnosable {
	return doctor.DiagnosableFunc(func(ctx context.Context, d *doctor.Diagnostic) {
		client := genericcli.GetResourceInterface(ctx, component.Type, dynamicclient.Get(ctx), tc.p.Space)

		obj, err := client.Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			d.Fatalf("couldn't get resource %q: %v", resourceName, err)
			return
		}

		component.DiagnoseSingle(ctx, d, obj)
	})
}
