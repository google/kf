module github.com/GoogleCloudPlatform/kf

require (
	contrib.go.opencensus.io/exporter/stackdriver v0.12.1 // indirect
	github.com/MakeNowJust/heredoc v0.0.0-20171113091838-e9091a26100e // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/docker/docker v0.7.3-0.20190307005417-54dddadc7d5d // indirect
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/elazarl/goproxy v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/elazarl/goproxy/ext v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/evanphx/json-patch v4.1.0+incompatible // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/golang/mock v1.3.1
	github.com/google/go-containerregistry v0.0.0-20190306174256-678f6c51f585
	github.com/google/uuid v1.1.1 // indirect
	github.com/google/wire v0.2.2
	github.com/imdario/mergo v0.3.7
	github.com/knative/build v0.6.0
	github.com/knative/pkg v0.0.0-20190621200921-9c5d970cbc9e
	github.com/knative/serving v0.6.1
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a // indirect
	github.com/mattn/go-runewidth v0.0.4 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/pkg/errors v0.8.1 // indirect
	github.com/poy/kontext v0.0.0-20190322194304-59ced15e96b1
	github.com/poy/service-catalog v0.0.0-20190305064623-db385b1d332c
	github.com/segmentio/textio v1.2.0
	github.com/sirupsen/logrus v1.3.0 // indirect
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3 // indirect
	go.opencensus.io v0.22.0 // indirect
	go.uber.org/atomic v1.3.2 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.9.1
	golang.org/x/crypto v0.0.0-20190426145343-a29dc8fdc734 // indirect
	golang.org/x/net v0.0.0-20190503192946-f4e77d36d62c // indirect
	golang.org/x/sys v0.0.0-20190502175342-a43fa875dd82 // indirect
	google.golang.org/appengine v1.5.0 // indirect
	gopkg.in/yaml.v2 v2.2.2
	gotest.tools v2.2.0+incompatible // indirect
	k8s.io/api v0.0.0-20181121191454-a61488babbd6
	k8s.io/apimachinery v0.0.0-20181127025237-2b1284ed4c93
	k8s.io/client-go v2.0.0-alpha.0.0.20190226174127-78295b709ec6+incompatible
	k8s.io/klog v0.3.0 // indirect
	k8s.io/kubernetes v1.14.2
)

// opencensus and go-cmp are fixed to satisfy unspecified dependencies in
// knative/pkg; update once https://github.com/knative/pkg/pull/475 goes through
replace go.opencensus.io => go.opencensus.io v0.20.2

replace contrib.go.opencensus.io/exporter/stackdriver => contrib.go.opencensus.io/exporter/stackdriver v0.9.2

replace github.com/google/go-cmp => github.com/google/go-cmp v0.3.0
