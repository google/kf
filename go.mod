module github.com/google/kf

require (
	contrib.go.opencensus.io/exporter/stackdriver v0.12.1 // indirect
	github.com/Azure/go-autorest v11.1.2+incompatible // indirect
	github.com/MakeNowJust/heredoc v0.0.0-20171113091838-e9091a26100e
	github.com/blang/semver v3.5.1+incompatible
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/evanphx/json-patch v0.0.0-20190203023257-5858425f7550 // indirect
	github.com/fatih/color v1.7.0
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/golang/mock v1.3.1
	github.com/google/go-containerregistry v0.0.0-20190306174256-678f6c51f585
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/gofuzz v0.0.0-20170612174753-24818f796faf // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/google/wire v0.2.2
	github.com/gorilla/mux v1.7.0
	github.com/imdario/mergo v0.3.7
	github.com/knative/build v0.7.0
	github.com/knative/pkg v0.0.0-20190621200921-9c5d970cbc9e
	github.com/knative/serving v0.7.1-0.20190701162519-7ca25646a186
	github.com/manifoldco/promptui v0.3.2
	github.com/markbates/inflect v1.0.4 // indirect
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a // indirect
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/poy/kontext v0.0.0-20190801225340-1f98414f4e12
	github.com/poy/service-catalog v0.0.0-20190305064623-db385b1d332c
	github.com/rogpeppe/go-internal v1.3.0 // indirect
	github.com/russross/blackfriday v1.5.2
	github.com/segmentio/textio v1.2.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	github.com/stretchr/testify v1.3.0 // indirect
	go.opencensus.io v0.22.0 // indirect
	go.uber.org/zap v1.9.1
	golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a // indirect
	google.golang.org/appengine v1.5.0 // indirect
	k8s.io/api v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/cli-runtime v0.0.0
	k8s.io/client-go v2.0.0-alpha.0.0.20190226174127-78295b709ec6+incompatible
	k8s.io/code-generator v0.0.0
	k8s.io/gengo v0.0.0-20190116091435-f8a0810f38af // indirect
	k8s.io/klog v0.3.1 // indirect
	k8s.io/kube-openapi v0.0.0-20190228160746-b3a7cee44a30 // indirect
	knative.dev/pkg v0.0.0-20190626215608-1104d6c75533
	knative.dev/serving v0.8.0
	sigs.k8s.io/yaml v1.1.0
)

// opencensus and go-cmp are fixed to satisfy unspecified dependencies in
// knative/pkg; update once https://github.com/knative/pkg/pull/475 goes through
replace go.opencensus.io => go.opencensus.io v0.20.2

// Remove once https://github.com/google/kf/issues/238 is resolved
replace github.com/knative/pkg => github.com/poy/knative-pkg v99.0.0+incompatible

replace (
	contrib.go.opencensus.io/exporter/stackdriver => contrib.go.opencensus.io/exporter/stackdriver v0.9.2
	github.com/google/go-cmp => github.com/google/go-cmp v0.3.0
	k8s.io/api => k8s.io/api v0.0.0-20190528110122-9ad12a4af326
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190528110544-fa58353d80f3
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190221084156-01f179d85dbc
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190528110248-2d60c3dee270
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190528110732-ad79ea2fbc0f
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190226174127-78295b709ec6
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20181009000525-95810021865e
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20181003203521-0cc92547a631
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20181128191024-b1289fc74931
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190107200011-1e2bcba2af7f
	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190326082326-5c2568eea0b8
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20190109195450-94d98b9371d9
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190528110328-0ab90e449f7e
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20190528111014-463e5d26aa13
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20190528110839-96abc4c8d1a4
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20190528110942-86bc7e94eb9a
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20190528110910-f5f997cd2103
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20190416155406-4c85c9b0ae06
	k8s.io/metrics => k8s.io/metrics v0.0.0-20190528110627-05eb8901940c
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20190528110419-48d5cc0538c7
)

exclude (
	github.com/alecthomas/gometalinter v2.0.11+incompatible
	gopkg.in/alecthomas/kingpin.v3-unstable v3.0.0-20180810215634-df19058c872c
)
