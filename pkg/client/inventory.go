package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type PredicateExpression struct {
	Type      string `json:"type"`      // "PredicateExpression"
	Operation string `json:"operation"` // "contains", "equals"
	Property  string `json:"property"`  // "name"
	Value     string `json:"value"`
}

type InventoryBrowserFilter struct {
	HierarchyType string               `json:"hierarchyType,omitempty"` // "VmsAndTemplates", "VmsAndTags", "HostsAndClusters"
	Filter        *PredicateExpression `json:"filter,omitempty"`
}

type InventoryBrowserResult struct {
	Data []InventoryItem `json:"data"`
}

type InventoryItem struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	HostName string `json:"hostName,omitempty"`
	ObjectID string `json:"objectId,omitempty"`
	URN      string `json:"urn,omitempty"`
}

// GetInventory items from Veeam VBR inventory browser
func (c *Client) GetInventory(ctx context.Context, hostName, hierarchyType, nameFilter string) ([]InventoryItem, error) {
	if hierarchyType == "" {
		hierarchyType = "VmsAndTemplates"
	}

	filterSpec := InventoryBrowserFilter{
		HierarchyType: hierarchyType,
	}

	if nameFilter != "" {
		filterSpec.Filter = &PredicateExpression{
			Type:      "PredicateExpression",
			Operation: "contains",
			Property:  "name",
			Value:     nameFilter,
		}
	}

	payload, err := json.Marshal(filterSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal inventory request: %w", err)
	}

	apiPath := "/api/v1/inventory"
	if hostName != "" {
		apiPath = fmt.Sprintf("/api/v1/inventory/%s", url.PathEscape(hostName))
	}

	resp, err := c.DoRequest(ctx, http.MethodPost, apiPath, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch inventory: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read inventory response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to browse inventory, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var result InventoryBrowserResult
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to parse inventory result: %w", err)
	}

	return result.Data, nil
}
