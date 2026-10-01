data "gitea_releases" "example" {
  user = "my-org"
  repo = "my-repo"
}

# tag_filter is matched server-side and supports "*" as a wildcard
# (requires Gitea >= 28.0.0).
data "gitea_releases" "stable_only" {
  user       = "my-org"
  repo       = "my-repo"
  tag_filter = "v*"
}
