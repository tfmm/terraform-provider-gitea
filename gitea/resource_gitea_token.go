package gitea

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	TokenName      string = "name"
	TokenHash      string = "token"
	TokenLastEight string = "last_eight"
	TokenScopes    string = "scopes"
)

var errTokenNotFound = errors.New("token not found")

// validScopes contains the valid scopes for tokens as listed
// at https://docs.gitea.com/development/oauth2-provider#scopes
var validScopes = map[string]bool{
	"all":                true,
	"read:activitypub":   true,
	"write:activitypub":  true,
	"read:admin":         true,
	"write:admin":        true,
	"read:issue":         true,
	"write:issue":        true,
	"read:misc":          true,
	"write:misc":         true,
	"read:notification":  true,
	"write:notification": true,
	"read:organization":  true,
	"write:organization": true,
	"read:package":       true,
	"write:package":      true,
	"read:repository":    true,
	"write:repository":   true,
	"read:user":          true,
	"write:user":         true,
}

func searchTokenById(c *gitea.Client, id int64) (res *gitea.AccessToken, err error) {
	page := 1

	for {
		tokens, _, err := c.ListAccessTokens(gitea.ListAccessTokensOptions{
			ListOptions: gitea.ListOptions{
				Page:     page,
				PageSize: 50,
			},
		})
		if err != nil {
			return nil, err
		}

		if len(tokens) == 0 {
			return nil, errTokenNotFound
		}

		for _, token := range tokens {
			if token.ID == id {
				return token, nil
			}
		}

		page += 1
	}
}

func resourceTokenCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error

	client := meta.(*GiteaClient).Client

	// Create a list of valid scopes. Thrown an error if an invalid scope is found
	var scopes []gitea.AccessTokenScope
	for _, s := range d.Get(TokenScopes).(*schema.Set).List() {
		s := s.(string)
		if validScopes[s] {
			scopes = append(scopes, gitea.AccessTokenScope(s))
		} else {
			return diag.FromErr(fmt.Errorf("Invalid token scope: '%s'", s))
		}
	}

	opts := gitea.CreateAccessTokenOption{
		Name:   d.Get(TokenName).(string),
		Scopes: scopes,
	}

	token, _, err := client.CreateAccessToken(opts)

	if err != nil {
		return diag.FromErr(err)
	}

	err = setTokenResourceData(token, d)

	return diag.FromErr(err)
}

func resourceTokenRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error

	client := meta.(*GiteaClient).Client

	var token *gitea.AccessToken

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	token, err = searchTokenById(client, id)

	if err != nil {
		if errors.Is(err, errTokenNotFound) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	err = setTokenResourceData(token, d)

	return diag.FromErr(err)
}

func resourceTokenDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error

	client := meta.(*GiteaClient).Client
	var resp *gitea.Response

	if id, parseErr := strconv.ParseInt(d.Id(), 10, 64); parseErr == nil {
		resp, err = client.DeleteAccessToken(id)
	} else {
		resp, err = client.DeleteAccessToken(d.Get(TokenName).(string))
	}

	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return diag.FromErr(err)
		} else {
			return diag.FromErr(err)
		}
	}

	return diag.FromErr(err)
}

func setTokenResourceData(token *gitea.AccessToken, d *schema.ResourceData) (err error) {

	d.SetId(fmt.Sprintf("%d", token.ID))
	if err := d.Set(TokenName, token.Name); err != nil {
		return err
	}
	if token.Token != "" {
		if err := d.Set(TokenHash, token.Token); err != nil {
			return err
		}
	}
	if err := d.Set(TokenLastEight, token.TokenLastEight); err != nil {
		return err
	}
	if err := d.Set(TokenScopes, token.Scopes); err != nil {
		return err
	}

	return
}

func resourceGiteaToken() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceTokenRead,
		CreateContext: resourceTokenCreate,
		DeleteContext: resourceTokenDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the Access Token",
			},
			"token": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The actual Access Token",
			},
			"last_eight": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"scopes": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required:    true,
				ForceNew:    true,
				Description: "List of string representations of scopes for the token",
			},
		},
		Description: "`gitea_token` manages gitea Access Tokens.\n\n" +
			"Due to upstream limitations (see https://gitea.com/gitea/go-sdk/issues/610) this resource\n" +
			"can only be used with username/password provider configuration.\n\n" +
			"WARNING:\n" +
			"Tokens will be stored in the terraform state!",
	}
}
