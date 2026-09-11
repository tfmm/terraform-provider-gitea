# Using `import` blocks in Terraform v1.5.0 and later:
# import {
#   to = gitea_repository_actions_variable.example
#   id = "<owner>/<repo>/<variable_name>"
# }

# Using `terraform import` in Terraform v1.4.0 and earlier:
terraform import gitea_repository_actions_variable.example <owner>/<repo>/<variable_name>
