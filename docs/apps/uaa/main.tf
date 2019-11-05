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

provider "google-beta" {
  credentials = "${file("account.json")}"
  project     = "kf-source"
}

provider "google" {
  credentials = "${file("account.json")}"
  project     = "kf-source"
}

data "google_compute_network" "default" {
  name = "${var.vpc_network_name}"
}

data "google_compute_subnetwork" "default" {
  name   = "${var.vpc_subnet_name}"
  region = "${var.region}"
}
