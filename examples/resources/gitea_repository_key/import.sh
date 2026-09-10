# Using `import` blocks in Terraform v1.5.0 and later:
# import {
#   to = gitea_repository_key.example
#   id = "<username>/<repo>/<key_id>"
# }

# Using `terraform import` in Terraform v1.4.0 and earlier:
terraform import gitea_repository_key.example <username>/<repo>/<key_id>
