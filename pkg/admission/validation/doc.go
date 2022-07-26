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

// Package kfvalidation contains validation webhook callbacks for Kf components
// that require cross-component lookups.
//
// These webhooks shouldn't be the _only_ thing protecting invalid state,
// instead, they should be used to provide better UX by failing commands
// fast.
package kfvalidation
