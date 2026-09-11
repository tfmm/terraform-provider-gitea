data "gitea_repository_webhook" "example" {
  username = "my-org"
  name     = "my-repo"
  id       = 1
}
