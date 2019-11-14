package dockerutil

import (
	"fmt"
	"os"
	"path/filepath"
)

func ExampleReadConfig() {
	cfg, err := ReadConfig(filepath.Join("testdata", "credhelpers"))
	if err != nil {
		panic(err)
	}

	if _, err := fmt.Println(cfg.CredentialHelpers); err != nil {
		panic(err)
	}

	// Output: map[asia.gcr.io:gcloud eu.gcr.io:gcloud gcr.io:gcloud us.gcr.io:gcloud]
}

func ExampleDescribeConfig_credHelpers() {
	cfg, err := ReadConfig(filepath.Join("testdata", "credhelpers"))
	if err != nil {
		panic(err)
	}

	if err := DescribeConfig(os.Stdout, cfg); err != nil {
		panic(err)
	}

	// Output: Docker config:
	//   Auth:
	//     <none>
	//   Credential helpers:
	//     Registry     Helper
	//     asia.gcr.io  gcloud
	//     eu.gcr.io    gcloud
	//     gcr.io       gcloud
	//     us.gcr.io    gcloud
}

func ExampleDescribeConfig_customAuth() {
	cfg, err := ReadConfig(filepath.Join("testdata", "customauth"))
	if err != nil {
		panic(err)
	}

	if err := DescribeConfig(os.Stdout, cfg); err != nil {
		panic(err)
	}

	// Output: Docker config:
	//   Auth:
	//     Registry        Username   Email
	//     https://gcr.io  _json_key  not@val.id
	//   Credential helpers:
	//     <none>
}
