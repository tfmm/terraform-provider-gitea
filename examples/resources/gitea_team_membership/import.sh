# Using `import` blocks in Terraform v1.5.0 and later:
# import {
#   to = gitea_team_membership.example
#   id = "<team_id>/<username>"
# }

# Using `terraform import` in Terraform v1.4.0 and earlier:
terraform import gitea_team_membership.example <team_id>/<username>
