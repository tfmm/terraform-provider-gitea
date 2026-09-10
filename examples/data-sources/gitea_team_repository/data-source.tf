data "gitea_team_repository" "example" {
  team_id          = 123
  repository_owner = "my-org"
  repository       = "my-repo"
}
