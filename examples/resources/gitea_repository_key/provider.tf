terraform {
  required_providers {
    gitea = {
      source  = "tfmm/gitea"
      version = ">= 0.9.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "4.0.4"
    }
  }
}
