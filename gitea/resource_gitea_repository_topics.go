package gitea

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceRepositoryTopicsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*GiteaClient).Client
	user := d.Get("user").(string)
	repoName := d.Get("repo").(string)

	repo, resp, err := client.GetRepo(user, repoName)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	if err := d.Set("topics", repo.Topics); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceRepositoryTopicsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*GiteaClient).Client
	user := d.Get("user").(string)
	repoName := d.Get("repo").(string)

	rawTopics := d.Get("topics").(*schema.Set).List()
	topics := make([]string, 0, len(rawTopics))
	for _, t := range rawTopics {
		topics = append(topics, t.(string))
	}

	_, err := client.SetRepoTopics(user, repoName, topics)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s/%s", user, repoName))
	return resourceRepositoryTopicsRead(ctx, d, meta)
}

func resourceRepositoryTopicsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceRepositoryTopicsCreate(ctx, d, meta)
}

func resourceRepositoryTopicsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*GiteaClient).Client
	user := d.Get("user").(string)
	repoName := d.Get("repo").(string)

	_, err := client.SetRepoTopics(user, repoName, []string{})
	return diag.FromErr(err)
}

func resourceGiteaRepositoryTopics() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceRepositoryTopicsRead,
		CreateContext: resourceRepositoryTopicsCreate,
		UpdateContext: resourceRepositoryTopicsUpdate,
		DeleteContext: resourceRepositoryTopicsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				parts := strings.Split(d.Id(), "/")
				if len(parts) != 2 {
					return nil, fmt.Errorf("unexpected ID format (%q), expected <user>/<repo>", d.Id())
				}
				if err := d.Set("user", parts[0]); err != nil {
					return nil, err
				}
				if err := d.Set("repo", parts[1]); err != nil {
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
			"topics": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required:    true,
				Description: "Set of topics to apply to the repository",
			},
		},
		Description: "This resource manages topics assigned to a Gitea repository.",
	}
}
