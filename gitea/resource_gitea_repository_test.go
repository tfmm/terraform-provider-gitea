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

// TestResourceRepoCreateSyncsIgnoreWhitespaceConflicts guards against a
// drift bug: gitea.CreateRepoOption has no ignore_whitespace_conflicts
// field, so a freshly created repo always starts at Gitea's own
// server-side default (false), regardless of what's configured. Without a
// follow-up EditRepo call, a config that relies on this resource's
// `true` schema default would read back false right after creation and
// show a perpetual, unapplied diff on every subsequent plan.
func TestResourceRepoCreateSyncsIgnoreWhitespaceConflicts(t *testing.T) {
	var editBody gitea.EditRepoOption
	var editCalled bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/orgs/testowner":
			http.NotFound(w, r)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/admin/users":
			_ = json.NewEncoder(w).Encode([]*gitea.User{{ID: 1, UserName: "testowner"}})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/admin/users/testowner/repos":
			// Gitea's real create-repo response: ignore_whitespace_conflicts
			// always comes back false here, since CreateRepoOption has no
			// such field to request otherwise.
			_ = json.NewEncoder(w).Encode(gitea.Repository{
				ID:                        1,
				Name:                      "testrepo",
				Owner:                     &gitea.User{UserName: "testowner"},
				IgnoreWhitespaceConflicts: false,
			})
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/repos/testowner/testrepo":
			editCalled = true
			_ = json.NewDecoder(r.Body).Decode(&editBody)
			_ = json.NewEncoder(w).Encode(gitea.Repository{
				ID:                        1,
				Name:                      "testrepo",
				Owner:                     &gitea.User{UserName: "testowner"},
				IgnoreWhitespaceConflicts: editBody.IgnoreWhitespaceConflicts != nil && *editBody.IgnoreWhitespaceConflicts,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	gc := newTestGiteaClient(t, server)

	d := schema.TestResourceDataRaw(t, resourceGiteaRepository().Schema, map[string]interface{}{
		"username": "testowner",
		"name":     "testrepo",
		// ignore_whitespace_conflicts deliberately left unset, relying on
		// the schema's `true` default - this is the scenario that used to
		// never converge.
	})

	if diags := resourceRepoCreate(context.Background(), d, gc); diags.HasError() {
		t.Fatalf("unexpected create error: %v", diags)
	}

	if !editCalled {
		t.Fatal("expected a follow-up EditRepo call to sync ignore_whitespace_conflicts")
	}
	if editBody.IgnoreWhitespaceConflicts == nil || !*editBody.IgnoreWhitespaceConflicts {
		t.Errorf("expected the follow-up edit to request ignore_whitespace_conflicts=true, got %+v", editBody.IgnoreWhitespaceConflicts)
	}
	if !d.Get(repoIgnoreWhitespace).(bool) {
		t.Errorf("expected final state ignore_whitespace_conflicts=true, got %v", d.Get(repoIgnoreWhitespace))
	}
}

// TestResourceRepoCreateSkipsEditWhenAlreadyMatching asserts the follow-up
// call is skipped (not just harmless) when the created repo's value
// already matches what's configured, to avoid an unconditional extra
// request on every single repo creation.
func TestResourceRepoCreateSkipsEditWhenAlreadyMatching(t *testing.T) {
	editCalled := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/orgs/testowner":
			http.NotFound(w, r)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/admin/users":
			_ = json.NewEncoder(w).Encode([]*gitea.User{{ID: 1, UserName: "testowner"}})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/admin/users/testowner/repos":
			_ = json.NewEncoder(w).Encode(gitea.Repository{
				ID:                        1,
				Name:                      "testrepo",
				Owner:                     &gitea.User{UserName: "testowner"},
				IgnoreWhitespaceConflicts: false,
			})
		case r.Method == http.MethodPatch:
			editCalled = true
			http.Error(w, "unexpected edit call", http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	gc := newTestGiteaClient(t, server)

	d := schema.TestResourceDataRaw(t, resourceGiteaRepository().Schema, map[string]interface{}{
		"username":                    "testowner",
		"name":                        "testrepo",
		"ignore_whitespace_conflicts": false,
	})

	if diags := resourceRepoCreate(context.Background(), d, gc); diags.HasError() {
		t.Fatalf("unexpected create error: %v", diags)
	}
	if editCalled {
		t.Error("expected no follow-up edit when the created value already matches config")
	}
}
