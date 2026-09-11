resource "gitea_org_webhook" "example" {
  org          = "my-org"
  type         = "gitea"
  url          = "https://example.com/webhook"
  content_type = "json"
  events       = ["push", "repository"]
  active       = true
}
