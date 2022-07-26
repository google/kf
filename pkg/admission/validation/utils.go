// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kfvalidation

// ServiceInstanceBindingInformerKey is used for associating the ServiceInstanceBindingInformer inside the context.Context.
type ServiceInstanceBindingInformerKey struct{}

// SpaceInformerKey is used for associating the SpaceInformer inside the context.Context.
type SpaceInformerKey struct{}

// AppInformerKey is used for associating the AppInformer inside the context.Context.
type AppInformerKey struct{}

// ServiceInstanceInformerKey is used for associating the ServiceInstanceInformer inside the context.Context.
type ServiceInstanceInformerKey struct{}
