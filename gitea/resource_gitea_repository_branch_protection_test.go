package gitea

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestRepositoryBranchProtectionImporterPreservesSlashInRuleName(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceGiteaRepositoryBranchProtection().Schema, map[string]interface{}{})
	d.SetId("alice/project/release/1.0/feature")

	res, err := resourceRepositoryBranchProtectionImport(context.Background(), d, nil)
	if err != nil {
		t.Fatalf("unexpected import error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected one resource data result, got %d", len(res))
	}

	got := res[0]
	if got.Get("username").(string) != "alice" {
		t.Fatalf("expected username alice, got %v", got.Get("username"))
	}
	if got.Get("name").(string) != "project" {
		t.Fatalf("expected repo project, got %v", got.Get("name"))
	}
	if got.Get("rule_name").(string) != "release/1.0/feature" {
		t.Fatalf("expected slash-preserving rule name, got %v", got.Get("rule_name"))
	}
	if got.Id() != "release/1.0/feature" {
		t.Fatalf("expected normalized id to be the rule name, got %q", got.Id())
	}
}

func TestGenerateWhitelistForcePushAndBypassAllowlists(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceGiteaRepositoryBranchProtection().Schema, map[string]interface{}{
		"force_push_allowlist_users":       []interface{}{"alice"},
		"force_push_allowlist_deploy_keys": false,
		"bypass_allowlist_teams":           []interface{}{"owners"},
	})

	if enabled, users, _ := generateWhitelist(d, "force_push_allowlist"); !enabled || len(users) != 1 || users[0] != "alice" {
		t.Fatalf("expected force_push_allowlist enabled with [alice], got enabled=%v users=%v", enabled, users)
	}
	if enabled, _, teams := generateWhitelist(d, "bypass_allowlist"); !enabled || len(teams) != 1 || teams[0] != "owners" {
		t.Fatalf("expected bypass_allowlist enabled with [owners], got enabled=%v teams=%v", enabled, teams)
	}

	d2 := schema.TestResourceDataRaw(t, resourceGiteaRepositoryBranchProtection().Schema, map[string]interface{}{
		"force_push_allowlist_deploy_keys": true,
	})
	if enabled, _, _ := generateWhitelist(d2, "force_push_allowlist"); !enabled {
		t.Fatal("expected force_push_allowlist enabled by force_push_allowlist_deploy_keys alone")
	}
}

func TestBranchProtectionCreateSendsNewFieldsAndFollowsUpWithEdit(t *testing.T) {
	var postBody, patchBody branchProtection
	var postCount, patchCount int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/repos/owner/repo/branch_protections":
			postCount++
			_ = json.NewDecoder(r.Body).Decode(&postBody)
			// Mirror Gitea 28.0.0's real create behavior: force-push fields
			// are silently dropped on create (see editBranchProtectionRaw's
			// caller in resourceRepositoryBranchProtectionCreate).
			resp := postBody
			resp.EnableForcePush = false
			resp.EnableForcePushAllowlist = false
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/repos/owner/repo/branch_protections/main":
			patchCount++
			_ = json.NewDecoder(r.Body).Decode(&patchBody)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(patchBody)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := gitea.NewClient(server.URL, gitea.SetGiteaVersion("28.0.0"))
	if err != nil {
		t.Fatalf("unexpected client error: %v", err)
	}
	gc := &GiteaClient{Client: client, baseURL: server.URL, httpClient: server.Client()}

	d := schema.TestResourceDataRaw(t, resourceGiteaRepositoryBranchProtection().Schema, map[string]interface{}{
		"username":                   "owner",
		"name":                       "repo",
		"rule_name":                  "main",
		"priority":                   5,
		"block_on_codeowner_reviews": true,
		"ignore_stale_approvals":     true,
		"enable_force_push":          true,
		"force_push_allowlist_users": []interface{}{"alice"},
	})

	if diags := resourceRepositoryBranchProtectionCreate(context.Background(), d, gc); diags.HasError() {
		t.Fatalf("unexpected create error: %v", diags)
	}

	if postCount != 1 || patchCount != 1 {
		t.Fatalf("expected exactly one POST and one follow-up PATCH, got POST=%d PATCH=%d", postCount, patchCount)
	}
	if !postBody.BlockOnCodeownerReviews {
		t.Error("expected block_on_codeowner_reviews to be sent on create")
	}
	if postBody.Priority == nil || *postBody.Priority != 5 {
		t.Errorf("expected priority 5 on create, got %v", postBody.Priority)
	}
	if !postBody.IgnoreStaleApprovals {
		t.Error("expected ignore_stale_approvals to be sent on create")
	}
	if !patchBody.EnableForcePush || !patchBody.EnableForcePushAllowlist {
		t.Errorf("expected the follow-up edit to carry enable_force_push/allowlist, got %+v", patchBody)
	}

	if !d.Get("enable_force_push").(bool) {
		t.Error("expected state to reflect enable_force_push=true after the follow-up edit, not the create response")
	}
	if !d.Get("block_on_codeowner_reviews").(bool) {
		t.Error("expected state to reflect block_on_codeowner_reviews=true")
	}
}

func TestBranchProtectionReadParsesGitea28Fields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/v1/repos/owner/repo/branch_protections/main" {
			w.Header().Set("Content-Type", "application/json")
			priority := int64(3)
			_ = json.NewEncoder(w).Encode(branchProtection{
				BranchName:               "main",
				RuleName:                 "main",
				Priority:                 &priority,
				BlockOnCodeownerReviews:  true,
				IgnoreStaleApprovals:     true,
				EnableBypassAllowlist:    true,
				BypassAllowlistUsernames: []string{"alice"},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	client, err := gitea.NewClient(server.URL, gitea.SetGiteaVersion("28.0.0"))
	if err != nil {
		t.Fatalf("unexpected client error: %v", err)
	}
	gc := &GiteaClient{Client: client, baseURL: server.URL, httpClient: server.Client()}

	d := schema.TestResourceDataRaw(t, resourceGiteaRepositoryBranchProtection().Schema, map[string]interface{}{
		"username":  "owner",
		"name":      "repo",
		"rule_name": "main",
	})

	if diags := resourceRepositoryBranchProtectionRead(context.Background(), d, gc); diags.HasError() {
		t.Fatalf("unexpected read error: %v", diags)
	}

	if !d.Get("block_on_codeowner_reviews").(bool) {
		t.Error("expected block_on_codeowner_reviews true")
	}
	if !d.Get("ignore_stale_approvals").(bool) {
		t.Error("expected ignore_stale_approvals true")
	}
	if d.Get("priority").(int) != 3 {
		t.Errorf("expected priority 3, got %v", d.Get("priority"))
	}
	if !d.Get("enable_bypass_allowlist").(bool) {
		t.Error("expected enable_bypass_allowlist true")
	}
	users := d.Get("bypass_allowlist_users").([]interface{})
	if len(users) != 1 || users[0].(string) != "alice" {
		t.Errorf("expected bypass_allowlist_users [alice], got %v", users)
	}
}

func TestBranchProtectionReadNotFoundClearsState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	client, err := gitea.NewClient(server.URL, gitea.SetGiteaVersion("28.0.0"))
	if err != nil {
		t.Fatalf("unexpected client error: %v", err)
	}
	gc := &GiteaClient{Client: client, baseURL: server.URL, httpClient: server.Client()}

	d := schema.TestResourceDataRaw(t, resourceGiteaRepositoryBranchProtection().Schema, map[string]interface{}{
		"username":  "owner",
		"name":      "repo",
		"rule_name": "main",
	})
	d.SetId("main")

	if diags := resourceRepositoryBranchProtectionRead(context.Background(), d, gc); diags.HasError() {
		t.Fatalf("unexpected read error: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("expected empty ID after 404, got %q", d.Id())
	}
}
