package gitea

import (
	"context"
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceRepositoryTagRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*GiteaClient).Client
	user := d.Get("user").(string)
	repo := d.Get("repo").(string)
	name := d.Get("name").(string)

	tag, resp, err := client.GetTag(user, repo, name)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	if err := d.Set("name", tag.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("commit_sha", tag.Commit.SHA); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("message", tag.Message); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRepositoryTagCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*GiteaClient).Client
	user := d.Get("user").(string)
	repo := d.Get("repo").(string)
	name := d.Get("name").(string)

	opts := gitea.CreateTagOption{
		TagName: name,
		Target:  d.Get("target").(string),
		Message: d.Get("message").(string),
	}

	_, _, err := client.CreateTag(user, repo, opts)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s/%s/%s", user, repo, name))
	return resourceRepositoryTagRead(ctx, d, meta)
}

func resourceRepositoryTagDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*GiteaClient).Client
	user := d.Get("user").(string)
	repo := d.Get("repo").(string)
	name := d.Get("name").(string)

	_, err := client.DeleteTag(user, repo, name)
	return diag.FromErr(err)
}

func resourceGiteaRepositoryTag() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceRepositoryTagRead,
		CreateContext: resourceRepositoryTagCreate,
		DeleteContext: resourceRepositoryTagDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				parts := strings.Split(d.Id(), "/")
				if len(parts) != 3 {
					return nil, fmt.Errorf("unexpected ID format (%q), expected <user>/<repo>/<tag_name>", d.Id())
				}
				if err := d.Set("user", parts[0]); err != nil {
					return nil, err
				}
				if err := d.Set("repo", parts[1]); err != nil {
					return nil, err
				}
				if err := d.Set("name", parts[2]); err != nil {
					return nil, err
				}
				d.SetId(d.Id())
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
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Tag name",
			},
			"target": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Target commit SHA or branch name (defaults to default branch)",
			},
			"message": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Tag message for annotated tags",
			},
			"commit_sha": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The SHA of the commit that the tag points to",
			},
		},
		Description: "This resource manages git tags in a Gitea repository.",
	}
}
