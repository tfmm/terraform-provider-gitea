package gitea

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMilestoneRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}

	user := d.Get("user").(string)
	repo := d.Get("repo").(string)

	milestone, resp, err := client.GetMilestone(user, repo, id)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("title", milestone.Title)
	d.Set("description", milestone.Description)
	d.Set("state", string(milestone.State))
	if milestone.Deadline != nil {
		d.Set("due_on", milestone.Deadline.Format(time.RFC3339))
	}

	return nil
}

func resourceMilestoneCreate(d *schema.ResourceData, meta interface{}) error {
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
				return fmt.Errorf("invalid due_on date format: %w", err)
			}
		}
		opts.Deadline = &t
	}

	milestone, _, err := client.CreateMilestone(user, repo, opts)
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(milestone.ID, 10))
	return resourceMilestoneRead(d, meta)
}

func resourceMilestoneUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
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
				return fmt.Errorf("invalid due_on date format: %w", err)
			}
		}
		opts.Deadline = &t
	}

	_, _, err = client.EditMilestone(user, repo, id, opts)
	if err != nil {
		return err
	}

	return resourceMilestoneRead(d, meta)
}

func resourceMilestoneDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}

	user := d.Get("user").(string)
	repo := d.Get("repo").(string)

	_, err = client.DeleteMilestone(user, repo, id)
	return err
}

func resourceGiteaMilestone() *schema.Resource {
	return &schema.Resource{
		Read:   resourceMilestoneRead,
		Create: resourceMilestoneCreate,
		Update: resourceMilestoneUpdate,
		Delete: resourceMilestoneDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				parts := strings.Split(d.Id(), "/")
				if len(parts) != 3 {
					return nil, fmt.Errorf("unexpected ID format (%q), expected <user>/<repo>/<milestone_id>", d.Id())
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
