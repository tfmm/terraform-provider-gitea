data "gitea_pull_request_by_base_head" "example" {
  owner = "my-org"
  repo  = "my-repo"
  base  = "main"
  head  = "feature"
}
