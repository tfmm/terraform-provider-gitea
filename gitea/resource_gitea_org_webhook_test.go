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
