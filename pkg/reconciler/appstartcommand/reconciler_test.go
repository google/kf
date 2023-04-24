// Copyright 2023 Google LLC
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

package appstartcommand

import (
	"errors"
	"testing"

	containerregistryv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis/duck/v1beta1"
)

func TestReconciler_updateStartCommand(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		app         *v1alpha1.App
		imageConfig *containerregistryv1.ConfigFile
		imageErr    error

		wantStartCommandStatus v1alpha1.StartCommandStatus
	}{
		"empty app doesn't update status": {
			app:      &v1alpha1.App{},
			imageErr: errors.New("should not happen"),

			wantStartCommandStatus: v1alpha1.StartCommandStatus{},
		},
		"nop image doesn't update status": {
			app: &v1alpha1.App{
				Status: v1alpha1.AppStatus{
					BuildStatusFields: v1alpha1.BuildStatusFields{
						Image: "gcr.io/kf-releases/nop:nop",
					},
				},
			},
			imageErr:               errors.New("should not happen"),
			wantStartCommandStatus: v1alpha1.StartCommandStatus{},
		},
		"empty image doesn't update status": {
			app: &v1alpha1.App{
				Status: v1alpha1.AppStatus{
					BuildStatusFields: v1alpha1.BuildStatusFields{
						Image: "",
					},
				},
			},
			imageErr:               errors.New("should not happen"),
			wantStartCommandStatus: v1alpha1.StartCommandStatus{},
		},
		"matching image doesn't update status": {
			app: &v1alpha1.App{
				Status: v1alpha1.AppStatus{
					BuildStatusFields: v1alpha1.BuildStatusFields{
						Image: "example.com/image:v1",
					},
					StartCommands: v1alpha1.StartCommandStatus{
						Image: "example.com/image:v1",
						Error: "old error",
					},
				},
			},
			imageErr: errors.New("new error"),
			wantStartCommandStatus: v1alpha1.StartCommandStatus{
				Image: "example.com/image:v1",
				Error: "old error",
			},
		},
		"mismatched image updates status error": {
			app: &v1alpha1.App{
				Status: v1alpha1.AppStatus{
					BuildStatusFields: v1alpha1.BuildStatusFields{
						Image: "example.com/image:v2",
					},
					StartCommands: v1alpha1.StartCommandStatus{
						Image: "example.com/image:v1",
						Error: "old error",
					},
				},
			},
			imageErr: errors.New("new error"),
			wantStartCommandStatus: v1alpha1.StartCommandStatus{
				Image: "example.com/image:v2",
				Error: "new error",
			},
		},
		"mismatched image updates status": {
			app: &v1alpha1.App{
				Status: v1alpha1.AppStatus{
					BuildStatusFields: v1alpha1.BuildStatusFields{
						Image: "example.com/image:v2",
					},
					StartCommands: v1alpha1.StartCommandStatus{
						Image: "example.com/image:v1",
						Error: "old error",
					},
				},
			},
			imageConfig: &containerregistryv1.ConfigFile{
				Config: containerregistryv1.Config{
					Entrypoint: []string{"/bin/sh", "-c"},
				},
			},
			wantStartCommandStatus: v1alpha1.StartCommandStatus{
				Image:     "example.com/image:v2",
				Container: []string{"/bin/sh", "-c"},
			},
		},
		"no command on image": {
			app: &v1alpha1.App{
				Status: v1alpha1.AppStatus{
					BuildStatusFields: v1alpha1.BuildStatusFields{
						Image: "example.com/image:v2",
					},
				},
			},
			imageConfig: &containerregistryv1.ConfigFile{},
			wantStartCommandStatus: v1alpha1.StartCommandStatus{
				Image:     "example.com/image:v2",
				Container: nil,
			},
		},
		"v2 bulidpack no entrypoint": {
			app: &v1alpha1.App{
				Status: v1alpha1.AppStatus{
					BuildStatusFields: v1alpha1.BuildStatusFields{
						Image: "example.com/image:v2",
					},
				},
			},
			imageConfig: &containerregistryv1.ConfigFile{
				Config: containerregistryv1.Config{
					Labels: map[string]string{
						"StartCommand": "java -jar some-file.jar",
					},
				},
			},
			wantStartCommandStatus: v1alpha1.StartCommandStatus{
				Image:     "example.com/image:v2",
				Container: nil,
				Buildpack: []string{"java -jar some-file.jar"},
			},
		},
		"v2 bulidpack with entrypoint": {
			app: &v1alpha1.App{
				Status: v1alpha1.AppStatus{
					BuildStatusFields: v1alpha1.BuildStatusFields{
						Image: "example.com/image:v2",
					},
				},
			},
			imageConfig: &containerregistryv1.ConfigFile{
				Config: containerregistryv1.Config{
					Entrypoint: []string{"/lifecycle/launcher"},
					Labels: map[string]string{
						"StartCommand": "java -jar some-file.jar",
					},
				},
			},
			wantStartCommandStatus: v1alpha1.StartCommandStatus{
				Image:     "example.com/image:v2",
				Container: []string{"/lifecycle/launcher"},
				Buildpack: []string{"java -jar some-file.jar"},
			},
		},
		"app status remains the same": {
			app: &v1alpha1.App{
				Status: v1alpha1.AppStatus{
					Status: v1beta1.Status{
						Conditions: v1beta1.Conditions{
							{Type: "Foo", Status: corev1.ConditionTrue},
						},
					},
					BuildStatusFields: v1alpha1.BuildStatusFields{
						Image: "example.com/image:v2",
					},
				},
			},
			imageConfig: &containerregistryv1.ConfigFile{
				Config: containerregistryv1.Config{
					Entrypoint: []string{"/lifecycle/launcher"},
					Labels: map[string]string{
						"StartCommand": "java -jar some-file.jar",
					},
				},
			},
			wantStartCommandStatus: v1alpha1.StartCommandStatus{
				Image:     "example.com/image:v2",
				Container: []string{"/lifecycle/launcher"},
				Buildpack: []string{"java -jar some-file.jar"},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			r := &Reconciler{}

			var mockImageFetcher ImageConfigFetcher = func(string) (*containerregistryv1.ConfigFile, error) {
				return tc.imageConfig, tc.imageErr
			}

			original := tc.app.DeepCopy()

			r.updateStartCommand(tc.app, mockImageFetcher)

			testutil.AssertEqual(
				t,
				"status.startCommands",
				tc.wantStartCommandStatus,
				tc.app.Status.StartCommands,
			)

			// It's imperative that other fields in the status don't change to
			// avoid infinite reconciliations with the app controller.
			{
				updated := tc.app.DeepCopy()
				// Mask off fields we want to be able to change.
				original.Status.StartCommands = v1alpha1.StartCommandStatus{}
				updated.Status.StartCommands = v1alpha1.StartCommandStatus{}

				testutil.AssertEqual(
					t,
					"status",
					original.Status,
					updated.Status,
				)
			}
		})
	}
}
