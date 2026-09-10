resource "gitea_repository_branch_protection" "example" {
  username    = "my-org"
  repository  = "my-repo"
  rule_name   = "main"
  enable_push = false
}
