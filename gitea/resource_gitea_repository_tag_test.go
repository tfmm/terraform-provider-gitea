package gitea

import (
	"testing"
)

func TestRepositoryTagImporterParsesID(t *testing.T) {
	resource := resourceGiteaRepositoryTag()
	d := resource.Data(nil)
	d.SetId("owner/repo/v1.0.0")

	results, err := resource.Importer.StateContext(nil, d, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results[0].Id() != "owner/repo/v1.0.0" {
		t.Errorf("expected Id owner/repo/v1.0.0, got %s", results[0].Id())
	}
	if results[0].Get("user").(string) != "owner" || results[0].Get("repo").(string) != "repo" || results[0].Get("name").(string) != "v1.0.0" {
		t.Errorf("expected owner/repo/v1.0.0, got %s/%s/%s", results[0].Get("user"), results[0].Get("repo"), results[0].Get("name"))
	}
}
