package gitea

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	userName                string = "username"
	userLoginName           string = "login_name"
	userEmail               string = "email"
	userFullName            string = "full_name"
	userPassword            string = "password"
	userMustChangePassword  string = "must_change_password"
	userSendNotification    string = "send_notification"
	userVisibility          string = "visibility"
	userDescription         string = "description"
	userLocation            string = "location"
	userActive              string = "active"
	userAdmin               string = "admin"
	userAllowGitHook        string = "allow_git_hook"
	userAllowLocalImport    string = "allow_import_local"
	userMaxRepoCreation     string = "max_repo_creation"
	userPhorbitLogin        string = "prohibit_login"
	userAllowCreateOrgs     string = "allow_create_organization"
	userRestricted          string = "restricted"
	userForcePasswordChange string = "force_password_change"
	userType                string = "user_type"
)

// Gitea's EditUserOption has no `type` field in code.gitea.io/sdk/gitea (it
// predates bot accounts, added in Gitea 28), so converting a user to/from a
// bot account needs a raw call. This mirrors the SDK's EditUserOption field
// for field, plus `Type`, with two deliberate differences: Password and
// LoginName are *string with omitempty instead of a bare string. The SDK's
// version has no omitempty on either, so it always serializes "password":""
// and "login_name":"<value>" even when not explicitly changing them; Gitea
// 28.0.0 rejects ANY edit carrying a password key (even an empty one) or a
// non-empty login_name against an account that is currently a bot ("a bot
// account cannot have a password or authentication source" - login_name is
// how Gitea identifies an external auth source like LDAP). Since login_name
// is a required field on this resource (so it's always non-empty in config,
// bot or not), that would otherwise make it impossible to ever again edit a
// bot account, including to convert it back to a regular user. Folding
// `type` into this same request (rather than a separate follow-up PATCH)
// also avoids an ordering hazard: a user->bot conversion done as
// edit-then-retype would still hit that same rejection on the edit half,
// since the account is still a regular user at that point.
type editUserOption struct {
	LoginName               *string `json:"login_name,omitempty"`
	Email                   *string `json:"email"`
	FullName                *string `json:"full_name"`
	Password                *string `json:"password,omitempty"`
	Description             *string `json:"description"`
	MustChangePassword      *bool   `json:"must_change_password"`
	Location                *string `json:"location"`
	Active                  *bool   `json:"active"`
	Admin                   *bool   `json:"admin"`
	AllowGitHook            *bool   `json:"allow_git_hook"`
	AllowImportLocal        *bool   `json:"allow_import_local"`
	MaxRepoCreation         *int    `json:"max_repo_creation"`
	ProhibitLogin           *bool   `json:"prohibit_login"`
	AllowCreateOrganization *bool   `json:"allow_create_organization"`
	Restricted              *bool   `json:"restricted"`
	Visibility              *string `json:"visibility"`
	Type                    string  `json:"type"`
}

type userTypeResponse struct {
	Type string `json:"type"`
}

func giteaUserType(userType string) string {
	if strings.EqualFold(userType, "bot") {
		return "Bot"
	}
	return "User"
}

func terraformUserType(giteaType string) string {
	if strings.EqualFold(giteaType, "Bot") {
		return "bot"
	}
	return "user"
}

func editUserRaw(client *GiteaClient, username string, opt editUserOption) error {
	return client.rawJSON("PATCH", fmt.Sprintf("/admin/users/%s", username), opt, nil)
}

// lookupUser fetches a user, preferring GetUserInfo(username) over
// GetUserByID(id). This matters for bot accounts: Gitea 28.0.0's
// GetUserByID goes through its general user-search index, which excludes
// bot-type accounts by default, so it reports an existing bot account as
// not found. GetUserInfo hits /users/{username} directly and works for
// every account type. The username is known on every call except a bare
// `terraform import <id>`, where it falls back to the ID-based lookup
// (which still works for regular accounts; importing a bot account by
// numeric ID is not supported).
func lookupUser(client *gitea.Client, d *schema.ResourceData) (*gitea.User, *gitea.Response, error) {
	if username := d.Get(userName).(string); username != "" {
		return client.GetUserInfo(username)
	}
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return nil, nil, err
	}
	return client.GetUserByID(id)
}

func getUserTypeRaw(client *GiteaClient, username string) (string, error) {
	var resp userTypeResponse
	if err := client.rawJSON("GET", fmt.Sprintf("/users/%s", username), nil, &resp); err != nil {
		return "", err
	}
	return terraformUserType(resp.Type), nil
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error
	giteaClient := meta.(*GiteaClient)
	client := giteaClient.Client

	var resp *gitea.Response
	var user *gitea.User

	user, resp, err = lookupUser(client, d)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		} else {
			return diag.FromErr(err)
		}
	}

	userTypeValue, err := getUserTypeRaw(giteaClient, user.UserName)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(userType, userTypeValue); err != nil {
		return diag.FromErr(err)
	}

	err = setUserResourceData(user, d)

	return diag.FromErr(err)
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error
	client := meta.(*GiteaClient).Client

	var user *gitea.User
	visibility := gitea.VisibleType(d.Get(userVisibility).(string))
	changePassword := d.Get(userMustChangePassword).(bool)

	opts := gitea.CreateUserOption{
		SourceID:           0,
		LoginName:          d.Get(userLoginName).(string),
		Username:           d.Get(userName).(string),
		FullName:           d.Get(userFullName).(string),
		Email:              d.Get(userEmail).(string),
		Password:           d.Get(userPassword).(string),
		MustChangePassword: &changePassword,
		SendNotify:         d.Get(userSendNotification).(bool),
		Visibility:         &visibility,
	}

	user, _, err = client.AdminCreateUser(opts)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d", user.ID))

	return resourceUserUpdate(ctx, d, meta)
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error
	giteaClient := meta.(*GiteaClient)
	client := giteaClient.Client

	var resp *gitea.Response
	var user *gitea.User

	user, resp, err = lookupUser(client, d)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return resourceUserCreate(ctx, d, meta)
		} else {
			return diag.FromErr(err)
		}
	}

	currentType, err := getUserTypeRaw(giteaClient, user.UserName)
	if err != nil {
		return diag.FromErr(err)
	}
	desiredType := d.Get(userType).(string)

	// Converting a bot account back to a regular user has to happen as its
	// own request before login_name/password can be sent: Gitea 28.0.0
	// validates a non-empty login_name (or a password) against the
	// account's *current* type, even within the same request that also sets
	// type back to "User" - see editUserOption's doc comment.
	if strings.EqualFold(currentType, "bot") && !strings.EqualFold(desiredType, "bot") {
		if err := editUserRaw(giteaClient, user.UserName, editUserOption{Type: giteaUserType(desiredType)}); err != nil {
			return diag.FromErr(err)
		}
	}

	mail := d.Get(userEmail).(string)
	fullName := d.Get(userFullName).(string)
	description := d.Get(userDescription).(string)
	changePassword := d.Get(userMustChangePassword).(bool)
	location := d.Get(userLocation).(string)
	active := d.Get(userActive).(bool)
	admin := d.Get(userAdmin).(bool)
	allowHook := d.Get(userAllowGitHook).(bool)
	allowImport := d.Get(userAllowLocalImport).(bool)
	maxRepoCreation := d.Get(userMaxRepoCreation).(int)
	accessDenied := d.Get(userPhorbitLogin).(bool)
	allowOrgs := d.Get(userAllowCreateOrgs).(bool)
	restricted := d.Get(userRestricted).(bool)
	visibility := d.Get(userVisibility).(string)

	opts := editUserOption{
		Email:                   &mail,
		FullName:                &fullName,
		Description:             &description,
		MustChangePassword:      &changePassword,
		Location:                &location,
		Active:                  &active,
		Admin:                   &admin,
		AllowGitHook:            &allowHook,
		AllowImportLocal:        &allowImport,
		MaxRepoCreation:         &maxRepoCreation,
		ProhibitLogin:           &accessDenied,
		AllowCreateOrganization: &allowOrgs,
		Restricted:              &restricted,
		Visibility:              &visibility,
		Type:                    giteaUserType(d.Get(userType).(string)),
	}

	// login_name is required by this resource's schema, so it's always
	// non-empty in config, but Gitea 28.0.0 rejects a non-empty login_name
	// on an edit against a bot account - see editUserOption's doc comment.
	if !strings.EqualFold(d.Get(userType).(string), "bot") {
		loginName := d.Get(userLoginName).(string)
		opts.LoginName = &loginName
	}

	// Gitea never returns the password on read, so Terraform can't detect
	// drift on it by itself. Send it whenever it changed, or whenever the
	// user explicitly forces it via force_password_change, otherwise a
	// changed `password` in config would silently never reach the server.
	// Left nil (omitted) otherwise - see editUserOption's doc comment for why
	// that matters for bot accounts specifically.
	if d.HasChange(userPassword) || d.Get(userForcePasswordChange).(bool) {
		password := d.Get(userPassword).(string)
		opts.Password = &password
	}

	if err := editUserRaw(giteaClient, d.Get(userName).(string), opts); err != nil {
		return diag.FromErr(err)
	}

	user, _, err = lookupUser(client, d)
	if err != nil {
		return diag.FromErr(err)
	}

	userTypeValue, err := getUserTypeRaw(giteaClient, user.UserName)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(userType, userTypeValue); err != nil {
		return diag.FromErr(err)
	}

	err = setUserResourceData(user, d)

	return diag.FromErr(err)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error
	client := meta.(*GiteaClient).Client

	var resp *gitea.Response

	resp, err = client.AdminDeleteUser(d.Get(userName).(string))
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return nil
		} else {
			return diag.FromErr(err)
		}
	}

	return diag.FromErr(err)
}

func setUserResourceData(user *gitea.User, d *schema.ResourceData) (err error) {
	d.SetId(fmt.Sprintf("%d", user.ID))
	if err := d.Set(userName, user.UserName); err != nil {
		return err
	}
	if err := d.Set(userEmail, user.Email); err != nil {
		return err
	}
	if err := d.Set(userFullName, user.FullName); err != nil {
		return err
	}
	if err := d.Set(userAdmin, user.IsAdmin); err != nil {
		return err
	}
	if err := d.Set(userLoginName, user.LoginName); err != nil {
		return err
	}
	if err := d.Set(userVisibility, string(user.Visibility)); err != nil {
		return err
	}
	if err := d.Set(userDescription, user.Description); err != nil {
		return err
	}
	if err := d.Set(userLocation, user.Location); err != nil {
		return err
	}
	if err := d.Set(userActive, user.IsActive); err != nil {
		return err
	}
	if err := d.Set(userPhorbitLogin, user.ProhibitLogin); err != nil {
		return err
	}
	if err := d.Set(userRestricted, user.Restricted); err != nil {
		return err
	}
	if err := d.Set(userMustChangePassword, d.Get(userMustChangePassword).(bool)); err != nil {
		return err
	}
	if err := d.Set(userSendNotification, d.Get(userSendNotification).(bool)); err != nil {
		return err
	}
	if err := d.Set(userAllowGitHook, d.Get(userAllowGitHook).(bool)); err != nil {
		return err
	}
	if err := d.Set(userAllowLocalImport, d.Get(userAllowLocalImport).(bool)); err != nil {
		return err
	}
	if err := d.Set(userMaxRepoCreation, d.Get(userMaxRepoCreation).(int)); err != nil {
		return err
	}
	if err := d.Set(userAllowCreateOrgs, d.Get(userAllowCreateOrgs).(bool)); err != nil {
		return err
	}
	if err := d.Set(userForcePasswordChange, d.Get(userForcePasswordChange).(bool)); err != nil {
		return err
	}

	return
}

func resourceGiteaUser() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceUserRead,
		CreateContext: resourceUserCreate,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Username of the user to be created",
			},
			"login_name": {
				Type:        schema.TypeString,
				Optional:    false,
				Required:    true,
				Description: "The login name can differ from the username",
			},
			"email": {
				Type:        schema.TypeString,
				Optional:    false,
				Required:    true,
				Description: "E-Mail Address of the user",
			},
			"full_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Required:    false,
				Description: "Full name of the user",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    false,
				Required:    true,
				Sensitive:   true,
				Description: "Password to be set for the user",
			},
			"must_change_password": {
				Type:        schema.TypeBool,
				Optional:    true,
				Required:    false,
				Default:     true,
				Description: "Flag if the user should change the password after first login",
			},
			"send_notification": {
				Type:        schema.TypeBool,
				Optional:    true,
				Required:    false,
				Default:     true,
				Description: "Flag to send a notification about the user creation to the defined `email`",
			},
			"visibility": {
				Type:        schema.TypeString,
				Optional:    true,
				Required:    false,
				Default:     "public",
				Description: "Visibility of the user. Can be `public`, `limited` or `private`",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Required:    false,
				Default:     "",
				Description: "A description of the user",
			},
			"location": {
				Type:     schema.TypeString,
				Optional: true,
				Required: false,
				Default:  "",
			},
			"active": {
				Type:        schema.TypeBool,
				Optional:    true,
				Required:    false,
				Default:     true,
				Description: "Flag if this user should be active or not",
			},
			"admin": {
				Type:        schema.TypeBool,
				Optional:    true,
				Required:    false,
				Default:     false,
				Description: "Flag if this user should be an administrator or not",
			},
			"allow_git_hook": {
				Type:     schema.TypeBool,
				Optional: true,
				Required: false,
				Default:  true,
			},
			"allow_import_local": {
				Type:     schema.TypeBool,
				Optional: true,
				Required: false,
				Default:  true,
			},
			"max_repo_creation": {
				Type:     schema.TypeInt,
				Optional: true,
				Required: false,
				Default:  -1,
			},
			"prohibit_login": {
				Type:        schema.TypeBool,
				Optional:    true,
				Required:    false,
				Default:     false,
				Description: "Flag if the user should not be allowed to log in (bot user)",
			},
			"allow_create_organization": {
				Type:     schema.TypeBool,
				Optional: true,
				Required: false,
				Default:  true,
			},
			"restricted": {
				Type:     schema.TypeBool,
				Optional: true,
				Required: false,
				Default:  false,
			},
			"force_password_change": {
				Type:        schema.TypeBool,
				Optional:    true,
				Required:    false,
				Default:     false,
				Description: "Flag if the user defined password should be overwritten or not",
			},
			"user_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "user",
				ValidateFunc: validation.StringInSlice([]string{"user", "bot"}, true),
				Description: "Whether this is a regular `user` or a `bot` account (Gitea >= 28.0.0). Bot accounts are " +
					"intended for automation/service use, e.g. CI tokens. `password` is still required by this resource " +
					"even for bot accounts, since Gitea user creation always needs one; prohibit_login is commonly set " +
					"to `true` alongside `user_type = \"bot\"` to prevent interactive login.",
			},
		},
		Description: "`gitea_user` manages a native gitea user.\n\n" +
			"If you are using OIDC or other kinds of authentication mechanisms you can still try to manage" +
			"ssh keys or other ressources this way.\n\n" +
			"Import requires `password` to remain configured because Gitea does not return passwords.",
	}
}
