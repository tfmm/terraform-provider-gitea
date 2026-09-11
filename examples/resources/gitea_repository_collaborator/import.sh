# Using `import` blocks in Terraform v1.5.0 and later:
# import {
#   to = gitea_repository_collaborator.example
#   id = "<owner>/<repo>/<username>"
# }

# Using `terraform import` in Terraform v1.4.0 and earlier:
terraform import gitea_repository_collaborator.example <owner>/<repo>/<username>
