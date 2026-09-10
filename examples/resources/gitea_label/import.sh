# Using `import` blocks in Terraform v1.5.0 and later:
# import {
#   to = gitea_label.repo_label
#   id = "<user>/<repo>/<id>"
# }

# Using `terraform import` in Terraform v1.4.0 and earlier:
terraform import gitea_label.repo_label <user>/<repo>/<id>
