package gitea

import (
	"fmt"
	"strconv"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGiteaRelease() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaReleaseRead,
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
			"id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Release ID",
			},
			"tag_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Tag name",
			},
			"target_commitish": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Target commitish",
			},
			"title": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Release title",
			},
			"note": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Release notes / body",
			},
			"draft": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the release is a draft",
			},
			"prerelease": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the release is a pre-release",
			},
		},
		Description: "Use this data source to retrieve details of a repository release in Gitea.",
	}
}

func dataSourceGiteaReleaseRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	user := d.Get("user").(string)
	repo := d.Get("repo").(string)
	id := int64(d.Get("id").(int))

	release, _, err := client.GetRelease(user, repo, id)
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(release.ID, 10))
	d.Set("tag_name", release.TagName)
	d.Set("target_commitish", release.Target)
	d.Set("title", release.Title)
	d.Set("note", release.Note)
	d.Set("draft", release.IsDraft)
	d.Set("prerelease", release.IsPrerelease)

	return nil
}

func dataSourceGiteaReleases() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaReleasesRead,
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
			"releases": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"tag_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"target_commitish": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"title": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"draft": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"prerelease": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
		Description: "Use this data source to list releases in a Gitea repository.",
	}
}

func dataSourceGiteaReleasesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	user := d.Get("user").(string)
	repo := d.Get("repo").(string)

	releases, _, err := client.ListReleases(user, repo, gitea.ListReleasesOptions{})
	if err != nil {
		return err
	}

	result := make([]map[string]interface{}, 0, len(releases))
	for _, rel := range releases {
		result = append(result, map[string]interface{}{
			"id":               rel.ID,
			"tag_name":         rel.TagName,
			"target_commitish": rel.Target,
			"title":            rel.Title,
			"draft":            rel.IsDraft,
			"prerelease":       rel.IsPrerelease,
		})
	}

	d.SetId(fmt.Sprintf("%s/%s/releases", user, repo))
	d.Set("releases", result)

	return nil
}
