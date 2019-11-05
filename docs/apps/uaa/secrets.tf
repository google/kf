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

resource "tls_private_key" "uaa_jwt_signing_key" {
  algorithm = "RSA"
}

resource "random_password" "uaa_admin_password" {
  length           = 16
  special          = true
  override_special = "_%@"
}

resource "random_password" "uaa_encryption_key_1" {
  length  = 16
  special = false
}

resource "random_password" "uaa_encryption_key_2" {
  length  = 16
  special = false
}

resource "tls_private_key" "uaa_provider_private_key" {
  algorithm   = "ECDSA"
  ecdsa_curve = "P384"
}

resource "tls_self_signed_cert" "uaa_provider_cert" {
  key_algorithm   = "ECDSA"
  private_key_pem = "${tls_private_key.uaa_provider_private_key.private_key_pem}"

  subject {
    common_name  = "${var.uaa_cert_common_name}"
    organization = "${var.uaa_cert_organization}"
  }

  validity_period_hours = 12

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}
