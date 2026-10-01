resource "gitea_user" "test" {
  username             = "user-test"
  login_name           = "user-test"
  password             = "Geheim1!"
  email                = "user-test@user.dev"
  must_change_password = false
}

# A bot account for automation/service use (e.g. a CI token holder),
# requires Gitea >= 28.0.0. prohibit_login is commonly set alongside
# user_type = "bot" to prevent interactive login.
resource "gitea_user" "ci_bot" {
  username             = "ci-bot"
  login_name           = "ci-bot"
  password             = "Geheim1!"
  email                = "ci-bot@user.dev"
  must_change_password = false
  user_type            = "bot"
  prohibit_login       = true
}