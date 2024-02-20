variable "kube_api_address" {
  type=string
}
variable "config_path" {
  type=string
}
variable "cert_dir" {
  type=string
}
variable "ca_cert_dir" {
  type=string
}
variable "users" {
  type = set(string)
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