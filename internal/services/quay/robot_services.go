package quay

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// <editor-fold desc="User Robot">

func (c *Client) CreateUserRobotAccount(robotShortName string, robotToCreate CreateRobotAccount) (RobotAccount, error) {
	body, err := json.Marshal(robotToCreate)
	if err != nil {
		return RobotAccount{}, err
	}

	req, err := http.NewRequest(http.MethodPut, c.BaseURL+"/api/"+c.APIVersion+"/user/robots/"+robotShortName, bytes.NewBuffer(body))
	if err != nil {
		return RobotAccount{}, err
	}

	logf.Log.Info("[Quay] Sending request", "method", req.Method, "url", req.URL.String())

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return RobotAccount{}, err
	}

	logf.Log.Info("[Quay] Request response", "status", resp.Status)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return RobotAccount{}, parseAPIError(resp)
	} else {
		// Get the updated robot account
		return c.GetUserRobotAccount(robotShortName)
	}
}

func (c *Client) GetUserRobotAccount(robotShortName string) (RobotAccount, error) {
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+"/api/"+c.APIVersion+"/user/robots/"+robotShortName, nil)
	if err != nil {
		return RobotAccount{}, err
	}

	logf.Log.Info("[Quay] Sending request", "method", req.Method, "url", req.URL.String())

	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return RobotAccount{}, err
	}

	logf.Log.Info("[Quay] Request response", "status", resp.Status)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return RobotAccount{}, parseAPIError(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return RobotAccount{}, err
	}

	robotAccount, err := ParseRobotAccount(body)
	if err != nil {
		return RobotAccount{}, err
	}

	return robotAccount, nil
}

func (c *Client) DeleteUserRobotAccount(robotShortName string) error {
	req, err := http.NewRequest(http.MethodDelete, c.BaseURL+"/api/"+c.APIVersion+"/user/robots/"+robotShortName, nil)
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

// </editor-fold>

// <editor-fold desc="Organization Robot">

func (c *Client) CreateOrganizationRobotAccount(organizationName string, robotShortName string, robotToCreate CreateRobotAccount) (RobotAccount, error) {
	body, err := json.Marshal(robotToCreate)
	if err != nil {
		return RobotAccount{}, err
	}

	req, err := http.NewRequest(http.MethodPut, c.BaseURL+"/api/"+c.APIVersion+"/organization/"+organizationName+"/robots/"+robotShortName, bytes.NewBuffer(body))
	if err != nil {
		return RobotAccount{}, err
	}

	logf.Log.Info("[Quay] Sending request", "method", req.Method, "url", req.URL.String())

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return RobotAccount{}, err
	}

	logf.Log.Info("[Quay] Request response", "status", resp.Status)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return RobotAccount{}, parseAPIError(resp)
	} else {
		// Get the updated robot account
		return c.GetOrganizationRobotAccount(organizationName, robotShortName)
	}
}

func (c *Client) GetOrganizationRobotAccount(organizationName string, robotShortName string) (RobotAccount, error) {
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+"/api/"+c.APIVersion+"/organization/"+organizationName+"/robots/"+robotShortName, nil)
	if err != nil {
		return RobotAccount{}, err
	}

	logf.Log.Info("[Quay] Sending request", "method", req.Method, "url", req.URL.String())

	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return RobotAccount{}, err
	}

	logf.Log.Info("[Quay] Request response", "status", resp.Status)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return RobotAccount{}, parseAPIError(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return RobotAccount{}, err
	}

	robotAccount, err := ParseRobotAccount(body)
	if err != nil {
		return RobotAccount{}, err
	}

	return robotAccount, nil
}

func (c *Client) DeleteOrganizationRobotAccount(organizationName string, robotShortName string) error {
	req, err := http.NewRequest(http.MethodDelete, c.BaseURL+"/api/"+c.APIVersion+"/organization/"+organizationName+"/robots/"+robotShortName, nil)
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

// </editor-fold>
