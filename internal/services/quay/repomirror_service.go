package quay

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func (c *Client) GetRepositoryMirror(repoName string) (RepositoryMirror, error) {
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+"/api/"+c.APIVersion+"/repository/"+repoName+"/mirror", nil)
	if err != nil {
		return RepositoryMirror{}, err
	}

	logf.Log.Info("[Quay] Sending request", "method", req.Method, "url", req.URL.String())

	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return RepositoryMirror{}, err
	}

	logf.Log.Info("[Quay] Request response", "status", resp.Status)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return RepositoryMirror{}, parseAPIError(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return RepositoryMirror{}, err
	}

	mirror, err := ParseRepositoryMirror(body)
	if err != nil {
		return RepositoryMirror{}, err
	}

	return mirror, nil
}

func (c *Client) CreateRepositoryMirror(repoName string, mirrorToCreate CreateRepositoryMirror) (RepositoryMirror, error) {
	// prepare body
	body, err := json.Marshal(mirrorToCreate)
	if err != nil {
		return RepositoryMirror{}, err
	}

	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/api/"+c.APIVersion+"/repository/"+repoName+"/mirror", bytes.NewBuffer(body))
	if err != nil {
		return RepositoryMirror{}, err
	}

	logf.Log.Info("[Quay] Sending request", "method", req.Method, "url", req.URL.String())

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return RepositoryMirror{}, err
	}

	logf.Log.Info("[Quay] Request response", "status", resp.Status)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return RepositoryMirror{}, parseAPIError(resp)
	} else {
		// Get the created repositoryMirror
		return c.GetRepositoryMirror(repoName)
	}
}
