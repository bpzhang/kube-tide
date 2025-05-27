package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// Namespace represents a Kubernetes namespace
type Namespace struct {
	ID          string    `json:"id" db:"id" validate:"required,uuid"`
	ClusterID   string    `json:"cluster_id" db:"cluster_id" validate:"required,uuid"`
	Name        string    `json:"name" db:"name" validate:"required,min=1,max=100"`
	Status      string    `json:"status" db:"status" validate:"required"`
	Phase       string    `json:"phase" db:"phase"`
	Labels      string    `json:"labels" db:"labels"`
	Annotations string    `json:"annotations" db:"annotations"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// NamespacePhase constants
const (
	NamespacePhaseActive      = "Active"
	NamespacePhaseTerminating = "Terminating"
)

// NamespaceStatus constants
const (
	NamespaceStatusActive      = "Active"
	NamespaceStatusTerminating = "Terminating"
)

// Validate validates the namespace model
func (n *Namespace) Validate() error {
	validate := validator.New()
	return validate.Struct(n)
}

// NamespaceUpdateRequest represents a request to update a namespace
type NamespaceUpdateRequest struct {
	Status      *string `json:"status,omitempty"`
	Phase       *string `json:"phase,omitempty"`
	Labels      *string `json:"labels,omitempty"`
	Annotations *string `json:"annotations,omitempty"`
}
