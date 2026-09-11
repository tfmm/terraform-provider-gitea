data "gitea_issue_attachments" "example" {
  owner       = "my-org"
  repo        = "my-repo"
  issue_index = 1
}
