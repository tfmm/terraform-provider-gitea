package gitea

import (
	"testing"
)

func TestReleaseImporterParsesID(t *testing.T) {
	resource := resourceGiteaRelease()
	d := resource.Data(nil)
	d.SetId("owner/repo/20")

	results, err := resource.Importer.StateContext(nil, d, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results[0].Id() != "20" {
		t.Errorf("expected Id 20, got %s", results[0].Id())
	}
	if results[0].Get("user").(string) != "owner" || results[0].Get("repo").(string) != "repo" {
		t.Errorf("expected owner/repo, got %s/%s", results[0].Get("user"), results[0].Get("repo"))
	}
}
