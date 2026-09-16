package gitea

import (
	"context"
	"net/http"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGiteaOrgActionsVariable() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGiteaOrgActionsVariableCreate,
		ReadContext:   resourceGiteaOrgActionsVariableRead,
		UpdateContext: resourceGiteaOrgActionsVariableUpdate,
		DeleteContext: resourceGiteaOrgActionsVariableDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			actionOrgField: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The organisation owning the Actions variable.",
			},
			"variable_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The Actions variable name.",
			},
			"value": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Actions variable value.",
			},
			descriptionField: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Actions variable description.",
			},
		},
		Description: "`gitea_org_actions_variable` manages an organisation-scoped Actions variable.\n\n" +
			"Import expects the resource ID in the form `org:variable_name`.",
	}
}

func resourceGiteaOrgActionsVariableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	org := strings.ToLower(d.Get(actionOrgField).(string))
	name := d.Get("variable_name").(string)
	_, err := client.CreateOrgActionVariable(org, name, gitea.CreateActionVariableOption{
		Value:       d.Get("value").(string),
		Description: d.Get(descriptionField).(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(buildTwoPartID(org, name))
	return resourceGiteaOrgActionsVariableRead(ctx, d, meta)
}

func resourceGiteaOrgActionsVariableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	org, name, err := parseTwoPartID(d.Id(), actionOrgField, "variable_name")
	if err != nil {
		return diag.FromErr(err)
	}
	variable, resp, err := client.GetOrgActionVariable(org, name)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	d.Set(actionOrgField, org)
	d.Set("variable_name", variable.Name)
	d.Set("value", variable.Data)
	d.Set(descriptionField, variable.Description)
	return nil
}

func resourceGiteaOrgActionsVariableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	org := strings.ToLower(d.Get(actionOrgField).(string))
	name := d.Get("variable_name").(string)
	_, err := client.UpdateOrgActionVariable(org, name, gitea.UpdateActionVariableOption{
		Name:        name,
		Value:       d.Get("value").(string),
		Description: d.Get(descriptionField).(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceGiteaOrgActionsVariableRead(ctx, d, meta)
}

func resourceGiteaOrgActionsVariableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	org, name, err := parseTwoPartID(d.Id(), actionOrgField, "variable_name")
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = client.DeleteOrgActionVariable(org, name)
	return diag.FromErr(err)
}

func resourceGiteaOrgActionsSecret() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGiteaOrgActionsSecretCreate,
		ReadContext:   resourceGiteaOrgActionsSecretRead,
		UpdateContext: resourceGiteaOrgActionsSecretUpdate,
		DeleteContext: resourceGiteaOrgActionsSecretDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			actionOrgField: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The organisation owning the Actions secret.",
			},
			"secret_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The Actions secret name.",
			},
			"secret_value": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The Actions secret value.",
			},
			descriptionField: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Actions secret description.",
			},
			createdAtField: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Actions secret creation timestamp.",
			},
		},
		Description: "`gitea_org_actions_secret` manages an organisation-scoped Actions secret.\n\n" +
			"Import expects the resource ID in the form `org:secret_name`.\n" +
			"Because Gitea does not return secret values, `secret_value` must still be configured when importing.\n\n" +
			"WARNING:\n" +
			"`secret_value` will be stored in the terraform state!",
	}
}

func resourceGiteaOrgActionsSecretCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	org := strings.ToLower(d.Get(actionOrgField).(string))
	name := d.Get("secret_name").(string)
	_, err := client.CreateOrgActionSecret(org, name, gitea.CreateOrUpdateSecretOption{
		Data:        d.Get("secret_value").(string),
		Description: d.Get(descriptionField).(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(buildTwoPartID(org, name))
	return resourceGiteaOrgActionsSecretRead(ctx, d, meta)
}

func resourceGiteaOrgActionsSecretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	org, secretName, err := parseTwoPartID(d.Id(), actionOrgField, "secret_name")
	if err != nil {
		return diag.FromErr(err)
	}
	secrets, err := collectPaginated(func(page int) ([]*gitea.Secret, error) {
		items, _, callErr := client.ListOrgActionSecret(org, gitea.ListOrgActionSecretOption{
			ListOptions: gitea.ListOptions{Page: page, PageSize: 100},
		})
		return items, callErr
	})
	if err != nil {
		return diag.FromErr(err)
	}
	var secret *gitea.Secret
	for _, item := range secrets {
		if item != nil && item.Name == secretName {
			secret = item
			break
		}
	}
	if secret == nil {
		d.SetId("")
		return nil
	}
	d.Set(actionOrgField, org)
	d.Set("secret_name", secret.Name)
	d.Set(descriptionField, secret.Description)
	d.Set(createdAtField, timeToString(secret.Created))
	return nil
}

func resourceGiteaOrgActionsSecretUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	org := strings.ToLower(d.Get(actionOrgField).(string))
	name := d.Get("secret_name").(string)
	_, err := client.CreateOrgActionSecret(org, name, gitea.CreateOrUpdateSecretOption{
		Data:        d.Get("secret_value").(string),
		Description: d.Get(descriptionField).(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceGiteaOrgActionsSecretRead(ctx, d, meta)
}

func resourceGiteaOrgActionsSecretDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	org, name, err := parseTwoPartID(d.Id(), actionOrgField, "secret_name")
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = client.DeleteOrgActionSecret(org, name)
	return diag.FromErr(err)
}

func resourceGiteaUserActionsVariable() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGiteaUserActionsVariableCreate,
		ReadContext:   resourceGiteaUserActionsVariableRead,
		UpdateContext: resourceGiteaUserActionsVariableUpdate,
		DeleteContext: resourceGiteaUserActionsVariableDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"variable_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The user-scoped Actions variable name.",
			},
			"value": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The user-scoped Actions variable value.",
			},
			descriptionField: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The user-scoped Actions variable description.",
			},
		},
		Description: "`gitea_user_actions_variable` manages a user-scoped Actions variable.\n\n" +
			"Import expects the resource ID in the form `variable_name`.",
	}
}

func resourceGiteaUserActionsVariableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	name := d.Get("variable_name").(string)
	_, err := client.CreateUserActionVariable(name, gitea.CreateActionVariableOption{
		Value:       d.Get("value").(string),
		Description: d.Get(descriptionField).(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(name)
	return resourceGiteaUserActionsVariableRead(ctx, d, meta)
}

func resourceGiteaUserActionsVariableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	variable, resp, err := client.GetUserActionVariable(d.Id())
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	d.Set("variable_name", variable.Name)
	d.Set("value", variable.Data)
	d.Set(descriptionField, variable.Description)
	return nil
}

func resourceGiteaUserActionsVariableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	name := d.Get("variable_name").(string)
	_, err := client.UpdateUserActionVariable(name, gitea.UpdateActionVariableOption{
		Name:        name,
		Value:       d.Get("value").(string),
		Description: d.Get(descriptionField).(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceGiteaUserActionsVariableRead(ctx, d, meta)
}

func resourceGiteaUserActionsVariableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	_, err := client.DeleteUserActionVariable(d.Id())
	return diag.FromErr(err)
}

func resourceGiteaUserActionsSecret() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGiteaUserActionsSecretCreate,
		ReadContext:   resourceGiteaUserActionsSecretRead,
		UpdateContext: resourceGiteaUserActionsSecretUpdate,
		DeleteContext: resourceGiteaUserActionsSecretDelete,
		Schema: map[string]*schema.Schema{
			"secret_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The user-scoped Actions secret name.",
			},
			"secret_value": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The user-scoped Actions secret value.",
			},
			descriptionField: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The user-scoped Actions secret description.",
			},
		},
		Description: "`gitea_user_actions_secret` manages a user-scoped Actions secret.\n\n" +
			"This resource is write-only because the Gitea API does not expose a read/list endpoint for user-scoped Actions secrets.\n" +
			"Import is intentionally unsupported.\n\n" +
			"WARNING:\n" +
			"`secret_value` will be stored in the terraform state!",
	}
}

func resourceGiteaUserActionsSecretCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	name := d.Get("secret_name").(string)
	_, err := client.CreateUserActionSecret(name, gitea.CreateOrUpdateSecretOption{
		Data:        d.Get("secret_value").(string),
		Description: d.Get(descriptionField).(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(name)
	return resourceGiteaUserActionsSecretRead(ctx, d, meta)
}

func resourceGiteaUserActionsSecretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.Set("secret_name", d.Get("secret_name").(string))
	d.Set(descriptionField, d.Get(descriptionField).(string))
	return nil
}

func resourceGiteaUserActionsSecretUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	name := d.Get("secret_name").(string)
	_, err := client.CreateUserActionSecret(name, gitea.CreateOrUpdateSecretOption{
		Data:        d.Get("secret_value").(string),
		Description: d.Get(descriptionField).(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceGiteaUserActionsSecretRead(ctx, d, meta)
}

func resourceGiteaUserActionsSecretDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	_, err := client.DeleteUserActionSecret(d.Id())
	return diag.FromErr(err)
}
