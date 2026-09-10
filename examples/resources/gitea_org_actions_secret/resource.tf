resource "gitea_org_actions_secret" "example" {
  org         = "my-org"
  secret_name = "MY_ORG_SECRET"
  value       = "secret_value"
}
