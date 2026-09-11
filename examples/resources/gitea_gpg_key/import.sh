# Using `import` blocks in Terraform v1.5.0 and later:
# import {
#   to = gitea_gpg_key.example
#   id = "<key_id>"
# }

# Using `terraform import` in Terraform v1.4.0 and earlier:
terraform import gitea_gpg_key.example <key_id>
