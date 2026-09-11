resource "gitea_repository_actions_workflow_state" "example" {
  owner       = "my-org"
  repo        = "my-repo"
  workflow_id = "build.yml"
  state       = "active"
}
