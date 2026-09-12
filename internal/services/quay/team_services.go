package quay

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func (c *Client) UpdateTeam(orgname string, teamname string, teamToUpdate *UpdateTeam) error {
	body, err := json.Marshal(teamToUpdate)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, c.BaseURL+"/api/"+c.APIVersion+"/organization/"+orgname+"/team/"+teamname, bytes.NewBuffer(body))
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

	if resp.StatusCode != http.StatusCreated {
		return parseAPIError(resp)
	} else {
		return nil
	}
}

func (c *Client) DeleteTeam(orgname string, teamname string) error {
	req, err := http.NewRequest(http.MethodDelete, c.BaseURL+"/api/"+c.APIVersion+"/organization/"+orgname+"/team/"+teamname, nil)
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

func (c *Client) ListTeamMembers(orgname string, teamname string) ([]Member, error) {
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+"/api/"+c.APIVersion+"/organization/"+orgname+"/team/"+teamname+"/members", nil)
	if err != nil {
		return []Member{}, err
	}

	logf.Log.Info("[Quay] Sending request", "method", req.Method, "url", req.URL.String())

	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return []Member{}, err
	}

	logf.Log.Info("[Quay] Request response", "status", resp.Status)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []Member{}, parseAPIError(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []Member{}, err
	}

	members, err := ParseMembers(body)
	if err != nil {
		return []Member{}, err
	}

	return members, nil
}

func (c *Client) AddTeamMember(orgname string, teamname string, membername string) error {
	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/api/"+c.APIVersion+"/organization/"+orgname+"/team/"+teamname+"/members/"+membername, nil)
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

	if resp.StatusCode != http.StatusOK {
		return parseAPIError(resp)
	}

	return nil
}

func (c *Client) RemoveTeamMember(orgname string, teamname string, membername string) error {
	req, err := http.NewRequest(http.MethodDelete, c.BaseURL+"/api/"+c.APIVersion+"/organization/"+orgname+"/team/"+teamname+"/members/"+membername, nil)
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
