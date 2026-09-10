package gitea

import (
	"fmt"
	"strconv"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGiteaLabel() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaLabelRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Label ID",
			},
			"org": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"user", "repo"},
				Description:   "Organization name",
			},
			"user": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"org"},
				RequiredWith:  []string{"repo"},
				Description:   "User or organization owner of the repository",
			},
			"repo": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"org"},
				RequiredWith:  []string{"user"},
				Description:   "Repository name",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Label name",
			},
			"color": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Label color",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Label description",
			},
			"exclusive": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the label is exclusive",
			},
		},
		Description: "Use this data source to retrieve details of a label in Gitea.",
	}
}

func dataSourceGiteaLabelRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	id := int64(d.Get("id").(int))

	var label *gitea.Label
	var err error

	if org, ok := d.GetOk("org"); ok && org.(string) != "" {
		label, _, err = client.GetOrgLabel(org.(string), id)
	} else {
		user := d.Get("user").(string)
		repo := d.Get("repo").(string)
		label, _, err = client.GetRepoLabel(user, repo, id)
	}

	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(label.ID, 10))
	d.Set("name", label.Name)
	d.Set("color", label.Color)
	d.Set("description", label.Description)
	d.Set("exclusive", label.Exclusive)

	return nil
}

func dataSourceGiteaLabels() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGiteaLabelsRead,
		Schema: map[string]*schema.Schema{
			"org": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"user", "repo"},
				Description:   "Organization name",
			},
			"user": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"org"},
				RequiredWith:  []string{"repo"},
				Description:   "User or organization owner of the repository",
			},
			"repo": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"org"},
				RequiredWith:  []string{"user"},
				Description:   "Repository name",
			},
			"labels": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"color": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"exclusive": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
		Description: "Use this data source to list labels in an organization or repository.",
	}
}

func dataSourceGiteaLabelsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)

	var labels []*gitea.Label
	var err error

	if org, ok := d.GetOk("org"); ok && org.(string) != "" {
		labels, _, err = client.ListOrgLabels(org.(string), gitea.ListOrgLabelsOptions{})
		d.SetId(fmt.Sprintf("org/%s/labels", org.(string)))
	} else {
		user := d.Get("user").(string)
		repo := d.Get("repo").(string)
		labels, _, err = client.ListRepoLabels(user, repo, gitea.ListLabelsOptions{})
		d.SetId(fmt.Sprintf("%s/%s/labels", user, repo))
	}

	if err != nil {
		return err
	}

	result := make([]map[string]interface{}, 0, len(labels))
	for _, label := range labels {
		result = append(result, map[string]interface{}{
			"id":          label.ID,
			"name":        label.Name,
			"color":       label.Color,
			"description": label.Description,
			"exclusive":   label.Exclusive,
		})
	}

	d.Set("labels", result)
	return nil
}
