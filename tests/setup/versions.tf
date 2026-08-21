terraform {
  required_providers {
    gitea = {
      source  = "tfmm/gitea"
      version = ">= 0.9.0"
    }
  }
  required_version = ">= 0.13"
}

provider "gitea" {
  base_url = "http://localhost:3000"
  username = "gitea_admin"
  password = "gitea_admin"
  insecure = true
}
