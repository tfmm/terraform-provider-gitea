package gitea

import (
	"testing"
)

func TestOrgWebhookImporterParsesID(t *testing.T) {
	resource := resourceGiteaOrgWebhook()
	d := resource.Data(nil)
	d.SetId("my-org/123")

	if resource.Importer == nil || resource.Importer.StateContext == nil {
		t.Fatal("expected importer state context to be defined")
	}

	results, err := resource.Importer.StateContext(nil, d, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	res := results[0]
	if res.Id() != "123" {
		t.Errorf("expected Id 123, got %s", res.Id())
	}
	if res.Get("org").(string) != "my-org" {
		t.Errorf("expected org my-org, got %s", res.Get("org").(string))
	}
}

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

func TestMilestoneImporterParsesID(t *testing.T) {
	resource := resourceGiteaMilestone()
	d := resource.Data(nil)
	d.SetId("owner/repo/10")

	results, err := resource.Importer.StateContext(nil, d, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results[0].Id() != "10" {
		t.Errorf("expected Id 10, got %s", results[0].Id())
	}
	if results[0].Get("user").(string) != "owner" || results[0].Get("repo").(string) != "repo" {
		t.Errorf("expected owner/repo, got %s/%s", results[0].Get("user"), results[0].Get("repo"))
	}
}

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

func TestNewResourcesRegisteredInProvider(t *testing.T) {
	p := Provider()
	expectedResources := []string{
		"gitea_org_webhook",
		"gitea_label",
		"gitea_milestone",
		"gitea_release",
		"gitea_repository_tag",
		"gitea_repository_topics",
	}

	for _, r := range expectedResources {
		if _, ok := p.ResourcesMap[r]; !ok {
			t.Errorf("resource %s is not registered in Provider", r)
		}
	}

	expectedDataSources := []string{
		"gitea_org_webhook",
		"gitea_org_webhooks",
		"gitea_label",
		"gitea_labels",
		"gitea_milestone",
		"gitea_milestones",
		"gitea_release",
		"gitea_releases",
		"gitea_repository_tag",
		"gitea_repository_tags",
		"gitea_repository_topics",
	}

	for _, ds := range expectedDataSources {
		if _, ok := p.DataSourcesMap[ds]; !ok {
			t.Errorf("data source %s is not registered in Provider", ds)
		}
	}
}
