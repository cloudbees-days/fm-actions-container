package cloudbees

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

// Client represents a CloudBees Platform API client
type Client struct {
	baseURL     string
	token       string
	orgID       string
	httpClient  *http.Client
	useOrgAsApp bool // Flag to determine if we use org ID as application ID for flags API
}

// Environment represents an environment
type Environment struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	ResourceID string `json:"resourceId"`
	IsDisabled bool   `json:"isDisabled"`
}

// ListEnvironmentsResponse represents the response when listing environments
type ListEnvironmentsResponse struct {
	Environments []Environment `json:"environments"`
}

// Flag represents a feature flag
type Flag struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	FlagType    string   `json:"flagType"`
	Variants    []string `json:"variants"`
	Description string   `json:"description"`
	IsPermanent bool     `json:"isPermanent"`
	ResourceID  string   `json:"resourceId"`
	CascURL     string   `json:"cascUrl"`
}

// GetFlagResponse represents the response when getting a flag
type GetFlagResponse struct {
	Flag Flag `json:"flag"`
}

// FlagConfiguration represents a flag configuration
type FlagConfiguration struct {
	Enabled            bool        `json:"enabled"`
	DefaultValue       interface{} `json:"defaultValue"`
	Conditions         interface{} `json:"conditions"`
	VariantsEnabled    bool        `json:"variantsEnabled"`
	StickinessProperty string      `json:"stickinessProperty,omitempty"`
}

// FlagConfigurationDetail represents detailed flag configuration
type FlagConfigurationDetail struct {
	FlagID        string            `json:"flagId"`
	FlagName      string            `json:"flagName"`
	Description   string            `json:"description"`
	Labels        []string          `json:"labels"`
	Created       string            `json:"created"`
	Updated       string            `json:"updated"`
	Configuration FlagConfiguration `json:"configuration"`
}

// GetFlagConfigurationResponse represents the response when getting flag configuration
type GetFlagConfigurationResponse struct {
	Configuration FlagConfiguration `json:"configuration"`
}

// UpdateFlagConfigurationRequest represents request to update flag configuration
type UpdateFlagConfigurationRequest struct {
	Configuration FlagConfiguration `json:"configuration"`
}

// CreateFlagRequest represents request to create a new flag
type CreateFlagRequest struct {
	Name        string   `json:"name"`
	FlagType    string   `json:"flagType"`
	Variants    []string `json:"variants"`
	Description string   `json:"description"`
	IsPermanent bool     `json:"isPermanent"`
}

// CreateFlagResponse represents response when creating a flag
type CreateFlagResponse struct {
	Flag Flag `json:"flag"`
}

// ListFlagsResponse represents response when listing flags
type ListFlagsResponse struct {
	Flags []Flag `json:"flags"`
}

// Application represents an application in CloudBees Platform
type Application struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	Description          string   `json:"description"`
	EndpointID           string   `json:"endpointId"`
	RepositoryURL        string   `json:"repositoryUrl"`
	DefaultBranch        string   `json:"defaultBranch"`
	OrganizationID       string   `json:"organizationId"`
	ServiceType          string   `json:"serviceType"`
	LinkedComponentIDs   []string `json:"linkedComponentIds"`
	LinkedEnvironmentIDs []string `json:"linkedEnvironmentIds"`
}

// ListApplicationsResponse represents the response when listing applications
type ListApplicationsResponse struct {
	Service []Application `json:"service"`
}

// NewClient creates a new CloudBees Platform API client
func NewClient(baseURL, token, orgID string) (*Client, error) {
	return NewClientWithOptions(baseURL, token, orgID, false)
}

// NewClientWithOptions creates a new CloudBees Platform API client with additional options
func NewClientWithOptions(baseURL, token, orgID string, useOrgAsApp bool) (*Client, error) {
	if baseURL == "" {
		baseURL = "https://api.cloudbees.io"
	}
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}
	if orgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	client := &Client{
		baseURL:     strings.TrimSuffix(baseURL, "/"),
		token:       token,
		orgID:       orgID,
		useOrgAsApp: useOrgAsApp,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	return client, nil
}

// makeRequest is a helper method to make HTTP requests
func (c *Client) makeRequest(method, url string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}

// ListEnvironments retrieves all environments for the organization
func (c *Client) ListEnvironments() ([]Environment, error) {
	url := fmt.Sprintf("%s/v2/organizations/%s/environments", c.baseURL, c.orgID)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response ListEnvironmentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Environments, nil
}

// GetFlagByName retrieves a flag by name from the organization
func (c *Client) GetFlagByName(applicationID, flagName string) (*Flag, error) {
	// Use org ID as application ID if the flag is set (legacy API), otherwise use the actual application ID
	apiAppID := applicationID
	if c.useOrgAsApp {
		apiAppID = c.orgID
	}
	url := fmt.Sprintf("%s/v2/applications/%s/flags/by-name/%s", c.baseURL, apiAppID, flagName)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response GetFlagResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response.Flag, nil
}

// GetFlagConfiguration retrieves flag configuration for a specific environment
func (c *Client) GetFlagConfiguration(applicationID, flagID, environmentID string) (*FlagConfigurationDetail, error) {
	// Use org ID as application ID if the flag is set (legacy API), otherwise use the actual application ID
	apiAppID := applicationID
	if c.useOrgAsApp {
		apiAppID = c.orgID
	}
	url := fmt.Sprintf("%s/v2/applications/%s/flags/%s/configuration/environments/%s",
		c.baseURL, apiAppID, flagID, environmentID)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response GetFlagConfigurationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	// Create a FlagConfigurationDetail with the response data
	config := &FlagConfigurationDetail{
		FlagID:        flagID,
		Configuration: response.Configuration,
	}

	return config, nil
}

// UpdateFlagConfiguration updates flag configuration for a specific environment
func (c *Client) UpdateFlagConfiguration(applicationID, flagID, environmentID string, config FlagConfiguration) error {
	fmt.Printf("DEBUG: UpdateFlagConfiguration called with appID=%s, flagID=%s, envID=%s\n", applicationID, flagID, environmentID)

	// Use org ID as application ID if the flag is set (legacy API), otherwise use the actual application ID
	apiAppID := applicationID
	if c.useOrgAsApp {
		apiAppID = c.orgID
	}
	url := fmt.Sprintf("%s/v2/applications/%s/flags/%s/configuration/environments/%s",
		c.baseURL, apiAppID, flagID, environmentID)

	request := UpdateFlagConfigurationRequest{
		Configuration: config,
	}

	// Debug: Log the request payload
	requestJSON, _ := json.Marshal(request)
	fmt.Printf("DEBUG: Sending PUT request to: %s\n", url)
	fmt.Printf("DEBUG: Request payload: %s\n", string(requestJSON))

	resp, err := c.makeRequest("PUT", url, request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// SetFlagConfiguration sets flag configuration using PUT with only specified fields
func (c *Client) SetFlagConfiguration(applicationID, flagID, environmentID string, config map[string]interface{}) error {
	// Use org ID as application ID if the flag is set (legacy API), otherwise use the actual application ID
	apiAppID := applicationID
	if c.useOrgAsApp {
		apiAppID = c.orgID
	}
	url := fmt.Sprintf("%s/v2/applications/%s/flags/%s/configuration/environments/%s",
		c.baseURL, apiAppID, flagID, environmentID)

	// Based on user testing, the API uses PUT for partial updates (opposite to REST conventions)
	resp, err := c.makeRequest("PUT", url, config)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ListFlags retrieves all flags for the application
func (c *Client) ListFlags(applicationID string) ([]Flag, error) {
	// Use org ID as application ID if the flag is set (legacy API), otherwise use the actual application ID
	apiAppID := applicationID
	if c.useOrgAsApp {
		apiAppID = c.orgID
	}
	url := fmt.Sprintf("%s/v2/applications/%s/flags", c.baseURL, apiAppID)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response ListFlagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Flags, nil
}

// CreateFlag creates a new feature flag
func (c *Client) CreateFlag(applicationID, name, flagType, description string, variants []string, isPermanent bool) (*Flag, error) {
	// Use org ID as application ID if the flag is set (legacy API), otherwise use the actual application ID
	apiAppID := applicationID
	if c.useOrgAsApp {
		apiAppID = c.orgID
	}
	url := fmt.Sprintf("%s/v2/applications/%s/flags", c.baseURL, apiAppID)

	request := CreateFlagRequest{
		Name:        name,
		FlagType:    flagType,
		Variants:    variants,
		Description: description,
		IsPermanent: isPermanent,
	}

	resp, err := c.makeRequest("POST", url, request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response CreateFlagResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response.Flag, nil
}

// DeleteFlag deletes a feature flag
func (c *Client) DeleteFlag(applicationID, flagID string) error {
	// Use org ID as application ID if the flag is set (legacy API), otherwise use the actual application ID
	apiAppID := applicationID
	if c.useOrgAsApp {
		apiAppID = c.orgID
	}
	url := fmt.Sprintf("%s/v2/applications/%s/flags/%s", c.baseURL, apiAppID, flagID)

	resp, err := c.makeRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ListApplications retrieves all applications for the organization
func (c *Client) ListApplications() ([]Application, error) {
	url := fmt.Sprintf("%s/v1/organizations/%s/services?typeFilter=APPLICATION_FILTER", c.baseURL, c.orgID)

	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response ListApplicationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Service, nil
}

// GetApplicationByName retrieves an application by its name
func (c *Client) GetApplicationByName(name string) (*Application, error) {
	applications, err := c.ListApplications()
	if err != nil {
		return nil, err
	}

	for _, app := range applications {
		if app.Name == name {
			return &app, nil
		}
	}

	return nil, fmt.Errorf("application '%s' not found", name)
}

// WriteOutput writes outputs in CloudBees format to $CLOUDBEES_OUTPUTS files
func WriteOutput(name, value string) {
	if outDir := os.Getenv("CLOUDBEES_OUTPUTS"); outDir != "" {
		filepath := path.Join(outDir, name)
		if err := os.WriteFile(filepath, []byte(value), 0640); err != nil {
			// Don't fail the whole operation if output writing fails, just log it
			fmt.Printf("Warning: failed to write CloudBees output %s: %v\n", name, err)
		}
	} else {
		fmt.Printf("Warning: CLOUDBEES_OUTPUTS environment variable not set, skipping output %s=%s\n", name, value)
	}
}
