package gitea

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"gitea": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()
}
func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("GITEA_TOKEN"); v == "" {
		t.Fatal("GITEA_TOKEN must be set for acceptance tests")
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

