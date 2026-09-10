resource "gitea_issue_comment_attachment" "example" {
  owner      = "my-org"
  repo       = "my-repo"
  comment_id = 123
  file_path  = "path/to/file.png"
}
