resource "gitea_repository_branch_protection" "example" {
  username    = "my-org"
  name        = "my-repo"
  rule_name   = "main"
  enable_push = false

  required_approvals         = 1
  block_on_codeowner_reviews = true # requires Gitea >= 28.0.0
}
