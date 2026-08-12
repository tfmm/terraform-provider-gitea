resource "gitea_user" "example" {
  username   = "mirror_example_user"
  login_name = "mirror_example_user"
  password   = "Geheim1!"
  email      = "mirror_example_user@user.dev"
}

resource "gitea_repository" "example" {
  username = gitea_user.example.username
  name     = "push-mirror-example-repo"
  private  = true
}

resource "gitea_push_mirror" "example" {
  owner           = gitea_user.example.username
  repo            = gitea_repository.example.name
  remote_address  = "https://github.com/example/target-repo.git"
  remote_username = "gituser"
  remote_password = "supersecretpassword"
  interval        = "8h0m0s"
  sync_on_commit  = true
}
