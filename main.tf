resource "tls_private_key" "user_key" {

  for_each  = toset([for user_permission in var.user_permissions : user_permission.user_name])
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_cert_request" "user_req" {

  for_each        = { for user_permission in var.user_permissions : user_permission.user_name => user_permission.user_group }
  private_key_pem = tls_private_key.user_key[each.key].private_key_pem
  subject {
    common_name  = each.key
    organization = each.value
  }
}

resource "kubernetes_certificate_signing_request_v1" "user_csr" {

  for_each = toset([for user_permission in var.user_permissions : user_permission.user_name])
  metadata {
    name = "${each.key}-csr"
  }
  spec {
    request     = tls_cert_request.user_req[each.key].cert_request_pem
    signer_name = "kubernetes.io/kube-apiserver-client"
    usages      = ["digital signature", "key encipherment", "client auth"]
  }
}

resource "kubernetes_role" "role" {

  for_each = { for user_permission in var.user_permissions : user_permission.user_name => user_permission.role_rules }
  metadata {
    name = each.key
    labels = {
      role = "${each.key}"
    }
  }

  dynamic "rule" {
    for_each = { for index, role_rule in each.value : index => role_rule }
    content {
      api_groups = rule.value["api_groups"]
      resources  = rule.value["resources"]
      verbs      = rule.value["verbs"]
    }
  }
}

resource "kubernetes_role_binding" "rolebinding" {

  for_each = toset([for user_permission in var.user_permissions : user_permission.user_name])
  metadata {
    name      = each.key
    namespace = "default"
  }
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "Role"
    name      = each.key
  }
  subject {
    kind      = "User"
    name      = each.key
    api_group = "rbac.authorization.k8s.io"
  }
}