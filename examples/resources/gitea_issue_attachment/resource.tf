resource "gitea_issue_attachment" "example" {
  owner       = "my-org"
  repo        = "my-repo"
  issue_index = 1
  file_path   = "path/to/file.png"
}
