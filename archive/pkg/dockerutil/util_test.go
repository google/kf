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

	fmt.Println(cfg.CredentialHelpers)

	// Output: map[asia.gcr.io:gcloud eu.gcr.io:gcloud gcr.io:gcloud us.gcr.io:gcloud]
}

func ExampleDescribeConfig_credHelpers() {
	cfg, err := ReadConfig(filepath.Join("testdata", "credhelpers"))
	if err != nil {
		panic(err)
	}

	DescribeConfig(os.Stdout, cfg)

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

	DescribeConfig(os.Stdout, cfg)

	// Output: Docker config:
	//   Auth:
	//     Registry        Username   Email
	//     https://gcr.io  _json_key  not@val.id
	//   Credential helpers:
	//     <none>
}
