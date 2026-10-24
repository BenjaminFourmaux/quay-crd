package quay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func (c *Client) ListRepositories(namespace string) ([]Repository, error) {
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+"/api/"+c.APIVersion+"/repository", nil)
	if err != nil {
		return []Repository{}, err
	}

	// Add query params
	query := req.URL.Query()
	if namespace != "" {
		query.Set("namespace", namespace)
	}
	req.URL.RawQuery = query.Encode()

	logf.Log.Info("[Quay] Sending request", "method", req.Method, "url", req.URL.String())

	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return []Repository{}, err
	}

	logf.Log.Info("[Quay] Request response", "status", resp.Status)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []Repository{}, parseAPIError(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []Repository{}, err
	}

	repos, err := ParseRepositories(body)
	if err != nil {
		return []Repository{}, err
	}

	return repos, nil
}

func (c *Client) GetRepository(name string) (Repository, error) {
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+"/api/"+c.APIVersion+"/repository/"+name, nil)
	if err != nil {
		return Repository{}, err
	}

	logf.Log.Info("[Quay] Sending request", "method", req.Method, "url", req.URL.String())

	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return Repository{}, err
	}

	logf.Log.Info("[Quay] Request response", "status", resp.Status)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Repository{}, parseAPIError(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Repository{}, err
	}

	repo, err := ParseRepository(body)
	if err != nil {
		return Repository{}, err
	}

	return repo, nil
}

func (c *Client) CreateRepository(repoToCreate *CreateRepository) (Repository, error) {
	// prepare body
	body, err := json.Marshal(repoToCreate)
	if err != nil {
		return Repository{}, err
	}

	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/api/"+c.APIVersion+"/repository", bytes.NewBuffer(body))
	if err != nil {
		return Repository{}, err
	}

	logf.Log.Info("[Quay] Sending request", "method", req.Method, "url", req.URL.String())

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return Repository{}, err
	}

	logf.Log.Info("[Quay] Request response", "status", resp.Status)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return Repository{}, parseAPIError(resp)
	} else {

		// reconstruct the full repo name
		var repoFullName string
		if repoToCreate.Namespace != "" {
			repoFullName = fmt.Sprintf("%s/%s", repoToCreate.Namespace, repoToCreate.Repository)
		} else {
			repoFullName = repoToCreate.Repository
		}

		// Get the created repo
		return c.GetRepository(repoFullName) // Quay doesn't return the created object... (not really RESTful API...)
	}
}

func (c *Client) UpdateRepository(repoName string, repoToUpdate *UpdateRepository) (Repository, error) {
	body, err := json.Marshal(repoToUpdate)
	if err != nil {
		return Repository{}, err
	}

	req, err := http.NewRequest(http.MethodPut, c.BaseURL+"/api/"+c.APIVersion+"/repository/"+repoName, bytes.NewBuffer(body))
	if err != nil {
		return Repository{}, err
	}

	logf.Log.Info("[Quay] Sending request", "method", req.Method, "url", req.URL.String())

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return Repository{}, err
	}

	logf.Log.Info("[Quay] Request response", "status", resp.Status)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Repository{}, parseAPIError(resp)
	} else {
		// Get the updated repo
		return c.GetRepository(repoName) // Quay doesn't return the created object... (not really RESTful API...)
	}
}

func (c *Client) ChangeRepositoryVisibility(repoName string, repoToUpdate *ChangeRepositoryVisibility) error {
	// prepare body
	body, err := json.Marshal(repoToUpdate)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/api/"+c.APIVersion+"/repository/"+repoName+"/changevisibility", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	logf.Log.Info("[Quay] Sending request", "method", req.Method, "url", req.URL.String())

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	logf.Log.Info("[Quay] Request response", "status", resp.Status)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return parseAPIError(resp)
	}
	return nil
}

func (c *Client) DeleteRepository(repoName string) error {
	req, err := http.NewRequest(http.MethodDelete, c.BaseURL+"/api/"+c.APIVersion+"/repository/"+repoName, nil)
	if err != nil {
		return err
	}

	logf.Log.Info("[Quay] Sending request", "method", req.Method, "url", req.URL.String())

	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	logf.Log.Info("[Quay] Request response", "status", resp.Status)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return parseAPIError(resp)
	}

	return nil
}
