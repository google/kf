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

package genericcli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	configlogging "github.com/google/kf/v2/pkg/kf/commands/config/logging"
	faketableclient "github.com/google/kf/v2/pkg/kf/injection/clients/tableclient/fake"
	fakeinjection "github.com/google/kf/v2/pkg/kf/injection/fake"
	"github.com/google/kf/v2/pkg/kf/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	fakedynamicclient "knative.dev/pkg/injection/clients/dynamicclient/fake"
)

func TestNewListCommand_docs(t *testing.T) {
	typ := &genericType{
		KubernetesType: KubernetesType{
			NsScoped: true,
			Group:    "kf.dev",
			Version:  "v1alpha1",
			Kind:     "App",
			Resource: "apps",
			KfName:   "KfApp",
		},
	}

	cases := map[string]struct {
		genericType *genericType
		opts        []ListOption

		wantUse     string
		wantShort   string
		wantLong    string
		wantAliases []string
		wantExample string
	}{
		"general": {
			genericType: typ,
			wantUse:     "kfapps",
			wantShort:   "List KfApps in the targeted Space.",
			wantLong:    "List KfApps in the targeted Space.",
			wantExample: "kf kfapps",
		},
		"override alias, command name, and plural": {
			genericType: typ,
			opts: []ListOption{
				WithListAliases([]string{"abc"}),
				WithListCommandName("list-foo"),
				WithListPluralFriendlyName("foos"),
			},
			wantUse:     "list-foo",
			wantShort:   "List foos in the targeted Space.",
			wantLong:    "List foos in the targeted Space.",
			wantAliases: []string{"abc"},
			wantExample: "kf list-foo",
		},
		"override short, long, and example": {
			genericType: typ,
			opts: []ListOption{
				WithListShort("override short"),
				WithListLong("override long"),
				WithListExample("override example"),
			},
			wantUse:     "kfapps",
			wantShort:   "override short",
			wantLong:    "override long",
			wantExample: "override example",
		},
		"multiple args": {
			genericType: typ,
			opts: []ListOption{
				WithListArgumentFilters([]ListArgumentFilter{
					{Name: "APP"},
					{Name: "IMAGE"},
				}),
			},
			wantUse:     "kfapps [APP [IMAGE]]",
			wantShort:   "List KfApps in the targeted Space.",
			wantLong:    "List KfApps in the targeted Space.",
			wantExample: "kf kfapps",
		},
		"required args": {
			genericType: typ,
			opts: []ListOption{
				WithListArgumentFilters([]ListArgumentFilter{
					{Name: "APP", Required: true},
					{Name: "IMAGE"},
				}),
			},
			wantUse:     "kfapps APP [IMAGE]",
			wantShort:   "List KfApps in the targeted Space.",
			wantLong:    "List KfApps in the targeted Space.",
			wantExample: "kf kfapps",
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			cmd := NewListCommand(tc.genericType, nil, tc.opts...)
			testutil.AssertEqual(t, "use", tc.wantUse, cmd.Use)
			testutil.AssertEqual(t, "short", tc.wantShort, cmd.Short)
			testutil.AssertEqual(t, "long", tc.wantLong, cmd.Long)
			testutil.AssertEqual(t, "aliases", tc.wantAliases, cmd.Aliases)
			testutil.AssertEqual(t, "example", tc.wantExample, cmd.Example)
		})
	}
}

func TestNewListCommand(t *testing.T) {
	t.Skip()
	nsType := &genericType{
		KubernetesType: KubernetesType{
			NsScoped: true,
			Group:    "kf.dev",
			Version:  "v1alpha1",
			Kind:     "App",
			Resource: "apps",
			KfName:   "App",
		},
	}

	clusterType := &genericType{
		KubernetesType: KubernetesType{
			NsScoped: false,
			Group:    "kf.dev",
			Version:  "v1alpha1",
			Kind:     "Space",
			Resource: "spaces",
			KfName:   "Space",
		},
	}

	type mocks struct {
		p    *config.KfParams
		opts []ListOption
	}

	cases := map[string]struct {
		t       Type
		args    []string
		setup   func(context.Context, *testing.T, *mocks)
		wantOut string
		wantErr error
	}{
		"wrong number of params": {
			t:       nsType,
			args:    []string{"some", "params", "here"},
			wantErr: errors.New("accepts 0 arg(s), received 3"),
		},
		"namespace no target": {
			t:    nsType,
			args: []string{},
			setup: func(ctx context.Context, t *testing.T, mocks *mocks) {
				mocks.p.Space = ""
			},
			wantErr: errors.New(config.EmptySpaceError),
		},
		"cluster good": {
			t:    clusterType,
			args: []string{},
			setup: func(ctx context.Context, t *testing.T, mocks *mocks) {
				obj := clusterType.NewTable("some-object-name")
				faketableclient.Get(ctx).EXPECT().Table(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(obj, nil)
			},
			wantOut: `Listing Spaces
Name
some-object-name
`,
		},
		"namespace good": {
			t:    nsType,
			args: []string{},
			setup: func(ctx context.Context, t *testing.T, mocks *mocks) {
				mocks.p.Space = "my-ns"
				obj := clusterType.NewTable("some-object-name")
				faketableclient.Get(ctx).EXPECT().Table(gomock.Any(), gomock.Any(), "my-ns", gomock.Any()).Return(obj, nil)
			},
			wantOut: `Listing Apps in Space: my-ns
Name
some-object-name
`,
		},
		"custom output cluster": {
			t:    clusterType,
			args: []string{"-o", "name"},
			setup: func(ctx context.Context, t *testing.T, mocks *mocks) {
				obj := clusterType.NewUnstructured("", "some-object-name")
				fakedynamicclient.Get(ctx).
					Resource(clusterType.GroupVersionResource(context.Background())).
					Create(ctx, obj, metav1.CreateOptions{})
			},
			wantOut: `Listing Spaces
space.kf.dev/some-object-name
`,
		},
		"GVK get set": {
			t:    clusterType,
			args: []string{"-o", "json"},
			setup: func(ctx context.Context, t *testing.T, mocks *mocks) {
				obj := clusterType.NewUnstructured("", "some-object-name")
				fakedynamicclient.Get(ctx).
					Resource(clusterType.GroupVersionResource(context.Background())).
					Create(ctx, obj, metav1.CreateOptions{})
			},
			wantOut: `Listing Spaces
{
    "apiVersion": "kf.dev/v1alpha1",
    "items": [
        {
            "apiVersion": "kf.dev/v1alpha1",
            "kind": "Space",
            "metadata": {
                "name": "some-object-name"
            }
        }
    ],
    "kind": "Space",
    "metadata": {
        "resourceVersion": ""
    }
}
`,
		},
		"label filters on table": {
			t:    nsType,
			args: []string{"--app", "some-object-name"},
			setup: func(ctx context.Context, t *testing.T, mocks *mocks) {
				obj := clusterType.NewTable("some-object-name")
				faketableclient.Get(ctx).EXPECT().Table(gomock.Any(), gomock.Any(), gomock.Any(), metav1.ListOptions{
					LabelSelector: "app-label=some-object-name",
				}).Return(obj, nil)
				mocks.opts = ListOptions{
					WithListLabelFilters(map[string]string{"app": "app-label", "unused": "unused-label"}),
				}
			},
			wantOut: `Listing Apps in Space: default
Name
some-object-name
`,
		},
		"label filters on list": {
			t:    nsType,
			args: []string{"--app", "app-label-value", "-o", "json"},
			setup: func(ctx context.Context, t *testing.T, mocks *mocks) {
				obj := nsType.NewUnstructured("default", "some-object-name")
				obj.SetLabels(map[string]string{"app-label-key": "app-label-value"})
				mismatchObj := nsType.NewUnstructured("default", "mismatch")
				fakedynamicclient.Get(ctx).
					Resource(nsType.GroupVersionResource(context.Background())).
					Namespace("default").
					Create(ctx, obj, metav1.CreateOptions{})
				fakedynamicclient.Get(ctx).
					Resource(nsType.GroupVersionResource(context.Background())).
					Namespace("default").
					Create(ctx, mismatchObj, metav1.CreateOptions{})

				mocks.opts = ListOptions{
					WithListLabelFilters(map[string]string{"app": "app-label-key"}),
				}
			},
			wantOut: `Listing Apps in Space: default
{
    "apiVersion": "kf.dev/v1alpha1",
    "items": [
        {
            "apiVersion": "kf.dev/v1alpha1",
            "kind": "App",
            "metadata": {
                "labels": {
                    "app-label-key": "app-label-value"
                },
                "name": "some-object-name",
                "namespace": "default"
            }
        }
    ],
    "kind": "App",
    "metadata": {
        "resourceVersion": ""
    }
}
`,
		},
		"label requirements on list": {
			t:    nsType,
			args: []string{"-o", "json"},
			setup: func(ctx context.Context, t *testing.T, mocks *mocks) {
				obj := nsType.NewUnstructured("default", "some-object-name")
				obj.SetLabels(map[string]string{"label-key": "label-value"})
				mismatchObj := nsType.NewUnstructured("default", "mismatch")
				mismatchObj.SetLabels(map[string]string{"label-key": "other-label-value"})
				fakedynamicclient.Get(ctx).
					Resource(nsType.GroupVersionResource(context.Background())).
					Namespace("default").
					Create(ctx, obj, metav1.CreateOptions{})
				fakedynamicclient.Get(ctx).
					Resource(nsType.GroupVersionResource(context.Background())).
					Namespace("default").
					Create(ctx, mismatchObj, metav1.CreateOptions{})

				req, _ := labels.NewRequirement("label-key", selection.Equals, []string{"label-value"})
				mocks.opts = ListOptions{
					WithListLabelRequirements([]labels.Requirement{*req}),
				}
			},
			wantOut: `Listing Apps in Space: default
{
    "apiVersion": "kf.dev/v1alpha1",
    "items": [
        {
            "apiVersion": "kf.dev/v1alpha1",
            "kind": "App",
            "metadata": {
                "labels": {
                    "label-key": "label-value"
                },
                "name": "some-object-name",
                "namespace": "default"
            }
        }
    ],
    "kind": "App",
    "metadata": {
        "resourceVersion": ""
    }
}
`,
		},
		"arg filters on table": {
			t:    nsType,
			args: []string{"some-object-name"},
			setup: func(ctx context.Context, t *testing.T, mocks *mocks) {
				obj := clusterType.NewTable("some-object-name")
				faketableclient.Get(ctx).EXPECT().Table(gomock.Any(), gomock.Any(), gomock.Any(), metav1.ListOptions{
					LabelSelector: "app-label=some-object-name",
				}).Return(obj, nil)
				mocks.opts = ListOptions{
					WithListArgumentFilters([]ListArgumentFilter{
						{
							Name:    "APP",
							Handler: NewAddLabelFilter("app-label"),
						},
						{
							Name: "UNUSED",
							Handler: func(argValue string, opts *metav1.ListOptions) error {
								return errors.New("shouldn't be used!")
							},
						},
					}),
				}
			},
			wantOut: `Listing Apps in Space: default
Name
some-object-name
`,
		},
		"arg filters error": {
			t:    nsType,
			args: []string{"some-object-name"},
			setup: func(ctx context.Context, t *testing.T, mocks *mocks) {
				mocks.opts = ListOptions{
					WithListArgumentFilters([]ListArgumentFilter{
						{
							Name: "ERRORS",
							Handler: func(argValue string, opts *metav1.ListOptions) error {
								return errors.New("shouldn't be used!")
							},
						},
					}),
				}
			},
			wantErr: errors.New(`couldn't parse argument 1 value: "some-object-name": shouldn't be used!`),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			mocks := &mocks{
				p: &config.KfParams{
					Space: "default",
				},
			}

			ctx := fakeinjection.WithInjection(context.Background(), t)
			defer fakeinjection.GetController(ctx).Finish()
			buf := new(bytes.Buffer)
			ctx = configlogging.SetupLogger(ctx, buf)

			if tc.setup != nil {
				tc.setup(ctx, t, mocks)
			}

			cmd := NewListCommand(tc.t, mocks.p, mocks.opts...)
			cmd.SetOutput(buf)
			cmd.SetArgs(tc.args)
			cmd.SetContext(ctx)

			_, actualErr := cmd.ExecuteC()
			if tc.wantErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.wantErr, actualErr)
				return
			}

			testutil.AssertEqual(t, "output", buf.String(), tc.wantOut)
		})
	}
}

func TestNewAddLabelFilter(t *testing.T) {
	cases := map[string]struct {
		key          string
		value        string
		opts         metav1.ListOptions
		wantSelector string
		wantErr      error
	}{
		"bad key": {
			key:     "bad key",
			value:   "good-value",
			wantErr: errors.New(`key: Invalid value: "bad key": name part must consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyName',  or 'my.name',  or '123-abc', regex used for validation is '([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]')`),
		},
		"bad value": {
			key:     "good-key",
			value:   strings.Repeat("z", 64),
			wantErr: errors.New(`values[0][good-key]: Invalid value: "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz": must be no more than 63 characters`),
		},
		"bad parse": {
			key:   "good-key",
			value: "good-value",
			opts: metav1.ListOptions{
				LabelSelector: "a/b/c ?=? def",
			},
			wantSelector: "a/b/c ?=? def",
			wantErr:      errors.New(`unable to parse requirement: <nil>: Invalid value: "a/b/c": a qualified name must consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyName',  or 'my.name',  or '123-abc', regex used for validation is '([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]') with an optional DNS subdomain prefix and '/' (e.g. 'example.com/MyName')`),
		},
		"good parse": {
			key:   "new-key",
			value: "new-value",
			opts: metav1.ListOptions{
				LabelSelector: "old-key=old-value",
			},
			wantSelector: "new-key=new-value,old-key=old-value",
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {

			gotErr := NewAddLabelFilter(tc.key)(tc.value, &tc.opts)

			testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
			testutil.AssertEqual(t, "LabelSelector", tc.wantSelector, tc.opts.LabelSelector)
		})
	}
}
