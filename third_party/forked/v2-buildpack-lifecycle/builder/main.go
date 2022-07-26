package main

import (
	"flag"
	"fmt"
	"os"

	"code.cloudfoundry.org/buildpackapplifecycle"
	"code.cloudfoundry.org/buildpackapplifecycle/buildpackrunner"
	"code.cloudfoundry.org/buildpackapplifecycle/credhub"
	"code.cloudfoundry.org/buildpackapplifecycle/databaseuri"
	"code.cloudfoundry.org/buildpackapplifecycle/platformoptions"
	"code.cloudfoundry.org/goshims/osshim"
)

func main() {
	config := buildpackapplifecycle.NewLifecycleBuilderConfig([]string{}, false, false)

	if err := config.Parse(os.Args[1:len(os.Args)]); err != nil {
		println(err.Error())
		os.Exit(1)
	}

	if err := config.Validate(); err != nil {
		println(err.Error())
		usage()
	}

	if platformOptions, err := platformoptions.Get(os.Getenv("VCAP_PLATFORM_OPTIONS")); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid platform options: %v", err)
		os.Exit(3)
	} else if platformOptions != nil && platformOptions.CredhubURI != "" {
		err = credhub.New(&osshim.OsShim{}).InterpolateServiceRefs(platformOptions.CredhubURI)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to interpolate credhub refs: %v", err)
			os.Exit(4)
		}
	}

	if os.Getenv("VCAP_SERVICES") != "" {
		dbUri := databaseuri.New()
		if creds, err := dbUri.Credentials([]byte(os.Getenv("VCAP_SERVICES"))); err == nil {
			databaseUrl := dbUri.Uri(creds)
			if databaseUrl != "" {
				os.Setenv("DATABASE_URL", databaseUrl)
			}
		}
	}
	if err := execRunner(&config); err != nil {
		println(err.Error())
		os.Exit(buildpackapplifecycle.ExitCodeFromError(err))
	}
}

func execRunner(config *buildpackapplifecycle.LifecycleBuilderConfig) error {
	runner := buildpackrunner.New(config)
	defer runner.CleanUp()

	_, err := runner.Run()
	return err
}

func usage() {
	flag.PrintDefaults()
	os.Exit(1)
}
