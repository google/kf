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
      spring_profiles: default
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
      # database:
      #   driverClassName: org.mariadb.jdbc.Driver
      #   url: jdbc:mysql://google_sql_database_instance.uaa.first_ip_address:3306/uaa
      #   username: uaa
      #   password: uaa
EOF

}
