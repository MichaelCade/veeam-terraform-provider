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

type ScaleOutExtentSpec struct {
	ID string `json:"id"`
}

type PerformanceTierSpec struct {
	PerformanceExtents []ScaleOutExtentSpec `json:"performanceExtents"`
}

type ScaleOutRepository struct {
	ID              string              `json:"id"`
	Name            string              `json:"name"`
	Description     string              `json:"description"`
	PerformanceTier PerformanceTierSpec `json:"performanceTier"`
}

type ScaleOutRepositoryListResponse struct {
	Data []ScaleOutRepository `json:"data"`
}

type CreateScaleOutRepositorySpec struct {
	Name            string              `json:"name"`
	Description     string              `json:"description,omitempty"`
	PerformanceTier PerformanceTierSpec `json:"performanceTier"`
}

// GetScaleOutRepositories retrieves all Scale-Out Backup Repositories (SOBR)
func (c *Client) GetScaleOutRepositories(ctx context.Context) ([]ScaleOutRepository, error) {
	resp, err := c.DoRequest(ctx, http.MethodGet, "/api/v1/backupInfrastructure/scaleOutRepositories", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch scale-out repositories: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get scale-out repositories, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var listResp ScaleOutRepositoryListResponse
	if err := json.Unmarshal(bodyBytes, &listResp); err != nil {
		return nil, fmt.Errorf("failed to parse scale-out repository response: %w", err)
	}

	return listResp.Data, nil
}

// GetScaleOutRepositoryByID retrieves a single SOBR by ID
func (c *Client) GetScaleOutRepositoryByID(ctx context.Context, id string) (*ScaleOutRepository, error) {
	path := fmt.Sprintf("/api/v1/backupInfrastructure/scaleOutRepositories/%s", id)
	resp, err := c.DoRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch scale-out repository %s: %w", id, err)
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
		return nil, fmt.Errorf("failed to get scale-out repository %s, status: %d, body: %s", id, resp.StatusCode, string(bodyBytes))
	}

	var sobr ScaleOutRepository
	if err := json.Unmarshal(bodyBytes, &sobr); err != nil {
		return nil, fmt.Errorf("failed to parse scale-out repository: %w", err)
	}

	return &sobr, nil
}

// CreateScaleOutRepository creates a new Scale-Out Backup Repository (SOBR)
func (c *Client) CreateScaleOutRepository(ctx context.Context, spec CreateScaleOutRepositorySpec) (*ScaleOutRepository, error) {
	jsonBytes, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal scale-out repository payload: %w", err)
	}

	targetName := spec.Name

	resp, err := c.DoRequest(ctx, http.MethodPost, "/api/v1/backupInfrastructure/scaleOutRepositories", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create scale-out repository: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("failed to create scale-out repository, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var created ScaleOutRepository
	if err := json.Unmarshal(bodyBytes, &created); err == nil && created.ID != "" && created.Name != "" {
		return &created, nil
	}

	var session struct {
		ID         string `json:"id"`
		ResourceId string `json:"resourceId"`
	}
	if err := json.Unmarshal(bodyBytes, &session); err == nil && session.ResourceId != "" {
		sobr, err := c.GetScaleOutRepositoryByID(ctx, session.ResourceId)
		if err == nil && sobr != nil {
			return sobr, nil
		}
	}

	if targetName != "" {
		for i := 0; i < 15; i++ {
			time.Sleep(1 * time.Second)
			sobrs, err := c.GetScaleOutRepositories(ctx)
			if err == nil {
				for _, s := range sobrs {
					if s.Name == targetName {
						return &s, nil
					}
				}
			}
		}
	}

	if session.ResourceId != "" {
		return &ScaleOutRepository{
			ID:   session.ResourceId,
			Name: targetName,
		}, nil
	}

	return nil, fmt.Errorf("scale-out repository creation initiated but '%s' could not be located in infrastructure list", targetName)
}

// DeleteScaleOutRepository removes a Scale-Out Backup Repository by ID
func (c *Client) DeleteScaleOutRepository(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/v1/backupInfrastructure/scaleOutRepositories/%s", id)
	resp, err := c.DoRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete scale-out repository %s: %w", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete scale-out repository %s, status: %d, body: %s", id, resp.StatusCode, string(bodyBytes))
	}

	return nil
}
