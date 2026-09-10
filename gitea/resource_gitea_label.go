package gitea

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceLabelRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}

	var label *gitea.Label
	var resp *gitea.Response

	if org, ok := d.GetOk("org"); ok && org.(string) != "" {
		label, resp, err = client.GetOrgLabel(org.(string), id)
	} else {
		user := d.Get("user").(string)
		repo := d.Get("repo").(string)
		label, resp, err = client.GetRepoLabel(user, repo, id)
	}

	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("name", label.Name)
	d.Set("color", label.Color)
	d.Set("description", label.Description)
	d.Set("exclusive", label.Exclusive)

	return nil
}

func resourceLabelCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)

	color := strings.TrimPrefix(d.Get("color").(string), "#")
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	exclusive := d.Get("exclusive").(bool)

	var label *gitea.Label
	var err error

	if org, ok := d.GetOk("org"); ok && org.(string) != "" {
		opts := gitea.CreateOrgLabelOption{
			Name:        name,
			Color:       color,
			Description: description,
			Exclusive:   exclusive,
		}
		label, _, err = client.CreateOrgLabel(org.(string), opts)
	} else {
		opts := gitea.CreateLabelOption{
			Name:        name,
			Color:       color,
			Description: description,
			Exclusive:   exclusive,
		}
		user := d.Get("user").(string)
		repo := d.Get("repo").(string)
		label, _, err = client.CreateLabel(user, repo, opts)
	}

	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(label.ID, 10))
	return resourceLabelRead(d, meta)
}

func resourceLabelUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}

	name := d.Get("name").(string)
	color := strings.TrimPrefix(d.Get("color").(string), "#")
	description := d.Get("description").(string)
	exclusive := d.Get("exclusive").(bool)

	if org, ok := d.GetOk("org"); ok && org.(string) != "" {
		opts := gitea.EditOrgLabelOption{
			Name:        &name,
			Color:       &color,
			Description: &description,
			Exclusive:   &exclusive,
		}
		_, _, err = client.EditOrgLabel(org.(string), id, opts)
	} else {
		opts := gitea.EditLabelOption{
			Name:        &name,
			Color:       &color,
			Description: &description,
			Exclusive:   &exclusive,
		}
		user := d.Get("user").(string)
		repo := d.Get("repo").(string)
		_, _, err = client.EditLabel(user, repo, id, opts)
	}

	if err != nil {
		return err
	}

	return resourceLabelRead(d, meta)
}

func resourceLabelDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}

	if org, ok := d.GetOk("org"); ok && org.(string) != "" {
		_, err = client.DeleteOrgLabel(org.(string), id)
	} else {
		user := d.Get("user").(string)
		repo := d.Get("repo").(string)
		_, err = client.DeleteLabel(user, repo, id)
	}

	return err
}

func resourceGiteaLabel() *schema.Resource {
	return &schema.Resource{
		Read:   resourceLabelRead,
		Create: resourceLabelCreate,
		Update: resourceLabelUpdate,
		Delete: resourceLabelDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				parts := strings.Split(d.Id(), "/")
				if len(parts) == 3 {
					d.Set("user", parts[0])
					d.Set("repo", parts[1])
					d.SetId(parts[2])
					return []*schema.ResourceData{d}, nil
				} else if len(parts) == 2 {
					if parts[0] == "org" {
						d.Set("org", parts[1])
					} else {
						d.Set("org", parts[0])
						d.SetId(parts[1])
						return []*schema.ResourceData{d}, nil
					}
				}
				if len(parts) == 3 && parts[0] == "org" {
					d.Set("org", parts[1])
					d.SetId(parts[2])
					return []*schema.ResourceData{d}, nil
				}
				return nil, fmt.Errorf("unexpected ID format (%q), expected <user>/<repo>/<id> or <org>/<id>", d.Id())
			},
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the label",
			},
			"color": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "6-character hex color code (e.g. `ff0000` or `#ff0000`)",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the label",
			},
			"exclusive": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether the label is scoped/exclusive",
			},
			"org": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"user", "repo"},
				Description:   "Organization name to create an org-level label",
			},
			"user": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"org"},
				RequiredWith:  []string{"repo"},
				Description:   "User or organization owner of the repository",
			},
			"repo": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"org"},
				RequiredWith:  []string{"user"},
				Description:   "Repository name to create a repo-level label",
			},
		},
		Description: "This resource manages labels for repositories or organizations in Gitea.",
	}
}
