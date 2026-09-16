package gitea

import (
	"context"
	"fmt"
	"log"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGiteaUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGiteaUserRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"username": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"email": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"full_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_admin": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"avatar_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"language": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_login": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		Description: "Use this data source to retrieve details of a Gitea user.",
	}
}

func dataSourceGiteaUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)

	var user *gitea.User
	var err error

	log.Printf("[INFO] Reading Gitea user")

	usernameData, usernameOk := d.GetOk("username")

	if !usernameOk {
		user, _, err = client.GetMyUserInfo()
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		username := strings.ToLower(usernameData.(string))

		user, _, err = client.GetUserInfo(username)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("id", user.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("username", user.UserName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("email", user.Email); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("full_name", user.FullName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_admin", user.IsAdmin); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created", user.Created); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("avatar_url", user.AvatarURL); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("last_login", user.LastLogin); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("language", user.Language); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d", user.ID))

	return nil
}
