data "gitea_issue_comment_attachments" "example" {
  owner      = "my-org"
  repo       = "my-repo"
  comment_id = 123
}
