package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ManagedServer struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Status      string `json:"status,omitempty"`
}

type ManagedServersResult struct {
	Data []ManagedServer `json:"data"`
}

type LinuxHostSSHSettingsModel struct {
	SSHPort int `json:"SSHPort"`
}

type CreateLinuxManagedServerSpec struct {
	Name                   string                     `json:"name"`
	Description            string                     `json:"description,omitempty"`
	Type                   string                     `json:"type"` // "LinuxHost"
	CredentialsStorageType string                     `json:"credentialsStorageType"` // "Stored"
	CredentialsID          string                     `json:"credentialsId"`
	SSHFingerprint         string                     `json:"sshFingerprint,omitempty"`
	SSHSettings            *LinuxHostSSHSettingsModel `json:"sshSettings,omitempty"`
}

type CreateWindowsManagedServerSpec struct {
	Name                   string `json:"name"`
	Description            string `json:"description,omitempty"`
	Type                   string `json:"type"` // "WindowsHost"
	CredentialsStorageType string `json:"credentialsStorageType"` // "Stored"
	CredentialsID          string `json:"credentialsId"`
}

// GetManagedServers fetches all managed infrastructure servers (vCenter, Windows, Linux, Hyper-V)
func (c *Client) GetManagedServers(ctx context.Context) ([]ManagedServer, error) {
	resp, err := c.DoRequest(ctx, http.MethodGet, "/api/v1/backupInfrastructure/managedServers", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch managed servers: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read managed servers response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get managed servers, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var result ManagedServersResult
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to parse managed servers result: %w", err)
	}

	return result.Data, nil
}

// GetManagedServerByID fetches a managed server by GUID
func (c *Client) GetManagedServerByID(ctx context.Context, id string) (*ManagedServer, error) {
	path := fmt.Sprintf("/api/v1/backupInfrastructure/managedServers/%s", id)
	resp, err := c.DoRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch managed server %s: %w", id, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read managed server response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get managed server %s, status: %d, body: %s", id, resp.StatusCode, string(bodyBytes))
	}

	var server ManagedServer
	if err := json.Unmarshal(bodyBytes, &server); err != nil {
		return nil, fmt.Errorf("failed to parse managed server: %w", err)
	}

	return &server, nil
}

// CreateManagedServer registers a new server (Linux or Windows) in Veeam Backup Infrastructure
func (c *Client) CreateManagedServer(ctx context.Context, payload interface{}) (*ManagedServer, error) {
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal managed server payload: %w", err)
	}

	var nameExtractor struct {
		Name string `json:"name"`
	}
	_ = json.Unmarshal(jsonBytes, &nameExtractor)
	targetName := nameExtractor.Name

	resp, err := c.DoRequest(ctx, http.MethodPost, "/api/v1/backupInfrastructure/managedServers", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create managed server: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read create managed server response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("failed to create managed server, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var created ManagedServer
	if err := json.Unmarshal(bodyBytes, &created); err == nil && created.ID != "" && created.Name != "" {
		return &created, nil
	}

	var session struct {
		ID         string `json:"id"`
		ResourceId string `json:"resourceId"`
	}
	if err := json.Unmarshal(bodyBytes, &session); err == nil && session.ResourceId != "" {
		srv, err := c.GetManagedServerByID(ctx, session.ResourceId)
		if err == nil && srv != nil {
			return srv, nil
		}
	}

	if targetName != "" {
		for i := 0; i < 15; i++ {
			time.Sleep(1 * time.Second)
			servers, err := c.GetManagedServers(ctx)
			if err == nil {
				for _, s := range servers {
					if s.Name == targetName {
						return &s, nil
					}
				}
			}
		}
	}

	if session.ResourceId != "" {
		return &ManagedServer{
			ID:   session.ResourceId,
			Name: targetName,
		}, nil
	}

	return nil, fmt.Errorf("managed server creation initiated but '%s' could not be located in infrastructure list", targetName)
}

// DeleteManagedServer removes a managed server from Veeam Backup Infrastructure by GUID
func (c *Client) DeleteManagedServer(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/v1/backupInfrastructure/managedServers/%s", id)
	resp, err := c.DoRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete managed server %s: %w", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete managed server %s, status: %d, body: %s", id, resp.StatusCode, string(bodyBytes))
	}

	return nil
}
