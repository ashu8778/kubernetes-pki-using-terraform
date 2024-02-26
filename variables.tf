variable "kube_api_address" {
  type=string
}

variable "kube_config_path" {
  # Only for local testing
  type=string
}
variable "cert_dir" {
  type=string
}
variable "ca_cert_dir" {
  type=string
  default = "/var/run/secrets/kubernetes.io/serviceaccount"
}
variable "users" {
  type = map(string)
}

variable "roles" {
 /* Example of role datatype
roles = {
  "user1" = {
    "pods"=["get","list"]
  },
  "user2" = {
    "pods"=["get","list"],
    "deployments"=["get","list"]
  }
}
*/
  type = map
}

variable "api_groups" {
  type = map(list(string))
}