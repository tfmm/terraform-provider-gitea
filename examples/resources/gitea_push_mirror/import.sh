# Using `import` blocks in Terraform v1.5.0 and later:
# import {
#   to = gitea_push_mirror.example
#   id = "<owner>/<repo>/<id>"
# }

# Using `terraform import` in Terraform v1.4.0 and earlier:
terraform import gitea_push_mirror.example <owner>/<repo>/<id>
