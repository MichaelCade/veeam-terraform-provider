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

type Repository struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Path        string `json:"path,omitempty"`
	HostID      string `json:"hostId,omitempty"`
}

type RepositoryListResponse struct {
	Data []Repository `json:"data"`
}

// 1. Linux Local / Hardened Repository Spec
type LinuxLocalRepositorySettings struct {
	Path                        string `json:"path"`
	UseFastCloningOnXFSVolumes bool   `json:"useFastCloningOnXFSVolumes,omitempty"`
	EnableGovernanceMode        bool   `json:"enableGovernanceMode,omitempty"`
	GovernanceModeRetentionDays int    `json:"governanceModeRetentionDays,omitempty"`
}

type CreateLinuxLocalRepositorySpec struct {
	Name        string                       `json:"name"`
	Description string                       `json:"description,omitempty"`
	Type        string                       `json:"type"` // "LinuxLocal" or "LinuxHardened"
	HostID      string                       `json:"hostId"`
	Repository  LinuxLocalRepositorySettings `json:"repository"`
}

// 2. Windows Local Repository Spec
type WindowsLocalRepositorySettings struct {
	Path string `json:"path"`
}

type CreateWindowsLocalRepositorySpec struct {
	Name        string                         `json:"name"`
	Description string                         `json:"description,omitempty"`
	Type        string                         `json:"type"` // "WinLocal"
	HostID      string                         `json:"hostId"`
	Repository  WindowsLocalRepositorySettings `json:"repository"`
}

// 3. SMB Repository Spec
type RepositoryShareGatewayModel struct {
	AutoSelectEnabled bool     `json:"autoSelectEnabled"`
	GatewayServerIDs  []string `json:"gatewayServerIds,omitempty"`
}

type SmbShareSettings struct {
	SharePath     string                      `json:"sharePath"`
	CredentialsID string                      `json:"credentialsId,omitempty"`
	GatewayServer RepositoryShareGatewayModel `json:"gatewayServer"`
}

type NetworkRepositorySettings struct {
	Path string `json:"path,omitempty"`
}

type CreateSmbRepositorySpec struct {
	Name        string                    `json:"name"`
	Description string                    `json:"description,omitempty"`
	Type        string                    `json:"type"` // "Smb"
	Share       SmbShareSettings          `json:"share"`
	Repository  NetworkRepositorySettings `json:"repository,omitempty"`
}

// 4. NFS Repository Spec
type NfsShareSettings struct {
	SharePath     string                      `json:"sharePath"`
	GatewayServer RepositoryShareGatewayModel `json:"gatewayServer"`
}

type CreateNfsRepositorySpec struct {
	Name        string                    `json:"name"`
	Description string                    `json:"description,omitempty"`
	Type        string                    `json:"type"` // "Nfs"
	Share       NfsShareSettings          `json:"share"`
	Repository  NetworkRepositorySettings `json:"repository,omitempty"`
}

// 5. S3 Compatible Repository Spec
type S3CompatibleAccountSettings struct {
	ServicePoint  string `json:"servicePoint"`
	RegionID      string `json:"regionId"`
	CredentialsID string `json:"credentialsId"`
}

type S3CompatibleBucketSettings struct {
	BucketName string `json:"bucketName"`
	FolderName string `json:"folderName"`
}

type CreateS3CompatibleRepositorySpec struct {
	Name        string                      `json:"name"`
	Description string                      `json:"description,omitempty"`
	Type        string                      `json:"type"` // "S3Compatible"
	Account     S3CompatibleAccountSettings `json:"account"`
	Bucket      S3CompatibleBucketSettings  `json:"bucket"`
}

// 6. Veeam Data Cloud Vault Spec
type VeeamDataCloudVaultInfo struct {
	ID string `json:"id"`
}

type VeeamDataCloudAccountSettings struct {
	Vault VeeamDataCloudVaultInfo `json:"vault"`
}

type ObjectStorageImmutabilitySettings struct {
	IsEnabled bool `json:"isEnabled"`
}

type VeeamDataCloudContainerSettings struct {
	Folder       string                            `json:"folder"`
	Immutability ObjectStorageImmutabilitySettings `json:"immutability,omitempty"`
}

type CreateVeeamDataCloudVaultRepositorySpec struct {
	Name        string                          `json:"name"`
	Description string                          `json:"description,omitempty"`
	Type        string                          `json:"type"` // "VeeamDataCloudVault"
	Account     VeeamDataCloudAccountSettings   `json:"account"`
	Container   VeeamDataCloudContainerSettings `json:"container"`
}

// GetRepositories retrieves all backup repositories from Veeam VBR
func (c *Client) GetRepositories(ctx context.Context) ([]Repository, error) {
	resp, err := c.DoRequest(ctx, http.MethodGet, "/api/v1/backupInfrastructure/repositories", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get repositories, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var listResp RepositoryListResponse
	if err := json.Unmarshal(bodyBytes, &listResp); err != nil {
		return nil, fmt.Errorf("failed to parse repository response: %w", err)
	}

	return listResp.Data, nil
}

// GetDefaultCacheRepositoryID retrieves the Default Backup Repository ID or the first available repository
func (c *Client) GetDefaultCacheRepositoryID(ctx context.Context) (string, error) {
	repos, err := c.GetRepositories(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to fetch repositories for cache repository lookup: %w", err)
	}

	if len(repos) == 0 {
		return "", fmt.Errorf("no backup repositories found in Veeam VBR to use as cache repository")
	}

	for _, r := range repos {
		if strings.EqualFold(r.Name, "Default Backup Repository") {
			return r.ID, nil
		}
	}

	return repos[0].ID, nil
}

// GetRepositoryByID fetches a single backup repository by ID
func (c *Client) GetRepositoryByID(ctx context.Context, id string) (*Repository, error) {
	path := fmt.Sprintf("/api/v1/backupInfrastructure/repositories/%s", id)
	resp, err := c.DoRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository %s: %w", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read repository body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get repository %s, status: %d, body: %s", id, resp.StatusCode, string(bodyBytes))
	}

	var repo Repository
	if err := json.Unmarshal(bodyBytes, &repo); err != nil {
		return nil, fmt.Errorf("failed to parse repository: %w", err)
	}

	return &repo, nil
}

type SessionResultModel struct {
	Result     string `json:"result,omitempty"`
	Message    string `json:"message,omitempty"`
	IsCanceled bool   `json:"isCanceled,omitempty"`
}

// SessionModel represents an async session status from Veeam VBR
type SessionModel struct {
	ID         string              `json:"id"`
	Name       string              `json:"name"`
	State      string              `json:"state"`
	ResourceId string              `json:"resourceId"`
	Result     *SessionResultModel `json:"result,omitempty"`
}

// GetSessionByID fetches async session status by ID
func (c *Client) GetSessionByID(ctx context.Context, id string) (*SessionModel, error) {
	path := fmt.Sprintf("/api/v1/sessions/%s", id)
	resp, err := c.DoRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch session %s: %w", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read session response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get session %s, status: %d, body: %s", id, resp.StatusCode, string(bodyBytes))
	}

	var session SessionModel
	if err := json.Unmarshal(bodyBytes, &session); err != nil {
		return nil, fmt.Errorf("failed to parse session: %w", err)
	}

	return &session, nil
}

// CreateRepository registers a new backup repository in Veeam VBR
func (c *Client) CreateRepository(ctx context.Context, payload interface{}) (*Repository, error) {
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal repository payload: %w", err)
	}

	var nameExtractor struct {
		Name string `json:"name"`
	}
	_ = json.Unmarshal(jsonBytes, &nameExtractor)
	targetName := nameExtractor.Name

	// 1. Pre-check: If repository already exists, bind and return it immediately
	if targetName != "" {
		if repos, err := c.GetRepositories(ctx); err == nil {
			for _, r := range repos {
				if r.Name == targetName {
					return &r, nil
				}
			}
		}
	}

	resp, err := c.DoRequest(ctx, http.MethodPost, "/api/v1/backupInfrastructure/repositories", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read create response: %w", err)
	}

	// 2. If 400 Bad Request returned because it "already exists", locate and adopt it
	if resp.StatusCode == http.StatusBadRequest && targetName != "" && strings.Contains(string(bodyBytes), "already exists") {
		if repos, err := c.GetRepositories(ctx); err == nil {
			for _, r := range repos {
				if r.Name == targetName {
					return &r, nil
				}
			}
		}
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("failed to create repository, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Response might be a Repository model directly
	var created Repository
	if err := json.Unmarshal(bodyBytes, &created); err == nil && created.ID != "" && created.Name != "" {
		return &created, nil
	}

	// Response is likely a SessionModel (async creation in Veeam VBR)
	var session SessionModel
	_ = json.Unmarshal(bodyBytes, &session)
	if session.ResourceId != "" {
		repo, err := c.GetRepositoryByID(ctx, session.ResourceId)
		if err == nil && repo != nil {
			return repo, nil
		}
	}

	// 3. Poll GetRepositories (and session status) for up to 60 seconds (30 retries, 2s sleep)
	if targetName != "" {
		for i := 0; i < 30; i++ {
			time.Sleep(2 * time.Second)

			// Search repositories list
			repos, err := c.GetRepositories(ctx)
			if err == nil {
				for _, r := range repos {
					if r.Name == targetName {
						return &r, nil
					}
				}
			}

			// If session ID is available, check session updates for resourceId
			if session.ID != "" {
				sessUpdate, err := c.GetSessionByID(ctx, session.ID)
				if err == nil && sessUpdate != nil && sessUpdate.ResourceId != "" {
					repo, err := c.GetRepositoryByID(ctx, sessUpdate.ResourceId)
					if err == nil && repo != nil {
						return repo, nil
					}
				}
			}
		}
	}

	if session.ResourceId != "" {
		return &Repository{
			ID:   session.ResourceId,
			Name: targetName,
		}, nil
	}

	return nil, fmt.Errorf("repository creation initiated but repository '%s' could not be located in infrastructure list within 60s timeout", targetName)
}

// DeleteRepository removes a backup repository by ID
func (c *Client) DeleteRepository(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/v1/backupInfrastructure/repositories/%s", id)
	resp, err := c.DoRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete repository %s: %w", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete repository %s, status: %d, body: %s", id, resp.StatusCode, string(bodyBytes))
	}

	return nil
}
