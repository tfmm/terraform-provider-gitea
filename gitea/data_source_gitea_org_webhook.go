package gitea

import (
	"fmt"
	"strconv"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGiteaOrgWebhook() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaOrgWebhookRead,
		Schema: map[string]*schema.Schema{
			"org": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Organization name",
			},
			"id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Organization webhook ID",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Webhook type",
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Target URL of the webhook",
			},
			"content_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The content type of the payload",
			},
			"active": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the webhook is active",
			},
			"events": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Events triggering the webhook",
			},
		},
		Description: "Use this data source to retrieve details of an organization webhook.",
	}
}

func dataSourceGiteaOrgWebhookRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	org := d.Get("org").(string)
	id := int64(d.Get("id").(int))

	hook, _, err := client.GetOrgHook(org, id)
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(hook.ID, 10))
	d.Set("type", hook.Type)
	d.Set("url", hookConfigValue(hook, "url"))
	d.Set("content_type", hookConfigValue(hook, "content_type"))
	d.Set("active", hook.Active)
	d.Set("events", hook.Events)

	return nil
}

func dataSourceGiteaOrgWebhooks() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaOrgWebhooksRead,
		Schema: map[string]*schema.Schema{
			"org": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Organization name",
			},
			"webhooks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"active": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
		Description: "Use this data source to list all webhooks for an organization.",
	}
}

func dataSourceGiteaOrgWebhooksRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	org := d.Get("org").(string)

	hooks, _, err := client.ListOrgHooks(org, gitea.ListHooksOptions{})
	if err != nil {
		return err
	}

	result := make([]map[string]interface{}, 0, len(hooks))
	for _, hook := range hooks {
		result = append(result, map[string]interface{}{
			"id":     hook.ID,
			"type":   hook.Type,
			"url":    hookConfigValue(hook, "url"),
			"active": hook.Active,
		})
	}

	d.SetId(fmt.Sprintf("org/%s/webhooks", org))
	d.Set("webhooks", result)

	return nil
}
