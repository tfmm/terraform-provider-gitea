package gitea

import (
	"fmt"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGiteaRepositoryTopics() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaRepositoryTopicsRead,
		Schema: map[string]*schema.Schema{
			"user": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "User or organization owner of the repository",
			},
			"repo": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Repository name",
			},
			"topics": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of topics set on the repository",
			},
		},
		Description: "Use this data source to retrieve topics of a Gitea repository.",
	}
}

func dataSourceGiteaRepositoryTopicsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	user := d.Get("user").(string)
	repoName := d.Get("repo").(string)

	repo, _, err := client.GetRepo(user, repoName)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s/%s/topics", user, repoName))
	d.Set("topics", repo.Topics)

	return nil
}
