package k8s

import "testing"

func TestValidatePrometheusURL(t *testing.T) {
	tests := []struct {
		url     string
		wantErr bool
	}{
		{"", false},
		{"http://prometheus.monitoring.svc.cluster.local:9090", false},
		{"http://prometheus.monitoring.svc:9090", false},
		{"https://prometheus.example.com", false},
		{"http://localhost:9090", true},
		{"http://127.0.0.1:9090", true},
		{"http://10.0.0.5:9090", true},
		{"http://mydevice.local:9090", true},
		{"ftp://prometheus.example.com", true},
	}
	for _, tt := range tests {
		err := ValidatePrometheusURL(tt.url)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidatePrometheusURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
		}
	}
}
