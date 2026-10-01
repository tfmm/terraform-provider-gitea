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

func newTestGiteaClient(t *testing.T, server *httptest.Server) *GiteaClient {
	t.Helper()
	client, err := gitea.NewClient(server.URL, gitea.SetGiteaVersion("28.0.0"))
	if err != nil {
		t.Fatalf("unexpected client error: %v", err)
	}
	return &GiteaClient{Client: client, baseURL: server.URL, httpClient: server.Client()}
}

func TestDeployTokenIDParts(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceGiteaRepositoryDeployToken().Schema, map[string]interface{}{})
	d.SetId("owner/repo/42")

	owner, repo, id, ok, err := deployTokenIDParts(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected ok=true for a well-formed ID")
	}
	if owner != "owner" || repo != "repo" || id != 42 {
		t.Fatalf("expected owner=owner repo=repo id=42, got owner=%s repo=%s id=%d", owner, repo, id)
	}

	d.SetId("")
	_, _, _, ok, err = deployTokenIDParts(d)
	if err != nil {
		t.Fatalf("unexpected error for empty ID: %v", err)
	}
	if ok {
		t.Fatal("expected ok=false for an empty ID")
	}
}

func TestResourceRepositoryDeployTokenCreate(t *testing.T) {
	var gotBody deployTokenCreateOption

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/v1/repos/owner/repo/keys/tokens" {
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(deployTokenResponse{
				ID:       7,
				KeyType:  "token",
				Title:    gotBody.Title,
				ReadOnly: gotBody.ReadOnly,
				Token:    "gdt_plaintext_example",
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()
	gc := newTestGiteaClient(t, server)

	d := schema.TestResourceDataRaw(t, resourceGiteaRepositoryDeployToken().Schema, map[string]interface{}{
		"username":  "owner",
		"name":      "repo",
		"title":     "ci-token",
		"read_only": false,
	})

	if diags := resourceRepositoryDeployTokenCreate(context.Background(), d, gc); diags.HasError() {
		t.Fatalf("unexpected create error: %v", diags)
	}

	if gotBody.Title != "ci-token" || gotBody.ReadOnly != false {
		t.Errorf("expected request body {ci-token false}, got %+v", gotBody)
	}
	if d.Id() != "owner/repo/7" {
		t.Errorf("expected ID 'owner/repo/7', got %q", d.Id())
	}
	if d.Get("token").(string) != "gdt_plaintext_example" {
		t.Errorf("expected token to be set from the create response, got %q", d.Get("token"))
	}
}

func TestResourceRepositoryDeployTokenRead(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/v1/repos/owner/repo/keys/7" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(deployTokenResponse{
				ID:       7,
				KeyType:  "token",
				Title:    "ci-token",
				ReadOnly: true,
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()
	gc := newTestGiteaClient(t, server)

	d := schema.TestResourceDataRaw(t, resourceGiteaRepositoryDeployToken().Schema, map[string]interface{}{
		"token": "gdt_should_survive_refresh",
	})
	d.SetId("owner/repo/7")

	if diags := resourceRepositoryDeployTokenRead(context.Background(), d, gc); diags.HasError() {
		t.Fatalf("unexpected read error: %v", diags)
	}

	if d.Get("username").(string) != "owner" || d.Get("name").(string) != "repo" {
		t.Errorf("expected username/name to be set from the ID, got username=%q name=%q", d.Get("username"), d.Get("name"))
	}
	if d.Get("title").(string) != "ci-token" {
		t.Errorf("expected title 'ci-token', got %q", d.Get("title"))
	}
	if !d.Get("read_only").(bool) {
		t.Error("expected read_only true")
	}
	// The plaintext token is only ever returned by the create call; Read
	// must not clobber it with an empty value on refresh.
	if d.Get("token").(string) != "gdt_should_survive_refresh" {
		t.Errorf("expected token to be left untouched by Read, got %q", d.Get("token"))
	}
}

func TestResourceRepositoryDeployTokenReadNotFoundClearsState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()
	gc := newTestGiteaClient(t, server)

	d := schema.TestResourceDataRaw(t, resourceGiteaRepositoryDeployToken().Schema, map[string]interface{}{})
	d.SetId("owner/repo/7")

	if diags := resourceRepositoryDeployTokenRead(context.Background(), d, gc); diags.HasError() {
		t.Fatalf("unexpected read error: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("expected empty ID after 404, got %q", d.Id())
	}
}

func TestResourceRepositoryDeployTokenDelete(t *testing.T) {
	deleted := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && r.URL.Path == "/api/v1/repos/owner/repo/keys/7" {
			deleted = true
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()
	gc := newTestGiteaClient(t, server)

	d := schema.TestResourceDataRaw(t, resourceGiteaRepositoryDeployToken().Schema, map[string]interface{}{})
	d.SetId("owner/repo/7")

	if diags := resourceRepositoryDeployTokenDelete(context.Background(), d, gc); diags.HasError() {
		t.Fatalf("unexpected delete error: %v", diags)
	}
	if !deleted {
		t.Error("expected the delete endpoint to be called")
	}
}

func TestResourceRepositoryDeployTokenImport(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceGiteaRepositoryDeployToken().Schema, map[string]interface{}{})
	d.SetId("owner/repo/7")

	out, err := resourceRepositoryDeployTokenImport(context.Background(), d, nil)
	if err != nil {
		t.Fatalf("unexpected import error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected one resource data result, got %d", len(out))
	}
	if out[0].Get("username").(string) != "owner" || out[0].Get("name").(string) != "repo" {
		t.Errorf("expected username=owner name=repo, got username=%v name=%v", out[0].Get("username"), out[0].Get("name"))
	}

	if _, err := resourceRepositoryDeployTokenImport(context.Background(), schema.TestResourceDataRaw(t, resourceGiteaRepositoryDeployToken().Schema, map[string]interface{}{}), nil); err == nil {
		t.Error("expected an error for an ID with no parts to import")
	}
}
