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

package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"testing/iotest"

	kf "github.com/GoogleCloudPlatform/kf/pkg/apis/kf/v1alpha1"
	ckf "github.com/GoogleCloudPlatform/kf/pkg/client/clientset/versioned/typed/kf/v1alpha1"
	pkf "github.com/GoogleCloudPlatform/kf/pkg/kf"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/apps"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	cbuild "github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CommandOverrideFetcher interface {
	FetchCommandOverrides() (map[string]*cobra.Command, error)
}

type commandOverrideFetcher struct {
	kfClient     ckf.KfV1alpha1Interface
	buildClient  cbuild.BuildV1alpha1Interface
	tailer       pkf.BuildTailer
	imageBuilder apps.SrcImageBuilder
	params       *config.KfParams
}

func NewCommandOverrideFetcher(
	kfClient ckf.KfV1alpha1Interface,
	buildClient cbuild.BuildV1alpha1Interface,
	tailer pkf.BuildTailer,
	imageBuilder apps.SrcImageBuilder,
	params *config.KfParams,
) CommandOverrideFetcher {
	return &commandOverrideFetcher{
		kfClient:     kfClient,
		buildClient:  buildClient,
		tailer:       tailer,
		imageBuilder: imageBuilder,
		params:       params,
	}
}

func (f *commandOverrideFetcher) FetchCommandOverrides() (map[string]*cobra.Command, error) {
	commandSet, err := f.kfClient.CommandSets(f.params.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("fetching CommandSets failed: %s", err)
	}

	if l := len(commandSet.Items); l != 1 {
		return nil, fmt.Errorf("cluster is not properly setup. Expected a single CommandSet but have %d", l)
	}

	results := map[string]*cobra.Command{}
	for _, spec := range commandSet.Items[0].Spec {
		cmd, err := f.buildCommand(
			f.params.Namespace,
			f.params.Verbose,
			commandSet.Items[0],
			spec,
			f.imageBuilder,
			f.tailer,
		)
		if err != nil {
			return nil, err
		}
		results[spec.Name] = cmd
	}

	return results, nil
}

func (f *commandOverrideFetcher) buildCommand(
	namespace string,
	verbose bool,
	set kf.CommandSet,
	spec kf.CommandSpec,
	imageBuilder apps.SrcImageBuilder,
	tailer pkf.BuildTailer,
) (*cobra.Command, error) {
	var (
		dir            string
		debugKeepBuild bool
		flagNames      = map[string]bool{}
	)

	var cc = &cobra.Command{
		Use:   spec.Use,
		Short: spec.Short,
		Long:  spec.Long,
		RunE: func(cmd *cobra.Command, args []string) error {
			encodedArgs, err := json.Marshal(args)
			if err != nil {
				panic(err)
			}

			fs := map[string][]string{}
			cmd.Flags().Visit(func(f *pflag.Flag) {
				if flagNames[f.Name] {
					fs[f.Name] = []string{f.Value.String()}
				}
			})

			encodedFlags, err := json.Marshal(fs)
			if err != nil {
				panic(err)
			}

			stdoutPrefix := fmt.Sprintf("stdout-%d ", rand.Uint64())
			stderrPrefix := fmt.Sprintf("stderr-%d ", rand.Uint64())

			benvs := []corev1.EnvVar{
				{
					Name:  "NAMESPACE",
					Value: namespace,
				},
				{
					Name:  "STDOUT_PREFIX",
					Value: stdoutPrefix,
				},
				{
					Name:  "STDERR_PREFIX",
					Value: stderrPrefix,
				},
			}

			if spec.UploadDir {
				srcImageName, err := uploadDir(
					dir,
					set,
					imageBuilder,
				)
				if err != nil {
					return err
				}
				benvs = append(benvs, corev1.EnvVar{
					Name:  "SOURCE_IMAGE",
					Value: srcImageName,
				})
			}

			bi := f.buildClient.Builds(namespace)
			b, err := bi.Create(&build.Build{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "kf-" + spec.Name,
					Namespace:    namespace,
				},
				Spec: build.BuildSpec{
					Template: &build.TemplateInstantiationSpec{
						Name: spec.BuildTemplate,
						Arguments: []build.ArgumentSpec{
							{
								Name:  "ARGS",
								Value: string(encodedArgs),
							},
							{
								Name:  "FLAGS",
								Value: string(encodedFlags),
							},
						},
						Env: benvs,
					},
				},
			})

			if err != nil {
				return fmt.Errorf("failed to create build: %s", err)
			}

			// Cleanup build
			defer func() {
				if debugKeepBuild {
					log.Printf("(KF_DEBUG_KEEP_BUILD=true) Keeping build %q", b.Name)
					return
				}

				prop := metav1.DeletePropagationForeground
				if err := bi.Delete(b.Name, &metav1.DeleteOptions{
					PropagationPolicy: &prop,
				}); err != nil {
					fmt.Fprintf(cmd.OutOrStderr(), "failed to cleanup build %q: %s", b.Name, err)
				}
			}()

			// initialize verboseWriter to a nop writer
			verboseWriter := iotest.TruncateWriter(os.Stderr, 0)
			if verbose {
				verboseWriter = os.Stderr
			}

			if err := tailer.Tail(
				context.Background(),
				NewPrefixFilter(map[string]io.Writer{
					stdoutPrefix: cmd.OutOrStdout(),
					stderrPrefix: cmd.OutOrStderr(),
				},
					verboseWriter,
				),
				b.Name,
				namespace,
			); err != nil {
				return err
			}

			return nil
		},
	}

	if spec.UploadDir {
		cc.Flags().StringVar(
			&dir,
			"path",
			"",
			"The directory path that will be uploaded. Defaults to current directory.",
		)
	}

	// Debug Flags
	cc.Flags().BoolVar(
		&debugKeepBuild,
		"debug-keep-build",
		false,
		"Keeps build that runs remote command. This is useful when trying to debug commands.",
	)
	flagNames["debug-keep-build"] = true

	// TODO: Add more types.
	for _, flag := range spec.Flags {
		flagNames[flag.Long] = true

		switch flag.Type {
		case "string":
			cc.Flags().StringP(
				flag.Long,
				flag.Short,
				flag.Default, // TODO
				flag.Description,
			)
		case "stringArray":
			cc.Flags().StringArrayP(
				flag.Long,
				flag.Short,
				nil, // TODO
				flag.Description,
			)
		case "bool":
			def, err := strconv.ParseBool(flag.Default)
			if err != nil && flag.Default != "" {
				log.Fatalf("Invalid flag default for type bool in CRD. Please contact operator: %s", err)
			}

			cc.Flags().BoolP(
				flag.Long,
				flag.Short,
				def,
				flag.Description,
			)
		default:
			log.Fatalf("Invalid flag type in CRD: %s. Please contact operator.", flag.Type)
		}
	}

	return cc, nil
}

func uploadDir(
	dir string,
	set kf.CommandSet,
	imageBuilder apps.SrcImageBuilder,
) (string, error) {
	if set.ContainerRegistry == "" {
		return "", errors.New("CRD failed to configure ContainerRegistry. Please contact your operator.")
	}

	if dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %s", err)
		}
		dir = cwd
	}

	// Kontext gets fussy if you don't hand it an absolute path
	var err error
	dir, err = filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	// TODO: Cleanup container
	srcImageName := fmt.Sprintf("src-%d", rand.Uint64())
	srcImageName = path.Join(set.ContainerRegistry, srcImageName)
	log.SetPrefix("\033[32m[upload-source-code]\033[0m ")
	err = imageBuilder.BuildSrcImage(dir, srcImageName)
	log.SetPrefix("")
	if err != nil {
		return "", fmt.Errorf("failed to upload directory: %s", err)
	}

	return srcImageName, nil
}
