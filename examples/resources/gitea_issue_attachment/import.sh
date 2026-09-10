# Using `import` blocks in Terraform v1.5.0 and later:
# import {
#   to = gitea_issue_attachment.example
#   id = "<owner>/<repo>/<issue_index>/<attachment_id>"
# }

# Using `terraform import` in Terraform v1.4.0 and earlier:
terraform import gitea_issue_attachment.example <owner>/<repo>/<issue_index>/<attachment_id>
