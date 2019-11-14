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
func DescribeConfig(w io.Writer, cfg *configfile.ConfigFile) error {
	return describe.SectionWriter(w, "Docker config", func(w io.Writer) error {
		if err := describe.SectionWriter(w, "Auth", func(w io.Writer) error {
			if len(cfg.AuthConfigs) == 0 {
				if _, err := fmt.Fprintln(w, "<none>"); err != nil {
					return err
				}

				return nil
			}

			var registries []string
			for registry := range cfg.AuthConfigs {
				registries = append(registries, registry)
			}
			sort.Strings(registries)

			if _, err := fmt.Fprintln(w, "Registry\tUsername\tEmail"); err != nil {
				return err
			}
			for _, registry := range registries {
				authConfig, ok := cfg.AuthConfigs[registry]
				if !ok {
					if _, err := fmt.Fprintf(w, "%s\t<none>\n", registry); err != nil {
						return err
					}
				} else {
					if _, err := fmt.Fprintf(w, "%s\t%s\t%s\n",
						registry,
						authConfig.Username,
						authConfig.Email,
					); err != nil {
						return err
					}
				}
			}

			return nil
		}); err != nil {
			return err
		}

		if err := describe.SectionWriter(w, "Credential helpers", func(w io.Writer) error {
			if len(cfg.CredentialHelpers) == 0 {
				if _, err := fmt.Fprintln(w, "<none>"); err != nil {
					return err
				}
				return nil
			}

			var registries []string
			for registry := range cfg.CredentialHelpers {
				registries = append(registries, registry)
			}
			sort.Strings(registries)

			if _, err := fmt.Fprintln(w, "Registry\tHelper"); err != nil {
				return err
			}
			for _, registry := range registries {
				credentialHelper, ok := cfg.CredentialHelpers[registry]
				if !ok {
					credentialHelper = "<none>"
				}
				if _, err := fmt.Fprintf(w, "%s\t%s\n", registry, credentialHelper); err != nil {
					return err
				}
			}

			return nil
		}); err != nil {
			return err
		}

		return nil
	})
}

// DescribeDefaultConfig writes debug info about the default docker
// configuration to the given writer.
func DescribeDefaultConfig(w io.Writer) error {
	cfg, err := ReadConfig("")
	if err != nil {
		return err
	} else {
		return DescribeConfig(w, cfg)
	}
}
