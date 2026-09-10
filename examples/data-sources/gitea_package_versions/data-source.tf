data "gitea_package_versions" "example" {
  owner        = "my-org"
  package_type = "npm"
  package_name = "my-package"
}
