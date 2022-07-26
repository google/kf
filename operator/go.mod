module kf-operator

go 1.16

require (
	cloud.google.com/go/compute v1.5.0
	github.com/Masterminds/semver/v3 v3.0.3
	github.com/docker/distribution v2.8.0+incompatible
	github.com/go-logr/zapr v1.2.2
	github.com/golang/mock v1.6.0
	github.com/google/go-cmp v0.5.7
	github.com/hashicorp/go-multierror v1.1.1
	github.com/manifestival/client-go-client v0.5.0
	github.com/manifestival/manifestival v0.7.1
	github.com/pkg/errors v0.9.1
	go.uber.org/zap v1.19.1
	k8s.io/api v0.23.8
	k8s.io/apiextensions-apiserver v0.23.8
	k8s.io/apimachinery v0.23.8
	k8s.io/client-go v0.23.8
	k8s.io/kube-aggregator v0.23.5
	k8s.io/kube-openapi v0.0.0-20220124234850-424119656bbf // indirect
	k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9
	knative.dev/operator v0.27.1-0.20211210172029-2f2c8b8fc83f
	knative.dev/pkg v0.0.0-20220621173822-9c5a7317fa9d
	sigs.k8s.io/yaml v1.3.0
)

replace github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.5.5
