resource "local_file" "user_cert" {

  for_each = toset([for user_permission in var.user_permissions : user_permission.user_name])
  content  = kubernetes_certificate_signing_request_v1.user_csr[each.key].certificate
  filename = "${var.cert_dir}/${each.key}/${each.key}.crt"
}
resource "local_file" "user_key" {

  for_each = toset([for user_permission in var.user_permissions : user_permission.user_name])
  content  = tls_private_key.user_key[each.key].private_key_pem
  filename = "${var.cert_dir}/${each.key}/${each.key}.key"
}

resource "local_file" "ca_crt" {

  for_each = toset([for user_permission in var.user_permissions : user_permission.user_name])
  content  = file("${var.cacert_path}")
  filename = "${var.cert_dir}/${each.key}/ca.crt"
}

resource "local_file" "kubeconfig" {

  for_each = toset([for user_permission in var.user_permissions : user_permission.user_name])
  content  = <<-EOF
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: ${base64encode("${file("${var.cacert_path}")}")}
    server: ${var.kube_config_api_address}
  name: my-cluster
contexts:
- context:
    cluster: my-cluster
    user: ${each.key}
  name: my-context
current-context: my-context
kind: Config
preferences: {}
users:
- name: ${each.key}
  user:
    client-certificate-data: ${base64encode("${kubernetes_certificate_signing_request_v1.user_csr[each.key].certificate}")}
    client-key-data: ${base64encode("${tls_private_key.user_key[each.key].private_key_pem}")}
EOF

  filename = "${var.cert_dir}/${each.key}/config"
}