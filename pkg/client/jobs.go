package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Job struct {
	ID              string                         `json:"id,omitempty"`
	Name            string                         `json:"name"`
	Description     string                         `json:"description,omitempty"`
	Type            string                         `json:"type"`
	IsDisabled      bool                           `json:"isDisabled,omitempty"`
	Storage         *BackupJobStorageModel         `json:"storage,omitempty"`
	GuestProcessing *BackupJobGuestProcessingModel `json:"guestProcessing,omitempty"`
	Schedule        *BackupScheduleModel           `json:"schedule,omitempty"`
}

type JobsResult struct {
	Data []Job `json:"data"`
}

// VMwareIncludeObject represents a VMware vSphere inventory object to back up
type VMwareIncludeObject struct {
	Platform string `json:"platform"` // "VSphere"
	Name     string `json:"name"`
	HostName string `json:"hostName"`
	Type     string `json:"type"` // "vCenterServer", "Host", "VirtualMachine", "Datacenter"
	ObjectID string `json:"objectId,omitempty"`
	URN      string `json:"urn,omitempty"`
}

type BackupJobVirtualMachinesSpec struct {
	Includes []VMwareIncludeObject `json:"includes"`
}

type BackupProxiesSettingsModel struct {
	AutoSelectEnabled bool `json:"autoSelectEnabled"`
}

type BackupJobRetentionPolicySettingsModel struct {
	Type     string `json:"type"` // "RestorePoints" or "Days"
	Quantity int    `json:"quantity"`
}

type GFSPolicySettingsWeeklyModel struct {
	IsEnabled            bool   `json:"isEnabled"`
	KeepForNumberOfWeeks int    `json:"keepForNumberOfWeeks,omitempty"`
	DesiredTime          string `json:"desiredTime,omitempty"`
}

type GFSPolicySettingsMonthlyModel struct {
	IsEnabled             bool   `json:"isEnabled"`
	KeepForNumberOfMonths int    `json:"keepForNumberOfMonths,omitempty"`
	DesiredTime           string `json:"desiredTime,omitempty"`
}

type GFSPolicySettingsYearlyModel struct {
	IsEnabled            bool   `json:"isEnabled"`
	KeepForNumberOfYears int    `json:"keepForNumberOfYears,omitempty"`
	DesiredMonth         string `json:"desiredMonth,omitempty"`
}

type GFSPolicySettingsModel struct {
	IsEnabled bool                           `json:"isEnabled"`
	Weekly    *GFSPolicySettingsWeeklyModel  `json:"weekly,omitempty"`
	Monthly   *GFSPolicySettingsMonthlyModel `json:"monthly,omitempty"`
	Yearly    *GFSPolicySettingsYearlyModel  `json:"yearly,omitempty"`
}

type BackupJobStorageModel struct {
	BackupRepositoryID string                                `json:"backupRepositoryId"`
	BackupProxies      BackupProxiesSettingsModel            `json:"backupProxies"`
	RetentionPolicy    BackupJobRetentionPolicySettingsModel `json:"retentionPolicy"`
	GFSPolicy          *GFSPolicySettingsModel               `json:"gfsPolicy,omitempty"`
}

type ScheduleDailyModel struct {
	IsEnabled bool     `json:"isEnabled"`
	LocalTime string   `json:"localTime,omitempty"` // "HH:MM"
	DailyKind string   `json:"dailyKind,omitempty"` // "Everyday", "WeekDays", "SelectedDays"
	Days      []string `json:"days,omitempty"`      // ["Monday", ...]
}

type BackupScheduleModel struct {
	RunAutomatically bool                `json:"runAutomatically"`
	Daily            *ScheduleDailyModel `json:"daily,omitempty"`
}

type BackupApplicationAwareProcessingModel struct {
	IsEnabled bool `json:"isEnabled"`
}

type GuestFileSystemIndexingModel struct {
	IsEnabled bool `json:"isEnabled"`
}

type SpecifiedGuestOsCredentialsModel struct {
	CredentialsID   string `json:"credentialsId"`
	CredentialsType string `json:"credentialsType,omitempty"`
}

type GuestOsCredentialsModel struct {
	UseAgentManagementCredentials bool                              `json:"useAgentManagementCredentials"`
	Credentials                   *SpecifiedGuestOsCredentialsModel `json:"credentials,omitempty"`
}

type BackupJobGuestProcessingModel struct {
	AppAwareProcessing BackupApplicationAwareProcessingModel `json:"appAwareProcessing"`
	GuestFSIndexing    GuestFileSystemIndexingModel          `json:"guestFSIndexing"`
	GuestCredentials   *GuestOsCredentialsModel              `json:"guestCredentials,omitempty"`
}

type CreateVSphereBackupJobSpec struct {
	ID              string                        `json:"id,omitempty"`
	Name            string                        `json:"name"`
	Description     string                        `json:"description,omitempty"`
	Type            string                        `json:"type"` // "VSphereBackup"
	IsHighPriority  bool                          `json:"isHighPriority"`
	VirtualMachines BackupJobVirtualMachinesSpec  `json:"virtualMachines"`
	Storage         BackupJobStorageModel         `json:"storage"`
	GuestProcessing BackupJobGuestProcessingModel `json:"guestProcessing"`
	Schedule        *BackupScheduleModel          `json:"schedule,omitempty"`
}

// GetJobs fetches all backup jobs from Veeam VBR
func (c *Client) GetJobs(ctx context.Context) ([]Job, error) {
	resp, err := c.DoRequest(ctx, http.MethodGet, "/api/v1/jobs", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch jobs: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read jobs response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get jobs, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var result JobsResult
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to parse jobs result: %w", err)
	}

	return result.Data, nil
}

// GetJobByID fetches a specific backup job by GUID
func (c *Client) GetJobByID(ctx context.Context, id string) (*Job, error) {
	path := fmt.Sprintf("/api/v1/jobs/%s", id)
	resp, err := c.DoRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch job %s: %w", id, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read job response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get job %s, status: %d, body: %s", id, resp.StatusCode, string(bodyBytes))
	}

	var job Job
	if err := json.Unmarshal(bodyBytes, &job); err != nil {
		return nil, fmt.Errorf("failed to parse job: %w", err)
	}

	return &job, nil
}

// CreateVSphereJob creates a new VMware backup job in Veeam VBR
func (c *Client) CreateVSphereJob(ctx context.Context, spec CreateVSphereBackupJobSpec) (*Job, error) {
	spec.Type = "VSphereBackup"
	payload, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal job payload: %w", err)
	}

	resp, err := c.DoRequest(ctx, http.MethodPost, "/api/v1/jobs", bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create backup job: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read create job response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create backup job, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var created Job
	if err := json.Unmarshal(bodyBytes, &created); err != nil {
		return nil, fmt.Errorf("failed to parse created job: %w", err)
	}

	return &created, nil
}

// UpdateVSphereJob updates an existing VMware backup job by GUID
func (c *Client) UpdateVSphereJob(ctx context.Context, id string, spec CreateVSphereBackupJobSpec) (*Job, error) {
	spec.Type = "VSphereBackup"

	type updateSpec struct {
		ID string `json:"id"`
		CreateVSphereBackupJobSpec
	}

	fullSpec := updateSpec{
		ID:                         id,
		CreateVSphereBackupJobSpec: spec,
	}

	payload, err := json.Marshal(fullSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal job update payload: %w", err)
	}

	path := fmt.Sprintf("/api/v1/jobs/%s", id)
	resp, err := c.DoRequest(ctx, http.MethodPut, path, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to update backup job %s: %w", id, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read update job response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to update backup job %s, status: %d, body: %s", id, resp.StatusCode, string(bodyBytes))
	}

	var updated Job
	if err := json.Unmarshal(bodyBytes, &updated); err != nil {
		return nil, fmt.Errorf("failed to parse updated job: %w", err)
	}

	return &updated, nil
}

// DeleteJob deletes a job by GUID
func (c *Client) DeleteJob(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/v1/jobs/%s", id)
	resp, err := c.DoRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete job %s: %w", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete job %s, status: %d, body: %s", id, resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// HyperVIncludeObject represents a Microsoft Hyper-V inventory object to back up
type HyperVIncludeObject struct {
	Platform string `json:"platform"` // "HyperV"
	Name     string `json:"name"`
	HostName string `json:"hostName,omitempty"`
	Type     string `json:"type"` // "VirtualMachine", "Cluster", "Host"
	ObjectID string `json:"objectId,omitempty"`
}

type HyperVVirtualMachinesSpec struct {
	Includes []HyperVIncludeObject `json:"includes"`
}

type CreateHyperVJobSpec struct {
	Name            string                         `json:"name"`
	Description     string                         `json:"description,omitempty"`
	Type            string                         `json:"type"` // "HyperVBackup"
	IsHighPriority  bool                           `json:"isHighPriority"`
	VirtualMachines HyperVVirtualMachinesSpec      `json:"virtualMachines"`
	Storage         BackupJobStorageModel          `json:"storage"`
	GuestProcessing *BackupJobGuestProcessingModel `json:"guestProcessing,omitempty"`
	Schedule        BackupScheduleModel            `json:"schedule"`
}

// CreateHyperVJob creates a new Microsoft Hyper-V backup job
func (c *Client) CreateHyperVJob(ctx context.Context, spec CreateHyperVJobSpec) (*Job, error) {
	if spec.Type == "" {
		spec.Type = "HyperVBackup"
	}
	for i := range spec.VirtualMachines.Includes {
		spec.VirtualMachines.Includes[i].Platform = "HyperV"
	}

	payload, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal hyper-v job spec: %w", err)
	}

	resp, err := c.DoRequest(ctx, http.MethodPost, "/api/v1/jobs", bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create hyper-v backup job: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read create hyper-v job response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create hyper-v backup job, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var created Job
	if err := json.Unmarshal(bodyBytes, &created); err != nil {
		return nil, fmt.Errorf("failed to parse created hyper-v job: %w", err)
	}

	return &created, nil
}

// UpdateHyperVJob updates an existing Microsoft Hyper-V backup job
func (c *Client) UpdateHyperVJob(ctx context.Context, id string, spec CreateHyperVJobSpec) (*Job, error) {
	if spec.Type == "" {
		spec.Type = "HyperVBackup"
	}
	for i := range spec.VirtualMachines.Includes {
		spec.VirtualMachines.Includes[i].Platform = "HyperV"
	}

	type updateSpec struct {
		ID string `json:"id"`
		CreateHyperVJobSpec
	}

	fullSpec := updateSpec{
		ID:                  id,
		CreateHyperVJobSpec: spec,
	}

	payload, err := json.Marshal(fullSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal hyper-v update spec: %w", err)
	}

	path := fmt.Sprintf("/api/v1/jobs/%s", id)
	resp, err := c.DoRequest(ctx, http.MethodPut, path, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to update hyper-v backup job %s: %w", id, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read update hyper-v job response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to update hyper-v backup job %s, status: %d, body: %s", id, resp.StatusCode, string(bodyBytes))
	}

	var updated Job
	if err := json.Unmarshal(bodyBytes, &updated); err != nil {
		return nil, fmt.Errorf("failed to parse updated hyper-v job: %w", err)
	}

	return &updated, nil
}

// File Share Job Specs
type FileBackupItemModel struct {
	Type         string `json:"type"`                   // "FileServerShare", "FileServerRoot", "Directory", "File"
	FileServerID string `json:"fileServerId"`           // GUID of Unstructured Data Server
	Path         string `json:"path,omitempty"`
}

type UnstructuredDataRetentionPolicySettingsModel struct {
	Type     string `json:"type"`     // "Days", "Months"
	Quantity int    `json:"quantity"` // e.g. 7 or 28
}

type FileBackupPrimaryRepositoryModel struct {
	BackupRepositoryID string                                        `json:"backupRepositoryId"`
	RetentionPolicy    *UnstructuredDataRetentionPolicySettingsModel `json:"retentionPolicy,omitempty"`
}

type CreateFileShareJobSpec struct {
	Name             string                           `json:"name"`
	Description      string                           `json:"description,omitempty"`
	Type             string                           `json:"type"` // "FileBackup"
	IsHighPriority   bool                             `json:"isHighPriority"`
	Objects          []FileBackupItemModel            `json:"objects"`
	BackupRepository FileBackupPrimaryRepositoryModel `json:"backupRepository"`
	Schedule         BackupScheduleModel              `json:"schedule"`
}

func (c *Client) CreateFileShareJob(ctx context.Context, spec CreateFileShareJobSpec) (*Job, error) {
	if spec.Type == "" {
		spec.Type = "FileBackup"
	}

	payload, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal file share job spec: %w", err)
	}

	resp, err := c.DoRequest(ctx, http.MethodPost, "/api/v1/jobs", bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create file share backup job: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read create file share job response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create file share job, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var created Job
	if err := json.Unmarshal(bodyBytes, &created); err != nil {
		return nil, fmt.Errorf("failed to parse created file share job: %w", err)
	}

	return &created, nil
}

func (c *Client) UpdateFileShareJob(ctx context.Context, id string, spec CreateFileShareJobSpec) (*Job, error) {
	if spec.Type == "" {
		spec.Type = "FileBackup"
	}

	type updateSpec struct {
		ID string `json:"id"`
		CreateFileShareJobSpec
	}

	fullSpec := updateSpec{
		ID:                     id,
		CreateFileShareJobSpec: spec,
	}

	payload, err := json.Marshal(fullSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal file share job spec: %w", err)
	}

	path := fmt.Sprintf("/api/v1/jobs/%s", id)
	resp, err := c.DoRequest(ctx, http.MethodPut, path, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to update file share backup job %s: %w", id, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read update file share job response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to update file share job %s, status: %d, body: %s", id, resp.StatusCode, string(bodyBytes))
	}

	var updated Job
	if err := json.Unmarshal(bodyBytes, &updated); err != nil {
		return nil, fmt.Errorf("failed to parse updated file share job: %w", err)
	}

	return &updated, nil
}

// Proxmox VE Job Specs
type ProxmoxIncludeObject struct {
	Name     string `json:"name"`
	Type     string `json:"type"` // "VirtualMachine", "ProxmoxNode", "ProxmoxCluster"
	ObjectID string `json:"objectId,omitempty"`
}

type ProxmoxScopeModel struct {
	Includes []ProxmoxIncludeObject `json:"includes"`
}

type CreateProxmoxJobSpec struct {
	Name           string                `json:"name"`
	Description    string                `json:"description,omitempty"`
	Type           string                `json:"type"` // "ProxmoxBackupJob"
	IsHighPriority bool                  `json:"isHighPriority"`
	Proxmox        ProxmoxScopeModel     `json:"proxmox"`
	Storage        BackupJobStorageModel `json:"storage"`
	Schedule       BackupScheduleModel   `json:"schedule"`
}

func (c *Client) CreateProxmoxJob(ctx context.Context, spec CreateProxmoxJobSpec) (*Job, error) {
	if spec.Type == "" {
		spec.Type = "ProxmoxBackupJob"
	}

	payload, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal proxmox job spec: %w", err)
	}

	resp, err := c.DoRequest(ctx, http.MethodPost, "/api/v1/jobs", bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create proxmox backup job: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read create proxmox job response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create proxmox backup job, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var created Job
	if err := json.Unmarshal(bodyBytes, &created); err != nil {
		return nil, fmt.Errorf("failed to parse created proxmox job: %w", err)
	}

	return &created, nil
}

// Nutanix AHV Job Specs
type NutanixIncludeObject struct {
	Name     string `json:"name"`
	Type     string `json:"type"` // "VirtualMachine", "Cluster"
	ObjectID string `json:"objectId,omitempty"`
}

type NutanixScopeModel struct {
	Includes []NutanixIncludeObject `json:"includes"`
}

type CreateNutanixJobSpec struct {
	Name           string                `json:"name"`
	Description    string                `json:"description,omitempty"`
	Type           string                `json:"type"` // "NutanixAHVBackupJob"
	IsHighPriority bool                  `json:"isHighPriority"`
	Nutanix        NutanixScopeModel     `json:"nutanix"`
	Storage        BackupJobStorageModel `json:"storage"`
	Schedule       BackupScheduleModel   `json:"schedule"`
}

func (c *Client) CreateNutanixJob(ctx context.Context, spec CreateNutanixJobSpec) (*Job, error) {
	if spec.Type == "" {
		spec.Type = "NutanixAHVBackupJob"
	}

	payload, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal nutanix job spec: %w", err)
	}

	resp, err := c.DoRequest(ctx, http.MethodPost, "/api/v1/jobs", bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create nutanix backup job: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read create nutanix job response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create nutanix backup job, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var created Job
	if err := json.Unmarshal(bodyBytes, &created); err != nil {
		return nil, fmt.Errorf("failed to parse created nutanix job: %w", err)
	}

	return &created, nil
}

// StartJob triggers an immediate execution session for a backup job by GUID
func (c *Client) StartJob(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/v1/jobs/%s/start", id)
	resp, err := c.DoRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return fmt.Errorf("failed to start job %s: %w", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to start job %s, status: %d, body: %s", id, resp.StatusCode, string(bodyBytes))
	}

	return nil
}
