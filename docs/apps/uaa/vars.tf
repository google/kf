/*
 * Copyright 2019 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the License);
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an AS IS BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

variable "poject" {
  type = string
}

variable "region" {
  type = string
}

/*
  Name of the VPC network and subnetwork used for GKE

  The CloudSQL database must be deployed in the same region as the GKE cluster
  used for UAA, so the subnet is required to make sure the regions match.
*/
variable "vpc_network_name" {
  type = string
}
variable "vpc_subnet_name" {
  type = string
}

// Name to give the CloudSQL database for UAA
variable "uaa_db_instance_name" {
  type = string
}

// Name of the container image for UAA.
// Used when generating a Kf manifest for UAA.
variable "uaa_image" {
  type = string
}

/*
  Variables for the self-signed certificate.
*/
variable "uaa_cert_common_name" {
  type = string
}
variable "uaa_cert_organization" {
  type = string
}

variable "kf_domain_name" {
  type = string
}
variable "kf_space_name" {
  type = string
}
