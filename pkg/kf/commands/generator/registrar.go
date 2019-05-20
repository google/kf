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

package generator

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"path"
	"strings"

	"github.com/GoogleCloudPlatform/kf/pkg/apis/kf/v1alpha1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	yaml "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate go run ../../internal/tools/option-builder/option-builder.go options.yml register_options.go

type Command struct {
	Command *cobra.Command
	Opts    RegisterOptions
}

var commands []Command

type CommandFactory func() *cobra.Command

func Register(f CommandFactory, opts ...RegisterOption) {
	c := f()
	commands = append(commands, Command{Command: c, Opts: opts})
}

func List() []Command {
	result := make([]Command, len(commands))
	copy(result, commands)
	return result
}

func Convert(containerRegistry, commandSetName string, output io.Writer) {
	if containerRegistry == "" {
		log.Fatal("container-registry is required")
	}
	if commandSetName == "" {
		commandSetName = "kf"
	}

	encoder := yaml.NewEncoder(output)

	for _, y := range []interface{}{
		crd(),
		clusterRoleBinding("default-log-view", map[string]interface{}{
			"apiGroup": "rbac.authorization.k8s.io",
			"kind":     "ClusterRole",
			"name":     "view",
		}),
		clusterRoleBinding("default-view", map[string]interface{}{
			"apiGroup": "rbac.authorization.k8s.io",
			"kind":     "ClusterRole",
			"name":     "knative-serving-admin",
		}),
		jsonToYaml(buildCommandSet(List(), containerRegistry, commandSetName)),
	} {
		if err := encoder.Encode(y); err != nil {
			log.Fatalf("failed to encode YAML: %s", err)
		}
	}

	for _, cmd := range List() {
		if err := encoder.Encode(buildTemplate(containerRegistry, cmd)); err != nil {
			log.Fatalf("failed to encode YAML: %s", err)
		}
	}
}

func jsonToYaml(j interface{}) interface{} {
	buf := bytes.Buffer{}
	if err := json.NewEncoder(&buf).Encode(j); err != nil {
		log.Fatalf("failed tot encode to JSON: %s", err)
	}

	var y interface{}
	if err := yaml.NewDecoder(bytes.NewReader(buf.Bytes())).Decode(&y); err != nil {
		log.Fatalf("failed to decode YAML: %s", err)
	}

	return y
}

func extractCommandName(cmd Command) string {
	fields := strings.Fields(cmd.Command.Use)
	if len(fields) == 0 {
		log.Fatalf("%#v does not have a well formed usage (e.g., 'push APP_NAME')", cmd)
	}
	return fields[0]
}

// TODO: Find the actual type so we don't have to use a map[string]interface{}
func clusterRoleBinding(name string, roleRef map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"apiVersion": "rbac.authorization.k8s.io/v1",
		"kind":       "ClusterRoleBinding",
		"metadata": map[string]interface{}{
			"name": name,
		},
		"roleRef": roleRef,
		"subjects": []map[string]interface{}{
			map[string]interface{}{
				"kind":      "ServiceAccount",
				"name":      "default",
				"namespace": "default",
			},
		},
	}
}

// TODO: Find the actual type so we don't have to use a map[string]interface{}
func crd() map[string]interface{} {
	return map[string]interface{}{
		"apiVersion": "apiextensions.k8s.io/v1beta1",
		"kind":       "CustomResourceDefinition",
		"metadata": map[string]interface{}{
			"name": "commandsets.kf.dev",
		},
		"spec": map[string]interface{}{
			"group":   "kf.dev",
			"version": "v1alpha1",
			"scope":   "Namespaced",
			"names": map[string]interface{}{
				"plural":     "commandsets",
				"singular":   "commandset",
				"kind":       "CommandSet",
				"shortNames": []string{"kfcs"},
			},
		},
	}
}

// TODO: Find the actual type so we don't have to use a map[string]interface{}
func buildTemplate(containerRegistry string, cmd Command) map[string]interface{} {
	name := extractCommandName(cmd)
	imageName := path.Join(containerRegistry, "kf-"+name) + ":latest"

	return map[string]interface{}{
		"apiVersion": "build.knative.dev/v1alpha1",
		"kind":       "BuildTemplate",
		"metadata": map[string]interface{}{
			"name": name,
		},
		"spec": map[string]interface{}{
			"parameters": []map[string]interface{}{
				map[string]interface{}{
					"name":        "ARGS",
					"description": "The args JSON encoded (type []string)",
				},
				map[string]interface{}{
					"name":        "FLAGS",
					"description": "The flags JSON encoded (type map[string]interface{})",
				},
			},
			"steps": []map[string]interface{}{
				map[string]interface{}{
					"args": []string{
						"${ARGS}",
						"${FLAGS}",
					},
					"image": imageName,
					"name":  name,
				},
			},
		},
	}
}

func buildCommandSet(cmds []Command, containerRegistry, commandSetName string) v1alpha1.CommandSet {
	set := v1alpha1.CommandSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kf.dev/v1alpha1",
			Kind:       "CommandSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: commandSetName,
		},
		ContainerRegistry: containerRegistry,
	}

	for _, cmd := range cmds {
		set.Spec = append(set.Spec, buildCommandSpec(cmd))
	}

	return set
}

func buildCommandSpec(cmd Command) v1alpha1.CommandSpec {
	name := extractCommandName(cmd)
	return v1alpha1.CommandSpec{
		Name:          name,
		UploadDir:     cmd.Opts.UploadDir(),
		BuildTemplate: name,
		Use:           cmd.Command.Use,
		Short:         cmd.Command.Short,
		Long:          cmd.Command.Long,
		Flags:         buildCommandSpecFlags(cmd),
	}
}

func buildCommandSpecFlags(cmd Command) []v1alpha1.Flag {
	flags := []v1alpha1.Flag{}
	cmd.Command.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}

		// path is reserved if this command uploads the directory.
		if cmd.Opts.UploadDir() && f.Name == "path" {
			return
		}

		flags = append(flags, v1alpha1.Flag{
			Type:        f.Value.Type(),
			Default:     f.DefValue,
			Long:        f.Name,
			Short:       f.Shorthand,
			Description: f.Usage,
		})
	})
	return flags
}
