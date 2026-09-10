data "gitea_repository_tag" "example" {
  user = "my-org"
  repo = "my-repo"
  name = "v1.0.0"
}
