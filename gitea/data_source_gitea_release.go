package gitea

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGiteaRelease() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGiteaReleaseRead,
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

func dataSourceGiteaReleaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*GiteaClient).Client
	user := d.Get("user").(string)
	repo := d.Get("repo").(string)
	id := int64(d.Get("id").(int))

	release, _, err := client.GetRelease(user, repo, id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(release.ID, 10))
	if err := d.Set("tag_name", release.TagName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("target_commitish", release.Target); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("title", release.Title); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("note", release.Note); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("draft", release.IsDraft); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("prerelease", release.IsPrerelease); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func dataSourceGiteaReleases() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGiteaReleasesRead,
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
			"tag_filter": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter releases by tag, matched server-side. Supports \"*\" as a wildcard (e.g. \"v1*\", \"*beta\", \"*rc*\"). Requires Gitea >= 28.0.0.",
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

func dataSourceGiteaReleasesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*GiteaClient)
	user := d.Get("user").(string)
	repo := d.Get("repo").(string)
	tagFilter := d.Get("tag_filter").(string)

	var releases []*gitea.Release
	if tagFilter != "" {
		// tag_filter is not yet supported by code.gitea.io/sdk/gitea's
		// ListReleasesOptions, so this issues the request directly.
		path := fmt.Sprintf("/repos/%s/%s/releases?tag_filter=%s", url.PathEscape(user), url.PathEscape(repo), url.QueryEscape(tagFilter))
		if err := client.rawJSON("GET", path, nil, &releases); err != nil {
			return diag.FromErr(err)
		}
	} else {
		var err error
		releases, _, err = client.Client.ListReleases(user, repo, gitea.ListReleasesOptions{})
		if err != nil {
			return diag.FromErr(err)
		}
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

	d.SetId(fmt.Sprintf("%s/%s/releases/%s", user, repo, tagFilter))
	if err := d.Set("releases", result); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
