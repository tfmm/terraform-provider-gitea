package gitea

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	repoWebhookUsername            string = "username"
	repoWebhookName                string = "name"
	repoWebhookType                string = "type"
	repoWebhookUrl                 string = "url"
	repoWebhookContentType         string = "content_type"
	repoWebhookSecret              string = "secret"
	repoWebhookAuthorizationHeader string = "authorization_header"
	repoWebhookEvents              string = "events"
	repoWebhookBranchFilter        string = "branch_filter"
	repoWebhookActive              string = "active"
	repoWebhookCreatedAt           string = "created_at"
	repoWebhookChannel             string = "channel"
	repoWebhookSlackUsername       string = "slack_username"
	repoWebhookIconUrl             string = "icon_url"
	repoWebhookColor               string = "color"
	repoWebhookHttpMethod          string = "http_method"
	repoWebhookConfig              string = "config"
)

func resourceRepositoryWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error
	client := meta.(*GiteaClient).Client

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	user := d.Get(repoWebhookUsername).(string)
	repo := d.Get(repoWebhookName).(string)

	hook, resp, err := client.GetRepoHook(user, repo, id)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return diag.FromErr(err)
		} else {
			return diag.FromErr(err)
		}
	}

	err = setRepositoryWebhookData(hook, d)

	return diag.FromErr(err)
}

func buildWebhookConfigMap(d *schema.ResourceData) map[string]string {
	config := make(map[string]string)

	if rawConfig, ok := d.GetOk(repoWebhookConfig); ok {
		for k, v := range rawConfig.(map[string]interface{}) {
			if strVal, ok := v.(string); ok {
				config[k] = strVal
			}
		}
	}

	if v, ok := d.GetOk(repoWebhookUrl); ok && v.(string) != "" {
		config["url"] = v.(string)
	}
	if v, ok := d.GetOk(repoWebhookContentType); ok && v.(string) != "" {
		config["content_type"] = v.(string)
	} else if wType, ok := d.GetOk(repoWebhookType); ok {
		t := strings.ToLower(wType.(string))
		if (t == "gitea" || t == "gogs") && config["content_type"] == "" {
			config["content_type"] = "json"
		}
	}
	if v, ok := d.GetOk(repoWebhookSecret); ok && v.(string) != "" {
		config["secret"] = v.(string)
	}
	if v, ok := d.GetOk(repoWebhookHttpMethod); ok && v.(string) != "" {
		config["http_method"] = v.(string)
	}
	if v, ok := d.GetOk(repoWebhookChannel); ok && v.(string) != "" {
		config["channel"] = v.(string)
	}
	if v, ok := d.GetOk(repoWebhookSlackUsername); ok && v.(string) != "" {
		config["username"] = v.(string)
	}
	if v, ok := d.GetOk(repoWebhookIconUrl); ok && v.(string) != "" {
		config["icon_url"] = v.(string)
	}
	if v, ok := d.GetOk(repoWebhookColor); ok && v.(string) != "" {
		config["color"] = v.(string)
	}

	return config
}

func resourceRepositoryWebhookCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error
	client := meta.(*GiteaClient).Client

	user := d.Get(repoWebhookUsername).(string)
	repo := d.Get(repoWebhookName).(string)

	config := buildWebhookConfigMap(d)
	events := extractEvents(d)

	hookOption := gitea.CreateHookOption{
		Type:                gitea.HookType(d.Get(repoWebhookType).(string)),
		Config:              config,
		Events:              events,
		BranchFilter:        d.Get(repoWebhookBranchFilter).(string),
		Active:              d.Get(repoWebhookActive).(bool),
		AuthorizationHeader: d.Get(repoWebhookAuthorizationHeader).(string),
	}

	hook, _, err := client.CreateRepoHook(user, repo, hookOption)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setRepositoryWebhookData(hook, d)

	return diag.FromErr(err)
}

func resourceRepositoryWebhookUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error
	client := meta.(*GiteaClient).Client

	user := d.Get(repoWebhookUsername).(string)
	repo := d.Get(repoWebhookName).(string)
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	config := buildWebhookConfigMap(d)
	events := extractEvents(d)
	active := d.Get(repoWebhookActive).(bool)

	hookOption := gitea.EditHookOption{
		Config:              config,
		Events:              events,
		BranchFilter:        d.Get(repoWebhookBranchFilter).(string),
		Active:              &active,
		AuthorizationHeader: d.Get(repoWebhookAuthorizationHeader).(string),
	}

	_, err = client.EditRepoHook(user, repo, id, hookOption)
	if err != nil {
		return diag.FromErr(err)
	}

	hook, _, err := client.GetRepoHook(user, repo, id)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setRepositoryWebhookData(hook, d)

	return diag.FromErr(err)
}

func resourceRepositoryWebhookDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error
	client := meta.(*GiteaClient).Client

	user := d.Get(repoWebhookUsername).(string)
	repo := d.Get(repoWebhookName).(string)
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = client.DeleteRepoHook(user, repo, id)
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(err)
}

func extractEvents(d *schema.ResourceData) []string {
	eventsMap := make(map[string]bool)

	if raw, ok := d.GetOk(repoWebhookEvents); ok {
		switch v := raw.(type) {
		case *schema.Set:
			for _, element := range v.List() {
				if str, ok := element.(string); ok {
					eventsMap[str] = true
				}
			}
		case []interface{}:
			for _, element := range v {
				if str, ok := element.(string); ok {
					eventsMap[str] = true
				}
			}
		}
	}

	rawConfig := d.GetRawConfig()
	if rawConfig.IsKnown() && !rawConfig.IsNull() {
		if eventsVal, ok := rawConfig.AsValueMap()["events"]; ok && eventsVal.IsKnown() && !eventsVal.IsNull() {
			if eventsVal.Type().IsSetType() || eventsVal.Type().IsListType() || eventsVal.Type().IsTupleType() {
				for _, elem := range eventsVal.AsValueSlice() {
					if elem.IsKnown() && !elem.IsNull() && elem.Type() == cty.String {
						eventsMap[elem.AsString()] = true
					}
				}
			}
		}
	}

	events := make([]string, 0, len(eventsMap))
	for str := range eventsMap {
		events = append(events, str)
	}
	return events
}

func setRepositoryWebhookData(hook *gitea.Hook, d *schema.ResourceData) (err error) {
	d.SetId(strconv.FormatInt(hook.ID, 10))
	if err := d.Set(repoWebhookUsername, d.Get(repoWebhookUsername).(string)); err != nil {
		return err
	}
	if err := d.Set(repoWebhookName, d.Get(repoWebhookName).(string)); err != nil {
		return err
	}
	if err := d.Set(repoWebhookType, hook.Type); err != nil {
		return err
	}
	if err := d.Set(repoWebhookUrl, hookConfigValue(hook, "url")); err != nil {
		return err
	}
	if err := d.Set(repoWebhookContentType, hookConfigValue(hook, "content_type")); err != nil {
		return err
	}

	secret := hookConfigValue(hook, "secret")
	if secret == "" {
		secret = d.Get(repoWebhookSecret).(string)
	}
	if secret != "" {
		if err := d.Set(repoWebhookSecret, secret); err != nil {
			return err
		}
	}
	if err := d.Set(repoWebhookEvents, hook.Events); err != nil {
		return err
	}
	if err := d.Set(repoWebhookBranchFilter, hook.BranchFilter); err != nil {
		return err
	}
	if err := d.Set(repoWebhookActive, hook.Active); err != nil {
		return err
	}
	if err := d.Set(repoWebhookCreatedAt, hook.Created.Format("2006-01-02T15:04:05Z07:00")); err != nil {
		return err
	}
	if err := d.Set(repoWebhookAuthorizationHeader, hook.AuthorizationHeader); err != nil {
		return err
	}

	if v := hookConfigValue(hook, "http_method"); v != "" {
		if err := d.Set(repoWebhookHttpMethod, v); err != nil {
			return err
		}
	}
	if v := hookConfigValue(hook, "channel"); v != "" {
		if err := d.Set(repoWebhookChannel, v); err != nil {
			return err
		}
	}
	if v := hookConfigValue(hook, "username"); v != "" {
		if err := d.Set(repoWebhookSlackUsername, v); err != nil {
			return err
		}
	}
	if v := hookConfigValue(hook, "icon_url"); v != "" {
		if err := d.Set(repoWebhookIconUrl, v); err != nil {
			return err
		}
	}
	if v := hookConfigValue(hook, "color"); v != "" {
		if err := d.Set(repoWebhookColor, v); err != nil {
			return err
		}
	}

	if isConfigConfiguredInHCL(d) {
		rawConfig := d.Get(repoWebhookConfig)
		userConfig := rawConfig.(map[string]interface{})
		newConfigMap := make(map[string]string)

		if hook.Config != nil {
			for k, v := range hook.Config {
				if v != "" {
					newConfigMap[k] = v
				} else if existingVal, exists := userConfig[k]; exists {
					if strVal, isStr := existingVal.(string); isStr && strVal != "" {
						newConfigMap[k] = strVal
					}
				}
			}
		}

		for k, v := range userConfig {
			if _, exists := newConfigMap[k]; !exists {
				if strVal, isStr := v.(string); isStr && strVal != "" {
					newConfigMap[k] = strVal
				}
			}
		}
		if err := d.Set(repoWebhookConfig, newConfigMap); err != nil {
			return err
		}
	} else {
		if err := d.Set(repoWebhookConfig, nil); err != nil {
			return err
		}
	}

	return
}

func isConfigConfiguredInHCL(d *schema.ResourceData) bool {
	rawConfig := d.GetRawConfig()
	if rawConfig.IsKnown() && !rawConfig.IsNull() {
		configVal, ok := rawConfig.AsValueMap()["config"]
		if ok && configVal.IsKnown() && !configVal.IsNull() {
			return configVal.LengthInt() > 0
		}
		return false
	}

	// Fallback when rawConfig is NullVal (during Read or unit tests):
	// If config map in state contains keys that are NOT top-level schema attributes,
	// then config was explicitly configured in HCL.
	if rawConfigMap, ok := d.GetOk(repoWebhookConfig); ok {
		if m, isMap := rawConfigMap.(map[string]interface{}); isMap && len(m) > 0 {
			topLevelKeys := map[string]bool{
				"url":          true,
				"content_type": true,
				"secret":       true,
				"http_method":  true,
				"channel":      true,
				"username":     true,
				"icon_url":     true,
				"color":        true,
			}
			for k := range m {
				if !topLevelKeys[k] {
					return true
				}
			}
		}
	}
	return false
}

func hookConfigValue(hook *gitea.Hook, key string) string {
	if hook == nil || hook.Config == nil {
		return ""
	}
	return hook.Config[key]
}

func stringSliceToInterfaceSlice(values []string) []interface{} {
	result := make([]interface{}, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	return result
}

func resourceGiteaRepositoryWebhook() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceRepositoryWebhookRead,
		CreateContext: resourceRepositoryWebhookCreate,
		UpdateContext: resourceRepositoryWebhookUpdate,
		DeleteContext: resourceRepositoryWebhookDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				parts := strings.Split(d.Id(), "/")
				if len(parts) != 3 {
					return nil, fmt.Errorf("unexpected ID format (%q), expected <username>/<repo>/<webhook_id>", d.Id())
				}
				if err := d.Set("username", parts[0]); err != nil {
					return nil, err
				}
				if err := d.Set("name", parts[1]); err != nil {
					return nil, err
				}
				d.SetId(parts[2])
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "User name or organization name",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Repository name",
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
				Type:             schema.TypeMap,
				Optional:         true,
				Elem:             &schema.Schema{Type: schema.TypeString},
				DiffSuppressFunc: webhookConfigDiffSuppressFunc,
				Description:      "Additional key-value configuration options for webhooks",
			},
		},
		Description: "This resource allows you to create and manage webhooks for repositories.",
	}
}

func webhookConfigDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	key := strings.TrimPrefix(k, "config.")

	if old != "" && new == "" {
		topLevelKeys := map[string]string{
			"url":          repoWebhookUrl,
			"content_type": repoWebhookContentType,
			"secret":       repoWebhookSecret,
			"http_method":  repoWebhookHttpMethod,
			"channel":      repoWebhookChannel,
			"username":     repoWebhookSlackUsername,
			"icon_url":     repoWebhookIconUrl,
			"color":        repoWebhookColor,
		}
		if topField, isTop := topLevelKeys[key]; isTop {
			if v, ok := d.GetOk(topField); ok && v.(string) != "" {
				return true
			}
		}
	}
	return false
}
