data "gitea_repository_signing_key" "example" {
  repository_owner = "my-org"
  repository       = "my-repo"
}
