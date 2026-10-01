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

func dataSourceGiteaOrg() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGiteaOrgRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"full_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"avatar_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"website": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"visibility": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		Description: "Use this data source to retrieve details of a Gitea organization.",
	}
}

func dataSourceGiteaOrgRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*GiteaClient).Client

	var org *gitea.Organization
	var err error

	log.Printf("[INFO] Reading Gitea Org")

	nameData, nameOk := d.GetOk("name")

	if !nameOk {
		return diag.FromErr(fmt.Errorf("name of org must be passed"))
	}
	name := strings.ToLower(nameData.(string))

	org, _, err = client.GetOrg(name)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("id", org.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", org.UserName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("full_name", org.FullName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("avatar_url", org.AvatarURL); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("location", org.Location); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("website", org.Website); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", org.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("visibility", org.Visibility); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d", org.ID))

	return nil
}
