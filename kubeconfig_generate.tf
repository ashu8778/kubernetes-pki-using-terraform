resource "local_file" "kubeconfig" {

  for_each = var.users
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
    user: ${each.key}
  name: my-context
current-context: my-context
kind: Config
preferences: {}
users:
- name: ${each.key}
  user:
    client-certificate: "${each.key}.crt"
    client-key: "${each.key}.key"
EOF

  filename = "${var.cert_dir}/${each.key}/config"
}