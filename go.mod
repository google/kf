module github.com/google/kf/v2

require (
	cloud.google.com/go/compute v1.5.0
	cloud.google.com/go/logging v1.0.0
	code.cloudfoundry.org/buildpackapplifecycle v0.0.0-00010101000000-000000000000
	github.com/MakeNowJust/heredoc v0.0.0-20171113091838-e9091a26100e
	github.com/alessio/shellescape v1.4.1
	github.com/aws/smithy-go v1.11.0
	github.com/moby/term v0.5.0
	github.com/docker/cli v20.10.12+incompatible
	github.com/docker/docker v20.10.24+incompatible
	github.com/emicklei/go-restful v2.16.0+incompatible
	github.com/fatih/color v1.13.0
	github.com/gofrs/flock v0.8.1
	github.com/golang/mock v1.6.0
	github.com/google/go-cmp v0.5.7
	github.com/google/go-containerregistry v0.8.1-0.20220216220642-00c59d91847c
	github.com/google/k8s-stateless-subresource v0.0.0-00010101000000-000000000000
	github.com/google/licenseclassifier v0.0.0-20190926221455-842c0d70d702
	github.com/google/subcommands v1.0.1
	github.com/google/wire v0.4.0
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/go-version v1.3.0
	github.com/imdario/mergo v0.3.12
	github.com/mitchellh/go-wordwrap v1.0.0
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron/v3 v3.0.1
	github.com/russross/blackfriday v1.6.0
	github.com/russross/blackfriday/v2 v2.1.0
	github.com/sabhiram/go-gitignore v0.0.0-20180611051255-d3107576ba94
	github.com/segmentio/textio v1.2.0
	github.com/spf13/cobra v1.3.0
	github.com/spf13/pflag v1.0.5
	github.com/tektoncd/pipeline v0.47.1
	go.uber.org/zap v1.19.1
	golang.org/x/sync v0.0.0-20220722155255-886fb9371eb4
	golang.org/x/sys v0.5.0
	google.golang.org/api v0.70.0
	google.golang.org/genproto v0.0.0-20220303160752-862486edd9cc
	google.golang.org/protobuf v1.28.0
	istio.io/api v0.0.0-20220602172134-34645c49f9d9
	k8s.io/api v0.24.2
	k8s.io/apimachinery v0.24.2
	k8s.io/apiserver v0.24.2
	k8s.io/cli-runtime v0.24.2
	k8s.io/client-go v0.24.2
	k8s.io/code-generator v0.24.2
	k8s.io/kube-aggregator v0.24.2
	knative.dev/pkg v0.0.0-20221123011941-9d7bd235ceed
	sigs.k8s.io/controller-runtime v0.8.0
	sigs.k8s.io/controller-tools v0.6.2
	sigs.k8s.io/go-open-service-broker-client/v2 v2.0.0-20200911103215-9787cad28392
	sigs.k8s.io/yaml v1.3.0
)

replace (
	github.com/google/go-cmp => github.com/google/go-cmp v0.3.0
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20190528110419-48d5cc0538c7
)

// Uploads API server
replace github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.5.5

go 1.17

replace knative.dev/pkg => knative.dev/pkg v0.0.0-20221123011941-9d7bd235ceed

replace github.com/tektoncd/pipeline => github.com/tektoncd/pipeline v0.47.1

replace istio.io/api => istio.io/api v0.0.0-20220602172134-34645c49f9d9

replace github.com/google/k8s-stateless-subresource => ./first_party/k8s-stateless-subresource

replace github.com/spf13/cobra => ./third_party/forked/cobra

replace code.cloudfoundry.org/buildpackapplifecycle => ./third_party/forked/v2-buildpack-lifecycle
