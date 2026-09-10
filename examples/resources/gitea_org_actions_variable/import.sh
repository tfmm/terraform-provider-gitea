# Using `import` blocks in Terraform v1.5.0 and later:
# import {
#   to = gitea_org_actions_variable.example
#   id = "<org>/<variable_name>"
# }

# Using `terraform import` in Terraform v1.4.0 and earlier:
terraform import gitea_org_actions_variable.example <org>/<variable_name>
