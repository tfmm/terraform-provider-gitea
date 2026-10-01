package gitea

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceGiteaReleasesWithoutTagFilter(t *testing.T) {
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/v1/repos/owner/repo/releases" {
			gotQuery = r.URL.RawQuery
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]*gitea.Release{
				{ID: 1, TagName: "v1.0.0"},
				{ID: 2, TagName: "v2.0.0-beta"},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()
	gc := newTestGiteaClient(t, server)

	d := schema.TestResourceDataRaw(t, dataSourceGiteaReleases().Schema, map[string]interface{}{
		"user": "owner",
		"repo": "repo",
	})

	if diags := dataSourceGiteaReleasesRead(context.Background(), d, gc); diags.HasError() {
		t.Fatalf("unexpected read error: %v", diags)
	}

	if strings.Contains(gotQuery, "tag_filter") {
		t.Errorf("expected no tag_filter query param to be sent, got query %q", gotQuery)
	}
	releases := d.Get("releases").([]interface{})
	if len(releases) != 2 {
		t.Fatalf("expected 2 releases, got %d", len(releases))
	}
	if d.Id() != "owner/repo/releases/" {
		t.Errorf("expected id 'owner/repo/releases/', got %q", d.Id())
	}
}

func TestDataSourceGiteaReleasesWithTagFilter(t *testing.T) {
	var gotPath, gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/v1/repos/owner/repo/releases" {
			gotPath = r.URL.Path
			gotQuery = r.URL.RawQuery
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]*gitea.Release{
				{ID: 1, TagName: "v1.0.0"},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()
	gc := newTestGiteaClient(t, server)

	d := schema.TestResourceDataRaw(t, dataSourceGiteaReleases().Schema, map[string]interface{}{
		"user":       "owner",
		"repo":       "repo",
		"tag_filter": "v1*",
	})

	if diags := dataSourceGiteaReleasesRead(context.Background(), d, gc); diags.HasError() {
		t.Fatalf("unexpected read error: %v", diags)
	}

	if gotPath != "/api/v1/repos/owner/repo/releases" {
		t.Errorf("expected path /api/v1/repos/owner/repo/releases, got %q", gotPath)
	}
	if gotQuery != "tag_filter=v1%2A" {
		t.Errorf("expected query tag_filter=v1%%2A, got %q", gotQuery)
	}
	releases := d.Get("releases").([]interface{})
	if len(releases) != 1 {
		t.Fatalf("expected 1 filtered release, got %d", len(releases))
	}
	if got := releases[0].(map[string]interface{})["tag_name"]; got != "v1.0.0" {
		t.Errorf("expected tag_name v1.0.0, got %v", got)
	}
	if d.Id() != "owner/repo/releases/v1*" {
		t.Errorf("expected id to incorporate the tag_filter, got %q", d.Id())
	}
}
