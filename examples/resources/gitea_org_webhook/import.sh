# Using `import` blocks in Terraform v1.5.0 and later:
# import {
#   to = gitea_org_webhook.example
#   id = "<org>/<webhook_id>"
# }

# Using `terraform import` in Terraform v1.4.0 and earlier:
terraform import gitea_org_webhook.example <org>/<webhook_id>
