package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// Node represents a Kubernetes node
type Node struct {
	ID                string    `json:"id" db:"id" validate:"required,uuid"`
	ClusterID         string    `json:"cluster_id" db:"cluster_id" validate:"required,uuid"`
	Name              string    `json:"name" db:"name" validate:"required,min=1,max=100"`
	Status            string    `json:"status" db:"status" validate:"required"`
	Roles             string    `json:"roles" db:"roles"`
	Age               string    `json:"age" db:"age"`
	Version           string    `json:"version" db:"version"`
	InternalIP        string    `json:"internal_ip" db:"internal_ip"`
	ExternalIP        string    `json:"external_ip" db:"external_ip"`
	OSImage           string    `json:"os_image" db:"os_image"`
	KernelVersion     string    `json:"kernel_version" db:"kernel_version"`
	ContainerRuntime  string    `json:"container_runtime" db:"container_runtime"`
	CPUCapacity       string    `json:"cpu_capacity" db:"cpu_capacity"`
	MemoryCapacity    string    `json:"memory_capacity" db:"memory_capacity"`
	CPUAllocatable    string    `json:"cpu_allocatable" db:"cpu_allocatable"`
	MemoryAllocatable string    `json:"memory_allocatable" db:"memory_allocatable"`
	Conditions        string    `json:"conditions" db:"conditions"`
	Labels            string    `json:"labels" db:"labels"`
	Annotations       string    `json:"annotations" db:"annotations"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// NodeStatus constants
const (
	NodeStatusReady    = "Ready"
	NodeStatusNotReady = "NotReady"
	NodeStatusUnknown  = "Unknown"
)

// Validate validates the node model
func (n *Node) Validate() error {
	validate := validator.New()
	return validate.Struct(n)
}

// NodeUpdateRequest represents a request to update a node
type NodeUpdateRequest struct {
	Status            *string `json:"status,omitempty"`
	Roles             *string `json:"roles,omitempty"`
	Age               *string `json:"age,omitempty"`
	Version           *string `json:"version,omitempty"`
	InternalIP        *string `json:"internal_ip,omitempty"`
	ExternalIP        *string `json:"external_ip,omitempty"`
	OSImage           *string `json:"os_image,omitempty"`
	KernelVersion     *string `json:"kernel_version,omitempty"`
	ContainerRuntime  *string `json:"container_runtime,omitempty"`
	CPUCapacity       *string `json:"cpu_capacity,omitempty"`
	MemoryCapacity    *string `json:"memory_capacity,omitempty"`
	CPUAllocatable    *string `json:"cpu_allocatable,omitempty"`
	MemoryAllocatable *string `json:"memory_allocatable,omitempty"`
	Conditions        *string `json:"conditions,omitempty"`
	Labels            *string `json:"labels,omitempty"`
	Annotations       *string `json:"annotations,omitempty"`
}
