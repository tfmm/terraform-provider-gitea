resource "gitea_user_actions_secret" "example" {
  secret_name = "MY_USER_SECRET"
  value       = "secret_value"
}
