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

// Package route contains the reconciler for three types:
//
// * VirtualServices
// * Routes
// * Apps
//
// Unlike traditional Kubernetes resources which project out, this reconciler
// is a funnel in. There are N:1 Apps to Route and N:1 Routes
// to VirtualServices. A VirtualService represents all the hosts on a single
// domain.
//
// In order to trigger updates, this reconciler watches Routes, Apps and
// VirtualServices. It then enqueues the domain of the VirtualService that needs
// to be updated.
//
// Routes are a resource that's not controlled. They are created by Apps
// but the App doesn't retain ownership. Deleting a Route should
// (but currently doesn't) cascade to remove the route binding on Apps.
//
// VirtualServices are created by this reconciler, but aren't controlled by any
// one Route, instead they're owned by all meaning that if all are deleted,
// the VirtualService will be as well.
package route
