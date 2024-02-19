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


