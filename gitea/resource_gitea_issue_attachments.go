package gitea

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func issueAttachmentSchema(subjectField, subjectDescription string) map[string]*schema.Schema {
	return mergeSchemaMaps(repositoryIdentitySchema(), map[string]*schema.Schema{
		subjectField: {
			Type:        schema.TypeInt,
			Required:    true,
			ForceNew:    true,
			Description: subjectDescription,
		},
		sourcePathField: {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The local source file to upload.",
		},
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The attachment name. Defaults to the source file basename.",
		},
		"attachment_id": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The attachment ID.",
		},
		"size": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The attachment size in bytes.",
		},
		"download_count": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The attachment download count.",
		},
		"uuid": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The attachment UUID.",
		},
		"download_url": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The attachment download URL.",
		},
		createdAtField: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The attachment creation timestamp.",
		},
	})
}

func setIssueAttachmentData(d *schema.ResourceData, owner, repo string, attachment *gitea.Attachment, subjectField string, subjectID int64) error {
	if err := d.Set(repositoryOwnerField, owner); err != nil {
		return err
	}
	if err := d.Set(repositoryNameField, repo); err != nil {
		return err
	}
	if err := d.Set(subjectField, int(subjectID)); err != nil {
		return err
	}
	if err := d.Set("attachment_id", int(attachment.ID)); err != nil {
		return err
	}
	if err := d.Set("name", attachment.Name); err != nil {
		return err
	}
	if err := d.Set("size", int(attachment.Size)); err != nil {
		return err
	}
	if err := d.Set("download_count", int(attachment.DownloadCount)); err != nil {
		return err
	}
	if err := d.Set("uuid", attachment.UUID); err != nil {
		return err
	}
	if err := d.Set("download_url", attachment.DownloadURL); err != nil {
		return err
	}
	if err := d.Set(createdAtField, timeToString(attachment.Created)); err != nil {
		return err
	}
	return nil
}

func parseInt64IDPart(value string, field string) (int64, error) {
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: %w", field, value, err)
	}
	return parsed, nil
}

func resourceGiteaIssueAttachment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGiteaIssueAttachmentCreate,
		ReadContext:   resourceGiteaIssueAttachmentRead,
		UpdateContext: resourceGiteaIssueAttachmentUpdate,
		DeleteContext: resourceGiteaIssueAttachmentDelete,
		Importer:      issueAttachmentImporter("issue_index"),
		Schema:        issueAttachmentSchema("issue_index", "The issue index."),
		Description: "`gitea_issue_attachment` manages an issue attachment.\n\n" +
			"Import expects the resource ID in the form `owner:repo:issue_index:attachment_id`.",
	}
}

func resourceGiteaIssueAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	owner := strings.ToLower(d.Get(repositoryOwnerField).(string))
	repo := strings.ToLower(d.Get(repositoryNameField).(string))
	issueIndex := int64(d.Get("issue_index").(int))

	content, err := readLocalFile(d.Get(sourcePathField).(string))
	if err != nil {
		return diag.FromErr(err)
	}
	name := basenameOrValue(d.Get(sourcePathField).(string), d.Get("name").(string))
	attachment, _, err := client.CreateIssueAttachment(owner, repo, issueIndex, bytes.NewReader(content), name)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(buildFourPartID(owner, repo, strconv.FormatInt(issueIndex, 10), strconv.FormatInt(attachment.ID, 10)))
	return resourceGiteaIssueAttachmentRead(ctx, d, meta)
}

func resourceGiteaIssueAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	owner, repo, issueIndexPart, attachmentIDPart, err := parseFourPartID(d.Id(), repositoryOwnerField, repositoryNameField, "issue_index", "attachment_id")
	if err != nil {
		return diag.FromErr(err)
	}
	issueIndex, err := parseInt64IDPart(issueIndexPart, "issue_index")
	if err != nil {
		return diag.FromErr(err)
	}
	attachmentID, err := parseInt64IDPart(attachmentIDPart, "attachment_id")
	if err != nil {
		return diag.FromErr(err)
	}
	attachment, resp, err := client.GetIssueAttachment(owner, repo, issueIndex, attachmentID)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	return diag.FromErr(setIssueAttachmentData(d, owner, repo, attachment, "issue_index", issueIndex))
}

func resourceGiteaIssueAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	owner, repo, issueIndexPart, attachmentIDPart, err := parseFourPartID(d.Id(), repositoryOwnerField, repositoryNameField, "issue_index", "attachment_id")
	if err != nil {
		return diag.FromErr(err)
	}
	issueIndex, err := parseInt64IDPart(issueIndexPart, "issue_index")
	if err != nil {
		return diag.FromErr(err)
	}
	attachmentID, err := parseInt64IDPart(attachmentIDPart, "attachment_id")
	if err != nil {
		return diag.FromErr(err)
	}
	_, _, err = client.EditIssueAttachment(owner, repo, issueIndex, attachmentID, gitea.EditAttachmentOptions{Name: d.Get("name").(string)})
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceGiteaIssueAttachmentRead(ctx, d, meta)
}

func resourceGiteaIssueAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	owner, repo, issueIndexPart, attachmentIDPart, err := parseFourPartID(d.Id(), repositoryOwnerField, repositoryNameField, "issue_index", "attachment_id")
	if err != nil {
		return diag.FromErr(err)
	}
	issueIndex, err := parseInt64IDPart(issueIndexPart, "issue_index")
	if err != nil {
		return diag.FromErr(err)
	}
	attachmentID, err := parseInt64IDPart(attachmentIDPart, "attachment_id")
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = client.DeleteIssueAttachment(owner, repo, issueIndex, attachmentID)
	return diag.FromErr(err)
}

func resourceGiteaIssueCommentAttachment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGiteaIssueCommentAttachmentCreate,
		ReadContext:   resourceGiteaIssueCommentAttachmentRead,
		UpdateContext: resourceGiteaIssueCommentAttachmentUpdate,
		DeleteContext: resourceGiteaIssueCommentAttachmentDelete,
		Importer:      issueAttachmentImporter("comment_id"),
		Schema:        issueAttachmentSchema("comment_id", "The issue comment ID."),
		Description: "`gitea_issue_comment_attachment` manages an issue comment attachment.\n\n" +
			"Import expects the resource ID in the form `owner:repo:comment_id:attachment_id`.",
	}
}

func resourceGiteaIssueCommentAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	owner := strings.ToLower(d.Get(repositoryOwnerField).(string))
	repo := strings.ToLower(d.Get(repositoryNameField).(string))
	commentID := int64(d.Get("comment_id").(int))

	content, err := readLocalFile(d.Get(sourcePathField).(string))
	if err != nil {
		return diag.FromErr(err)
	}
	name := basenameOrValue(d.Get(sourcePathField).(string), d.Get("name").(string))
	attachment, _, err := client.CreateIssueCommentAttachment(owner, repo, commentID, bytes.NewReader(content), name)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(buildFourPartID(owner, repo, strconv.FormatInt(commentID, 10), strconv.FormatInt(attachment.ID, 10)))
	return resourceGiteaIssueCommentAttachmentRead(ctx, d, meta)
}

func resourceGiteaIssueCommentAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	owner, repo, commentIDPart, attachmentIDPart, err := parseFourPartID(d.Id(), repositoryOwnerField, repositoryNameField, "comment_id", "attachment_id")
	if err != nil {
		return diag.FromErr(err)
	}
	commentID, err := parseInt64IDPart(commentIDPart, "comment_id")
	if err != nil {
		return diag.FromErr(err)
	}
	attachmentID, err := parseInt64IDPart(attachmentIDPart, "attachment_id")
	if err != nil {
		return diag.FromErr(err)
	}
	attachment, resp, err := client.GetIssueCommentAttachment(owner, repo, commentID, attachmentID)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	return diag.FromErr(setIssueAttachmentData(d, owner, repo, attachment, "comment_id", commentID))
}

func resourceGiteaIssueCommentAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	owner, repo, commentIDPart, attachmentIDPart, err := parseFourPartID(d.Id(), repositoryOwnerField, repositoryNameField, "comment_id", "attachment_id")
	if err != nil {
		return diag.FromErr(err)
	}
	commentID, err := parseInt64IDPart(commentIDPart, "comment_id")
	if err != nil {
		return diag.FromErr(err)
	}
	attachmentID, err := parseInt64IDPart(attachmentIDPart, "attachment_id")
	if err != nil {
		return diag.FromErr(err)
	}
	_, _, err = client.EditIssueCommentAttachment(owner, repo, commentID, attachmentID, gitea.EditAttachmentOptions{Name: d.Get("name").(string)})
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceGiteaIssueCommentAttachmentRead(ctx, d, meta)
}

func resourceGiteaIssueCommentAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	owner, repo, commentIDPart, attachmentIDPart, err := parseFourPartID(d.Id(), repositoryOwnerField, repositoryNameField, "comment_id", "attachment_id")
	if err != nil {
		return diag.FromErr(err)
	}
	commentID, err := parseInt64IDPart(commentIDPart, "comment_id")
	if err != nil {
		return diag.FromErr(err)
	}
	attachmentID, err := parseInt64IDPart(attachmentIDPart, "attachment_id")
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = client.DeleteIssueCommentAttachment(owner, repo, commentID, attachmentID)
	return diag.FromErr(err)
}

func dataSourceGiteaIssueAttachments() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGiteaIssueAttachmentsRead,
		Schema: mergeSchemaMaps(repositoryIdentitySchema(), map[string]*schema.Schema{
			"issue_index": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The issue index.",
			},
			"attachments": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Issue attachments.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":             {Type: schema.TypeInt, Computed: true},
						"name":           {Type: schema.TypeString, Computed: true},
						"size":           {Type: schema.TypeInt, Computed: true},
						"download_count": {Type: schema.TypeInt, Computed: true},
						createdAtField:   {Type: schema.TypeString, Computed: true},
						"uuid":           {Type: schema.TypeString, Computed: true},
						"download_url":   {Type: schema.TypeString, Computed: true},
					},
				},
			},
		}),
		Description: "Lists issue attachments.",
	}
}

func dataSourceGiteaIssueAttachmentsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	owner := strings.ToLower(d.Get(repositoryOwnerField).(string))
	repo := strings.ToLower(d.Get(repositoryNameField).(string))
	issueIndex := int64(d.Get("issue_index").(int))
	attachments, _, err := client.ListIssueAttachments(owner, repo, issueIndex)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("attachments", flattenAttachments(attachments)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting attachments: %w", err))
	}
	d.SetId(buildResourceID(owner, repo, strconv.FormatInt(issueIndex, 10)))
	return nil
}

func dataSourceGiteaIssueCommentAttachments() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGiteaIssueCommentAttachmentsRead,
		Schema: mergeSchemaMaps(repositoryIdentitySchema(), map[string]*schema.Schema{
			"comment_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The issue comment ID.",
			},
			"attachments": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Issue comment attachments.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":             {Type: schema.TypeInt, Computed: true},
						"name":           {Type: schema.TypeString, Computed: true},
						"size":           {Type: schema.TypeInt, Computed: true},
						"download_count": {Type: schema.TypeInt, Computed: true},
						createdAtField:   {Type: schema.TypeString, Computed: true},
						"uuid":           {Type: schema.TypeString, Computed: true},
						"download_url":   {Type: schema.TypeString, Computed: true},
					},
				},
			},
		}),
		Description: "Lists issue comment attachments.",
	}
}

func dataSourceGiteaIssueCommentAttachmentsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitea.Client)
	owner := strings.ToLower(d.Get(repositoryOwnerField).(string))
	repo := strings.ToLower(d.Get(repositoryNameField).(string))
	commentID := int64(d.Get("comment_id").(int))
	attachments, _, err := client.ListIssueCommentAttachments(owner, repo, commentID)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("attachments", flattenAttachments(attachments)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting attachments: %w", err))
	}
	d.SetId(buildResourceID(owner, repo, strconv.FormatInt(commentID, 10)))
	return nil
}

func issueAttachmentImporter(subjectField string) *schema.ResourceImporter {
	return &schema.ResourceImporter{
		StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
			parts := strings.Split(d.Id(), ":")
			if len(parts) != 4 {
				return nil, fmt.Errorf("unexpected ID format (%q), expected owner:repo:%s:attachment_id", d.Id(), subjectField)
			}
			if err := d.Set(repositoryOwnerField, parts[0]); err != nil {
				return nil, err
			}
			if err := d.Set(repositoryNameField, parts[1]); err != nil {
				return nil, err
			}
			id, err := strconv.Atoi(parts[2])
			if err != nil {
				return nil, err
			}
			if _, err := strconv.Atoi(parts[3]); err != nil {
				return nil, err
			}
			if err := d.Set(subjectField, id); err != nil {
				return nil, err
			}
			d.SetId(buildFourPartID(parts[0], parts[1], parts[2], parts[3]))
			return []*schema.ResourceData{d}, nil
		},
	}
}
