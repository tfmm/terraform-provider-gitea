package gitea

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGiteaRepo() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGiteaRepoRead,

		Schema: map[string]*schema.Schema{
			"username": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"full_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"fork": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"mirror": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"html_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ssh_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"clone_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"website": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"stars": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"forks": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"watchers": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"open_issue_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"default_branch": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"permission_admin": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"permission_push": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"permission_pull": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"default_merge_style": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Default merge style for pull requests in the repository",
			},
		},
	}
}

func dataSourceGiteaRepoRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*GiteaClient).Client

	usernameData, usernameOk := d.GetOk("username")
	if !usernameOk {
		return diag.FromErr(fmt.Errorf("name of repo owner must be passed"))
	}
	username := strings.ToLower(usernameData.(string))

	nameData, nameOk := d.GetOk("name")
	if !nameOk {
		return diag.FromErr(fmt.Errorf("name of repo must be passed"))
	}
	name := strings.ToLower(nameData.(string))

	repo, _, err := client.GetRepo(username, name)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d", repo.ID))
	if err := d.Set("name", repo.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", repo.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("full_name", repo.FullName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", repo.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("private", repo.Private); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("fork", repo.Fork); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("mirror", repo.Mirror); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("size", repo.Size); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("html_url", repo.HTMLURL); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("ssh_url", repo.SSHURL); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("clone_url", repo.CloneURL); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("website", repo.Website); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("stars", repo.Stars); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("forks", repo.Forks); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("watchers", repo.Watchers); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("open_issue_count", repo.OpenIssues); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("default_branch", repo.DefaultBranch); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("default_merge_style", string(repo.DefaultMergeStyle)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created", timeToString(repo.Created)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("updated", timeToString(repo.Updated)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("permission_admin", repo.Permissions.Admin); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("permission_push", repo.Permissions.Push); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("permission_pull", repo.Permissions.Pull); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
