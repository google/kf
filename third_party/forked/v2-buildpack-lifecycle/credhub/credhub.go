package credhub

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/buildpackapplifecycle/containerpath"
	api "code.cloudfoundry.org/credhub-cli/credhub"
	"code.cloudfoundry.org/goshims/osshim"
)

type Credhub struct {
	os      osshim.Os
	pathFor func(path ...string) string
}

func New(os osshim.Os) *Credhub {
	return &Credhub{
		os:      os,
		pathFor: containerpath.New(os.Getenv("USERPROFILE")).For,
	}
}

func (c *Credhub) InterpolateServiceRefs(credhubURI string) error {
	if c.os.Getenv("CREDHUB_SKIP_INTERPOLATION") != "" {
		return nil
	}
	if !strings.Contains(c.os.Getenv("VCAP_SERVICES"), `"credhub-ref"`) {
		return nil
	}
	ch, err := c.credhubClient(credhubURI)
	if err != nil {
		return fmt.Errorf("Unable to set up credhub client: %v", err)
	}
	interpolatedServices, err := ch.InterpolateString(c.os.Getenv("VCAP_SERVICES"))
	if err != nil {
		return fmt.Errorf("Unable to interpolate credhub references: %v", err)
	}
	if err := c.os.Setenv("VCAP_SERVICES", interpolatedServices); err != nil {
		return fmt.Errorf("Unable to update VCAP_SERVICES with interpolated credhub references: %v", err)
	}
	return nil
}

func (c *Credhub) credhubClient(credhubURI string) (*api.CredHub, error) {
	if c.os.Getenv("CF_INSTANCE_CERT") == "" || c.os.Getenv("CF_INSTANCE_KEY") == "" {
		return nil, fmt.Errorf("Missing CF_INSTANCE_CERT and/or CF_INSTANCE_KEY")
	}
	if c.os.Getenv("CF_SYSTEM_CERT_PATH") == "" {
		return nil, fmt.Errorf("Missing CF_SYSTEM_CERT_PATH")
	}

	systemCertsPath := c.pathFor(c.os.Getenv("CF_SYSTEM_CERT_PATH"))
	caCerts := []string{}
	files, err := ioutil.ReadDir(systemCertsPath)
	if err != nil {
		return nil, fmt.Errorf("Can't read contents of system cert path: %v", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".crt") {
			contents, err := ioutil.ReadFile(filepath.Join(systemCertsPath, file.Name()))
			if err != nil {
				return nil, fmt.Errorf("Can't read contents of cert in system cert path: %v", err)
			}
			caCerts = append(caCerts, string(contents))
		}
	}

	return api.New(
		credhubURI,
		api.ClientCert(c.pathFor(c.os.Getenv("CF_INSTANCE_CERT")), c.pathFor(c.os.Getenv("CF_INSTANCE_KEY"))),
		api.CaCerts(caCerts...),
	)
}
