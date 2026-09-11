# Using `import` blocks in Terraform v1.5.0 and later:
# import {
#   to = gitea_user_actions_secret.example
#   id = "<secret_name>"
# }

# Using `terraform import` in Terraform v1.4.0 and earlier:
terraform import gitea_user_actions_secret.example <secret_name>
