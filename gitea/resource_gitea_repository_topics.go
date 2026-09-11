package gitea

import (
	"context"
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceRepositoryTopicsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	user := d.Get("user").(string)
	repoName := d.Get("repo").(string)

	repo, resp, err := client.GetRepo(user, repoName)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("topics", repo.Topics)
	return nil
}

func resourceRepositoryTopicsCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	user := d.Get("user").(string)
	repoName := d.Get("repo").(string)

	rawTopics := d.Get("topics").(*schema.Set).List()
	topics := make([]string, 0, len(rawTopics))
	for _, t := range rawTopics {
		topics = append(topics, t.(string))
	}

	_, err := client.SetRepoTopics(user, repoName, topics)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s/%s", user, repoName))
	return resourceRepositoryTopicsRead(d, meta)
}

func resourceRepositoryTopicsUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceRepositoryTopicsCreate(d, meta)
}

func resourceRepositoryTopicsDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitea.Client)
	user := d.Get("user").(string)
	repoName := d.Get("repo").(string)

	_, err := client.SetRepoTopics(user, repoName, []string{})
	return err
}

func resourceGiteaRepositoryTopics() *schema.Resource {
	return &schema.Resource{
		Read:   resourceRepositoryTopicsRead,
		Create: resourceRepositoryTopicsCreate,
		Update: resourceRepositoryTopicsUpdate,
		Delete: resourceRepositoryTopicsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				parts := strings.Split(d.Id(), "/")
				if len(parts) != 2 {
					return nil, fmt.Errorf("unexpected ID format (%q), expected <user>/<repo>", d.Id())
				}
				d.Set("user", parts[0])
				d.Set("repo", parts[1])
				d.SetId(d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"user": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "User or organization owner of the repository",
			},
			"repo": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Repository name",
			},
			"topics": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required:    true,
				Description: "Set of topics to apply to the repository",
			},
		},
		Description: "This resource manages topics assigned to a Gitea repository.",
	}
}
