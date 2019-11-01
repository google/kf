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
  address_type = "INTERNAL"
  prefix_length = 16
  network       = "${data.google_compute_network.default.self_link}"
}

resource "google_service_networking_connection" "uaa_db_instance_private_vpc_connection" {
  provider = "google-beta"

  network       = "${data.google_compute_network.default.self_link}"
  service       = "servicenetworking.googleapis.com"
  reserved_peering_ranges = ["${google_compute_global_address.uaa_db_instance_ip.name}"]
}

/*
resource "google_sql_database_instance" "uaa" {

  name = "${var.uaa_db_instance_name}"
  region = "${data.google_compute_subnetwork.default.region}"

  database_version = "MYSQL_5_7"

  depends_on = [
    "google_service_networking_connection.uaa_db_instance_private_vpc_connection"
  ]

  settings {
    # Second-generation instance tiers are based on the machine
    # type. See argument reference below.
    tier = "db-f1-micro"
    ip_configuration {
      require_ssl = false
      ipv4_enabled = true
      private_network = "${data.google_compute_network.default.self_link}"
    }
  }
}

resource "google_sql_user" "uaa" {
  name     = "uaa"
  instance = "${google_sql_database_instance.uaa.name}"
  password = "uaa"
}
*/
/*
resource "google_sql_ssl_cert" "uaa_client" {
  common_name = "uaa-client"
  instance    = "${google_sql_database_instance.uaa.name}"
}

resource "local_file" "server_ca" {
    content     = google_sql_ssl_cert.uaa_client.server_ca_cert
    filename = "server-ca.pem"
}

resource "local_file" "client_cert" {
    content     = google_sql_ssl_cert.uaa_client.cert
    filename = "client-cert.pem"
}

resource "local_file" "client_key" {
    content     = google_sql_ssl_cert.uaa_client.private_key
    filename = "client-key.pem"
}

*/
