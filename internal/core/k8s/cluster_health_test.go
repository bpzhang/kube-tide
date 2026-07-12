package k8s

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestWorkloadHealth(t *testing.T) {
	h := workloadHealth("coredns", 2, 2)
	if h.Status != "healthy" {
		t.Fatalf("expected healthy, got %s", h.Status)
	}

	h = workloadHealth("coredns", 1, 2)
	if h.Status != "degraded" {
		t.Fatalf("expected degraded, got %s", h.Status)
	}
}

func TestNodeReady(t *testing.T) {
	node := &corev1.Node{}
	node.Status.Conditions = []corev1.NodeCondition{
		{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
	}
	if !nodeReady(node) {
		t.Fatal("expected node ready")
	}
}
