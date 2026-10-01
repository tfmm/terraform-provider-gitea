package gitea

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGiteaRepositoryActionsVariable() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGiteaRepositoryActionsVariableCreate,
		ReadContext:   resourceGiteaRepositoryActionsVariableRead,
		UpdateContext: resourceGiteaRepositoryActionsVariableUpdate,
		DeleteContext: resourceGiteaRepositoryActionsVariableDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"repository_owner": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Owner of the repository.",
			},
			"repository": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the repository.",
			},
			"variable_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the variable.",
			},
			"value": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Value of the variable.",
			},
		},
	}
}

func resourceGiteaRepositoryActionsVariableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*GiteaClient).Client

	repoOwnerData, usernameOk := d.GetOk("repository_owner")
	if !usernameOk {
		return diag.FromErr(fmt.Errorf("name of repo owner must be passed"))
	}
	repoOwner := strings.ToLower(repoOwnerData.(string))

	nameData, nameOk := d.GetOk("repository")
	if !nameOk {
		return diag.FromErr(fmt.Errorf("CREATE name of repo must be passed"))
	}
	name := strings.ToLower(nameData.(string))

	variableNameData, nameOk := d.GetOk("variable_name")
	if !nameOk {
		return diag.FromErr(fmt.Errorf("variable_name of repo must be passed"))
	}
	variableName := variableNameData.(string)

	valueData, nameOk := d.GetOk("value")
	if !nameOk {
		return diag.FromErr(fmt.Errorf("value must be passed"))
	}
	value := valueData.(string)

	_, err := client.CreateRepoActionVariable(repoOwner, name, variableName, value)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(buildThreePartID(repoOwner, name, variableName))

	return resourceGiteaRepositoryActionsVariableRead(ctx, d, meta)
}

func resourceGiteaRepositoryActionsVariableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*GiteaClient).Client

	repoOwnerData, usernameOk := d.GetOk("repository_owner")
	if !usernameOk {
		return diag.FromErr(fmt.Errorf("name of repo owner must be passed"))
	}
	repoOwner := strings.ToLower(repoOwnerData.(string))

	repositoryData, nameOk := d.GetOk("repository")
	if !nameOk {
		return diag.FromErr(fmt.Errorf("READ name of repo must be passed"))
	}
	repository := strings.ToLower(repositoryData.(string))

	variableNameData, nameOk := d.GetOk("variable_name")
	if !nameOk {
		return diag.FromErr(fmt.Errorf("READ variable_name of repo must be passed"))
	}
	variableName := variableNameData.(string)

	valueData, nameOk := d.GetOk("value")
	if !nameOk {
		return diag.FromErr(fmt.Errorf("value must be passed"))
	}
	value := valueData.(string)

	_, err := client.UpdateRepoActionVariable(repoOwner, repository, variableName, value)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceGiteaRepositoryActionsVariableRead(ctx, d, meta)
}

func resourceGiteaRepositoryActionsVariableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*GiteaClient).Client

	repoOwner, repository, variableName, err := parseThreePartID(d.Id(), "repository_owner", "repository", "variable_name")
	if err != nil {
		return diag.FromErr(err)
	}

	variable, resp, err := client.GetRepoActionVariable(repoOwner, repository, variableName)

	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		} else {
			return diag.FromErr(err)
		}
	}

	if err = d.Set("repository_owner", repoOwner); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("repository", repository); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("variable_name", variableName); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("value", variable.Value); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceGiteaRepositoryActionsVariableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*GiteaClient).Client

	repoOwnerData, usernameOk := d.GetOk("repository_owner")
	if !usernameOk {
		return diag.FromErr(fmt.Errorf("name of repo owner must be passed"))
	}
	repoOwner := strings.ToLower(repoOwnerData.(string))

	repositoryData, nameOk := d.GetOk("repository")
	if !nameOk {
		return diag.FromErr(fmt.Errorf("name of repo must be passed"))
	}
	repository := strings.ToLower(repositoryData.(string))

	variableNameData, nameOk := d.GetOk("variable_name")
	if !nameOk {
		return diag.FromErr(fmt.Errorf("variable_name must be passed"))
	}
	variableName := strings.ToLower(variableNameData.(string))

	_, err := client.DeleteRepoActionVariable(repoOwner, repository, variableName)

	return diag.FromErr(err)
}
