package gitea

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Gitea 28 introduced repository-scoped deploy tokens that authenticate git
// over HTTPS (as opposed to SSH deploy keys). They are created through a
// dedicated endpoint but share the same underlying deploy-key storage, so
// they are read and deleted through the existing /keys/{id} endpoints.
// The code.gitea.io/sdk/gitea package has no bindings for any of this yet
// (its DeployKey struct predates the key_type/token fields), so this
// resource talks to the API directly via GiteaClient.rawJSON.

const (
	deployTokenUsername string = "username"
	deployTokenRepo     string = "name"
	deployTokenTitle    string = "title"
	deployTokenReadOnly string = "read_only"
	deployTokenToken    string = "token"
)

type deployTokenCreateOption struct {
	Title    string `json:"title"`
	ReadOnly bool   `json:"read_only"`
}

type deployTokenResponse struct {
	ID       int64  `json:"id"`
	KeyType  string `json:"key_type"`
	Title    string `json:"title"`
	ReadOnly bool   `json:"read_only"`
	Token    string `json:"token"`
}

func deployTokenIDParts(d *schema.ResourceData) (owner, repo string, id int64, ok bool, err error) {
	parts := strings.SplitN(d.Id(), "/", 3)
	if len(parts) != 3 {
		return "", "", 0, false, nil
	}
	id, err = strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return "", "", 0, false, err
	}
	return parts[0], parts[1], id, true, nil
}

func resourceRepositoryDeployTokenCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*GiteaClient)

	owner := d.Get(deployTokenUsername).(string)
	repo := d.Get(deployTokenRepo).(string)

	opt := deployTokenCreateOption{
		Title:    d.Get(deployTokenTitle).(string),
		ReadOnly: d.Get(deployTokenReadOnly).(bool),
	}

	var dt deployTokenResponse
	if err := client.rawJSON("POST", fmt.Sprintf("/repos/%s/%s/keys/tokens", owner, repo), opt, &dt); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s/%s/%d", owner, repo, dt.ID))
	if err := d.Set(deployTokenTitle, dt.Title); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(deployTokenReadOnly, dt.ReadOnly); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(deployTokenToken, dt.Token); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceRepositoryDeployTokenRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*GiteaClient)

	owner, repo, id, ok, err := deployTokenIDParts(d)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ok {
		d.SetId("")
		return nil
	}

	var dt deployTokenResponse
	if err := client.rawJSON("GET", fmt.Sprintf("/repos/%s/%s/keys/%d", owner, repo, id), nil, &dt); err != nil {
		if strings.Contains(err.Error(), "status 404") {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if err := d.Set(deployTokenUsername, owner); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(deployTokenRepo, repo); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(deployTokenTitle, dt.Title); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(deployTokenReadOnly, dt.ReadOnly); err != nil {
		return diag.FromErr(err)
	}
	// The plaintext token is only ever returned by the create call; it is
	// intentionally left untouched here so it is not wiped from state on refresh.
	return nil
}

func resourceRepositoryDeployTokenDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*GiteaClient)

	owner, repo, id, ok, err := deployTokenIDParts(d)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ok {
		return nil
	}

	if err := client.rawJSON("DELETE", fmt.Sprintf("/repos/%s/%s/keys/%d", owner, repo, id), nil, nil); err != nil {
		if !strings.Contains(err.Error(), "status 404") {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourceRepositoryDeployTokenImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	owner, repo, _, ok, err := deployTokenIDParts(d)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("unexpected ID format (%q), expected <username>/<repo>/<id>", d.Id())
	}
	if err := d.Set(deployTokenUsername, owner); err != nil {
		return nil, err
	}
	if err := d.Set(deployTokenRepo, repo); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func resourceGiteaRepositoryDeployToken() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRepositoryDeployTokenCreate,
		ReadContext:   resourceRepositoryDeployTokenRead,
		DeleteContext: resourceRepositoryDeployTokenDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRepositoryDeployTokenImport,
		},
		Schema: map[string]*schema.Schema{
			deployTokenUsername: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Owner (user or organization) of the repository.",
			},
			deployTokenRepo: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Repository name.",
			},
			deployTokenTitle: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the deploy token.",
			},
			deployTokenReadOnly: {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     true,
				Description: "Whether this token only allows read access (true) or read/write access (false).",
			},
			deployTokenToken: {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The plaintext token used to authenticate git over HTTPS. Only known at creation time; Gitea never returns it again, so it is not refreshed on subsequent reads.",
			},
		},
		Description: "`gitea_repository_deploy_token` manages an HTTPS deploy token for a repository (Gitea >= 28.0.0). " +
			"Unlike SSH deploy keys (`gitea_repository_key`), deploy tokens authenticate git and LFS operations over HTTPS. " +
			"The token value is only available in state right after creation; it cannot be retrieved again from Gitea, " +
			"so losing it (e.g. by tainting state) requires replacing the resource to get a new one.",
	}
}
