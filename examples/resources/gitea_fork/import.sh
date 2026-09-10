# Using `import` blocks in Terraform v1.5.0 and later:
# import {
#   to = gitea_fork.example
#   id = "<owner>/<repo>"
# }

# Using `terraform import` in Terraform v1.4.0 and earlier:
terraform import gitea_fork.example <owner>/<repo>
