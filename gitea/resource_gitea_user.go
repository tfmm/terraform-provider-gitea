package gitea

import (
	"context"
	"fmt"
	"strconv"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
)

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error
	client := meta.(*GiteaClient).Client

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	var resp *gitea.Response
	var user *gitea.User

	user, resp, err = client.GetUserByID(id)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		} else {
			return diag.FromErr(err)
		}
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
	client := meta.(*GiteaClient).Client

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}
	var resp *gitea.Response
	var user *gitea.User

	user, resp, err = client.GetUserByID(id)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return resourceUserCreate(ctx, d, meta)
		} else {
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
	visibility := gitea.VisibleType(d.Get(userVisibility).(string))

	opts := gitea.EditUserOption{
		SourceID:                0,
		LoginName:               d.Get(userLoginName).(string),
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
	}

	// Gitea never returns the password on read, so Terraform can't detect
	// drift on it by itself. Send it whenever it changed, or whenever the
	// user explicitly forces it via force_password_change, otherwise a
	// changed `password` in config would silently never reach the server.
	if d.HasChange(userPassword) || d.Get(userForcePasswordChange).(bool) {
		opts.Password = d.Get(userPassword).(string)
	}

	_, err = client.AdminEditUser(d.Get(userName).(string), opts)
	if err != nil {
		return diag.FromErr(err)
	}

	user, _, err = client.GetUserByID(id)
	if err != nil {
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
		},
		Description: "`gitea_user` manages a native gitea user.\n\n" +
			"If you are using OIDC or other kinds of authentication mechanisms you can still try to manage" +
			"ssh keys or other ressources this way.\n\n" +
			"Import requires `password` to remain configured because Gitea does not return passwords.",
	}
}
