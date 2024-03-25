# TODO: Update to remote backend
terraform {
  backend "local" {
    path = "config/terraform.tfstate"
  }
}