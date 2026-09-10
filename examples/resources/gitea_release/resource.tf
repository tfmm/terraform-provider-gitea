resource "gitea_release" "example" {
  user             = "my-org"
  repo             = "my-repo"
  tag_name         = "v1.0.0"
  target_commitish = "main"
  title            = "v1.0.0 Initial Release"
  note             = "Release notes for version 1.0.0"
  draft            = false
  prerelease       = false
}
