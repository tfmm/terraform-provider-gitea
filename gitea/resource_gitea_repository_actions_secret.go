package gitea

import (
	"context"
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGiteaRepositoryActionsSecret() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGiteaRepositoryActionsSecretCreate,
		ReadContext:   resourceGiteaRepositoryActionsSecretRead,
		UpdateContext: resourceGiteaRepositoryActionsSecretUpdate,
		DeleteContext: resourceGiteaRepositoryActionsSecretDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: mergeSchemaMaps(map[string]*schema.Schema{
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
			"secret_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the secret.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date of 'actions_secret' creation.",
			},
		}, writeOnlySecretValueSchema("Value of the secret.")),
		Description: "`gitea_repository_actions_secret` manages a repository actions secret.\n\n" +
			"Import expects the resource ID in the form `owner:repository:secret_name`.\n" +
			"Because Gitea does not return secret values, `secret_value` or `secret_value_wo` must still be configured when importing.\n\n" +
			"WARNING:\n" +
			"`secret_value` will be stored in the terraform state! Use `secret_value_wo` instead to avoid that.",
	}
}

func resourceGiteaRepositoryActionsSecretCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)

	repoOwnerData, usernameOk := d.GetOk("repository_owner")
	if !usernameOk {
		return diag.FromErr(fmt.Errorf("name of repo owner must be passed"))
	}
	repoOwner := strings.ToLower(repoOwnerData.(string))

	repositoryData, nameOk := d.GetOk("repository")
	if !nameOk {
		return diag.FromErr(fmt.Errorf("CREATE name of repo must be passed"))
	}
	repository := strings.ToLower(repositoryData.(string))

	secretNameData, nameOk := d.GetOk("secret_name")
	if !nameOk {
		return diag.FromErr(fmt.Errorf("secret_name must be passed"))
	}
	secretName := secretNameData.(string)

	value, err := resolveSecretValue(d)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = client.CreateRepoActionSecret(repoOwner, repository, secretName, gitea.CreateOrUpdateSecretOption{
		Data: value,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(buildThreePartID(repoOwner, repository, secretName))

	return resourceGiteaRepositoryActionsSecretRead(ctx, d, meta)
}

func resourceGiteaRepositoryActionsSecretUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)

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

	variableNameData, nameOk := d.GetOk("secret_name")
	if !nameOk {
		return diag.FromErr(fmt.Errorf("secret_name of repo must be passed"))
	}
	variableName := variableNameData.(string)

	value, err := resolveSecretValue(d)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = client.CreateRepoActionSecret(repoOwner, repository, variableName, gitea.CreateOrUpdateSecretOption{
		Data: value,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceGiteaRepositoryActionsSecretRead(ctx, d, meta)
}

func resourceGiteaRepositoryActionsSecretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)

	repoOwner, repository, secretName, err := parseThreePartID(d.Id(), "repository_owner", "repository", "secret_name")
	if err != nil {
		return diag.FromErr(err)
	}

	var notFound bool
	secrets, err := collectPaginated(func(page int) ([]*gitea.Secret, error) {
		items, resp, callErr := client.ListRepoActionSecret(repoOwner, repository, gitea.ListRepoActionSecretOption{
			ListOptions: gitea.ListOptions{
				Page:     page,
				PageSize: 100,
			},
		})
		if callErr != nil && resp != nil && resp.StatusCode == 404 {
			notFound = true
		}
		return items, callErr
	})
	if notFound {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	var requestedSecret *gitea.Secret
	for _, secret := range secrets {
		if secret != nil && secret.Name == secretName {
			requestedSecret = secret
			break
		}
	}

	createdAtData, dateOk := d.GetOk("created_at")

	if requestedSecret == nil {
		d.SetId("")
		return nil
	}

	if dateOk {
		if requestedSecret.Created.String() != createdAtData.(string) {
			d.SetId("")
			return nil
		}
	}

	if err := d.Set("repository_owner", repoOwner); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("repository", repository); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("secret_name", secretName); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("created_at", requestedSecret.Created.String()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceGiteaRepositoryActionsSecretDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)

	repoOwner, repository, secretName, _ := parseThreePartID(d.Id(), "repository_owner", "repository", "secret_name")

	_, err := client.DeleteRepoActionSecret(repoOwner, repository, secretName)

	return diag.FromErr(err)
}
