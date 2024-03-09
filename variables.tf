variable "kube_config_api_address" {
  type = string
  # for local testing in containers - used in generated kubeconfig.
  default = "https://kubernetes.default.svc"
}

# Only for local testing - remove
variable "kube_config_path" {
  type = string
}

variable "cacert_path" {
  type = string
  # for local testing in containers - used in generated kubeconfig
  default = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
}

variable "cert_dir" {
  type    = string
  default = ".user_certificates"
}

# to add user permissions
variable "user_permissions" {
  type = list(
    object({
      user_name  = string,
      user_group = string,
      role_rules = list(
        object({
          api_groups = list(string),
          resources  = list(string),
          verbs      = list(string)
        })
      )
    })
  )
}