//go:build tools
// +build tools

package tools

import (
	_ "github.com/golang/mock/mockgen"
	_ "k8s.io/code-generator/cmd/client-gen"
	_ "k8s.io/code-generator/cmd/deepcopy-gen"
	_ "k8s.io/code-generator/cmd/informer-gen"
	_ "k8s.io/code-generator/cmd/lister-gen"
	_ "knative.dev/pkg/codegen/cmd/injection-gen"
)
