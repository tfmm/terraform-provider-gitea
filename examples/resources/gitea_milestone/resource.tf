resource "gitea_milestone" "example" {
  user        = "my-org"
  repo        = "my-repo"
  title       = "v1.0 Release"
  description = "First major release milestone"
  due_on      = "2026-12-31T23:59:59Z"
  state       = "open"
}
