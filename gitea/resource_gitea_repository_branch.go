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

const (
	repoBranchName string = "name"
	repoBranchRepo string = "repository"
)

func resourceRepoBranchIdParts(d *schema.ResourceData) (hasId bool, repoId int64, branchId string, err error) {
	parts := strings.SplitN(d.Id(), "/", 2)
	if len(parts) != 2 {
		return false, 0, "", nil
	}

	repoId, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return false, 0, "", err
	}

	branchId = parts[1]
	return true, repoId, branchId, err
}

func resourceRepoBranchRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error
	client := meta.(*gitea.Client)

	hasId, repoId, branchId, err := resourceRepoBranchIdParts(d)
	if err != nil {
		return diag.FromErr(err)
	}
	if !hasId {
		d.SetId("")
		return nil
	}

	repo, resp, err := client.GetRepoByID(repoId)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		} else {
			return diag.FromErr(err)
		}
	}

	branch, resp, err := client.GetRepoBranch(repo.Owner.UserName, repo.Name, branchId)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		} else {
			return diag.FromErr(err)
		}
	}

	err = setRepoBranchResourceData(branch, repoId, d)

	return diag.FromErr(err)
}

func resourceRepoBranchCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error
	client := meta.(*gitea.Client)

	repo, _, err := client.GetRepoByID(int64(d.Get(repoBranchRepo).(int)))

	if err != nil {
		return diag.FromErr(err)
	}

	rb, _, err := client.CreateBranch(repo.Owner.UserName, repo.Name, gitea.CreateBranchOption{
		BranchName: d.Get(repoBranchName).(string),
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d/%s", repo.ID, d.Get(repoBranchName).(string)))

	err = setRepoBranchResourceData(rb, repo.ID, d)
	return diag.FromErr(err)
}

func resourceRepoBranchDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error
	client := meta.(*gitea.Client)
	hasId, repoId, branchId, err := resourceRepoBranchIdParts(d)
	if err != nil {
		return diag.FromErr(err)
	}
	if !hasId {
		d.SetId("")
		return nil
	}

	repo, resp, err := client.GetRepoByID(repoId)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	deleted, resp, err := client.DeleteRepoBranch(repo.Owner.UserName, repo.Name, branchId)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	if !deleted {
		return diag.FromErr(fmt.Errorf("branch %q was not deleted", branchId))
	}
	return nil
}

func setRepoBranchResourceData(rb *gitea.Branch, repoId int64, d *schema.ResourceData) (err error) {
	if err := d.Set(repoBranchName, rb.Name); err != nil {
		return err
	}
	if err := d.Set(repoBranchRepo, repoId); err != nil {
		return err
	}
	return
}

func resourceGiteaRepositoryBranch() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceRepoBranchRead,
		CreateContext: resourceRepoBranchCreate,
		DeleteContext: resourceRepoBranchDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			repoBranchName: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the created branch",
			},
			repoBranchRepo: {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the target repository",
			},
		},
		Description: "`gitea_repository_branch` manages a branch for a single `gitea_repository`.",
	}
}
