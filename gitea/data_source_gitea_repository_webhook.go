package gitea

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGiteaRepositoryWebhook() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGiteaRepositoryWebhookRead,

		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Owner of the repository",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Repository name",
			},
			"id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Webhook ID",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Webhook type, e.g. `gitea`, `slack`, etc.",
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Target URL of the webhook",
			},
			"content_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Payload content type (`json` or `form`)",
			},
			"secret": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Webhook secret",
			},
			"authorization_header": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Webhook authorization header",
			},
			"http_method": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "HTTP method used for the webhook",
			},
			"events": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of events that trigger the webhook",
			},
			"branch_filter": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Branch filter on the webhook",
			},
			"active": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the webhook is active",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Webhook creation timestamp",
			},
			"channel": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Channel name for Slack webhooks",
			},
			"slack_username": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Bot username for Slack webhooks",
			},
			"icon_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Icon URL for Slack or Discord webhooks",
			},
			"color": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Hex color code for Slack webhooks",
			},
			"config": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Key-value configuration map for the webhook",
			},
		},
		Description: "Fetches details of a specific repository webhook.",
	}
}

func dataSourceGiteaRepositoryWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*GiteaClient).Client

	owner := strings.ToLower(d.Get("username").(string))
	repo := strings.ToLower(d.Get("name").(string))
	id := int64(d.Get("id").(int))

	hook, resp, err := client.GetRepoHook(owner, repo, id)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return diag.FromErr(fmt.Errorf("webhook with id %d not found for repo %s/%s", id, owner, repo))
		}
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(hook.ID, 10))
	if err := d.Set("username", owner); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", repo); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", hook.Type); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("url", hookConfigValue(hook, "url")); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("content_type", hookConfigValue(hook, "content_type")); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("secret", hookConfigValue(hook, "secret")); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("authorization_header", hook.AuthorizationHeader); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("events", stringSliceToInterfaceSlice(hook.Events)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("branch_filter", hook.BranchFilter); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("active", hook.Active); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", hook.Created.Format("2006-01-02T15:04:05Z07:00")); err != nil {
		return diag.FromErr(err)
	}

	if v := hookConfigValue(hook, "http_method"); v != "" {
		if err := d.Set("http_method", v); err != nil {
			return diag.FromErr(err)
		}
	}
	if v := hookConfigValue(hook, "channel"); v != "" {
		if err := d.Set("channel", v); err != nil {
			return diag.FromErr(err)
		}
	}
	if v := hookConfigValue(hook, "username"); v != "" {
		if err := d.Set("slack_username", v); err != nil {
			return diag.FromErr(err)
		}
	}
	if v := hookConfigValue(hook, "icon_url"); v != "" {
		if err := d.Set("icon_url", v); err != nil {
			return diag.FromErr(err)
		}
	}
	if v := hookConfigValue(hook, "color"); v != "" {
		if err := d.Set("color", v); err != nil {
			return diag.FromErr(err)
		}
	}
	if hook.Config != nil {
		if err := d.Set("config", hook.Config); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
