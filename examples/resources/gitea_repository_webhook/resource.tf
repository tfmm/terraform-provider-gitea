resource "gitea_repository_webhook" "example" {
  username     = "my-org"
  name         = "my-repo"
  type         = "gitea"
  url          = "https://example.com/webhook"
  content_type = "json"
  events       = ["push"]
  active       = true
}
