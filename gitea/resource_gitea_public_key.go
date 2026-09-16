package gitea

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	PublicKeyUser         string = "username"
	PublicKey             string = "key"
	PublicKeyReadOnlyFlag string = "read_only"
	PublicKeyTitle        string = "title"
	PublicKeyId           string = "id"
	PublicKeyFingerprint  string = "fingerprint"
	PublicKeyCreated      string = "created"
	PublicKeyType         string = "type"
)

// normalizeSSHKey standardizes SSH key format for consistent comparison
func normalizeSSHKey(key string) string {
	// Remove leading/trailing whitespace
	normalized := strings.TrimSpace(key)
	// Normalize line endings to Unix format
	normalized = strings.ReplaceAll(normalized, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	// Remove any trailing newlines
	normalized = strings.TrimRight(normalized, "\n")
	return normalized
}

// sshKeyDiffSuppressFunc compares SSH keys after normalization
func sshKeyDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	return normalizeSSHKey(old) == normalizeSSHKey(new)
}

func resourcePublicKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error
	client := meta.(*gitea.Client)

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	var resp *gitea.Response
	var pubKey *gitea.PublicKey

	pubKey, resp, err = client.GetPublicKey(id)

	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		} else {
			return diag.FromErr(err)
		}
	}

	err = setPublicKeyResourceData(pubKey, d)

	return diag.FromErr(err)
}

func resourcePublicKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error
	client := meta.(*gitea.Client)

	var pubKey *gitea.PublicKey

	opts := gitea.CreateKeyOption{
		Title:    d.Get(PublicKeyTitle).(string),
		Key:      d.Get(PublicKey).(string),
		ReadOnly: d.Get(PublicKeyReadOnlyFlag).(bool),
	}

	pubKey, _, err = client.AdminCreateUserPublicKey(d.Get(PublicKeyUser).(string), opts)

	err = setPublicKeyResourceData(pubKey, d)

	return diag.FromErr(err)
}

func resourcePublicKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// update = recreate
	if diags := resourcePublicKeyDelete(ctx, d, meta); diags.HasError() {
		return diags
	}
	return resourcePublicKeyCreate(ctx, d, meta)
}

func resourcePublicKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error
	client := meta.(*gitea.Client)

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	var resp *gitea.Response

	resp, err = client.AdminDeleteUserPublicKey(d.Get(PublicKeyUser).(string), int(id))

	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return diag.FromErr(err)
		} else {
			return diag.FromErr(err)
		}
	}

	return diag.FromErr(err)
}

func setPublicKeyResourceData(pubKey *gitea.PublicKey, d *schema.ResourceData) (err error) {
	d.SetId(fmt.Sprintf("%d", pubKey.ID))
	if err := d.Set(PublicKeyUser, pubKey.Owner.UserName); err != nil {
		return err
	}
	if err := d.Set(PublicKey, pubKey.Key); err != nil {
		return err
	}
	if err := d.Set(PublicKeyTitle, pubKey.Title); err != nil {
		return err
	}
	if err := d.Set(PublicKeyReadOnlyFlag, pubKey.ReadOnly); err != nil {
		return err
	}
	if err := d.Set(PublicKeyCreated, timeToString(pubKey.Created)); err != nil {
		return err
	}
	if err := d.Set(PublicKeyFingerprint, pubKey.Fingerprint); err != nil {
		return err
	}
	if err := d.Set(PublicKeyType, pubKey.KeyType); err != nil {
		return err
	}
	return
}

func resourceGiteaPublicKey() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourcePublicKeyRead,
		CreateContext: resourcePublicKeyCreate,
		UpdateContext: resourcePublicKeyUpdate,
		DeleteContext: resourcePublicKeyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"title": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Title of the key to add",
			},
			"key": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Sensitive:        true,
				Description:      "An armored SSH key to add",
				DiffSuppressFunc: sshKeyDiffSuppressFunc,
			},
			"read_only": {
				Type:        schema.TypeBool,
				Required:    false,
				Optional:    true,
				Default:     false,
				Description: "Describe if the key has only read access or read/write",
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Optional:    false,
				ForceNew:    true,
				Description: "User to associate with the added key",
			},
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		Description: "`gitea_public_key` manages ssh key that are associated with users.",
	}
}
