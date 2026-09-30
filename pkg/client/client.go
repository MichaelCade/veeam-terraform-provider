package client

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Client handles communication with the Veeam Backup & Replication REST API.
type Client struct {
	Endpoint   string
	Username   string
	Password   string
	APIVersion string
	Insecure   bool
	HTTPClient *http.Client

	mu           sync.RWMutex
	accessToken  string
	refreshToken string
	expiresAt    time.Time
}

// Config options for creating a Client
type Config struct {
	Endpoint   string
	Username   string
	Password   string
	APIVersion string
	Insecure   bool
}

// NewClient creates a new Veeam API Client instance
func NewClient(cfg Config) (*Client, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("veeam client endpoint cannot be empty")
	}

	endpoint := strings.TrimRight(cfg.Endpoint, "/")

	apiVersion := cfg.APIVersion
	if apiVersion == "" {
		apiVersion = "1.1-rev0" // Default API version for VBR v12+
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.Insecure,
		},
	}

	httpClient := &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}

	return &Client{
		Endpoint:   endpoint,
		Username:   cfg.Username,
		Password:   cfg.Password,
		APIVersion: apiVersion,
		Insecure:   cfg.Insecure,
		HTTPClient: httpClient,
	}, nil
}
