# Using `import` blocks in Terraform v1.5.0 and later:
# import {
#   to = gitea_milestone.example
#   id = "<user>/<repo>/<milestone_id>"
# }

# Using `terraform import` in Terraform v1.4.0 and earlier:
terraform import gitea_milestone.example <user>/<repo>/<milestone_id>
