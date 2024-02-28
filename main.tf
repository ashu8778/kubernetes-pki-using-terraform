resource "tls_private_key" "user_key" {

  for_each = var.users
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_cert_request" "user_req" {
  
  for_each = var.users
  private_key_pem = tls_private_key.user_key[each.key].private_key_pem
  subject {
    common_name  = "${each.key}"
    organization = "${each.value}"
  }
}

resource "kubernetes_certificate_signing_request_v1" "user_csr" {

  for_each = var.users
  metadata {
    name = "${each.key}-csr"
  }
  spec {
    request = tls_cert_request.user_req[each.key].cert_request_pem
    signer_name = "kubernetes.io/kube-apiserver-client"
    usages    = ["digital signature", "key encipherment", "client auth"]
  }
}

resource "local_file" "user_cert" {

  for_each = var.users
  content = "${kubernetes_certificate_signing_request_v1.user_csr[each.key].certificate}"
  filename = "${var.cert_dir}/${each.key}/${each.key}.crt"
}
resource "local_file" "user_key" {

  for_each = var.users
  content = "${tls_private_key.user_key[each.key].private_key_pem}"
  filename = "${var.cert_dir}/${each.key}/${each.key}.key"
}

resource "local_file" "ca_crt" {

  for_each = var.users
  content = "${file("${var.ca_cert_dir}/ca.crt")}"
  filename = "${var.cert_dir}/${each.key}/ca.crt"
}

resource "kubernetes_role" "role" {

  for_each = var.roles
  metadata {
    name = "${each.key}"
    labels = {
      role = "${each.key}"
    }
  }

   dynamic "rule" {
    for_each = each.value
    content {
    api_groups     = var.api_groups[rule.key]
    resources      = [rule.key]
    verbs          = rule.value
  }
   }
}

resource "kubernetes_role_binding" "rolebinding" {

  for_each = var.users
  metadata {
    name      = "${each.key}"
    namespace = "default"
  }
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "Role"
    name      = "${each.key}"
  }
  subject {
    kind      = "User"
    name      = "${each.key}"
    api_group = "rbac.authorization.k8s.io"
  }
}