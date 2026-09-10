# Using `import` blocks in Terraform v1.5.0 and later:
# import {
#   to = gitea_user.example
#   id = "<user_id_or_username>"
# }

# Using `terraform import` in Terraform v1.4.0 and earlier:
terraform import gitea_user.example <user_id_or_username>
