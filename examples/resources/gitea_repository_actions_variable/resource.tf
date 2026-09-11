resource "gitea_repository_actions_variable" "example" {
  owner         = "my-org"
  repository    = "my-repo"
  variable_name = "MY_VAR"
  value         = "my_value"
}
