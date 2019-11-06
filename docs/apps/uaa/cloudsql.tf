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

resource "google_compute_global_address" "uaa_db_instance_ip" {
  provider = "google-beta"

  name          = "${var.uaa_db_instance_name}-ip"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = "${data.google_compute_network.default.self_link}"
}

resource "google_service_networking_connection" "uaa_db_instance_private_vpc_connection" {
  provider = "google-beta"

  network                 = "${data.google_compute_network.default.self_link}"
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = ["${google_compute_global_address.uaa_db_instance_ip.name}"]
}

resource "random_id" "db_name_suffix" {
  byte_length = 4
}

resource "google_sql_database_instance" "uaa" {

  name   = "uaa-${random_id.db_name_suffix.hex}"
  region = "${data.google_compute_subnetwork.default.region}"

  database_version = "MYSQL_5_7"

  depends_on = [
    "google_service_networking_connection.uaa_db_instance_private_vpc_connection"
  ]

  settings {
    # Second-generation instance tiers are based on the machine
    # type. See argument reference below.
    tier              = "db-n1-standard-2"
    availability_type = "REGIONAL"

    ip_configuration {
      require_ssl  = false
      ipv4_enabled = false

      private_network = "${data.google_compute_network.default.self_link}"
    }

    backup_configuration {
      binary_log_enabled = true
      enabled            = true
    }
  }
}

resource "google_sql_database" "uaa" {
  name     = "uaa"
  instance = "${google_sql_database_instance.uaa.name}"
}

resource "google_sql_user" "uaa" {
  name     = "uaa"
  instance = "${google_sql_database_instance.uaa.name}"
  password = "${random_password.uaa_db_user_password.result}"
  host     = "%"

}
