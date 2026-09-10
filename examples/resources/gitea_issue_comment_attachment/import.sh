# Using `import` blocks in Terraform v1.5.0 and later:
# import {
#   to = gitea_issue_comment_attachment.example
#   id = "<owner>/<repo>/<comment_id>/<attachment_id>"
# }

# Using `terraform import` in Terraform v1.4.0 and earlier:
terraform import gitea_issue_comment_attachment.example <owner>/<repo>/<comment_id>/<attachment_id>
