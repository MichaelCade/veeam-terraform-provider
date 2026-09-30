package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Credential struct {
	ID                 string `json:"id,omitempty"`
	Username           string `json:"username"`
	Password           string `json:"password,omitempty"`
	Description        string `json:"description,omitempty"`
	Type               string `json:"type"`                         // e.g. "Standard", "Linux", "ManagedService"
	AuthenticationType string `json:"authenticationType,omitempty"` // e.g. "Password", "SshKey" for Linux credentials
	SSHPort            int    `json:"SSHPort,omitempty"`
	CreationTime       string `json:"creationTime,omitempty"`
}

type CredentialListResponse struct {
	Data []Credential `json:"data"`
}

// GetCredentials fetches all stored credentials
func (c *Client) GetCredentials(ctx context.Context) ([]Credential, error) {
	resp, err := c.DoRequest(ctx, http.MethodGet, "/api/v1/credentials", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch credentials: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get credentials, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var listResp CredentialListResponse
	if err := json.Unmarshal(bodyBytes, &listResp); err != nil {
		return nil, fmt.Errorf("failed to parse credentials list: %w", err)
	}

	return listResp.Data, nil
}

// GetCredentialByID fetches a single credential by ID
func (c *Client) GetCredentialByID(ctx context.Context, id string) (*Credential, error) {
	path := fmt.Sprintf("/api/v1/credentials/%s", id)
	resp, err := c.DoRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch credential %s: %w", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read credential body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get credential %s, status: %d, body: %s", id, resp.StatusCode, string(bodyBytes))
	}

	var cred Credential
	if err := json.Unmarshal(bodyBytes, &cred); err != nil {
		return nil, fmt.Errorf("failed to parse credential: %w", err)
	}

	return &cred, nil
}

// CreateCredential creates a new credential entry in VBR
func (c *Client) CreateCredential(ctx context.Context, cred Credential) (*Credential, error) {
	// If type is Linux and authenticationType is empty, default to Password
	if cred.Type == "Linux" && cred.AuthenticationType == "" {
		cred.AuthenticationType = "Password"
	}

	payload, err := json.Marshal(cred)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal credential payload: %w", err)
	}

	resp, err := c.DoRequest(ctx, http.MethodPost, "/api/v1/credentials", bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read create response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create credential, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var created Credential
	if err := json.Unmarshal(bodyBytes, &created); err != nil {
		return nil, fmt.Errorf("failed to parse created credential: %w", err)
	}

	return &created, nil
}

// DeleteCredential deletes a credential entry by ID
func (c *Client) DeleteCredential(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/v1/credentials/%s", id)
	resp, err := c.DoRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete credential: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete credential %s, status: %d, body: %s", id, resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// UpdateCredential updates an existing credential entry in VBR
func (c *Client) UpdateCredential(ctx context.Context, id string, cred Credential) (*Credential, error) {
	if cred.Type == "Linux" && cred.AuthenticationType == "" {
		cred.AuthenticationType = "Password"
	}
	cred.ID = id

	if cred.CreationTime == "" {
		existing, err := c.GetCredentialByID(ctx, id)
		if err == nil && existing != nil {
			cred.CreationTime = existing.CreationTime
		}
	}

	payload, err := json.Marshal(cred)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal credential payload: %w", err)
	}

	path := fmt.Sprintf("/api/v1/credentials/%s", id)
	resp, err := c.DoRequest(ctx, http.MethodPut, path, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to update credential %s: %w", id, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read update response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to update credential %s, status: %d, body: %s", id, resp.StatusCode, string(bodyBytes))
	}

	var updated Credential
	if err := json.Unmarshal(bodyBytes, &updated); err != nil {
		return nil, fmt.Errorf("failed to parse updated credential: %w", err)
	}

	return &updated, nil
}

