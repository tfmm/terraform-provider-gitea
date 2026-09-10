data "gitea_repository_subscribers" "example" {
  repository_owner = "my-org"
  repository       = "my-repo"
}
