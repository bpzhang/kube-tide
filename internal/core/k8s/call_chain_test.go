package k8s

import "testing"

func TestSplitHubbleEndpoint(t *testing.T) {
	ns, name := splitHubbleEndpoint("default/nginx-abc123")
	if ns != "default" || name != "nginx-abc123" {
		t.Fatalf("unexpected: %s/%s", ns, name)
	}
	ns, name = splitHubbleEndpoint("coredns")
	if ns != "" || name != "coredns" {
		t.Fatalf("unexpected bare name: %s/%s", ns, name)
	}
}

func TestTrimPodHash(t *testing.T) {
	if got := trimPodHash("nginx-7d4f8b9c6d-xk2mz"); got != "nginx" {
		t.Fatalf("expected nginx, got %s", got)
	}
	if got := trimPodHash("simple-name"); got != "simple-name" {
		t.Fatalf("expected unchanged, got %s", got)
	}
}
