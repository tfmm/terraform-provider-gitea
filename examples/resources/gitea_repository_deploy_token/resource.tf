resource "gitea_repository_deploy_token" "example" {
  username  = "my-org"
  name      = "my-repo"
  title     = "ci-deploy-token"
  read_only = true
}

# The plaintext token is only available right after creation, since Gitea
# never returns it again afterwards.
output "example_deploy_token" {
  value     = gitea_repository_deploy_token.example.token
  sensitive = true
}
