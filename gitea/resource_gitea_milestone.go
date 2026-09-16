package gitea

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMilestoneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	user := d.Get("user").(string)
	repo := d.Get("repo").(string)

	milestone, resp, err := client.GetMilestone(user, repo, id)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	if err := d.Set("title", milestone.Title); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", milestone.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("state", string(milestone.State)); err != nil {
		return diag.FromErr(err)
	}
	if milestone.Deadline != nil {
		if err := d.Set("due_on", milestone.Deadline.Format(time.RFC3339)); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceMilestoneCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	user := d.Get("user").(string)
	repo := d.Get("repo").(string)

	opts := gitea.CreateMilestoneOption{
		Title:       d.Get("title").(string),
		Description: d.Get("description").(string),
		State:       gitea.StateType(d.Get("state").(string)),
	}

	if dueOnStr, ok := d.GetOk("due_on"); ok && dueOnStr.(string) != "" {
		t, err := time.Parse(time.RFC3339, dueOnStr.(string))
		if err != nil {
			// Try YYYY-MM-DD
			t, err = time.Parse("2006-01-02", dueOnStr.(string))
			if err != nil {
				return diag.FromErr(fmt.Errorf("invalid due_on date format: %w", err))
			}
		}
		opts.Deadline = &t
	}

	milestone, _, err := client.CreateMilestone(user, repo, opts)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(milestone.ID, 10))
	return resourceMilestoneRead(ctx, d, meta)
}

func resourceMilestoneUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	user := d.Get("user").(string)
	repo := d.Get("repo").(string)

	title := d.Get("title").(string)
	description := d.Get("description").(string)
	state := gitea.StateType(d.Get("state").(string))

	opts := gitea.EditMilestoneOption{
		Title:       title,
		Description: &description,
		State:       &state,
	}

	if dueOnStr, ok := d.GetOk("due_on"); ok && dueOnStr.(string) != "" {
		t, err := time.Parse(time.RFC3339, dueOnStr.(string))
		if err != nil {
			t, err = time.Parse("2006-01-02", dueOnStr.(string))
			if err != nil {
				return diag.FromErr(fmt.Errorf("invalid due_on date format: %w", err))
			}
		}
		opts.Deadline = &t
	}

	_, _, err = client.EditMilestone(user, repo, id, opts)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceMilestoneRead(ctx, d, meta)
}

func resourceMilestoneDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	user := d.Get("user").(string)
	repo := d.Get("repo").(string)

	_, err = client.DeleteMilestone(user, repo, id)
	return diag.FromErr(err)
}

func resourceGiteaMilestone() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceMilestoneRead,
		CreateContext: resourceMilestoneCreate,
		UpdateContext: resourceMilestoneUpdate,
		DeleteContext: resourceMilestoneDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				parts := strings.Split(d.Id(), "/")
				if len(parts) != 3 {
					return nil, fmt.Errorf("unexpected ID format (%q), expected <user>/<repo>/<milestone_id>", d.Id())
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
			"title": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Milestone title",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Milestone description",
			},
			"state": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "open",
				Description: "Milestone state: `open` or `closed`",
			},
			"due_on": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Due date/timestamp for the milestone (e.g. `2026-12-31T23:59:59Z` or `2026-12-31`)",
			},
		},
		Description: "This resource manages repository milestones in Gitea.",
	}
}
