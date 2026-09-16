package gitea

import (
	"context"
	"fmt"
	"strconv"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	forkOwner        string = "owner"
	forkRepo         string = "repo"
	forkOrganization string = "organization"
)

func resourceForkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error
	client := meta.(*gitea.Client)

	var opts gitea.CreateForkOption
	var org string
	org = d.Get(forkOrganization).(string)
	if org != "" {
		opts.Organization = &org
	}

	repo, _, err := client.CreateFork(d.Get(forkOwner).(string),
		d.Get(forkRepo).(string),
		opts)
	if err == nil {
		err = setForkResourceData(repo, client, d)
	}
	return diag.FromErr(err)
}

func resourceForkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error
	client := meta.(*gitea.Client)

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	var resp *gitea.Response

	if err != nil {
		return diag.FromErr(err)
	}

	repo, resp, err := client.GetRepoByID(id)

	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		} else {
			return diag.FromErr(err)
		}
	}

	err = setForkResourceData(repo, client, d)

	return diag.FromErr(err)
}

func resourceForkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error
	client := meta.(*gitea.Client)

	id, err := strconv.ParseInt(d.Id(), 10, 64)

	if err != nil {
		return diag.FromErr(err)
	}

	repo, resp, err := client.GetRepoByID(id)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	resp, err = client.DeleteRepo(repo.Owner.UserName, repo.Name)

	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return diag.FromErr(err)
		} else {
			return diag.FromErr(err)
		}
	}

	return diag.FromErr(err)
}

func setForkResourceData(repo *gitea.Repository, client *gitea.Client, d *schema.ResourceData) (err error) {
	d.SetId(fmt.Sprintf("%d", repo.ID))

	owner := d.Get(forkOwner).(string)
	name := d.Get(forkRepo).(string)
	if repo.Parent != nil {
		if repo.Parent.Owner != nil {
			owner = repo.Parent.Owner.UserName
		}
		name = repo.Parent.Name
	}
	if err := d.Set(forkOwner, owner); err != nil {
		return err
	}
	if err := d.Set(forkRepo, name); err != nil {
		return err
	}

	organization := ""
	if repo.Owner != nil {
		_, resp, orgErr := client.GetOrg(repo.Owner.UserName)
		if orgErr != nil {
			if resp != nil && resp.StatusCode != 404 {
				return orgErr
			}
		} else {
			organization = repo.Owner.UserName
		}
	}
	if err := d.Set(forkOrganization, organization); err != nil {
		return err
	}

	return
}

func resourceGiteaFork() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceForkRead,
		CreateContext: resourceForkCreate,
		DeleteContext: resourceForkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"owner": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The owner or owning organization of the repository to fork",
			},
			"repo": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the repository to fork",
			},
			"organization": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				ForceNew:    true,
				Description: "The organization that owns the forked repo",
			},
		},
		Description: "`gitea_fork` manages repository fork to the current user or an organisation\n" +
			"Forking a repository to a dedicated user is currently unsupported\n" +
			"Creating a fork using this resource without an organisation will create the fork in the executors name",
	}
}
