resource "gitea_repository_actions_secret" "example" {
  owner       = "my-org"
  repository  = "my-repo"
  secret_name = "MY_SECRET"
  value       = "secret_value"
}
