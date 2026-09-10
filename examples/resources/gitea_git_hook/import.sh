# Using `import` blocks in Terraform v1.5.0 and later:
# import {
#   to = gitea_git_hook.example
#   id = "<owner>/<repo>/<hook_name>"
# }

# Using `terraform import` in Terraform v1.4.0 and earlier:
terraform import gitea_git_hook.example <owner>/<repo>/<hook_name>
