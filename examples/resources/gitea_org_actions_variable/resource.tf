resource "gitea_org_actions_variable" "example" {
  org           = "my-org"
  variable_name = "MY_ORG_VAR"
  value         = "my_value"
}
