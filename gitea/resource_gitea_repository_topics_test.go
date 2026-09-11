package gitea

import (
	"testing"
)

func TestRepositoryTopicsImporterParsesID(t *testing.T) {
	resource := resourceGiteaRepositoryTopics()
	d := resource.Data(nil)
	d.SetId("owner/repo")

	results, err := resource.Importer.StateContext(nil, d, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results[0].Id() != "owner/repo" {
		t.Errorf("expected Id owner/repo, got %s", results[0].Id())
	}
	if results[0].Get("user").(string) != "owner" || results[0].Get("repo").(string) != "repo" {
		t.Errorf("expected owner/repo, got %s/%s", results[0].Get("user"), results[0].Get("repo"))
	}
}
