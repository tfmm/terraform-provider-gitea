package gitea

import (
	"testing"
)

func TestLabelImporterParsesID(t *testing.T) {
	resource := resourceGiteaLabel()

	// Test repo label ID format
	dRepo := resource.Data(nil)
	dRepo.SetId("owner/repo/456")

	resultsRepo, err := resource.Importer.StateContext(nil, dRepo, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resultsRepo[0].Id() != "456" {
		t.Errorf("expected Id 456, got %s", resultsRepo[0].Id())
	}
	if resultsRepo[0].Get("user").(string) != "owner" || resultsRepo[0].Get("repo").(string) != "repo" {
		t.Errorf("expected owner/repo, got %s/%s", resultsRepo[0].Get("user"), resultsRepo[0].Get("repo"))
	}

	// Test org label ID format
	dOrg := resource.Data(nil)
	dOrg.SetId("my-org/789")

	resultsOrg, err := resource.Importer.StateContext(nil, dOrg, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resultsOrg[0].Id() != "789" {
		t.Errorf("expected Id 789, got %s", resultsOrg[0].Id())
	}
	if resultsOrg[0].Get("org").(string) != "my-org" {
		t.Errorf("expected org my-org, got %s", resultsOrg[0].Get("org"))
	}
}
