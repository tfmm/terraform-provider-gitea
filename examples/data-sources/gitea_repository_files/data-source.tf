data "gitea_repository_files" "example" {
  username = "my-org"
  name     = "my-repo"
  ref      = "main"
}
