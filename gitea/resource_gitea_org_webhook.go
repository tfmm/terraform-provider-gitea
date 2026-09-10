package gitea

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	orgWebhookOrg                 string = "org"
	orgWebhookType                string = "type"
	orgWebhookUrl                 string = "url"
	orgWebhookContentType         string = "content_type"
	orgWebhookSecret              string = "secret"
	orgWebhookAuthorizationHeader string = "authorization_header"
	orgWebhookEvents              string = "events"
	orgWebhookBranchFilter        string = "branch_filter"
	orgWebhookActive              string = "active"
	orgWebhookCreatedAt           string = "created_at"
	orgWebhookChannel             string = "channel"
	orgWebhookSlackUsername       string = "slack_username"
	orgWebhookIconUrl             string = "icon_url"
	orgWebhookColor               string = "color"
	orgWebhookHttpMethod          string = "http_method"
	orgWebhookConfig              string = "config"
)

func resourceOrgWebhookRead(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}

	org := d.Get(orgWebhookOrg).(string)

	hook, resp, err := client.GetOrgHook(org, id)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return err
	}

	err = setOrgWebhookData(org, hook, d)
	return
}

func buildOrgWebhookConfigMap(d *schema.ResourceData) map[string]string {
	config := make(map[string]string)

	if rawConfig, ok := d.GetOk(orgWebhookConfig); ok {
		for k, v := range rawConfig.(map[string]interface{}) {
			if strVal, ok := v.(string); ok {
				config[k] = strVal
			}
		}
	}

	if v, ok := d.GetOk(orgWebhookUrl); ok && v.(string) != "" {
		config["url"] = v.(string)
	}
	if v, ok := d.GetOk(orgWebhookContentType); ok && v.(string) != "" {
		config["content_type"] = v.(string)
	} else if wType, ok := d.GetOk(orgWebhookType); ok {
		t := strings.ToLower(wType.(string))
		if (t == "gitea" || t == "gogs") && config["content_type"] == "" {
			config["content_type"] = "json"
		}
	}
	if v, ok := d.GetOk(orgWebhookSecret); ok && v.(string) != "" {
		config["secret"] = v.(string)
	}
	if v, ok := d.GetOk(orgWebhookHttpMethod); ok && v.(string) != "" {
		config["http_method"] = v.(string)
	}
	if v, ok := d.GetOk(orgWebhookChannel); ok && v.(string) != "" {
		config["channel"] = v.(string)
	}
	if v, ok := d.GetOk(orgWebhookSlackUsername); ok && v.(string) != "" {
		config["username"] = v.(string)
	}
	if v, ok := d.GetOk(orgWebhookIconUrl); ok && v.(string) != "" {
		config["icon_url"] = v.(string)
	}
	if v, ok := d.GetOk(orgWebhookColor); ok && v.(string) != "" {
		config["color"] = v.(string)
	}

	return config
}

func resourceOrgWebhookCreate(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	org := d.Get(orgWebhookOrg).(string)

	config := buildOrgWebhookConfigMap(d)
	events := extractEvents(d)

	hookOption := gitea.CreateHookOption{
		Type:                gitea.HookType(d.Get(orgWebhookType).(string)),
		Config:              config,
		Events:              events,
		BranchFilter:        d.Get(orgWebhookBranchFilter).(string),
		Active:              d.Get(orgWebhookActive).(bool),
		AuthorizationHeader: d.Get(orgWebhookAuthorizationHeader).(string),
	}

	hook, _, err := client.CreateOrgHook(org, hookOption)
	if err != nil {
		return err
	}

	err = setOrgWebhookData(org, hook, d)
	return
}

func resourceOrgWebhookUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	org := d.Get(orgWebhookOrg).(string)
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}

	config := buildOrgWebhookConfigMap(d)
	events := extractEvents(d)
	active := d.Get(orgWebhookActive).(bool)

	hookOption := gitea.EditHookOption{
		Config:              config,
		Events:              events,
		BranchFilter:        d.Get(orgWebhookBranchFilter).(string),
		Active:              &active,
		AuthorizationHeader: d.Get(orgWebhookAuthorizationHeader).(string),
	}

	_, err = client.EditOrgHook(org, id, hookOption)
	if err != nil {
		return err
	}

	hook, _, err := client.GetOrgHook(org, id)
	if err != nil {
		return err
	}

	err = setOrgWebhookData(org, hook, d)
	return
}

func resourceOrgWebhookDelete(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitea.Client)

	org := d.Get(orgWebhookOrg).(string)
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}

	_, err = client.DeleteOrgHook(org, id)
	return err
}

func setOrgWebhookData(org string, hook *gitea.Hook, d *schema.ResourceData) (err error) {
	d.SetId(strconv.FormatInt(hook.ID, 10))

	d.Set(orgWebhookOrg, org)
	d.Set(orgWebhookType, hook.Type)
	d.Set(orgWebhookUrl, hookConfigValue(hook, "url"))
	d.Set(orgWebhookContentType, hookConfigValue(hook, "content_type"))

	secret := hookConfigValue(hook, "secret")
	if secret == "" {
		secret = d.Get(orgWebhookSecret).(string)
	}
	if secret != "" {
		d.Set(orgWebhookSecret, secret)
	}

	d.Set(orgWebhookEvents, hook.Events)
	d.Set(orgWebhookBranchFilter, hook.BranchFilter)
	d.Set(orgWebhookActive, hook.Active)
	d.Set(orgWebhookCreatedAt, hook.Created.Format("2006-01-02T15:04:05Z07:00"))
	d.Set(orgWebhookAuthorizationHeader, hook.AuthorizationHeader)

	if v := hookConfigValue(hook, "http_method"); v != "" {
		d.Set(orgWebhookHttpMethod, v)
	}
	if v := hookConfigValue(hook, "channel"); v != "" {
		d.Set(orgWebhookChannel, v)
	}
	if v := hookConfigValue(hook, "username"); v != "" {
		d.Set(orgWebhookSlackUsername, v)
	}
	if v := hookConfigValue(hook, "icon_url"); v != "" {
		d.Set(orgWebhookIconUrl, v)
	}
	if v := hookConfigValue(hook, "color"); v != "" {
		d.Set(orgWebhookColor, v)
	}

	return nil
}

func resourceGiteaOrgWebhook() *schema.Resource {
	return &schema.Resource{
		Read:   resourceOrgWebhookRead,
		Create: resourceOrgWebhookCreate,
		Update: resourceOrgWebhookUpdate,
		Delete: resourceOrgWebhookDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				parts := strings.Split(d.Id(), "/")
				if len(parts) != 2 {
					return nil, fmt.Errorf("unexpected ID format (%q), expected <org>/<webhook_id>", d.Id())
				}
				d.Set("org", parts[0])
				d.SetId(parts[1])
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"org": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Organization name",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Webhook type, e.g. `gitea`, `gogs`, `slack`, `discord`, `dingtalk`, `msteams`, `telegram`, `feishu`, `matrix`, `wechatwork`, `packagist`",
			},
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Target URL of the webhook",
			},
			"content_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The content type of the payload. It can be `json`, or `form`",
			},
			"secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Webhook secret",
			},
			"authorization_header": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Webhook authorization header",
			},
			"events": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required:    true,
				Description: "A list of events that will trigger the webhook, e.g. `[\"push\"]`",
			},
			"branch_filter": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "*",
				Description: "Set branch filter on the webhook, e.g. `\"*\"`",
			},
			"active": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Set webhook to active, e.g. `true`",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Webhook creation timestamp",
			},
			"channel": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Channel name for Slack webhooks (e.g. `#general` or `@username`)",
			},
			"slack_username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Bot username for Slack webhooks",
			},
			"icon_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Icon URL for Slack or Discord webhooks",
			},
			"color": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Hex color code for Slack webhooks (e.g. `#ff0000`)",
			},
			"http_method": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "HTTP method used for the webhook",
			},
			"config": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Additional key-value configuration options for webhooks",
			},
		},
		Description: "This resource allows you to create and manage webhooks for organizations.",
	}
}
