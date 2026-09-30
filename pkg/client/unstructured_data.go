package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type UnstructuredDataServer struct {
	ID                  string `json:"id"`
	Name                string `json:"name,omitempty"`
	Type                string `json:"type"`
	Path                string `json:"path,omitempty"`
	AccessCredentialsID string `json:"accessCredentialsId,omitempty"`
	CacheRepositoryID   string `json:"cacheRepositoryId,omitempty"`
}

type UnstructuredDataServersResult struct {
	Data []UnstructuredDataServer `json:"data"`
}

type UnstructuredDataProcessingModel struct {
	CacheRepositoryID string                     `json:"cacheRepositoryId,omitempty"`
	BackupProxies     BackupProxiesSettingsModel `json:"backupProxies"`
}

type CreateSmbShareServerSpec struct {
	Type                      string                           `json:"type"` // "SMBShare"
	Path                      string                           `json:"path"`
	AccessCredentialsRequired bool                             `json:"accessCredentialsRequired,omitempty"`
	AccessCredentialsID       string                           `json:"accessCredentialsId,omitempty"`
	Processing                UnstructuredDataProcessingModel `json:"processing"`
}

type CreateNfsShareServerSpec struct {
	Type       string                           `json:"type"` // "NFSShare"
	Path       string                           `json:"path"`
	Processing UnstructuredDataProcessingModel `json:"processing"`
}

// GetUnstructuredDataServers retrieves all unstructured data sources (SMB, NFS, etc.)
func (c *Client) GetUnstructuredDataServers(ctx context.Context) ([]UnstructuredDataServer, error) {
	resp, err := c.DoRequest(ctx, http.MethodGet, "/api/v1/inventory/unstructuredDataServers", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch unstructured data servers: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get unstructured data servers, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var result UnstructuredDataServersResult
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to parse unstructured data servers: %w", err)
	}

	return result.Data, nil
}

// GetUnstructuredDataServerByID fetches a single unstructured data server by ID
func (c *Client) GetUnstructuredDataServerByID(ctx context.Context, id string) (*UnstructuredDataServer, error) {
	path := fmt.Sprintf("/api/v1/inventory/unstructuredDataServers/%s", id)
	resp, err := c.DoRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch unstructured data server %s: %w", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get unstructured data server %s, status: %d, body: %s", id, resp.StatusCode, string(bodyBytes))
	}

	var server UnstructuredDataServer
	if err := json.Unmarshal(bodyBytes, &server); err != nil {
		return nil, fmt.Errorf("failed to parse unstructured data server: %w", err)
	}

	return &server, nil
}

// CreateUnstructuredDataServer adds a new SMB or NFS file share as a source in Veeam VBR inventory
func (c *Client) CreateUnstructuredDataServer(ctx context.Context, payload interface{}) (*UnstructuredDataServer, error) {
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal unstructured data server payload: %w", err)
	}

	var pathExtractor struct {
		Path string `json:"path"`
		Name string `json:"name"`
	}
	_ = json.Unmarshal(jsonBytes, &pathExtractor)
	targetPath := pathExtractor.Path

	// 1. Pre-check: Check if source with targetPath already exists
	if targetPath != "" {
		if servers, err := c.GetUnstructuredDataServers(ctx); err == nil {
			for _, s := range servers {
				if s.Path == targetPath || s.Name == targetPath {
					return &s, nil
				}
			}
		}
	}

	resp, err := c.DoRequest(ctx, http.MethodPost, "/api/v1/inventory/unstructuredDataServers", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create unstructured data server: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read create response: %w", err)
	}

	// 2. Auto-adopt if 400 Bad Request returned due to "already exists"
	if resp.StatusCode == http.StatusBadRequest && targetPath != "" && strings.Contains(string(bodyBytes), "already exists") {
		if servers, err := c.GetUnstructuredDataServers(ctx); err == nil {
			for _, s := range servers {
				if s.Path == targetPath || s.Name == targetPath {
					return &s, nil
				}
			}
		}
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("failed to create unstructured data server, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var created UnstructuredDataServer
	if err := json.Unmarshal(bodyBytes, &created); err == nil && created.ID != "" {
		return &created, nil
	}

	var session SessionModel
	_ = json.Unmarshal(bodyBytes, &session)
	if session.ResourceId != "" {
		srv, err := c.GetUnstructuredDataServerByID(ctx, session.ResourceId)
		if err == nil && srv != nil {
			return srv, nil
		}
	}

	// 3. Poll GetUnstructuredDataServers for up to 60s
	if targetPath != "" {
		for i := 0; i < 30; i++ {
			time.Sleep(2 * time.Second)
			servers, err := c.GetUnstructuredDataServers(ctx)
			if err == nil {
				for _, s := range servers {
					if s.Path == targetPath || s.Name == targetPath {
						return &s, nil
					}
				}
			}

			if session.ID != "" {
				sessUpdate, err := c.GetSessionByID(ctx, session.ID)
				if err == nil && sessUpdate != nil && sessUpdate.ResourceId != "" {
					srv, err := c.GetUnstructuredDataServerByID(ctx, sessUpdate.ResourceId)
					if err == nil && srv != nil {
						return srv, nil
					}
				}
			}
		}
	}

	if session.ResourceId != "" {
		return &UnstructuredDataServer{
			ID:   session.ResourceId,
			Path: targetPath,
		}, nil
	}

	return nil, fmt.Errorf("unstructured data server creation initiated but '%s' could not be located in inventory list within 60s timeout", targetPath)
}

// DeleteUnstructuredDataServer removes an unstructured data server from Veeam inventory by ID
func (c *Client) DeleteUnstructuredDataServer(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/v1/inventory/unstructuredDataServers/%s", id)
	resp, err := c.DoRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete unstructured data server %s: %w", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusAccepted {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete unstructured data server %s, status: %d, body: %s", id, resp.StatusCode, string(bodyBytes))
	}

	return nil
}
