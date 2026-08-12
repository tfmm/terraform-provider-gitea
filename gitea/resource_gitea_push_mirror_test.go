package gitea

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestPushMirrorImporterParsesCompositeID(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceGiteaPushMirror().Schema, map[string]interface{}{})

	// Test colon format
	d.SetId("owner:repo:remote1")
	out, err := resourceGiteaPushMirror().Importer.StateContext(context.Background(), d, nil)
	if err != nil {
		t.Fatalf("unexpected import error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 resource state, got %d", len(out))
	}
	if out[0].Id() != "owner:repo:remote1" {
		t.Errorf("expected ID 'owner:repo:remote1', got '%s'", out[0].Id())
	}
	if out[0].Get("owner") != "owner" || out[0].Get("repo") != "repo" {
		t.Errorf("expected owner='owner', repo='repo', got owner='%v', repo='%v'", out[0].Get("owner"), out[0].Get("repo"))
	}

	// Test slash format
	d.SetId("owner/repo/remote2")
	out, err = resourceGiteaPushMirror().Importer.StateContext(context.Background(), d, nil)
	if err != nil {
		t.Fatalf("unexpected import error for slash format: %v", err)
	}
	if out[0].Id() != "owner:repo:remote2" {
		t.Errorf("expected normalized ID 'owner:repo:remote2', got '%s'", out[0].Id())
	}
	if out[0].Get("owner") != "owner" || out[0].Get("repo") != "repo" {
		t.Errorf("expected owner='owner', repo='repo', got owner='%v', repo='%v'", out[0].Get("owner"), out[0].Get("repo"))
	}
}

func TestResourcePushMirrorReadSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/v1/repos/owner/repo/push_mirrors/remote1" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(gitea.PushMirrorResponse{
				Created:       "2026-01-01T00:00:00Z",
				Interval:      "8h0m0s",
				LastError:     "",
				LastUpdate:    "2026-01-01T01:00:00Z",
				RemoteAddress: "https://git.example.com/target.git",
				RemoteName:    "remote1",
				RepoName:      "repo",
				SyncONCommit:  true,
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	client, err := gitea.NewClient(server.URL, gitea.SetGiteaVersion("1.25.0"))
	if err != nil {
		t.Fatalf("unexpected client error: %v", err)
	}

	d := schema.TestResourceDataRaw(t, resourceGiteaPushMirror().Schema, map[string]interface{}{})
	d.SetId("owner:repo:remote1")

	err = resourceGiteaPushMirrorRead(d, client)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}

	if d.Id() != "owner:repo:remote1" {
		t.Errorf("expected ID 'owner:repo:remote1', got '%s'", d.Id())
	}
	if d.Get("remote_address") != "https://git.example.com/target.git" {
		t.Errorf("expected remote_address 'https://git.example.com/target.git', got '%v'", d.Get("remote_address"))
	}
	if d.Get("interval") != "8h0m0s" {
		t.Errorf("expected interval '8h0m0s', got '%v'", d.Get("interval"))
	}
	if d.Get("sync_on_commit") != true {
		t.Errorf("expected sync_on_commit true, got '%v'", d.Get("sync_on_commit"))
	}
}

func TestResourcePushMirrorReadNotFoundClearsState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	client, err := gitea.NewClient(server.URL, gitea.SetGiteaVersion("1.25.0"))
	if err != nil {
		t.Fatalf("unexpected client error: %v", err)
	}

	d := schema.TestResourceDataRaw(t, resourceGiteaPushMirror().Schema, map[string]interface{}{})
	d.SetId("owner:repo:nonexistent")

	err = resourceGiteaPushMirrorRead(d, client)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	if d.Id() != "" {
		t.Errorf("expected empty ID after 404, got '%s'", d.Id())
	}
}

func TestResourcePushMirrorCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/v1/repos/owner/repo/push_mirrors" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(gitea.PushMirrorResponse{
				Created:       "2026-01-01T00:00:00Z",
				Interval:      "8h0m0s",
				LastError:     "",
				LastUpdate:    "2026-01-01T00:00:00Z",
				RemoteAddress: "https://git.example.com/target.git",
				RemoteName:    "gitea-push-mirror-1",
				RepoName:      "repo",
				SyncONCommit:  true,
			})
			return
		}
		if r.Method == http.MethodGet && r.URL.Path == "/api/v1/repos/owner/repo/push_mirrors/gitea-push-mirror-1" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(gitea.PushMirrorResponse{
				Created:       "2026-01-01T00:00:00Z",
				Interval:      "8h0m0s",
				LastError:     "",
				LastUpdate:    "2026-01-01T00:00:00Z",
				RemoteAddress: "https://git.example.com/target.git",
				RemoteName:    "gitea-push-mirror-1",
				RepoName:      "repo",
				SyncONCommit:  true,
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	client, err := gitea.NewClient(server.URL, gitea.SetGiteaVersion("1.25.0"))
	if err != nil {
		t.Fatalf("unexpected client error: %v", err)
	}

	d := schema.TestResourceDataRaw(t, resourceGiteaPushMirror().Schema, map[string]interface{}{
		"owner":          "owner",
		"repo":           "repo",
		"remote_address": "https://git.example.com/target.git",
		"interval":       "8h0m0s",
		"sync_on_commit": true,
	})

	err = resourceGiteaPushMirrorCreate(d, client)
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}

	if d.Id() != "owner:repo:gitea-push-mirror-1" {
		t.Errorf("expected ID 'owner:repo:gitea-push-mirror-1', got '%s'", d.Id())
	}
}

func TestResourcePushMirrorDelete(t *testing.T) {
	deleted := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && r.URL.Path == "/api/v1/repos/owner/repo/push_mirrors/remote1" {
			deleted = true
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	client, err := gitea.NewClient(server.URL, gitea.SetGiteaVersion("1.25.0"))
	if err != nil {
		t.Fatalf("unexpected client error: %v", err)
	}

	d := schema.TestResourceDataRaw(t, resourceGiteaPushMirror().Schema, map[string]interface{}{})
	d.SetId("owner:repo:remote1")

	err = resourceGiteaPushMirrorDelete(d, client)
	if err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}
	if !deleted {
		t.Error("expected delete endpoint to be called")
	}
}

func TestAccGiteaPushMirror_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGiteaPushMirrorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGiteaPushMirrorConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("gitea_push_mirror.test", "owner", "testuser"),
					resource.TestCheckResourceAttr("gitea_push_mirror.test", "repo", "testrepo"),
					resource.TestCheckResourceAttr("gitea_push_mirror.test", "remote_address", "https://github.com/example/target.git"),
					resource.TestCheckResourceAttrSet("gitea_push_mirror.test", "remote_name"),
				),
			},
		},
	})
}

func testAccCheckGiteaPushMirrorDestroy(s *terraform.State) error {
	client, err := testAccNewGiteaClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "gitea_push_mirror" {
			continue
		}

		owner, repo, remoteName, err := parsePushMirrorID(rs.Primary.ID)
		if err != nil {
			return err
		}

		pm, _, err := client.GetPushMirrorByRemoteName(owner, repo, remoteName)
		if err == nil && pm != nil {
			return fmt.Errorf("push mirror %s still exists on %s/%s", remoteName, owner, repo)
		}
	}

	return nil
}

func testAccGiteaPushMirrorConfig() string {
	return `
resource "gitea_user" "test" {
  username   = "pushmirroruser"
  login_name = "pushmirroruser"
  password   = "Geheim1!"
  email      = "pushmirroruser@user.dev"
}

resource "gitea_repository" "test" {
  name     = "pushmirrorrepo"
  username = gitea_user.test.username
  private  = true
}

resource "gitea_push_mirror" "test" {
  owner          = gitea_user.test.username
  repo           = gitea_repository.test.name
  remote_address = "https://github.com/example/target.git"
  interval       = "8h0m0s"
  sync_on_commit = true
}
`
}
