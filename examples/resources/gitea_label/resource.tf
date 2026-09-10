resource "gitea_label" "repo_label" {
  user        = "my-org"
  repo        = "my-repo"
  name        = "bug"
  color       = "#ff0000"
  description = "Something isn't working"
}

resource "gitea_label" "org_label" {
  org         = "my-org"
  name        = "feature"
  color       = "#00ff00"
  description = "New feature request"
}
