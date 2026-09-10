# Using `import` blocks in Terraform v1.5.0 and later:
# import {
#   to = gitea_repository_actions_workflow_state.example
#   id = "<owner>/<repo>/<workflow_id_or_file>"
# }

# Using `terraform import` in Terraform v1.4.0 and earlier:
terraform import gitea_repository_actions_workflow_state.example <owner>/<repo>/<workflow_id_or_file>
