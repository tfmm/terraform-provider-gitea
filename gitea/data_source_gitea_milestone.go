package gitea

import (
	"fmt"
	"strconv"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGiteaMilestone() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaMilestoneRead,
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
				Description: "Milestone ID",
			},
			"title": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Milestone title",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Milestone description",
			},
			"state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Milestone state (`open` or `closed`)",
			},
			"due_on": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Due date/timestamp",
			},
			"open_issues": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of open issues in milestone",
			},
			"closed_issues": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of closed issues in milestone",
			},
		},
		Description: "Use this data source to retrieve details of a repository milestone in Gitea.",
	}
}

func dataSourceGiteaMilestoneRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	user := d.Get("user").(string)
	repo := d.Get("repo").(string)
	id := int64(d.Get("id").(int))

	milestone, _, err := client.GetMilestone(user, repo, id)
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(milestone.ID, 10))
	d.Set("title", milestone.Title)
	d.Set("description", milestone.Description)
	d.Set("state", string(milestone.State))
	d.Set("open_issues", milestone.OpenIssues)
	d.Set("closed_issues", milestone.ClosedIssues)

	if milestone.Deadline != nil {
		d.Set("due_on", milestone.Deadline.Format(time.RFC3339))
	}

	return nil
}

func dataSourceGiteaMilestones() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaMilestonesRead,
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
			"state": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "all",
				Description: "Filter state: `open`, `closed`, or `all`",
			},
			"milestones": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"title": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"state": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"open_issues": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"closed_issues": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
		Description: "Use this data source to list milestones in a Gitea repository.",
	}
}

func dataSourceGiteaMilestonesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	user := d.Get("user").(string)
	repo := d.Get("repo").(string)
	state := d.Get("state").(string)

	milestones, _, err := client.ListRepoMilestones(user, repo, gitea.ListMilestoneOption{
		State: gitea.StateType(state),
	})
	if err != nil {
		return err
	}

	result := make([]map[string]interface{}, 0, len(milestones))
	for _, ms := range milestones {
		result = append(result, map[string]interface{}{
			"id":            ms.ID,
			"title":         ms.Title,
			"description":   ms.Description,
			"state":         string(ms.State),
			"open_issues":   ms.OpenIssues,
			"closed_issues": ms.ClosedIssues,
		})
	}

	d.SetId(fmt.Sprintf("%s/%s/milestones", user, repo))
	d.Set("milestones", result)

	return nil
}
