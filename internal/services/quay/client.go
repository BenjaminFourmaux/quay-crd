package quay

import (
	"github.com/go-logr/logr"
	"net/http"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type Client struct {
	BaseURL    string
	APIVersion string
	Token      string
	HTTPClient *http.Client
	logger     logr.Logger
}

/*
NewClient is the Quay Client constructor
*/
func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL:    baseURL,
		APIVersion: "v1", // Fixed API version
		Token:      token,
		HTTPClient: http.DefaultClient,
		logger:     logf.Log.WithName("quay-service"),
	}
}

func (c *Client) Ping() error {
	user, err := c.GetCurrentUser()

	if err == nil {
		c.logger.Info("Successfully pinged Quay API", "user is super user", user.SuperUser)
	}

	return err
}
