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

package envutil

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	v1alpha1 "github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
)

// EnvVarsToMap constructs a map of environment name to value from a slice of
// env vars. Vars with duplicate names will be resolved to the latest one in the
// list.
func EnvVarsToMap(envs []corev1.EnvVar) map[string]string {
	out := make(map[string]string)

	for _, env := range envs {
		out[env.Name] = env.Value
	}

	return out
}

// MapToEnvVars converts a map of name/value pairs into environment variables.
// The list will be sorted lexicographically based on name.
func MapToEnvVars(envMap map[string]string) []corev1.EnvVar {
	var out []corev1.EnvVar

	for n, v := range envMap {
		out = append(out, corev1.EnvVar{Name: n, Value: v})
	}

	SortEnvVars(out)

	return out
}

// SortEnvVars sorts the environment variable list in order by name.
func SortEnvVars(toSort []corev1.EnvVar) {
	sort.Slice(toSort, func(i, j int) bool {
		return toSort[i].Name < toSort[j].Name
	})
}

// RemoveEnvVars removes the environment variables with the given names from the
// list.
func RemoveEnvVars(varsToRemove []string, envs []corev1.EnvVar) []corev1.EnvVar {
	m := EnvVarsToMap(envs)

	for _, n := range varsToRemove {
		delete(m, n)
	}

	return MapToEnvVars(m)
}

// ParseCLIEnvVars turns a slice of strings formatted as NAME=VALUE into a map.
// The logic is taken from os/exec.dedupEnvCase with a few differences:
// malformed strings create an error, and case insensitivity is always assumed
// false.
func ParseCLIEnvVars(cliEnv []string) ([]corev1.EnvVar, error) {
	out := make(map[string]string)

	for _, kv := range cliEnv {
		parts := strings.Split(kv, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("malformed environment variable: %s", kv)
		}

		out[parts[0]] = parts[1]
	}

	return MapToEnvVars(out), nil
}

// DeduplicateEnvVars deduplicates environment variables and returns the
// canonical version of them (last environment variable takes preccidence).
func DeduplicateEnvVars(env []corev1.EnvVar) []corev1.EnvVar {
	return MapToEnvVars(EnvVarsToMap(env))
}

// NewJSONEnvVar converts a value to a JSON string and sets it on the
// environment variable.
func NewJSONEnvVar(key string, value interface{}) (corev1.EnvVar, error) {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return corev1.EnvVar{}, err
	}

	return corev1.EnvVar{Name: key, Value: string(valueBytes)}, nil
}

// GetAppEnvVars reads the environment variables off a app.
// Prefer using this function directly rather than accessing nested objects
// on app so kf can adapt to future changes.
func GetAppEnvVars(app *v1alpha1.App) []corev1.EnvVar {
	if app == nil {
		return nil
	}

	if len(app.Spec.Template.Spec.Containers) == 0 {
		return nil
	}

	return app.Spec.Template.Spec.Containers[0].Env
}

// SetAppEnvVars sets environment variables on a app.
// Prefer using this function directly rather than accessing nested objects
// on app so kf can adapt to future changes.
func SetAppEnvVars(app *v1alpha1.App, env []corev1.EnvVar) {
	if len(app.Spec.Template.Spec.Containers) == 0 {
		app.Spec.Template.Spec.Containers = []v1.Container{{}}
	}

	app.Spec.Template.Spec.Containers[0].Env = env
}
