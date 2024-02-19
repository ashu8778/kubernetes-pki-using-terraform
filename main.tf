resource "tls_private_key" "user_keys" {

  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_cert_request" "user_req" {

  private_key_pem = tls_private_key.user_keys.private_key_pem
  subject {
    common_name  = "tf"
    organization = "tf-org"
  }
}

resource "kubernetes_certificate_signing_request_v1" "user_csr" {

  metadata {
    name = "tf-csr"
  }

  spec {
    request = tls_cert_request.user_req.cert_request_pem
    signer_name = "kubernetes.io/kube-apiserver-client"
    usages    = ["digital signature", "key encipherment", "client auth"]
  }
}

resource "local_file" "kubeconfig" {

  content = <<-EOF
apiVersion: v1
clusters:
- cluster:
    certificate-authority: ca.crt 
    server: ${var.kube_api_address}
  name: my-cluster
contexts:
- context:
    cluster: my-cluster
    user: my-user
  name: my-context
current-context: my-context
kind: Config
preferences: {}
users:
- name: my-user
  user:
    client-certificate: "user.crt"
    client-key: "user.key"
EOF

  filename = "${var.cert_dir}/config"
}

resource "local_file" "user_cert" {
  content = "${kubernetes_certificate_signing_request_v1.user_csr.certificate}"
  filename = "${var.cert_dir}/user.crt"
}
resource "local_file" "user_key" {
  content = "${tls_private_key.user_keys.private_key_pem}"
  filename = "${var.cert_dir}/user.key"
}
resource "local_file" "ca_crt" {
  content = "${file("${var.ca_cert_dir}/ca.crt")}"
  filename = "${var.cert_dir}/ca.crt"
}

resource "kubernetes_role" "example" {
  metadata {
    name = "tf"
    labels = {
      test = "MyRole"
    }
  }

  rule {
    api_groups     = [""]
    resources      = ["pods"]
    verbs          = ["get", "list", "watch"]
  }
}

resource "kubernetes_role_binding" "example" {
  metadata {
    name      = "tf"
    namespace = "default"
  }
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "Role"
    name      = "tf"
  }
  subject {
    kind      = "User"
    name      = "tf"
    api_group = "rbac.authorization.k8s.io"
  }
}