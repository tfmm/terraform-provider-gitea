terraform {
  required_providers {
    gitea = {
      source  = "tfmm/gitea"
      version = ">= 0.9.0"
    }
  }
}

provider "gitea" {
  base_url = var.gitea_url
  username = var.gitea_username
  password = var.gitea_password
}
