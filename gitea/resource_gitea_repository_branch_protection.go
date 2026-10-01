package gitea

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	repoBPUsername string = "username"
	repoBPName     string = "name"
	repoBPRuleName string = "rule_name"

	repoBPProtectedFilePatterns   string = "protected_file_patterns"
	repoBPUnprotectedFilePatterns string = "unprotected_file_patterns"

	repoBPPriority string = "priority"

	repoBPEnablePush              string = "enable_push"
	repoBPEnablePushWhitelist     string = "enable_push_whitelist"
	repoBPPushWhitelistUsers      string = "push_whitelist_users"
	repoBPPushWhitelistTeams      string = "push_whitelist_teams"
	repoBPPushWhitelistDeployKeys string = "push_whitelist_deploy_keys"

	repoBPEnableForcePush              string = "enable_force_push"
	repoBPEnableForcePushAllowlist     string = "enable_force_push_allowlist"
	repoBPForcePushAllowlistUsers      string = "force_push_allowlist_users"
	repoBPForcePushAllowlistTeams      string = "force_push_allowlist_teams"
	repoBPForcePushAllowlistDeployKeys string = "force_push_allowlist_deploy_keys"

	repoBPEnableBypassAllowlist string = "enable_bypass_allowlist"
	repoBPBypassAllowlistUsers  string = "bypass_allowlist_users"
	repoBPBypassAllowlistTeams  string = "bypass_allowlist_teams"

	repoBPRequireSignedCommits string = "require_signed_commits"

	repoBPRequiredApprovals       string = "required_approvals"
	repoBPEnableApprovalWhitelist string = "enable_approval_whitelist"
	repoBPApprovalWhitelistUsers  string = "approval_whitelist_users"
	repoBPApprovalWhitelistTeams  string = "approval_whitelist_teams"
	repoBPDismissStaleApprovals   string = "dismiss_stale_approvals"
	repoBPIgnoreStaleApprovals    string = "ignore_stale_approvals"

	repoBPEnableStatusCheck   string = "enable_status_check"
	repoBPStatusCheckPatterns string = "status_check_patterns"

	repoBPEnableMergeWhitelist string = "enable_merge_whitelist"
	repoBPMergeWhitelistUsers  string = "merge_whitelist_users"
	repoBPMergeWhitelistTeams  string = "merge_whitelist_teams"

	repoBPBlockMergeOnRejectedReviews        string = "block_merge_on_rejected_reviews"
	repoBPBlockMergeOnOfficialReviewRequests string = "block_merge_on_official_review_requests"
	repoBPBlockMergeOnOutdatedBranch         string = "block_merge_on_outdated_branch"
	repoBPBlockOnCodeownerReviews            string = "block_on_codeowner_reviews"
	repoBPBlockAdminMergeOverride            string = "block_admin_merge_override"

	repoBPUpdatedAt string = "updated_at"
	repoBPCreatedAt string = "created_at"
)

// branchProtection mirrors the Gitea BranchProtection API object. It is kept
// local (instead of using code.gitea.io/sdk/gitea's BranchProtection) because
// the SDK has not yet caught up with fields introduced by Gitea, such as
// block_on_codeowner_reviews (Gitea 28) and the bypass/force-push allowlists.
type branchProtection struct {
	BranchName string `json:"branch_name"`
	RuleName   string `json:"rule_name"`
	// Priority is a pointer with omitempty deliberately, unlike every other
	// field here: Gitea auto-assigns it sequentially (1, 2, 3, ...) by
	// creation order when it's left out of a create request, but a PATCH
	// that explicitly includes it (even priority:0) overwrites that
	// assignment outright - verified against a live Gitea 28.0.0 server.
	// Since every other field in this struct is unconditionally sent as
	// part of a full-sync payload, a plain int64 here would silently reset
	// every rule's priority to 0 on every create and every edit unless the
	// user had explicitly configured one.
	Priority                      *int64    `json:"priority,omitempty"`
	EnablePush                    bool      `json:"enable_push"`
	EnablePushWhitelist           bool      `json:"enable_push_whitelist"`
	PushWhitelistUsernames        []string  `json:"push_whitelist_usernames"`
	PushWhitelistTeams            []string  `json:"push_whitelist_teams"`
	PushWhitelistDeployKeys       bool      `json:"push_whitelist_deploy_keys"`
	EnableForcePush               bool      `json:"enable_force_push"`
	EnableForcePushAllowlist      bool      `json:"enable_force_push_allowlist"`
	ForcePushAllowlistUsernames   []string  `json:"force_push_allowlist_usernames"`
	ForcePushAllowlistTeams       []string  `json:"force_push_allowlist_teams"`
	ForcePushAllowlistDeployKeys  bool      `json:"force_push_allowlist_deploy_keys"`
	EnableBypassAllowlist         bool      `json:"enable_bypass_allowlist"`
	BypassAllowlistUsernames      []string  `json:"bypass_allowlist_usernames"`
	BypassAllowlistTeams          []string  `json:"bypass_allowlist_teams"`
	EnableMergeWhitelist          bool      `json:"enable_merge_whitelist"`
	MergeWhitelistUsernames       []string  `json:"merge_whitelist_usernames"`
	MergeWhitelistTeams           []string  `json:"merge_whitelist_teams"`
	EnableStatusCheck             bool      `json:"enable_status_check"`
	StatusCheckContexts           []string  `json:"status_check_contexts"`
	RequiredApprovals             int64     `json:"required_approvals"`
	EnableApprovalsWhitelist      bool      `json:"enable_approvals_whitelist"`
	ApprovalsWhitelistUsernames   []string  `json:"approvals_whitelist_username"`
	ApprovalsWhitelistTeams       []string  `json:"approvals_whitelist_teams"`
	BlockOnRejectedReviews        bool      `json:"block_on_rejected_reviews"`
	BlockOnOfficialReviewRequests bool      `json:"block_on_official_review_requests"`
	BlockOnOutdatedBranch         bool      `json:"block_on_outdated_branch"`
	BlockOnCodeownerReviews       bool      `json:"block_on_codeowner_reviews"`
	DismissStaleApprovals         bool      `json:"dismiss_stale_approvals"`
	IgnoreStaleApprovals          bool      `json:"ignore_stale_approvals"`
	RequireSignedCommits          bool      `json:"require_signed_commits"`
	ProtectedFilePatterns         string    `json:"protected_file_patterns"`
	UnprotectedFilePatterns       string    `json:"unprotected_file_patterns"`
	BlockAdminMergeOverride       bool      `json:"block_admin_merge_override"`
	Created                       time.Time `json:"created_at"`
	Updated                       time.Time `json:"updated_at"`
}

func branchProtectionPath(owner, repo, name string) string {
	if name == "" {
		return fmt.Sprintf("/repos/%s/%s/branch_protections", owner, repo)
	}
	return fmt.Sprintf("/repos/%s/%s/branch_protections/%s", owner, repo, name)
}

func getBranchProtectionRaw(client *GiteaClient, owner, repo, name string) (*branchProtection, error) {
	bp := new(branchProtection)
	if err := client.rawJSON("GET", branchProtectionPath(owner, repo, name), nil, bp); err != nil {
		return nil, err
	}
	return bp, nil
}

func createBranchProtectionRaw(client *GiteaClient, owner, repo string, opt *branchProtection) (*branchProtection, error) {
	bp := new(branchProtection)
	if err := client.rawJSON("POST", branchProtectionPath(owner, repo, ""), opt, bp); err != nil {
		return nil, err
	}
	return bp, nil
}

func editBranchProtectionRaw(client *GiteaClient, owner, repo, name string, opt *branchProtection) (*branchProtection, error) {
	bp := new(branchProtection)
	if err := client.rawJSON("PATCH", branchProtectionPath(owner, repo, name), opt, bp); err != nil {
		return nil, err
	}
	return bp, nil
}

func deleteBranchProtectionRaw(client *GiteaClient, owner, repo, name string) error {
	return client.rawJSON("DELETE", branchProtectionPath(owner, repo, name), nil, nil)
}

func resourceRepositoryBranchProtectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*GiteaClient)

	user := d.Get(repoBPUsername).(string)
	repo := d.Get(repoBPName).(string)
	ruleName := d.Get(repoBPRuleName).(string)

	bp, err := getBranchProtectionRaw(client, user, repo, ruleName)
	if err != nil {
		if strings.Contains(err.Error(), "status 404") {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return diag.FromErr(setRepositoryBranchProtectionData(bp, user, repo, d))
}

func generateWhitelist(d *schema.ResourceData, listname string) (enabled bool, users []string, teams []string) {
	u := d.Get(listname + "_users")
	users = make([]string, 0)
	if u != nil {
		for _, element := range u.([]interface{}) {
			users = append(users, element.(string))
		}
	}

	t := d.Get(listname + "_teams")
	teams = make([]string, 0)
	if u != nil {
		for _, element := range t.([]interface{}) {
			teams = append(teams, element.(string))
		}
	}

	if c := len(users) + len(teams); c > 0 {
		enabled = true
	}
	if (listname == "push_whitelist" || listname == "force_push_allowlist") && d.Get(listname+"_deploy_keys").(bool) {
		enabled = true
	}

	log.Println("enabled?:", enabled, listname)
	return enabled, users, teams
}

func branchProtectionFromResourceData(d *schema.ResourceData, ruleName string) *branchProtection {
	enablePushWhitelist, pushWhitelistUsernames, pushWhitelistTeams := generateWhitelist(d, "push_whitelist")
	enableForcePushAllowlist, forcePushAllowlistUsernames, forcePushAllowlistTeams := generateWhitelist(d, "force_push_allowlist")
	enableBypassAllowlist, bypassAllowlistUsernames, bypassAllowlistTeams := generateWhitelist(d, "bypass_allowlist")
	enableMergeWhitelist, mergeWhitelistUsernames, mergeWhitelistTeams := generateWhitelist(d, "merge_whitelist")
	enableApprovalsWhitelist, approvalsWhitelistUsernames, approvalsWhitelistTeams := generateWhitelist(d, "approval_whitelist")

	statusCheckContexts := make([]string, 0)
	for _, element := range d.Get(repoBPStatusCheckPatterns).([]interface{}) {
		statusCheckContexts = append(statusCheckContexts, element.(string))
	}
	enableStatusCheck := len(statusCheckContexts) > 0

	enablePush := d.Get(repoBPEnablePush).(bool) || enablePushWhitelist

	return &branchProtection{
		// BranchName is deprecated in gitea, but still required by the API, therefore using RuleName
		BranchName:                    ruleName,
		RuleName:                      ruleName,
		Priority:                      optionalInt64Value(d, repoBPPriority),
		EnablePush:                    enablePush,
		EnablePushWhitelist:           enablePushWhitelist,
		PushWhitelistUsernames:        pushWhitelistUsernames,
		PushWhitelistTeams:            pushWhitelistTeams,
		PushWhitelistDeployKeys:       d.Get(repoBPPushWhitelistDeployKeys).(bool),
		EnableForcePush:               d.Get(repoBPEnableForcePush).(bool),
		EnableForcePushAllowlist:      enableForcePushAllowlist,
		ForcePushAllowlistUsernames:   forcePushAllowlistUsernames,
		ForcePushAllowlistTeams:       forcePushAllowlistTeams,
		ForcePushAllowlistDeployKeys:  d.Get(repoBPForcePushAllowlistDeployKeys).(bool),
		EnableBypassAllowlist:         enableBypassAllowlist,
		BypassAllowlistUsernames:      bypassAllowlistUsernames,
		BypassAllowlistTeams:          bypassAllowlistTeams,
		EnableMergeWhitelist:          enableMergeWhitelist,
		MergeWhitelistUsernames:       mergeWhitelistUsernames,
		MergeWhitelistTeams:           mergeWhitelistTeams,
		EnableStatusCheck:             enableStatusCheck,
		StatusCheckContexts:           statusCheckContexts,
		RequiredApprovals:             int64(d.Get(repoBPRequiredApprovals).(int)),
		EnableApprovalsWhitelist:      enableApprovalsWhitelist,
		ApprovalsWhitelistUsernames:   approvalsWhitelistUsernames,
		ApprovalsWhitelistTeams:       approvalsWhitelistTeams,
		BlockOnRejectedReviews:        d.Get(repoBPBlockMergeOnRejectedReviews).(bool),
		BlockOnOfficialReviewRequests: d.Get(repoBPBlockMergeOnOfficialReviewRequests).(bool),
		BlockOnOutdatedBranch:         d.Get(repoBPBlockMergeOnOutdatedBranch).(bool),
		BlockOnCodeownerReviews:       d.Get(repoBPBlockOnCodeownerReviews).(bool),
		DismissStaleApprovals:         d.Get(repoBPDismissStaleApprovals).(bool),
		IgnoreStaleApprovals:          d.Get(repoBPIgnoreStaleApprovals).(bool),
		RequireSignedCommits:          d.Get(repoBPRequireSignedCommits).(bool),
		ProtectedFilePatterns:         d.Get(repoBPProtectedFilePatterns).(string),
		UnprotectedFilePatterns:       d.Get(repoBPUnprotectedFilePatterns).(string),
		BlockAdminMergeOverride:       d.Get(repoBPBlockAdminMergeOverride).(bool),
	}
}

func resourceRepositoryBranchProtectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*GiteaClient)

	user := d.Get(repoBPUsername).(string)
	repo := d.Get(repoBPName).(string)
	ruleName := d.Get(repoBPRuleName).(string)

	opt := branchProtectionFromResourceData(d, ruleName)

	bp, err := createBranchProtectionRaw(client, user, repo, opt)
	if err != nil {
		return diag.FromErr(err)
	}

	// Gitea 28.0.0's create endpoint silently ignores enable_force_push and
	// enable_force_push_allowlist (the edit endpoint honors both), so a
	// second call normalizes state with what the server actually applies.
	bp, err = editBranchProtectionRaw(client, user, repo, ruleName, opt)
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(setRepositoryBranchProtectionData(bp, user, repo, d))
}

func resourceRepositoryBranchProtectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*GiteaClient)

	user := d.Get(repoBPUsername).(string)
	repo := d.Get(repoBPName).(string)
	ruleName := d.Id()

	opt := branchProtectionFromResourceData(d, ruleName)

	bp, err := editBranchProtectionRaw(client, user, repo, ruleName, opt)
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(setRepositoryBranchProtectionData(bp, user, repo, d))
}

func resourceRepositoryBranchProtectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*GiteaClient)

	user := d.Get(repoBPUsername).(string)
	repo := d.Get(repoBPName).(string)
	ruleName := d.Id()

	return diag.FromErr(deleteBranchProtectionRaw(client, user, repo, ruleName))
}

func setRepositoryBranchProtectionData(bp *branchProtection, user string, repo string, d *schema.ResourceData) (err error) {
	d.SetId(bp.RuleName)
	if err := d.Set(repoBPUsername, user); err != nil {
		return err
	}
	if err := d.Set(repoBPName, repo); err != nil {
		return err
	}
	if err := d.Set(repoBPProtectedFilePatterns, bp.ProtectedFilePatterns); err != nil {
		return err
	}
	if err := d.Set(repoBPUnprotectedFilePatterns, bp.UnprotectedFilePatterns); err != nil {
		return err
	}
	var priority int64
	if bp.Priority != nil {
		priority = *bp.Priority
	}
	if err := d.Set(repoBPPriority, priority); err != nil {
		return err
	}
	if err := d.Set(repoBPEnablePush, bp.EnablePush); err != nil {
		return err
	}
	if err := d.Set(repoBPEnablePushWhitelist, bp.EnablePushWhitelist); err != nil {
		return err
	}
	if err := d.Set(repoBPPushWhitelistUsers, bp.PushWhitelistUsernames); err != nil {
		return err
	}
	if err := d.Set(repoBPPushWhitelistTeams, bp.PushWhitelistTeams); err != nil {
		return err
	}
	if err := d.Set(repoBPPushWhitelistDeployKeys, bp.PushWhitelistDeployKeys); err != nil {
		return err
	}
	if err := d.Set(repoBPEnableForcePush, bp.EnableForcePush); err != nil {
		return err
	}
	if err := d.Set(repoBPEnableForcePushAllowlist, bp.EnableForcePushAllowlist); err != nil {
		return err
	}
	if err := d.Set(repoBPForcePushAllowlistUsers, bp.ForcePushAllowlistUsernames); err != nil {
		return err
	}
	if err := d.Set(repoBPForcePushAllowlistTeams, bp.ForcePushAllowlistTeams); err != nil {
		return err
	}
	if err := d.Set(repoBPForcePushAllowlistDeployKeys, bp.ForcePushAllowlistDeployKeys); err != nil {
		return err
	}
	if err := d.Set(repoBPEnableBypassAllowlist, bp.EnableBypassAllowlist); err != nil {
		return err
	}
	if err := d.Set(repoBPBypassAllowlistUsers, bp.BypassAllowlistUsernames); err != nil {
		return err
	}
	if err := d.Set(repoBPBypassAllowlistTeams, bp.BypassAllowlistTeams); err != nil {
		return err
	}
	if err := d.Set(repoBPRequireSignedCommits, bp.RequireSignedCommits); err != nil {
		return err
	}
	if err := d.Set(repoBPRequiredApprovals, bp.RequiredApprovals); err != nil {
		return err
	}
	if err := d.Set(repoBPEnableApprovalWhitelist, bp.EnableApprovalsWhitelist); err != nil {
		return err
	}
	if err := d.Set(repoBPApprovalWhitelistUsers, bp.ApprovalsWhitelistUsernames); err != nil {
		return err
	}
	if err := d.Set(repoBPApprovalWhitelistTeams, bp.ApprovalsWhitelistTeams); err != nil {
		return err
	}
	if err := d.Set(repoBPDismissStaleApprovals, bp.DismissStaleApprovals); err != nil {
		return err
	}
	if err := d.Set(repoBPIgnoreStaleApprovals, bp.IgnoreStaleApprovals); err != nil {
		return err
	}
	if err := d.Set(repoBPEnableStatusCheck, bp.EnableStatusCheck); err != nil {
		return err
	}
	if err := d.Set(repoBPStatusCheckPatterns, bp.StatusCheckContexts); err != nil {
		return err
	}
	if err := d.Set(repoBPEnableMergeWhitelist, bp.EnableMergeWhitelist); err != nil {
		return err
	}
	if err := d.Set(repoBPMergeWhitelistUsers, bp.MergeWhitelistUsernames); err != nil {
		return err
	}
	if err := d.Set(repoBPMergeWhitelistTeams, bp.MergeWhitelistTeams); err != nil {
		return err
	}
	if err := d.Set(repoBPBlockMergeOnRejectedReviews, bp.BlockOnRejectedReviews); err != nil {
		return err
	}
	if err := d.Set(repoBPBlockMergeOnOfficialReviewRequests, bp.BlockOnOfficialReviewRequests); err != nil {
		return err
	}
	if err := d.Set(repoBPBlockMergeOnOutdatedBranch, bp.BlockOnOutdatedBranch); err != nil {
		return err
	}
	if err := d.Set(repoBPBlockOnCodeownerReviews, bp.BlockOnCodeownerReviews); err != nil {
		return err
	}
	if err := d.Set(repoBPBlockAdminMergeOverride, bp.BlockAdminMergeOverride); err != nil {
		return err
	}
	if err := d.Set(repoBPUpdatedAt, timeToString(bp.Updated)); err != nil {
		return err
	}
	if err := d.Set(repoBPCreatedAt, timeToString(bp.Created)); err != nil {
		return err
	}

	return nil
}

func resourceRepositoryBranchProtectionImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.SplitN(d.Id(), "/", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("unexpected ID format (%q), expected <username>/<repo>/<rule_name>", d.Id())
	}
	if err := d.Set("username", parts[0]); err != nil {
		return nil, err
	}
	if err := d.Set("name", parts[1]); err != nil {
		return nil, err
	}
	if err := d.Set("rule_name", parts[2]); err != nil {
		return nil, err
	}
	d.SetId(parts[2])
	return []*schema.ResourceData{d}, nil
}

func resourceGiteaRepositoryBranchProtection() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceRepositoryBranchProtectionRead,
		CreateContext: resourceRepositoryBranchProtectionCreate,
		UpdateContext: resourceRepositoryBranchProtectionUpdate,
		DeleteContext: resourceRepositoryBranchProtectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRepositoryBranchProtectionImport,
		},
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "User name or organization name",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Repository name",
			},
			"rule_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Protected Branch Name Pattern",
			},
			"priority": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: false,
				Description: "Priority of this branch protection rule when multiple rules match the same branch. Lower values are evaluated " +
					"first. Leave unset to let Gitea assign it automatically (sequentially, by creation order); setting it explicitly " +
					"takes over that assignment.",
			},
			"protected_file_patterns": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Default:     "",
				Description: "Protected file patterns (separated using semicolon ';')",
			},
			"unprotected_file_patterns": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Default:     "",
				Description: "Unprotected file patterns (separated using semicolon ';')",
			},
			"enable_push": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  false,
				Description: `Anyone with write access will be allowed to push to this branch
								(but not force push), add a whitelist users or teams to limit
								access.`,
			},
			"enable_push_whitelist": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if a push whitelist is used.",
			},
			"push_whitelist_users": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				RequiredWith: []string{"enable_push"},
				Optional:     true,
				ForceNew:     false,
				Description:  "Allowlisted users for pushing. Requires enable_push to be set to true.",
			},
			"push_whitelist_teams": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				RequiredWith: []string{"enable_push"},
				Optional:     true,
				ForceNew:     false,
				Description:  "Allowlisted teams for pushing. Requires enable_push to be set to true.",
			},
			"push_whitelist_deploy_keys": {
				Type:         schema.TypeBool,
				RequiredWith: []string{"enable_push"},
				Optional:     true,
				ForceNew:     false,
				Default:      false,
				Description:  "Allow deploy keys with write access to push. Requires enable_push to be set to true.",
			},
			"enable_force_push": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  false,
				Description: `Allow force pushes to this branch by anyone with push access. Mutually
								exclusive with the force push allowlist: if force_push_allowlist_users
								or force_push_allowlist_teams is set, Gitea restricts force pushing to
								that allowlist and this field reads back as false.`,
			},
			"enable_force_push_allowlist": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if a force push allowlist is used.",
			},
			"force_push_allowlist_users": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				ForceNew:    false,
				Description: "Allowlisted users who may force push to this branch.",
			},
			"force_push_allowlist_teams": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				ForceNew:    false,
				Description: "Allowlisted teams who may force push to this branch.",
			},
			"force_push_allowlist_deploy_keys": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Default:     false,
				Description: "Allow deploy keys with write access to force push.",
			},
			"enable_bypass_allowlist": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if a bypass allowlist is used.",
			},
			"bypass_allowlist_users": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				ForceNew:    false,
				Description: "Allowlisted users who may bypass this branch protection rule entirely.",
			},
			"bypass_allowlist_teams": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				ForceNew:    false,
				Description: "Allowlisted teams who may bypass this branch protection rule entirely.",
			},
			"require_signed_commits": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Default:     false,
				Description: "Reject pushes to this branch if they are unsigned or unverifiable.",
			},
			"required_approvals": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    false,
				Default:     0,
				Description: "Allow only to merge pull request with enough positive reviews.",
			},
			"enable_approval_whitelist": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if a approval whitelist is used.",
			},
			"approval_whitelist_users": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				ForceNew: false,
				Description: `Only reviews from allowlisted users will count to the required
								approvals. Without approval allowlist, reviews from anyone with
								write access count to the required approvals.`,
			},
			"approval_whitelist_teams": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				ForceNew: false,
				Description: `Only reviews from allowlisted teams will count to the required
								approvals. Without approval allowlist, reviews from anyone with
								write access count to the required approvals.`,
			},
			"dismiss_stale_approvals": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  false,
				Description: `When new commits that change the content of the pull request
								are pushed to the branch, old approvals will be dismissed.`,
			},
			"ignore_stale_approvals": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  false,
				Description: `Do not count approvals that were made on older commits (stale
								reviews) towards how many approvals the PR has. Irrelevant if
								stale reviews are already dismissed.`,
			},
			"enable_status_check": {
				Type:     schema.TypeBool,
				Computed: true,
				Description: `Require status checks to pass before merging. When enabled,
								commits must first be pushed to another branch, then merged
								or pushed directly to a branch that matches this rule after
								status checks have passed. If no contexts are matched, the
								last commit must be successful regardless of context`,
			},
			"status_check_patterns": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				ForceNew: false,
				Description: `Enter patterns to specify which status checks must pass before
								branches can be merged into a branch that matches this rule.
								Each line specifies a pattern. Patterns cannot be empty.`,
			},
			"enable_merge_whitelist": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if a merge whitelist is used.",
			},
			"merge_whitelist_users": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				ForceNew:    false,
				Description: "Allow only allowlisted users to merge pull requests into this branch.",
			},
			"merge_whitelist_teams": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				ForceNew:    false,
				Description: "Allow only allowlisted teams to merge pull requests into this branch.",
			},
			"block_merge_on_rejected_reviews": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  false,
				Description: `Merging will not be possible when changes are
								requested by official reviewers, even if there are enough
								approvals.`,
			},
			"block_merge_on_official_review_requests": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  false,
				Description: `Merging will not be possible when it has official
								review requests, even if there are enough approvals.`,
			},
			"block_merge_on_outdated_branch": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Default:     false,
				Description: "Merging will not be possible when head branch is behind base branch.",
			},
			"block_on_codeowner_reviews": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Default:     false,
				Description: "Merging will not be possible until all code owners (per CODEOWNERS) have approved. Requires Gitea >= 28.0.0.",
			},
			"block_admin_merge_override": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Default:     false,
				Description: "Prevent admins from bypassing branch protection rules when merging.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Webhook creation timestamp",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Webhook creation timestamp",
			},
		},
		Description: "This resource allows you to create and manage branch protections for repositories.",
	}
}
