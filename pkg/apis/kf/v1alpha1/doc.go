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

// +k8s:deepcopy-gen=package
// +k8s:defaulter-gen=TypeMeta
// +groupName=kf.dev

//go:generate go run ../../../kf/internal/tools/conditiongen/generator.go --pkg v1alpha1 --status-type AppStatus --prefix App Build Service ServiceAccount Deployment Space Route EnvVarSecret ServiceInstanceBindings HorizontalPodAutoscaler
//go:generate go run ../../../kf/internal/tools/conditiongen/generator.go --pkg v1alpha1 --status-type SpaceStatus --prefix Space Namespace BuildServiceAccount BuildSecret BuildRole BuildRoleBinding IngressGateway RuntimeConfig NetworkConfig BuildConfig BuildNetworkPolicy AppNetworkPolicy RoleBindings ClusterRole ClusterRoleBindings IAMPolicy
//go:generate go run ../../../kf/internal/tools/conditiongen/generator.go --pkg v1alpha1 --status-type BuildStatus --prefix Build --batch=true Space TaskRun SourcePackage
//go:generate go run ../../../kf/internal/tools/conditiongen/generator.go --pkg v1alpha1 --status-type ServiceInstanceStatus --prefix ServiceInstance Space BackingResource ParamsSecret ParamsSecretPopulated
//go:generate go run ../../../kf/internal/tools/conditiongen/generator.go --pkg v1alpha1 --status-type ServiceInstanceBindingStatus --prefix ServiceInstanceBinding ServiceInstance BackingResource ParamsSecret ParamsSecretPopulated CredentialsSecret VolumeParamsPopulated
//go:generate go run ../../../kf/internal/tools/conditiongen/generator.go --pkg v1alpha1 --status-type RouteStatus --prefix Route VirtualService SpaceDomain RouteService
//go:generate go run ../../../kf/internal/tools/conditiongen/generator.go --pkg v1alpha1 --status-type CommonServiceBrokerStatus --prefix CommonServiceBroker CredsSecret CredsSecretPopulated Catalog

// Tasks don't technically need the PipelineRun status, but we need to keep it until v2.8.0 as we required it before and
// don't want any tasks while we upgrade to fail.
//go:generate go run ../../../kf/internal/tools/conditiongen/generator.go --pkg v1alpha1 --status-type TaskStatus --prefix Task --batch=true Space PipelineRun TaskRun App Config

//go:generate go run ../../../kf/internal/tools/conditiongen/generator.go --pkg v1alpha1 --status-type SourcePackageStatus --prefix SourcePackage --batch=true Upload
//go:generate go run ../../../kf/internal/tools/conditiongen/generator.go --pkg v1alpha1 --status-type TaskScheduleStatus --prefix TaskSchedule Space

package v1alpha1
