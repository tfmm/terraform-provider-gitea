data "gitea_actions_jobs" "example" {
  owner  = "my-org"
  repo   = "my-repo"
  run_id = 123
}
