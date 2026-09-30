package gitea

import (
	"context"
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	collabOwner      string = "owner"
	collabRepo       string = "repo"
	collabUsername   string = "username"
	collabPermission string = "permission"
)

func resourceRepositoryCollaboratorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error
	client := meta.(*GiteaClient).Client

	owner := d.Get(collabOwner).(string)
	repo := d.Get(collabRepo).(string)
	username := d.Get(collabUsername).(string)
	permission := gitea.AccessMode(d.Get(collabPermission).(string))

	_, err = client.AddCollaborator(owner, repo, username, gitea.AddCollaboratorOption{
		Permission: &permission,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = setRepositoryCollaboratorData(d, owner, repo, username, string(permission))

	return diag.FromErr(err)
}

func resourceRepositoryCollaboratorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error
	client := meta.(*GiteaClient).Client

	owner, repo, username, parseErr := parseCollaboratorID(d.Id())
	if parseErr != nil {
		return diag.FromErr(parseErr)
	}

	isCollab, resp, err := client.IsCollaborator(owner, repo, username)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	if !isCollab {
		d.SetId("")
		return nil
	}

	permResult, _, err := client.CollaboratorPermission(owner, repo, username)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setRepositoryCollaboratorData(d, owner, repo, username, string(permResult.Permission))

	return diag.FromErr(err)
}

func resourceRepositoryCollaboratorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error
	client := meta.(*GiteaClient).Client

	owner := d.Get(collabOwner).(string)
	repo := d.Get(collabRepo).(string)
	username := d.Get(collabUsername).(string)
	permission := gitea.AccessMode(d.Get(collabPermission).(string))

	_, err = client.AddCollaborator(owner, repo, username, gitea.AddCollaboratorOption{
		Permission: &permission,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = setRepositoryCollaboratorData(d, owner, repo, username, string(permission))

	return diag.FromErr(err)
}

func resourceRepositoryCollaboratorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error
	client := meta.(*GiteaClient).Client

	owner := d.Get(collabOwner).(string)
	repo := d.Get(collabRepo).(string)
	username := d.Get(collabUsername).(string)

	_, err = client.DeleteCollaborator(owner, repo, username)

	return diag.FromErr(err)
}

func parseCollaboratorID(id string) (owner, repo, username string, err error) {
	parts := strings.SplitN(id, "/", 3)
	if len(parts) != 3 {
		err = fmt.Errorf("invalid collaborator ID format: %s (expected owner/repo/username)", id)
		return
	}
	owner = parts[0]
	repo = parts[1]
	username = parts[2]
	return
}

func setRepositoryCollaboratorData(d *schema.ResourceData, owner, repo, username, permission string) (err error) {
	d.SetId(fmt.Sprintf("%s/%s/%s", owner, repo, username))
	if err := d.Set(collabOwner, owner); err != nil {
		return err
	}
	if err := d.Set(collabRepo, repo); err != nil {
		return err
	}
	if err := d.Set(collabUsername, username); err != nil {
		return err
	}
	if err := d.Set(collabPermission, permission); err != nil {
		return err
	}
	return
}

func resourceGiteaRepositoryCollaborator() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceRepositoryCollaboratorRead,
		CreateContext: resourceRepositoryCollaboratorCreate,
		UpdateContext: resourceRepositoryCollaboratorUpdate,
		DeleteContext: resourceRepositoryCollaboratorDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"owner": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The owner of the repository (user or organization).",
			},
			"repo": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the repository.",
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The username of the collaborator to add.",
			},
			"permission": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "read",
				Description: "The permission level for the collaborator: `read`, `write`, or `admin`.",
			},
		},
		Description: "`gitea_repository_collaborator` manages a single collaborator's access to a repository without requiring team creation.",
	}
}
