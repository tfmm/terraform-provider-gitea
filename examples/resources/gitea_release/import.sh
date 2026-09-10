# Using `import` blocks in Terraform v1.5.0 and later:
# import {
#   to = gitea_release.example
#   id = "<user>/<repo>/<release_id>"
# }

# Using `terraform import` in Terraform v1.4.0 and earlier:
terraform import gitea_release.example <user>/<repo>/<release_id>
