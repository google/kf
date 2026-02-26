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

package dependencies

import (
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	documentationOnly = `NOTE: this command is for documentation purposes only
	and may change format or structure without warning.`
)

// versionResolver returns a version for a dependency or an error
type versionResolver func() (string, error)

// urlResolver resolves the URL for a dependency given a version
type urlResolver func(version string) (string, error)

type dependency struct {
	// Name is the name of the dependency.
	Name string
	// ShortNames is the list of short names that can
	// resolve to this dependency.
	ShortNames []string
	// InfoURL contains the project URL page for the
	// dependency.
	InfoURL string
	// ResolveVersion is a function that can fetch the
	// current version
	// for the dependency.
	ResolveVersion versionResolver
	// ResolveURL is a function that can generate a URL
	// for the dependency given the version.
	ResolveURL urlResolver
}

func (d *dependency) names() sets.String {
	ss := sets.NewString(d.ShortNames...)
	ss.Insert(d.Name)
	return ss
}

func (d *dependency) ResolveAll() (version, url string, err error) {
	version, err = d.ResolveVersion()
	if err != nil {
		return
	}

	url, err = d.ResolveURL(version)
	return
}

func moduleVersionResolver(modulePath string) versionResolver {
	return func() (string, error) {
		bi, ok := debug.ReadBuildInfo()
		if !ok {
			return "", errors.New("couldn't get build info")
		}

		for _, mod := range bi.Deps {
			if mod.Path == modulePath {
				// descend into replace statements because the
				// first import might not be the real version if
				// it was aliased by a later replace
				for ; mod.Replace != nil; mod = mod.Replace {
				}

				return mod.Version, nil
			}
		}

		return "", fmt.Errorf("couldn't find module %q", modulePath)
	}
}

func staticVersionResolver(version string) versionResolver {
	return func() (string, error) {
		return version, nil
	}
}

func newDependencies() []dependency {
	return []dependency{
		{
			Name:       "Tekton",
			ShortNames: []string{"tekton"},
			InfoURL:    "https://tekton.dev/",
			// XXX: Ideally, we would use moduleVersionResolver, however our
			// dep matrix right now is fairly impossible. To ensure we still
			// testing against the right version though, we are going to hard
			// code this.
			ResolveVersion: staticVersionResolver("v1.9.0"),
			ResolveURL: func(version string) (string, error) {
				const URL = "https://infra.tekton.dev/tekton-releases/pipeline/previous/%s/release.yaml"
				return fmt.Sprintf(URL, version), nil
			},
		},
		{
			Name:       "Anthos Service Mesh",
			ShortNames: []string{"asm"},
			InfoURL:    "https://cloud.google.com/service-mesh/docs/gke-install-overview",
			// This version is fetched from the asmcli script. It needs to be
			// updated by hand until we have a programtic way to fetch it.
			ResolveVersion: staticVersionResolver("1.28.2-asm.4"),
			ResolveURL: func(version string) (string, error) {
				const URL = "https://github.com/GoogleCloudPlatform/anthos-service-mesh-packages/releases/tag/%s"
				return fmt.Sprintf(URL, version), nil
			},
		},
		{
			Name:       "Config Connector",
			ShortNames: []string{"kcc"},
			InfoURL:    "https://cloud.google.com//config-connector/docs/how-to/advanced-install",
			// This version needs to be updated by hand until we have a
			// programtic way to fetch it.
			ResolveVersion: staticVersionResolver("1.144.0"),
			ResolveURL: func(version string) (string, error) {
				const URL = "gs://configconnector-operator/%s/release-bundle.tar.gz"
				return fmt.Sprintf(URL, version), nil
			},
		},
	}
}

// NewDependencyCommand returns a top-level command that holds runtime
// dependencies.
func NewDependencyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Hidden: true,
		Annotations: map[string]string{
			config.SkipVersionCheckAnnotation: "",
		},
		Use:   "dependencies",
		Short: "Get Kf dependencies and links",
		Long: `Get dependencies for the Kf server side components.
		` + documentationOnly,
		SilenceUsage: true,
	}

	deps := newDependencies()

	cmd.AddCommand(
		newMatrixCommand(deps),
		newURLCommand(deps),
		newVersionCommand(deps),
	)

	return cmd
}
