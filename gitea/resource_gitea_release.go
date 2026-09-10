package gitea

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceReleaseRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}

	user := d.Get("user").(string)
	repo := d.Get("repo").(string)

	release, resp, err := client.GetRelease(user, repo, id)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("tag_name", release.TagName)
	d.Set("target_commitish", release.Target)
	d.Set("title", release.Title)
	d.Set("note", release.Note)
	d.Set("draft", release.IsDraft)
	d.Set("prerelease", release.IsPrerelease)

	return nil
}

func resourceReleaseCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
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
		return err
	}

	d.SetId(strconv.FormatInt(release.ID, 10))
	return resourceReleaseRead(d, meta)
}

func resourceReleaseUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
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
		return err
	}

	return resourceReleaseRead(d, meta)
}

func resourceReleaseDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}

	user := d.Get("user").(string)
	repo := d.Get("repo").(string)

	_, err = client.DeleteRelease(user, repo, id)
	return err
}

func resourceGiteaRelease() *schema.Resource {
	return &schema.Resource{
		Read:   resourceReleaseRead,
		Create: resourceReleaseCreate,
		Update: resourceReleaseUpdate,
		Delete: resourceReleaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				parts := strings.Split(d.Id(), "/")
				if len(parts) != 3 {
					return nil, fmt.Errorf("unexpected ID format (%q), expected <user>/<repo>/<release_id>", d.Id())
				}
				d.Set("user", parts[0])
				d.Set("repo", parts[1])
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
