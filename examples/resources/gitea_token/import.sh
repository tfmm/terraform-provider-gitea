# Using `import` blocks in Terraform v1.5.0 and later:
# import {
#   to = gitea_token.example
#   id = "<username>/<token_name>/<token_id>"
# }

# Using `terraform import` in Terraform v1.4.0 and earlier:
terraform import gitea_token.example <username>/<token_name>/<token_id>
