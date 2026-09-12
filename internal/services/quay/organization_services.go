package quay

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

func (c *Client) CreateOrganization(orgToCreate *CreateOrganization) (Organization, error) {
	// prepare body
	body, err := json.Marshal(orgToCreate)
	if err != nil {
		return Organization{}, err
	}

	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/api/"+c.APIVersion+"/organization/", bytes.NewBuffer(body))
	if err != nil {
		return Organization{}, err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return Organization{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return Organization{}, parseAPIError(resp)
	} else {
		// Get the created org
		return c.GetOrganization(orgToCreate.Name) // Quay doesn't return the created object... (not really RESTful API...)
	}
}

func (c *Client) GetOrganization(name string) (Organization, error) {
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+"/api/"+c.APIVersion+"/organization/"+name, nil)
	if err != nil {
		return Organization{}, err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return Organization{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Organization{}, parseAPIError(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Organization{}, err
	}

	org, err := ParseOrganization(body)
	if err != nil {
		return Organization{}, err
	}

	return org, nil
}

func (c *Client) UpdateOrganization(orgname string, orgToUpdate *UpdateOrganization) (Organization, error) {
	body, err := json.Marshal(orgToUpdate)
	if err != nil {
		return Organization{}, err
	}

	req, err := http.NewRequest(http.MethodPut, c.BaseURL+"/api/"+c.APIVersion+"/organization/"+orgname, bytes.NewBuffer(body))
	if err != nil {
		return Organization{}, err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return Organization{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return Organization{}, parseAPIError(resp)
	} else {
		// Get the updated org
		return c.GetOrganization(orgname) // Quay doesn't return the created object... (not really RESTful API...)
	}
}

func (c *Client) DeleteOrganization(orgname string) error {
	req, err := http.NewRequest(http.MethodDelete, c.BaseURL+"/api/"+c.APIVersion+"/organization/"+orgname, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return parseAPIError(resp)
	}

	return nil
}
