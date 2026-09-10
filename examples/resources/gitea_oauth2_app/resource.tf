resource "gitea_oauth2_app" "example" {
  name          = "my-oauth-app"
  redirect_uris = ["https://example.com/oauth/callback"]
}
