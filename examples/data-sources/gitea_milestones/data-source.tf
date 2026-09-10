data "gitea_milestones" "example" {
  user  = "my-org"
  repo  = "my-repo"
  state = "open"
}
