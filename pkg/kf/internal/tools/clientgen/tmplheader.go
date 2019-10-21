// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package clientgen

import (
	"text/template"

	"github.com/google/kf/pkg/kf/internal/tools/generator"
)

var headerTemplate = template.Must(template.New("").Funcs(generator.TemplateFuncs()).Parse(`
{{genlicense}}

{{gennotice "functions.go"}}

package {{.Package}}

// Generator defined imports
import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"knative.dev/pkg/kmp"
	{{ if .SupportsConditions }}"knative.dev/pkg/apis"
	corev1 "k8s.io/api/core/v1"{{ end }}
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)


// User defined imports
{{genimports .Imports}}

`))
