// +build !ignore_autogenerated

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *App) DeepCopyInto(out *App) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new App.
func (in *App) DeepCopy() *App {
	if in == nil {
		return nil
	}
	out := new(App)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *App) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppList) DeepCopyInto(out *AppList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]App, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppList.
func (in *AppList) DeepCopy() *AppList {
	if in == nil {
		return nil
	}
	out := new(AppList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AppList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppSpec) DeepCopyInto(out *AppSpec) {
	*out = *in
	in.Source.DeepCopyInto(&out.Source)
	in.Template.DeepCopyInto(&out.Template)
	in.Instances.DeepCopyInto(&out.Instances)
	if in.Routes != nil {
		in, out := &in.Routes, &out.Routes
		*out = make([]RouteSpecFields, len(*in))
		copy(*out, *in)
	}
	if in.ServiceBindings != nil {
		in, out := &in.ServiceBindings, &out.ServiceBindings
		*out = make([]AppSpecServiceBinding, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppSpec.
func (in *AppSpec) DeepCopy() *AppSpec {
	if in == nil {
		return nil
	}
	out := new(AppSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppSpecInstances) DeepCopyInto(out *AppSpecInstances) {
	*out = *in
	if in.Exactly != nil {
		in, out := &in.Exactly, &out.Exactly
		*out = new(int)
		**out = **in
	}
	if in.Min != nil {
		in, out := &in.Min, &out.Min
		*out = new(int)
		**out = **in
	}
	if in.Max != nil {
		in, out := &in.Max, &out.Max
		*out = new(int)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppSpecInstances.
func (in *AppSpecInstances) DeepCopy() *AppSpecInstances {
	if in == nil {
		return nil
	}
	out := new(AppSpecInstances)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppSpecServiceBinding) DeepCopyInto(out *AppSpecServiceBinding) {
	*out = *in
	out.InstanceRef = in.InstanceRef
	if in.Parameters != nil {
		in, out := &in.Parameters, &out.Parameters
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppSpecServiceBinding.
func (in *AppSpecServiceBinding) DeepCopy() *AppSpecServiceBinding {
	if in == nil {
		return nil
	}
	out := new(AppSpecServiceBinding)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppSpecTemplate) DeepCopyInto(out *AppSpecTemplate) {
	*out = *in
	in.Spec.DeepCopyInto(&out.Spec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppSpecTemplate.
func (in *AppSpecTemplate) DeepCopy() *AppSpecTemplate {
	if in == nil {
		return nil
	}
	out := new(AppSpecTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppStatus) DeepCopyInto(out *AppStatus) {
	*out = *in
	in.Status.DeepCopyInto(&out.Status)
	out.SourceStatusFields = in.SourceStatusFields
	out.ConfigurationStatusFields = in.ConfigurationStatusFields
	in.RouteStatusFields.DeepCopyInto(&out.RouteStatusFields)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppStatus.
func (in *AppStatus) DeepCopy() *AppStatus {
	if in == nil {
		return nil
	}
	out := new(AppStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in HTTPRoutes) DeepCopyInto(out *HTTPRoutes) {
	{
		in := &in
		*out = make(HTTPRoutes, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
		return
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPRoutes.
func (in HTTPRoutes) DeepCopy() HTTPRoutes {
	if in == nil {
		return nil
	}
	out := new(HTTPRoutes)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in OwnerReferences) DeepCopyInto(out *OwnerReferences) {
	{
		in := &in
		*out = make(OwnerReferences, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
		return
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OwnerReferences.
func (in OwnerReferences) DeepCopy() OwnerReferences {
	if in == nil {
		return nil
	}
	out := new(OwnerReferences)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Route) DeepCopyInto(out *Route) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Route.
func (in *Route) DeepCopy() *Route {
	if in == nil {
		return nil
	}
	out := new(Route)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Route) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RouteClaim) DeepCopyInto(out *RouteClaim) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RouteClaim.
func (in *RouteClaim) DeepCopy() *RouteClaim {
	if in == nil {
		return nil
	}
	out := new(RouteClaim)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RouteClaim) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RouteClaimList) DeepCopyInto(out *RouteClaimList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]RouteClaim, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RouteClaimList.
func (in *RouteClaimList) DeepCopy() *RouteClaimList {
	if in == nil {
		return nil
	}
	out := new(RouteClaimList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RouteClaimList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RouteClaimSpec) DeepCopyInto(out *RouteClaimSpec) {
	*out = *in
	out.RouteSpecFields = in.RouteSpecFields
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RouteClaimSpec.
func (in *RouteClaimSpec) DeepCopy() *RouteClaimSpec {
	if in == nil {
		return nil
	}
	out := new(RouteClaimSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RouteList) DeepCopyInto(out *RouteList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Route, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RouteList.
func (in *RouteList) DeepCopy() *RouteList {
	if in == nil {
		return nil
	}
	out := new(RouteList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RouteList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RouteSpec) DeepCopyInto(out *RouteSpec) {
	*out = *in
	out.RouteSpecFields = in.RouteSpecFields
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RouteSpec.
func (in *RouteSpec) DeepCopy() *RouteSpec {
	if in == nil {
		return nil
	}
	out := new(RouteSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RouteSpecFields) DeepCopyInto(out *RouteSpecFields) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RouteSpecFields.
func (in *RouteSpecFields) DeepCopy() *RouteSpecFields {
	if in == nil {
		return nil
	}
	out := new(RouteSpecFields)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in RouteSpecFieldsSlice) DeepCopyInto(out *RouteSpecFieldsSlice) {
	{
		in := &in
		*out = make(RouteSpecFieldsSlice, len(*in))
		copy(*out, *in)
		return
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RouteSpecFieldsSlice.
func (in RouteSpecFieldsSlice) DeepCopy() RouteSpecFieldsSlice {
	if in == nil {
		return nil
	}
	out := new(RouteSpecFieldsSlice)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in Routes) DeepCopyInto(out *Routes) {
	{
		in := &in
		*out = make(Routes, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
		return
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Routes.
func (in Routes) DeepCopy() Routes {
	if in == nil {
		return nil
	}
	out := new(Routes)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Source) DeepCopyInto(out *Source) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Source.
func (in *Source) DeepCopy() *Source {
	if in == nil {
		return nil
	}
	out := new(Source)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Source) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SourceList) DeepCopyInto(out *SourceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Source, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SourceList.
func (in *SourceList) DeepCopy() *SourceList {
	if in == nil {
		return nil
	}
	out := new(SourceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SourceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SourceSpec) DeepCopyInto(out *SourceSpec) {
	*out = *in
	out.ContainerImage = in.ContainerImage
	in.BuildpackBuild.DeepCopyInto(&out.BuildpackBuild)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SourceSpec.
func (in *SourceSpec) DeepCopy() *SourceSpec {
	if in == nil {
		return nil
	}
	out := new(SourceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SourceSpecBuildpackBuild) DeepCopyInto(out *SourceSpecBuildpackBuild) {
	*out = *in
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make([]v1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SourceSpecBuildpackBuild.
func (in *SourceSpecBuildpackBuild) DeepCopy() *SourceSpecBuildpackBuild {
	if in == nil {
		return nil
	}
	out := new(SourceSpecBuildpackBuild)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SourceSpecContainerImage) DeepCopyInto(out *SourceSpecContainerImage) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SourceSpecContainerImage.
func (in *SourceSpecContainerImage) DeepCopy() *SourceSpecContainerImage {
	if in == nil {
		return nil
	}
	out := new(SourceSpecContainerImage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SourceStatus) DeepCopyInto(out *SourceStatus) {
	*out = *in
	in.Status.DeepCopyInto(&out.Status)
	out.SourceStatusFields = in.SourceStatusFields
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SourceStatus.
func (in *SourceStatus) DeepCopy() *SourceStatus {
	if in == nil {
		return nil
	}
	out := new(SourceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SourceStatusFields) DeepCopyInto(out *SourceStatusFields) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SourceStatusFields.
func (in *SourceStatusFields) DeepCopy() *SourceStatusFields {
	if in == nil {
		return nil
	}
	out := new(SourceStatusFields)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Space) DeepCopyInto(out *Space) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Space.
func (in *Space) DeepCopy() *Space {
	if in == nil {
		return nil
	}
	out := new(Space)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Space) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SpaceDomain) DeepCopyInto(out *SpaceDomain) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SpaceDomain.
func (in *SpaceDomain) DeepCopy() *SpaceDomain {
	if in == nil {
		return nil
	}
	out := new(SpaceDomain)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in SpaceDomains) DeepCopyInto(out *SpaceDomains) {
	{
		in := &in
		*out = make(SpaceDomains, len(*in))
		copy(*out, *in)
		return
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SpaceDomains.
func (in SpaceDomains) DeepCopy() SpaceDomains {
	if in == nil {
		return nil
	}
	out := new(SpaceDomains)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SpaceList) DeepCopyInto(out *SpaceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Space, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SpaceList.
func (in *SpaceList) DeepCopy() *SpaceList {
	if in == nil {
		return nil
	}
	out := new(SpaceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SpaceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SpaceSpec) DeepCopyInto(out *SpaceSpec) {
	*out = *in
	out.Security = in.Security
	in.BuildpackBuild.DeepCopyInto(&out.BuildpackBuild)
	in.Execution.DeepCopyInto(&out.Execution)
	in.ResourceLimits.DeepCopyInto(&out.ResourceLimits)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SpaceSpec.
func (in *SpaceSpec) DeepCopy() *SpaceSpec {
	if in == nil {
		return nil
	}
	out := new(SpaceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SpaceSpecBuildpackBuild) DeepCopyInto(out *SpaceSpecBuildpackBuild) {
	*out = *in
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make([]v1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SpaceSpecBuildpackBuild.
func (in *SpaceSpecBuildpackBuild) DeepCopy() *SpaceSpecBuildpackBuild {
	if in == nil {
		return nil
	}
	out := new(SpaceSpecBuildpackBuild)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SpaceSpecExecution) DeepCopyInto(out *SpaceSpecExecution) {
	*out = *in
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make([]v1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Domains != nil {
		in, out := &in.Domains, &out.Domains
		*out = make([]SpaceDomain, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SpaceSpecExecution.
func (in *SpaceSpecExecution) DeepCopy() *SpaceSpecExecution {
	if in == nil {
		return nil
	}
	out := new(SpaceSpecExecution)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SpaceSpecResourceLimits) DeepCopyInto(out *SpaceSpecResourceLimits) {
	*out = *in
	if in.SpaceQuota != nil {
		in, out := &in.SpaceQuota, &out.SpaceQuota
		*out = make(v1.ResourceList, len(*in))
		for key, val := range *in {
			(*out)[key] = val.DeepCopy()
		}
	}
	if in.ResourceDefaults != nil {
		in, out := &in.ResourceDefaults, &out.ResourceDefaults
		*out = make([]v1.LimitRangeItem, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SpaceSpecResourceLimits.
func (in *SpaceSpecResourceLimits) DeepCopy() *SpaceSpecResourceLimits {
	if in == nil {
		return nil
	}
	out := new(SpaceSpecResourceLimits)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SpaceSpecSecurity) DeepCopyInto(out *SpaceSpecSecurity) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SpaceSpecSecurity.
func (in *SpaceSpecSecurity) DeepCopy() *SpaceSpecSecurity {
	if in == nil {
		return nil
	}
	out := new(SpaceSpecSecurity)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SpaceStatus) DeepCopyInto(out *SpaceStatus) {
	*out = *in
	in.Status.DeepCopyInto(&out.Status)
	in.Quota.DeepCopyInto(&out.Quota)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SpaceStatus.
func (in *SpaceStatus) DeepCopy() *SpaceStatus {
	if in == nil {
		return nil
	}
	out := new(SpaceStatus)
	in.DeepCopyInto(out)
	return out
}
