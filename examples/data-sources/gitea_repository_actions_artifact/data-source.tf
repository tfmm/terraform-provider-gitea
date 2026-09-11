data "gitea_repository_actions_artifact" "example" {
  owner       = "my-org"
  repo        = "my-repo"
  artifact_id = 123
}
