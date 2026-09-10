# Using `import` blocks in Terraform v1.5.0 and later:
# import {
#   to = gitea_repository_branch_protection.example
#   id = "<username>/<repo>/<branch_name>"
# }

# Using `terraform import` in Terraform v1.4.0 and earlier:
terraform import gitea_repository_branch_protection.example <username>/<repo>/<branch_name>
