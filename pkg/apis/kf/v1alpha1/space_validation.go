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

package v1alpha1

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation"
	"knative.dev/pkg/apis"
)

// Validate makes sure that Space is properly configured.
func (space *Space) Validate(ctx context.Context) (errs *apis.FieldError) {
	// If we're specifically updating status, don't reject the change because
	// of a spec issue.
	if apis.IsInStatusUpdate(ctx) {
		return
	}

	// validate name
	if space.Name == "kf" || space.Name == "default" {
		errs = errs.Also(apis.ErrInvalidValue(space.Name, "name"))
	}

	errs = errs.Also(apis.ValidateObjectMetadata(space.GetObjectMeta()).ViaField("metadata"))
	errs = errs.Also(space.Spec.Validate(apis.WithinSpec(ctx)).ViaField("spec"))

	return errs
}

// Validate makes sure that SpaceSpec is properly configured.
func (s *SpaceSpec) Validate(ctx context.Context) (errs *apis.FieldError) {
	errs = errs.Also(s.BuildConfig.Validate(ctx).ViaField("buildConfig"))
	errs = errs.Also(s.NetworkConfig.Validate(ctx).ViaField("networkConfig"))
	errs = errs.Also(s.RuntimeConfig.Validate(ctx).ViaField("runtimeConfig"))

	return errs
}

// Validate makes sure that SpaceSpecBuildConfig is properly configured.
func (s *SpaceSpecBuildConfig) Validate(ctx context.Context) (errs *apis.FieldError) {
	if s.ContainerRegistry == "" {
		errs = errs.Also(apis.ErrMissingField("containerRegistry"))
	}

	if s.ServiceAccount == "" {
		errs = errs.Also(apis.ErrMissingField("serviceAccount"))
	}

	return errs
}

func errDuplicateValue(value interface{}, fieldPath string) *apis.FieldError {
	return &apis.FieldError{
		Message: fmt.Sprintf("duplicate value: %v", value),
		Paths:   []string{fieldPath},
	}
}

// Validate implements apis.Validatable.
func (s *SpaceSpecNetworkConfig) Validate(ctx context.Context) (errs *apis.FieldError) {

	foundDomains := sets.NewString()
	for idx, domain := range s.Domains {
		if foundDomains.Has(domain.Domain) {
			errs = errs.Also(errDuplicateValue(domain.Domain, "domain").ViaFieldIndex("domains", idx))
		}

		foundDomains.Insert(domain.Domain)
	}

	errs = errs.Also(s.ValidateDomainGateways(ctx))
	errs = errs.Also(s.AppNetworkPolicy.Validate(ctx).ViaField("appNetworkPolicy"))
	errs = errs.Also(s.BuildNetworkPolicy.Validate(ctx).ViaField("buildNetworkPolicy"))

	return errs
}

// ValidateDomainGateways ensures the Istio gateway names for domains are valid.
func (s *SpaceSpecNetworkConfig) ValidateDomainGateways(ctx context.Context) (errs *apis.FieldError) {

	for idx, domain := range s.Domains {

		var gatewayNamespace string
		var gatewayName string

		if domain.GatewayName == "" {
			errs = errs.Also(apis.ErrMissingField("gatewayName").ViaFieldIndex("domains", idx))

			// no validation can continue
			continue
		}

		i := strings.Index(domain.GatewayName, "/")
		if i == -1 {
			errs = errs.Also((&apis.FieldError{
				Message: "Invalid gatewayName",
				Details: "Namespace prefix is missing",
				Paths:   []string{"gatewayName"},
			}).ViaFieldIndex("domains", idx))

			// no validation can continue
			continue
		}

		gatewayNamespace = domain.GatewayName[:i]
		gatewayName = domain.GatewayName[i+1:]

		if gatewayNamespace == "" {
			errs = errs.Also((&apis.FieldError{
				Message: "Invalid gatewayName",
				Details: "Gateway Namespace was missing",
				Paths:   []string{"gatewayName"},
			}).ViaFieldIndex("domains", idx))
		} else {
			for _, errMsg := range validation.IsDNS1123Label(gatewayNamespace) {
				errs = errs.Also((&apis.FieldError{
					Message: "Invalid namespace for gatewayName",
					Details: errMsg,
					Paths:   []string{"gatewayName"},
				}).ViaFieldIndex("domains", idx))
			}
			if gatewayNamespace != "kf" {
				errs = errs.Also((&apis.FieldError{
					Message: "Invalid namespace for gatewayName",
					Details: "Only the kf namespace is allowed",
					Paths:   []string{"gatewayName"},
				}).ViaFieldIndex("domains", idx))
			}
		}

		if gatewayName == "" {
			errs = errs.Also((&apis.FieldError{
				Message: "Invalid gatewayName",
				Details: "Gateway name was missing",
				Paths:   []string{"gatewayName"},
			}).ViaFieldIndex("domains", idx))
		} else {
			for _, errMsg := range validation.IsDNS1123Label(gatewayName) {
				errs = errs.Also((&apis.FieldError{
					Message: "Invalid name for gatewayName",
					Details: errMsg,
					Paths:   []string{"gatewayName"},
				}).ViaFieldIndex("domains", idx))
			}
		}

	}

	return errs
}

// Validate implements apis.Validatable.
func (s *SpaceSpecNetworkConfigPolicy) Validate(ctx context.Context) (errs *apis.FieldError) {
	validPolicyTypes := sets.NewString(PermitAllNetworkPolicy, DenyAllNetworkPolicy)

	if !validPolicyTypes.Has(s.Ingress) {
		errs = errs.Also(ErrInvalidEnumValue(s.Ingress, "ingress", validPolicyTypes.List()))
	}

	if !validPolicyTypes.Has(s.Egress) {
		errs = errs.Also(ErrInvalidEnumValue(s.Egress, "egress", validPolicyTypes.List()))
	}

	return
}

// Validate implements apis.Validatable.
func (s *SpaceSpecRuntimeConfig) Validate(ctx context.Context) (errs *apis.FieldError) {
	// nothing to validate
	return errs
}
