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

package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiregistrationv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
)

// APIServiceOption enables further configuration of an APIService.
type APIServiceOption func(*apiregistrationv1.APIService)

// APIService creates an APIService and then applies APIServiceOptions to it.
func APIService(name string, do ...APIServiceOption) *apiregistrationv1.APIService {
	s := &apiregistrationv1.APIService{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	for _, opt := range do {
		opt(s)
	}
	return s
}

// WithAPIServiceInsecuretSkipTLSVerify creates a APIServiceOption that sets
// the InsecureSkipTLSVerify field.
func WithAPIServiceInsecuretSkipTLSVerify(val bool) APIServiceOption {
	return func(s *apiregistrationv1.APIService) {
		s.Spec.InsecureSkipTLSVerify = val
	}
}

// WithAPIServiceCABundle creates a APIServiceOption that sets the CABundle
// field.
func WithAPIServiceCABundle(val []byte) APIServiceOption {
	return func(s *apiregistrationv1.APIService) {
		s.Spec.CABundle = val
	}
}
