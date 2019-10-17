package dockerutil

import (
	"fmt"
	"io"
	"sort"

	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/google/kf/pkg/kf/describe"
)

// ReadConfig is a utility function to read a docker config from a path.
// if the path is blank the default config is used.
func ReadConfig(configPath string) (*configfile.ConfigFile, error) {
	return config.Load(configPath)
}

// DescribeConfig creates a function that can write information about a
// configuration file.
func DescribeConfig(w io.Writer, cfg *configfile.ConfigFile) {
	describe.SectionWriter(w, "Docker config", func(w io.Writer) {
		describe.SectionWriter(w, "Auth", func(w io.Writer) {
			if len(cfg.AuthConfigs) == 0 {
				fmt.Fprintln(w, "<none>")
				return
			}

			var registries []string
			for registry := range cfg.AuthConfigs {
				registries = append(registries, registry)
			}
			sort.Strings(registries)

			fmt.Fprintln(w, "Registry\tUsername\tEmail")
			for _, registry := range registries {
				authConfig := cfg.AuthConfigs[registry]

				fmt.Fprintf(w, "%s\t%s\t%s\n",
					registry,
					authConfig.Username,
					authConfig.Email,
				)
			}
		})

		describe.SectionWriter(w, "Credential helpers", func(w io.Writer) {
			if len(cfg.CredentialHelpers) == 0 {
				fmt.Fprintln(w, "<none>")
				return
			}

			var registries []string
			for registry := range cfg.CredentialHelpers {
				registries = append(registries, registry)
			}
			sort.Strings(registries)

			fmt.Fprintln(w, "Registry\tHelper")
			for _, registry := range registries {
				fmt.Fprintf(w, "%s\t%s\n", registry, cfg.CredentialHelpers[registry])
			}
		})
	})
}

// DescribeDefaultConfig writes debug info about the default docker
// configuration to the given writer.
func DescribeDefaultConfig(w io.Writer) {
	cfg, err := ReadConfig("")
	if err != nil {
		fmt.Fprintf(w, "couldn't read default docker config: %v\n", err)
	} else {
		DescribeConfig(w, cfg)
	}
}
