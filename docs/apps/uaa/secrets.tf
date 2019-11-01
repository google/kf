resource "tls_private_key" "uaa_jwt_signing_key" {
  algorithm   = "RSA"
}

resource "random_password" "uaa_admin_password" {
  length = 16
  special = true
  override_special = "_%@"
}

resource "random_password" "uaa_encryption_key_1" {
  length = 16
  special = false
}

resource "random_password" "uaa_encryption_key_2" {
  length = 16
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
