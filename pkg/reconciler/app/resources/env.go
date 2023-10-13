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

package resources

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/algorithms"
	cfutil "github.com/google/kf/v2/pkg/kf/cfutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/pkg/ptr"
)

// EnvRuntime describes the different environment variable runtimes Kf supports.
type EnvRuntime int

const (
	// CFRunning represents the App runtime where the App serves web traffic.
	CFRunning EnvRuntime = 1 << iota
	// CFStaging represents the App runtime where the App is being built.
	CFStaging EnvRuntime = 1 << iota
	// CFTask represents the App runtime where the App is executing as a one-time
	// Job.
	CFTask EnvRuntime = 1 << iota
)

func (e EnvRuntime) matches(o EnvRuntime) bool {
	return e&o > 0
}

var (
	memoryDivisor = resource.MustParse("1Mi")
	diskDivisor   = resource.MustParse("1Mi")
)

type envProducer func(app *v1alpha1.App) corev1.EnvVar

type runtimeEnvVar struct {
	name        string
	aliases     []string
	description string
	compute     envProducer
	runtime     EnvRuntime
}

type limits struct {
	Disk   string `json:"disk,omitempty"`
	Memory string `json:"mem,omitempty"`
}
type runtimeEnvVars []runtimeEnvVar

var _ algorithms.Interface = (runtimeEnvVars)(nil)

func (s runtimeEnvVars) Len() int {
	return len(s)
}

func (s runtimeEnvVars) Less(i, j int) bool {
	return s[i].name < s[j].name
}

func (s runtimeEnvVars) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s runtimeEnvVars) Set(i int, a algorithms.Interface, j int, b algorithms.Interface) {
	a.(runtimeEnvVars)[i] = b.(runtimeEnvVars)[j]
}

func (s runtimeEnvVars) Slice(i, j int) algorithms.Interface {
	return s[i:j]
}

func (s runtimeEnvVars) Append(a algorithms.Interface) algorithms.Interface {
	return append(runtimeEnvVars{}, s...)
}

func (s runtimeEnvVars) Clone() algorithms.Interface {
	return append(runtimeEnvVars{}, s...)
}

func staticValue(value string) envProducer {
	return func(_ *v1alpha1.App) corev1.EnvVar {
		return corev1.EnvVar{Value: value}
	}
}

func fieldRef(path string) envProducer {
	return func(_ *v1alpha1.App) corev1.EnvVar {
		return corev1.EnvVar{
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					// "v1" is the default value.
					APIVersion: "v1",
					FieldPath:  path,
				},
			},
		}
	}
}

func injectedSecretRef(path string, optional bool) envProducer {
	return func(app *v1alpha1.App) corev1.EnvVar {
		return corev1.EnvVar{
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: KfInjectedEnvSecretName(app),
					},
					Key:      path,
					Optional: ptr.Bool(optional),
				},
			},
		}
	}
}

func getRuntimeEnvVars(runtime EnvRuntime) runtimeEnvVars {
	// See: https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#envvar-v1-core
	// for the options available to K8s environment variables.

	// See: https://docs.cloudfoundry.org/devguide/deploy-apps/environment-variable.html
	// for the CFAR environment variables.

	// The following variables should be added by the buildpack:
	// PWD, USER, HOME
	// The following variables are unsupported currently:
	// TMPDIR, CF_INSTANCE_PORTS

	all := []runtimeEnvVar{
		{
			name:        "PORT",
			aliases:     []string{"VCAP_APP_PORT"},
			description: "The port the App should listen on for requests.",
			compute: func(app *v1alpha1.App) corev1.EnvVar {
				return corev1.EnvVar{
					Value: strconv.Itoa(int(getUserPort(app))),
				}
			},
			runtime: CFRunning,
		},
		{
			name:        "CF_INSTANCE_IP",
			aliases:     []string{"CF_INSTANCE_INTERNAL_IP", "VCAP_APP_HOST"},
			description: "The cluster-visible IP of the App instance.",
			compute:     fieldRef("status.podIP"),
			runtime:     CFRunning | CFStaging | CFTask,
		},
		{
			// This isn't an alias for PORT because the semantics are different even
			// though for now in Kf the result is the same.
			name:        "CF_INSTANCE_PORT",
			description: "The cluster-visible port of the App instance. In Kf this is the same as PORT.",
			compute: func(app *v1alpha1.App) corev1.EnvVar {
				return corev1.EnvVar{
					Value: strconv.Itoa(int(getUserPort(app))),
				}
			},
			runtime: CFRunning | CFStaging | CFTask,
		},
		{
			name:        "CF_INSTANCE_ADDR",
			description: "The cluster-visible IP:PORT of the App instance.",
			compute:     staticValue("$(CF_INSTANCE_IP):$(CF_INSTANCE_PORT)"),
			runtime:     CFRunning | CFStaging | CFTask,
		},
		{
			name:        "CF_INSTANCE_GUID",
			aliases:     []string{"INSTANCE_GUID"},
			description: "The UUID of the App instance.",
			compute:     fieldRef("metadata.uid"),
			runtime:     CFRunning | CFTask,
		},
		{
			name:        "CF_INSTANCE_INDEX",
			aliases:     []string{"INSTANCE_INDEX"},
			description: "The index number of the App instance, this will ALWAYS be 0.",
			compute:     staticValue("0"),
			runtime:     CFRunning,
		},
		{
			name:        "MEMORY_LIMIT",
			description: "The maximum amount of memory in MB the App can consume.",
			compute: func(app *v1alpha1.App) corev1.EnvVar {
				return corev1.EnvVar{
					ValueFrom: &corev1.EnvVarSource{
						ResourceFieldRef: &corev1.ResourceFieldSelector{
							Resource: "limits.memory",
							Divisor:  memoryDivisor,
						},
					},
				}
			},
			runtime: CFRunning | CFStaging | CFTask,
		},
		{
			name:        "DISK_LIMIT",
			description: "The maximum amount of disk storage in MB the App can use.",
			compute: func(app *v1alpha1.App) corev1.EnvVar {
				return corev1.EnvVar{
					ValueFrom: &corev1.EnvVarSource{
						ResourceFieldRef: &corev1.ResourceFieldSelector{
							Resource: "limits.ephemeral-storage",
							Divisor:  diskDivisor,
						},
					},
				}
			},
			runtime: CFRunning | CFStaging | CFTask,
		},
		{
			name:        "LANG",
			description: "Required by buildpacks to ensure consistent script load order.",
			compute:     staticValue("en_US.UTF-8"),
			runtime:     CFRunning | CFStaging | CFTask,
		},
		{
			name:        cfutil.VcapApplicationEnvVarName,
			description: "A JSON structure containing app metadata.",
			// Some VCAP_APPLICATION values are currently missing in Kf.
			// The full list of values can be found here:
			// https://docs.run.pivotal.io/devguide/deploy-apps/environment-variable.html
			compute: func(app *v1alpha1.App) corev1.EnvVar {
				appValues := cfutil.CreateVcapApplication(app)
				// add values that can only be computed at runtime
				appValues["limits"] = limits{
					Disk:   "$(DISK_LIMIT)",
					Memory: "$(MEMORY_LIMIT)",
				}

				valueBytes, _ := json.Marshal(appValues)
				jsonStr := string(valueBytes)

				// Replace limit values with unquoted env vars.
				// This ensures that disk and mem on the "limits" field are correctly represented in the JSON
				// as ints instead of strings.
				jsonWithInts := strings.ReplaceAll(jsonStr, `"$(MEMORY_LIMIT)"`, "$(MEMORY_LIMIT)")
				jsonWithInts = strings.ReplaceAll(jsonWithInts, `"$(DISK_LIMIT)"`, "$(DISK_LIMIT)")
				return corev1.EnvVar{
					Value: string(jsonWithInts),
				}
			},
			runtime: CFRunning | CFStaging | CFTask,
		},
		{
			name:        cfutil.VcapServicesEnvVarName,
			description: "A JSON structure specifying bound services.",
			compute:     injectedSecretRef(cfutil.VcapServicesEnvVarName, false),
			runtime:     CFRunning | CFStaging | CFTask,
		},
		{
			name:        cfutil.DatabaseURLEnvVarName,
			description: "The first URI found in a VCAP_SERVICES credential.",
			compute:     injectedSecretRef(cfutil.DatabaseURLEnvVarName, true),
			runtime:     CFRunning | CFTask,
		},
	}

	var out []runtimeEnvVar

	for _, varSet := range all {
		if varSet.runtime.matches(runtime) {
			out = append(out, varSet)
		}
	}

	return out
}

// BuildRuntimeEnvVars creates a list of environment variables that get injected
// when the app is running.
//
// Environment variables with aliases will be followed by others that reference
// the original.
func BuildRuntimeEnvVars(runtime EnvRuntime, app *v1alpha1.App) (out []corev1.EnvVar) {
	for _, varSet := range getRuntimeEnvVars(runtime) {
		variable := varSet.compute(app)
		variable.Name = varSet.name
		out = append(out, variable)

		// Use the variable resolution syntax for aliases so it's immediately
		// apparent that the value is just an alias of another.
		for _, alias := range varSet.aliases {
			out = append(out, corev1.EnvVar{
				Name:  alias,
				Value: fmt.Sprintf("$(%s)", varSet.name),
			})
		}
	}

	return
}

// RuntimeEnvVarList returns a list of variable names for all the built-in
// environment variables.
func RuntimeEnvVarList(runtime EnvRuntime) sets.String {
	out := sets.NewString()

	for _, varSet := range getRuntimeEnvVars(runtime) {
		out.Insert(varSet.name)
		out.Insert(varSet.aliases...)
	}

	return out
}

// RuntimeEnvVarDocs produces documentation for the injected runtime environment
// variables.
func RuntimeEnvVarDocs(runtime EnvRuntime) string {
	w := &bytes.Buffer{}

	fmt.Fprintln(w, "Kf provides the following runtime environment variables:")
	fmt.Fprintln(w, "")

	envVars := getRuntimeEnvVars(runtime)
	envVars = algorithms.Dedupe(envVars).(runtimeEnvVars)
	sort.Sort(envVars)

	for _, v := range envVars {
		fmt.Fprintf(w, " * %s: %s", v.name, v.description)
		fmt.Fprintln(w)

		for _, alias := range v.aliases {
			fmt.Fprintf(w, " * %s: Alias of %s", alias, v.name)
			fmt.Fprintln(w)
		}
	}

	return w.String()
}
