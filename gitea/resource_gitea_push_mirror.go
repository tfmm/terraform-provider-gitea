package gitea

import (
	"context"
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func parsePushMirrorID(id string) (owner, repo, remoteName string, err error) {
	var parts []string
	if strings.Contains(id, ":") {
		parts = strings.SplitN(id, ":", 3)
	} else if strings.Contains(id, "/") {
		parts = strings.SplitN(id, "/", 3)
	}
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("unexpected ID format (%q). Expected owner:repo:remote_name or owner/repo/remote_name", id)
	}
	return parts[0], parts[1], parts[2], nil
}

func resourceGiteaPushMirrorImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	owner, repo, remoteName, err := parsePushMirrorID(d.Id())
	if err != nil {
		return nil, err
	}
	d.SetId(buildThreePartID(owner, repo, remoteName))
	if err := d.Set("owner", owner); err != nil {
		return nil, err
	}
	if err := d.Set("repo", repo); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func resourceGiteaPushMirrorCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)

	owner := d.Get("owner").(string)
	repo := d.Get("repo").(string)

	opt := gitea.CreatePushMirrorOption{
		RemoteAddress:  d.Get("remote_address").(string),
		RemoteUsername: d.Get("remote_username").(string),
		RemotePassword: d.Get("remote_password").(string),
		Interval:       d.Get("interval").(string),
		SyncONCommit:   d.Get("sync_on_commit").(bool),
	}

	pm, _, err := client.PushMirrors(owner, repo, opt)
	if err != nil {
		return err
	}

	d.SetId(buildThreePartID(owner, repo, pm.RemoteName))
	return resourceGiteaPushMirrorRead(d, meta)
}

func resourceGiteaPushMirrorRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)

	owner, repo, remoteName, err := parsePushMirrorID(d.Id())
	if err != nil {
		return err
	}

	pm, resp, err := client.GetPushMirrorByRemoteName(owner, repo, remoteName)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return err
	}

	if pm == nil {
		d.SetId("")
		return nil
	}

	d.SetId(buildThreePartID(owner, repo, pm.RemoteName))
	_ = d.Set("owner", owner)
	_ = d.Set("repo", repo)
	_ = d.Set("remote_name", pm.RemoteName)
	_ = d.Set("remote_address", pm.RemoteAddress)
	_ = d.Set("interval", pm.Interval)
	_ = d.Set("sync_on_commit", pm.SyncONCommit)
	_ = d.Set("created", pm.Created)
	_ = d.Set("last_error", pm.LastError)
	_ = d.Set("last_update", pm.LastUpdate)

	return nil
}

func resourceGiteaPushMirrorDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)

	owner, repo, remoteName, err := parsePushMirrorID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.DeletePushMirror(owner, repo, remoteName)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return nil
		}
		return err
	}

	return nil
}

func resourceGiteaPushMirror() *schema.Resource {
	return &schema.Resource{
		Create: resourceGiteaPushMirrorCreate,
		Read:   resourceGiteaPushMirrorRead,
		Delete: resourceGiteaPushMirrorDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceGiteaPushMirrorImport,
		},
		Schema: map[string]*schema.Schema{
			"owner": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The owner (user or organization) of the repository.",
			},
			"repo": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the repository.",
			},
			"remote_address": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The remote repository URL to push to.",
			},
			"remote_username": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Username for authenticating with the remote repository.",
			},
			"remote_password": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Sensitive:   true,
				Description: "Password or authentication token for the remote repository.",
			},
			"interval": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "8h0m0s",
				Description: "Interval between mirror syncs (e.g. `8h0m0s`).",
			},
			"sync_on_commit": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     true,
				Description: "Whether to trigger mirror sync when new commits are pushed.",
			},
			"remote_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The remote name assigned to the push mirror by Gitea.",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the push mirror was created.",
			},
			"last_error": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last error message recorded during sync, if any.",
			},
			"last_update": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp of the last sync update.",
			},
		},
		Description: "`gitea_push_mirror` manages a repository push mirror in Gitea.",
	}
}
