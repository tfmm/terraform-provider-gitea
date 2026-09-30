variable "repo_name" {
  type    = string
  default = "test-repo"
}

variable "org_name" {
  type    = string
  default = "test-org"
}

variable "repo_user_name" {
  type    = string
  default = "test_user"
}

variable "user_name" {
  type    = string
  default = "test_user"
}

variable "rule_name" {
  type    = string
  default = "branch-protection"
}

variable "protected_file_patterns" {
  type    = string
  default = ""
}

variable "unprotected_file_patterns" {
  type    = string
  default = ""
}

variable "enable_push" {
  type    = bool
  default = true
}

variable "push_whitelist_users" {
  type    = list(string)
  default = []
}

variable "push_whitelist_teams" {
  type    = list(string)
  default = []
}

variable "push_whitelist_deploy_keys" {
  type    = bool
  default = false
}

variable "require_signed_commits" {
  type    = bool
  default = false
}

variable "required_approvals" {
  type    = number
  default = 0
}

variable "approval_whitelist_users" {
  type    = list(string)
  default = []
}

variable "approval_whitelist_teams" {
  type    = list(string)
  default = []
}

variable "dismiss_stale_approvals" {
  type    = bool
  default = false
}

variable "status_check_patterns" {
  type    = list(string)
  default = []
}

variable "merge_whitelist_users" {
  type    = list(string)
  default = []
}

variable "merge_whitelist_teams" {
  type    = list(string)
  default = []
}

variable "block_merge_on_rejected_reviews" {
  type    = bool
  default = false
}

variable "block_merge_on_official_review_requests" {
  type    = bool
  default = false
}

variable "block_merge_on_outdated_branch" {
  type    = bool
  default = false
}

variable "block_on_codeowner_reviews" {
  type    = bool
  default = false
}

variable "enable_force_push" {
  type    = bool
  default = false
}

variable "ignore_stale_approvals" {
  type    = bool
  default = false
}

variable "priority" {
  type    = number
  default = 0
}

variable "force_push_allowlist_users" {
  type    = list(string)
  default = []
}

variable "force_push_allowlist_teams" {
  type    = list(string)
  default = []
}

variable "bypass_allowlist_users" {
  type    = list(string)
  default = []
}

variable "bypass_allowlist_teams" {
  type    = list(string)
  default = []
}

