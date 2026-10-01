package gitea

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceReleaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*GiteaClient).Client
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	user := d.Get("user").(string)
	repo := d.Get("repo").(string)

	release, resp, err := client.GetRelease(user, repo, id)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
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

func resourceReleaseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*GiteaClient).Client
	user := d.Get("user").(string)
	repo := d.Get("repo").(string)

	opts := gitea.CreateReleaseOption{
		TagName:      d.Get("tag_name").(string),
		Target:       d.Get("target_commitish").(string),
		Title:        d.Get("title").(string),
		Note:         d.Get("note").(string),
		IsDraft:      d.Get("draft").(bool),
		IsPrerelease: d.Get("prerelease").(bool),
	}

	release, _, err := client.CreateRelease(user, repo, opts)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(release.ID, 10))
	return resourceReleaseRead(ctx, d, meta)
}

func resourceReleaseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*GiteaClient).Client
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	user := d.Get("user").(string)
	repo := d.Get("repo").(string)

	draft := d.Get("draft").(bool)
	prerelease := d.Get("prerelease").(bool)

	opts := gitea.EditReleaseOption{
		TagName:      d.Get("tag_name").(string),
		Target:       d.Get("target_commitish").(string),
		Title:        d.Get("title").(string),
		Note:         d.Get("note").(string),
		IsDraft:      &draft,
		IsPrerelease: &prerelease,
	}

	_, _, err = client.EditRelease(user, repo, id, opts)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceReleaseRead(ctx, d, meta)
}

func resourceReleaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*GiteaClient).Client
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	user := d.Get("user").(string)
	repo := d.Get("repo").(string)

	_, err = client.DeleteRelease(user, repo, id)
	return diag.FromErr(err)
}

func resourceGiteaRelease() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceReleaseRead,
		CreateContext: resourceReleaseCreate,
		UpdateContext: resourceReleaseUpdate,
		DeleteContext: resourceReleaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				parts := strings.Split(d.Id(), "/")
				if len(parts) != 3 {
					return nil, fmt.Errorf("unexpected ID format (%q), expected <user>/<repo>/<release_id>", d.Id())
				}
				if err := d.Set("user", parts[0]); err != nil {
					return nil, err
				}
				if err := d.Set("repo", parts[1]); err != nil {
					return nil, err
				}
				d.SetId(parts[2])
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"user": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "User or organization owner of the repository",
			},
			"repo": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Repository name",
			},
			"tag_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the tag for this release",
			},
			"target_commitish": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "main",
				Description: "Specifies the commitish value that determines where the Git tag is created from",
			},
			"title": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Title of the release",
			},
			"note": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Release notes / body content",
			},
			"draft": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether this release is a draft",
			},
			"prerelease": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether this release is a pre-release",
			},
		},
		Description: "This resource manages repository releases in Gitea.",
	}
}
