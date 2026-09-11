package gitea

import (
	"fmt"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGiteaRepositoryTag() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaRepositoryTagRead,
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
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Tag name",
			},
			"commit_sha": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Commit SHA of the tag",
			},
			"message": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Tag message",
			},
		},
		Description: "Use this data source to retrieve details of a specific tag in a Gitea repository.",
	}
}

func dataSourceGiteaRepositoryTagRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	user := d.Get("user").(string)
	repo := d.Get("repo").(string)
	name := d.Get("name").(string)

	tag, _, err := client.GetTag(user, repo, name)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s/%s/%s", user, repo, name))
	d.Set("commit_sha", tag.Commit.SHA)
	d.Set("message", tag.Message)

	return nil
}

func dataSourceGiteaRepositoryTags() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaRepositoryTagsRead,
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
			"tags": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"commit_sha": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"message": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
		Description: "Use this data source to list all tags in a Gitea repository.",
	}
}

func dataSourceGiteaRepositoryTagsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	user := d.Get("user").(string)
	repo := d.Get("repo").(string)

	tags, _, err := client.ListRepoTags(user, repo, gitea.ListRepoTagsOptions{})
	if err != nil {
		return err
	}

	result := make([]map[string]interface{}, 0, len(tags))
	for _, t := range tags {
		commitSHA := ""
		if t.Commit != nil {
			commitSHA = t.Commit.SHA
		}
		result = append(result, map[string]interface{}{
			"name":       t.Name,
			"commit_sha": commitSHA,
			"message":    t.Message,
		})
	}

	d.SetId(fmt.Sprintf("%s/%s/tags", user, repo))
	d.Set("tags", result)

	return nil
}
