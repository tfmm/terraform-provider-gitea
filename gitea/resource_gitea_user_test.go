package gitea

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestGiteaUserTypeRoundTrip(t *testing.T) {
	cases := []struct {
		terraform string
		gitea     string
	}{
		{"user", "User"},
		{"bot", "Bot"},
		{"Bot", "Bot"},
		{"", "User"},
	}
	for _, c := range cases {
		if got := giteaUserType(c.terraform); got != c.gitea {
			t.Errorf("giteaUserType(%q) = %q, want %q", c.terraform, got, c.gitea)
		}
	}

	if got := terraformUserType("Bot"); got != "bot" {
		t.Errorf("terraformUserType(Bot) = %q, want bot", got)
	}
	if got := terraformUserType("User"); got != "user" {
		t.Errorf("terraformUserType(User) = %q, want user", got)
	}
	if got := terraformUserType("Organization"); got != "user" {
		t.Errorf("terraformUserType(Organization) = %q, want user (unknown types default to user)", got)
	}
}

func TestLookupUserPrefersUsernameOverID(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/users/alice":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"id": 9, "login": "alice", "type": "User"})
		case "/api/v1/users/search":
			// The shape code.gitea.io/sdk/gitea's GetUserByID expects from
			// its search-based fallback.
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":   true,
				"data": []map[string]interface{}{{"id": 9, "login": "alice", "type": "User"}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	gc := newTestGiteaClient(t, server)

	// Username known (the normal post-create/refresh case): must hit
	// /users/{username}, not the ID-based search endpoint, since that
	// endpoint doesn't see bot accounts.
	d := schema.TestResourceDataRaw(t, resourceGiteaUser().Schema, map[string]interface{}{"username": "alice"})
	d.SetId("9")
	if _, _, err := lookupUser(gc.Client, d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v1/users/alice" {
		t.Errorf("expected lookup by username, got path %q", gotPath)
	}

	// Username unknown (cold `terraform import <id>`): falls back to the
	// ID-based (search) lookup.
	dImport := schema.TestResourceDataRaw(t, resourceGiteaUser().Schema, map[string]interface{}{})
	dImport.SetId("9")
	if _, _, err := lookupUser(gc.Client, dImport); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v1/users/search" {
		t.Errorf("expected the import fallback to use the search endpoint, got %q", gotPath)
	}
}

// userServerState is a tiny in-memory fake of the Gitea 28.0.0 admin-user
// edit endpoint, modeling the two server behaviors this resource works
// around: GET /users/{username} always reflects the current type, and a
// PATCH carrying a password or login_name is rejected outright while the
// account is currently a bot.
type userServerState struct {
	username string
	t        *testing.T
	typ      string // "User" or "Bot"
	patches  int
}

func (s *userServerState) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/users/"+s.username:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id": 9, "login": s.username, "type": s.typ,
			})
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/admin/users/"+s.username:
			s.patches++
			var body editUserOption
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				s.t.Fatalf("failed to decode PATCH body: %v", err)
			}
			if s.typ == "Bot" && (body.Password != nil || body.LoginName != nil) {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"message": "a bot account cannot have a password or authentication source",
				})
				return
			}
			if body.Type != "" {
				s.typ = body.Type
			}
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}
}

func TestResourceUserUpdateBotToUserFlipsTypeBeforeFullEdit(t *testing.T) {
	state := &userServerState{username: "ci-bot", t: t, typ: "Bot"}
	server := httptest.NewServer(state.handler())
	defer server.Close()
	gc := newTestGiteaClient(t, server)

	d := schema.TestResourceDataRaw(t, resourceGiteaUser().Schema, map[string]interface{}{
		"username":   "ci-bot",
		"login_name": "ci-bot",
		"email":      "ci-bot@example.com",
		"password":   "irrelevant-not-changing",
		"user_type":  "user",
	})
	d.SetId("9")

	if diags := resourceUserUpdate(context.Background(), d, gc); diags.HasError() {
		t.Fatalf("unexpected update error converting bot->user: %v", diags)
	}

	if state.patches != 2 {
		t.Fatalf("expected a type-only PATCH followed by the full edit (2 requests), got %d", state.patches)
	}
	if state.typ != "User" {
		t.Errorf("expected the account to end up as User, server state is %q", state.typ)
	}
	if d.Get("user_type").(string) != "user" {
		t.Errorf("expected state user_type=user, got %q", d.Get("user_type"))
	}
}

func TestResourceUserUpdateUserToBotIsOneRequest(t *testing.T) {
	state := &userServerState{username: "ci-bot", t: t, typ: "User"}
	server := httptest.NewServer(state.handler())
	defer server.Close()
	gc := newTestGiteaClient(t, server)

	d := schema.TestResourceDataRaw(t, resourceGiteaUser().Schema, map[string]interface{}{
		"username":   "ci-bot",
		"login_name": "ci-bot",
		"email":      "ci-bot@example.com",
		"password":   "irrelevant-not-changing",
		"user_type":  "bot",
	})
	d.SetId("9")

	if diags := resourceUserUpdate(context.Background(), d, gc); diags.HasError() {
		t.Fatalf("unexpected update error converting user->bot: %v", diags)
	}

	if state.patches != 1 {
		t.Fatalf("expected a single merged PATCH, got %d", state.patches)
	}
	if state.typ != "Bot" {
		t.Errorf("expected the account to end up as Bot, server state is %q", state.typ)
	}
	if d.Get("user_type").(string) != "bot" {
		t.Errorf("expected state user_type=bot, got %q", d.Get("user_type"))
	}
}

func TestResourceUserUpdateNoOpOnAlreadyBotDoesNotSendLoginNameOrPassword(t *testing.T) {
	state := &userServerState{username: "ci-bot", t: t, typ: "Bot"}
	server := httptest.NewServer(state.handler())
	defer server.Close()
	gc := newTestGiteaClient(t, server)

	// password deliberately omitted so schema.TestResourceDataRaw reports
	// HasChange("password") == false (mirroring a real refresh/update where
	// the config's password didn't change), and force_password_change is
	// false, so the resource must not try to resend it while the account is
	// still a bot (that would 400 per the server fake above).
	d := schema.TestResourceDataRaw(t, resourceGiteaUser().Schema, map[string]interface{}{
		"username":              "ci-bot",
		"login_name":            "ci-bot",
		"email":                 "ci-bot@example.com",
		"force_password_change": false,
		"user_type":             "bot",
	})
	d.SetId("9")

	if diags := resourceUserUpdate(context.Background(), d, gc); diags.HasError() {
		t.Fatalf("unexpected update error on a no-op bot update: %v", diags)
	}
	if state.patches != 1 {
		t.Fatalf("expected a single PATCH, got %d", state.patches)
	}
	if state.typ != "Bot" {
		t.Errorf("expected the account to remain Bot, server state is %q", state.typ)
	}
}
