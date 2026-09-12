package quay

import (
	"encoding/json"
	"io"
	"net/http"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func (c *Client) GetCurrentUser() (User, error) {
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+"/api/"+c.APIVersion+"/user/", nil)
	if err != nil {
		return User{}, err
	}

	logf.Log.Info("[Quay] Sending request", "method", req.Method, "url", req.URL.String())

	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return User{}, err
	}

	logf.Log.Info("[Quay] Request response", "status", resp.Status)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return User{}, parseAPIError(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return User{}, err
	}

	// Parse response
	var user User
	err = json.Unmarshal(body, &user)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (c *Client) GetUser(username string) (User, error) {
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+"/api/"+c.APIVersion+"/user/"+username, nil)
	if err != nil {
		return User{}, err
	}

	logf.Log.Info("[Quay] Sending request", "method", req.Method, "url", req.URL.String())

	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return User{}, err
	}

	logf.Log.Info("[Quay] Request response", "status", resp.Status)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return User{}, parseAPIError(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return User{}, err
	}

	// Parse response
	var user User
	err = json.Unmarshal(body, &user)
	if err != nil {
		return User{}, err
	}

	return user, nil
}
