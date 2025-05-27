package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// Service represents a Kubernetes service
type Service struct {
	ID              string    `json:"id" db:"id" validate:"required,uuid"`
	ClusterID       string    `json:"cluster_id" db:"cluster_id" validate:"required,uuid"`
	Namespace       string    `json:"namespace" db:"namespace" validate:"required,min=1,max=100"`
	Name            string    `json:"name" db:"name" validate:"required,min=1,max=100"`
	Type            string    `json:"type" db:"type" validate:"required"`
	ClusterIP       string    `json:"cluster_ip" db:"cluster_ip"`
	ExternalIPs     string    `json:"external_ips" db:"external_ips"`
	Ports           string    `json:"ports" db:"ports"`
	Selector        string    `json:"selector" db:"selector"`
	SessionAffinity string    `json:"session_affinity" db:"session_affinity"`
	Labels          string    `json:"labels" db:"labels"`
	Annotations     string    `json:"annotations" db:"annotations"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// ServiceType constants
const (
	ServiceTypeClusterIP    = "ClusterIP"
	ServiceTypeNodePort     = "NodePort"
	ServiceTypeLoadBalancer = "LoadBalancer"
	ServiceTypeExternalName = "ExternalName"
)

// SessionAffinity constants
const (
	SessionAffinityNone     = "None"
	SessionAffinityClientIP = "ClientIP"
)

// Validate validates the service model
func (s *Service) Validate() error {
	validate := validator.New()
	return validate.Struct(s)
}

// ServiceUpdateRequest represents a request to update a service
type ServiceUpdateRequest struct {
	Type            *string `json:"type,omitempty"`
	ClusterIP       *string `json:"cluster_ip,omitempty"`
	ExternalIPs     *string `json:"external_ips,omitempty"`
	Ports           *string `json:"ports,omitempty"`
	Selector        *string `json:"selector,omitempty"`
	SessionAffinity *string `json:"session_affinity,omitempty"`
	Labels          *string `json:"labels,omitempty"`
	Annotations     *string `json:"annotations,omitempty"`
}
