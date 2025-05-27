package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// Pod represents a Kubernetes pod
type Pod struct {
	ID              string    `json:"id" db:"id" validate:"required,uuid"`
	ClusterID       string    `json:"cluster_id" db:"cluster_id" validate:"required,uuid"`
	Namespace       string    `json:"namespace" db:"namespace" validate:"required,min=1,max=100"`
	Name            string    `json:"name" db:"name" validate:"required,min=1,max=100"`
	Status          string    `json:"status" db:"status" validate:"required"`
	Phase           string    `json:"phase" db:"phase"`
	NodeName        string    `json:"node_name" db:"node_name"`
	PodIP           string    `json:"pod_ip" db:"pod_ip"`
	HostIP          string    `json:"host_ip" db:"host_ip"`
	RestartCount    int       `json:"restart_count" db:"restart_count"`
	ReadyContainers int       `json:"ready_containers" db:"ready_containers"`
	TotalContainers int       `json:"total_containers" db:"total_containers"`
	CPURequests     string    `json:"cpu_requests" db:"cpu_requests"`
	MemoryRequests  string    `json:"memory_requests" db:"memory_requests"`
	CPULimits       string    `json:"cpu_limits" db:"cpu_limits"`
	MemoryLimits    string    `json:"memory_limits" db:"memory_limits"`
	Labels          string    `json:"labels" db:"labels"`
	Annotations     string    `json:"annotations" db:"annotations"`
	OwnerReferences string    `json:"owner_references" db:"owner_references"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// PodPhase constants
const (
	PodPhasePending   = "Pending"
	PodPhaseRunning   = "Running"
	PodPhaseSucceeded = "Succeeded"
	PodPhaseFailed    = "Failed"
	PodPhaseUnknown   = "Unknown"
)

// PodStatus constants
const (
	PodStatusRunning     = "Running"
	PodStatusPending     = "Pending"
	PodStatusSucceeded   = "Succeeded"
	PodStatusFailed      = "Failed"
	PodStatusUnknown     = "Unknown"
	PodStatusTerminating = "Terminating"
)

// Validate validates the pod model
func (p *Pod) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// PodUpdateRequest represents a request to update a pod
type PodUpdateRequest struct {
	Status          *string `json:"status,omitempty"`
	Phase           *string `json:"phase,omitempty"`
	NodeName        *string `json:"node_name,omitempty"`
	PodIP           *string `json:"pod_ip,omitempty"`
	HostIP          *string `json:"host_ip,omitempty"`
	RestartCount    *int    `json:"restart_count,omitempty"`
	ReadyContainers *int    `json:"ready_containers,omitempty"`
	TotalContainers *int    `json:"total_containers,omitempty"`
	CPURequests     *string `json:"cpu_requests,omitempty"`
	MemoryRequests  *string `json:"memory_requests,omitempty"`
	CPULimits       *string `json:"cpu_limits,omitempty"`
	MemoryLimits    *string `json:"memory_limits,omitempty"`
	Labels          *string `json:"labels,omitempty"`
	Annotations     *string `json:"annotations,omitempty"`
	OwnerReferences *string `json:"owner_references,omitempty"`
}
