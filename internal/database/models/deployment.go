package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// Deployment represents a Kubernetes deployment
type Deployment struct {
	ID                  string    `json:"id" db:"id" validate:"required,uuid"`
	ClusterID           string    `json:"cluster_id" db:"cluster_id" validate:"required,uuid"`
	Namespace           string    `json:"namespace" db:"namespace" validate:"required,min=1,max=100"`
	Name                string    `json:"name" db:"name" validate:"required,min=1,max=100"`
	Replicas            int       `json:"replicas" db:"replicas"`
	ReadyReplicas       int       `json:"ready_replicas" db:"ready_replicas"`
	AvailableReplicas   int       `json:"available_replicas" db:"available_replicas"`
	UnavailableReplicas int       `json:"unavailable_replicas" db:"unavailable_replicas"`
	UpdatedReplicas     int       `json:"updated_replicas" db:"updated_replicas"`
	StrategyType        string    `json:"strategy_type" db:"strategy_type"`
	Labels              string    `json:"labels" db:"labels"`
	Annotations         string    `json:"annotations" db:"annotations"`
	Selector            string    `json:"selector" db:"selector"`
	Template            string    `json:"template" db:"template"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

// DeploymentStrategyType constants
const (
	DeploymentStrategyRecreate      = "Recreate"
	DeploymentStrategyRollingUpdate = "RollingUpdate"
)

// Validate validates the deployment model
func (d *Deployment) Validate() error {
	validate := validator.New()
	return validate.Struct(d)
}

// DeploymentUpdateRequest represents a request to update a deployment
type DeploymentUpdateRequest struct {
	Replicas            *int    `json:"replicas,omitempty"`
	ReadyReplicas       *int    `json:"ready_replicas,omitempty"`
	AvailableReplicas   *int    `json:"available_replicas,omitempty"`
	UnavailableReplicas *int    `json:"unavailable_replicas,omitempty"`
	UpdatedReplicas     *int    `json:"updated_replicas,omitempty"`
	StrategyType        *string `json:"strategy_type,omitempty"`
	Labels              *string `json:"labels,omitempty"`
	Annotations         *string `json:"annotations,omitempty"`
	Selector            *string `json:"selector,omitempty"`
	Template            *string `json:"template,omitempty"`
}
