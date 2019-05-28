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

// Package apps provides a cf compatible way of managing Knative Services as
// cf apps.
package apps

//go:generate go run ../internal/tools/option-builder/option-builder.go --pkg apps ../internal/tools/clientgen/common-options.yml zz_generated.clientoptions.go
//go:generate go run ../internal/tools/clientgen/genclient.go client.yml
