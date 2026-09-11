# Using `import` blocks in Terraform v1.5.0 and later:
# import {
#   to = gitea_oauth2_app.example
#   id = "<user_or_org>/<app_id>"
# }

# Using `terraform import` in Terraform v1.4.0 and earlier:
terraform import gitea_oauth2_app.example <user_or_org>/<app_id>
