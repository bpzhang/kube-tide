package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// Cluster represents a Kubernetes cluster
type Cluster struct {
	ID          string    `json:"id" db:"id" validate:"required,uuid"`
	Name        string    `json:"name" db:"name" validate:"required,min=1,max=100"`
	Config      string    `json:"config" db:"config" validate:"required"`
	Status      string    `json:"status" db:"status" validate:"oneof=active inactive"`
	Description string    `json:"description" db:"description"`
	Kubeconfig  string    `json:"kubeconfig,omitempty" db:"kubeconfig"`
	Endpoint    string    `json:"endpoint" db:"endpoint"`
	Version     string    `json:"version" db:"version"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ClusterStatus constants
const (
	ClusterStatusActive   = "active"
	ClusterStatusInactive = "inactive"
)

// Validate validates the cluster model
func (c *Cluster) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

// ClusterFilters represents filters for cluster queries
type ClusterFilters struct {
	Status string `json:"status" form:"status"`
	Name   string `json:"name" form:"name"`
}

// ClusterCreateRequest represents a request to create a cluster
type ClusterCreateRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Config      string `json:"config" validate:"required"`
	Description string `json:"description"`
	Kubeconfig  string `json:"kubeconfig"`
	Endpoint    string `json:"endpoint"`
}

// ClusterUpdateRequest represents a request to update a cluster
type ClusterUpdateRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Config      *string `json:"config,omitempty"`
	Status      *string `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
	Description *string `json:"description,omitempty"`
	Kubeconfig  *string `json:"kubeconfig,omitempty"`
	Endpoint    *string `json:"endpoint,omitempty"`
	Version     *string `json:"version,omitempty"`
}
