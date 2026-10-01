package gitea

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
)

// Config is per-provider, specifies where to connect to gitea
type Config struct {
	Token      string
	Username   string
	Password   string
	BaseURL    string
	Insecure   bool
	CACertFile string
}

// GiteaClient wraps the official SDK client together with the raw HTTP
// plumbing needed to reach endpoints the SDK has no bindings for yet
// (e.g. newly introduced Gitea APIs). All existing code keeps using the
// embedded *gitea.Client exactly as before; only resources that need an
// endpoint missing from the SDK use the raw* helpers below.
type GiteaClient struct {
	*gitea.Client

	baseURL    string
	httpClient *http.Client

	token    string
	username string
	password string
}

// rawRequest issues an authenticated request against the Gitea API for
// endpoints not yet exposed by the code.gitea.io/sdk/gitea package.
func (c *GiteaClient) rawRequest(method, path string, body io.Reader) (*http.Response, error) {
	url := strings.TrimSuffix(c.baseURL, "/") + "/api/v1" + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	switch {
	case c.token != "":
		req.Header.Set("Authorization", "token "+c.token)
	case c.username != "":
		req.SetBasicAuth(c.username, c.password)
	}
	return c.httpClient.Do(req)
}

// rawJSON performs a raw API request and decodes a JSON response body into out.
// A nil out skips decoding (useful for 204/2xx responses with no body).
func (c *GiteaClient) rawJSON(method, path string, in, out interface{}) error {
	var body io.Reader
	if in != nil {
		encoded, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(encoded)
	}

	resp, err := c.rawRequest(method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gitea API request %s %s failed with status %d: %s", method, path, resp.StatusCode, string(respBody))
	}

	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// Client returns a *GiteaClient to interact with the configured gitea instance
func (c *Config) Client() (interface{}, error) {

	if c.Token == "" && c.Username == "" {
		return nil, fmt.Errorf("either a token or a username needs to be used")
	}
	// Configure TLS/SSL
	tlsConfig := tls.Config{MinVersion: tls.VersionTLS12}
	// If a CACertFile has been specified, use that for cert validation
	if c.CACertFile != "" {
		caCert, err := os.ReadFile(c.CACertFile)
		if err != nil {
			return nil, err
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}

	// If configured as insecure, turn off SSL verification
	tlsConfig.InsecureSkipVerify = c.Insecure

	t := http.DefaultTransport.(*http.Transport).Clone()
	t.TLSClientConfig = &tlsConfig
	t.MaxIdleConnsPerHost = 100
	t.TLSHandshakeTimeout = 10 * time.Second

	httpClient := &http.Client{
		Transport: logging.NewTransport("Gitea", t),
	}

	if c.BaseURL == "" {
		c.BaseURL = "https://gitea.com"
	}

	var client *gitea.Client
	var err error
	if c.Token != "" {
		client, err = gitea.NewClient(c.BaseURL, gitea.SetToken(c.Token), gitea.SetHTTPClient(httpClient))
		if err != nil {
			return nil, err
		}
	}

	if c.Username != "" {
		client, err = gitea.NewClient(c.BaseURL, gitea.SetBasicAuth(c.Username, c.Password), gitea.SetHTTPClient(httpClient))
		if err != nil {
			return nil, err
		}
	}

	// Test the credentials by checking we can get information about the authenticated user.
	_, _, err = client.GetMyUserInfo()
	if err != nil {
		return nil, err
	}

	return &GiteaClient{
		Client:     client,
		baseURL:    c.BaseURL,
		httpClient: httpClient,
		token:      c.Token,
		username:   c.Username,
		password:   c.Password,
	}, nil
}
