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

/*(
output "uaa_mysql_cmd" {
  value = "mysql -h ${google_sql_database_instance.uaa.first_ip_address} -u ${google_sql_user.uaa.name} -p --ssl-ca server-ca.pem --ssl-cert client-cert.pem --ssl-key client-key.pem"
}
*/

output "uaa_manifest" {
  value = <<EOF

---
applications:
- name: uaa
  docker:
    image: ${var.uaa_image}
  minScale: 1
  env:
    UAA_URL: http://localhost:8080
    LOGIN_URL: http://localhost:8080
    UAA_CONFIG_YAML: |
      uaa:
        url: http://localhost:8080/uaa
      issuer:
        uri: http://localhost:8080/uaa
      encryption:
        active_key_label: key-1
        encryption_keys:
          - label: key-1
            passphrase: ${random_password.uaa_encryption_key_1.result}
          - label: key-2
            passphrase: ${random_password.uaa_encryption_key_2.result}
      spring_profiles: mysql
      jwt:
        token:
          signing-key: |
            ${indent(12, tls_private_key.uaa_jwt_signing_key.private_key_pem)}
          verification-key: |
            ${indent(12, tls_private_key.uaa_jwt_signing_key.public_key_pem)}
      login:
        serviceProviderKey: |
          ${indent(10, tls_private_key.uaa_provider_private_key.private_key_pem)}
        serviceProviderKeyPassword:
        serviceProviderCertificate: |
          ${indent(10, tls_self_signed_cert.uaa_provider_cert.cert_pem)}
      oauth:
        # Always override clients on startup
        client:
          override: true
        # List of OAuth clients
        clients:
          admin:
            id: admin
            secret: ${random_password.uaa_admin_password.result}
            authorized-grant-types: client_credentials
            scope: none
            authorities: uaa.admin,clients.admin,clients.read,clients.write,clients.secret
      database:
        driverClassName: org.mariadb.jdbc.Driver
        url: jdbc:mysql://${google_sql_database_instance.uaa.first_ip_address}:3306/uaa
        username: uaa
        password: uaa
EOF

}
