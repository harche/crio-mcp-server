package openshift

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

// helper to replace run during tests
func withRunMock(f func(ctx context.Context, args ...string) ([]byte, error), test func()) {
	orig := Run
	Run = f
	defer func() { Run = orig }()
	test()
}

func TestNetworkLogs(t *testing.T) {
	expected := []string{"adm", "must-gather", "--dest-dir=test", "--", "/usr/bin/gather_network_logs"}
	withRunMock(func(ctx context.Context, args ...string) ([]byte, error) {
		if fmt.Sprint(args) != fmt.Sprint(expected) {
			t.Fatalf("unexpected args %v", args)
		}
		return []byte("logs"), nil
	}, func() {
		out, err := NetworkLogs(context.Background(), "test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out != "logs" {
			t.Fatalf("unexpected output %q", out)
		}
	})
}

func TestNetworkLogsError(t *testing.T) {
	withRunMock(func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte("bad"), errors.New("failure")
	}, func() {
		out, err := NetworkLogs(context.Background(), "")
		if err == nil {
			t.Fatal("expected error")
		}
		if out != "" {
			t.Fatalf("expected empty output, got %q", out)
		}
	})
}

func TestProfilingNode(t *testing.T) {
	expected := []string{"adm", "must-gather", "--dest-dir=/tmp", "--", "/usr/bin/gather_profiling_node"}
	withRunMock(func(ctx context.Context, args ...string) ([]byte, error) {
		if fmt.Sprint(args) != fmt.Sprint(expected) {
			t.Fatalf("unexpected args %v", args)
		}
		return []byte("prof"), nil
	}, func() {
		out, err := ProfilingNode(context.Background(), "/tmp")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out != "prof" {
			t.Fatalf("unexpected output %q", out)
		}
	})
}

func TestEvents(t *testing.T) {
	expected := []string{"get", "events", "-A"}
	withRunMock(func(ctx context.Context, args ...string) ([]byte, error) {
		if fmt.Sprint(args) != fmt.Sprint(expected) {
			t.Fatalf("unexpected args %v", args)
		}
		return []byte("events"), nil
	}, func() {
		out, err := Events(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out != "events" {
			t.Fatalf("unexpected output %q", out)
		}
	})
}

func TestPodLogs(t *testing.T) {
	expected := []string{"logs", "-n", "ns", "pod", "-c", "ctr", "--since", "2h"}
	withRunMock(func(ctx context.Context, args ...string) ([]byte, error) {
		if fmt.Sprint(args) != fmt.Sprint(expected) {
			t.Fatalf("unexpected args %v", args)
		}
		return []byte("pod logs"), nil
	}, func() {
		out, err := PodLogs(context.Background(), "ns", "pod", "ctr", "2h")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out != "pod logs" {
			t.Fatalf("unexpected output %q", out)
		}
	})
}

func TestNodeConfig(t *testing.T) {
	expected := []string{"debug", "node/testnode", "--", "chroot", "/host", "sh", "-c", "cat /etc/kubernetes/kubelet.conf && echo --- && cat /etc/crio/crio.conf"}
	withRunMock(func(ctx context.Context, args ...string) ([]byte, error) {
		if fmt.Sprint(args) != fmt.Sprint(expected) {
			t.Fatalf("unexpected args %v", args)
		}
		return []byte("cfg"), nil
	}, func() {
		out, err := NodeConfig(context.Background(), "testnode")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out != "cfg" {
			t.Fatalf("unexpected output %q", out)
		}
	})
}

func TestNodeMetrics(t *testing.T) {
	expected := []string{"adm", "top", "nodes"}
	withRunMock(func(ctx context.Context, args ...string) ([]byte, error) {
		if fmt.Sprint(args) != fmt.Sprint(expected) {
			t.Fatalf("unexpected args %v", args)
		}
		return []byte("metrics"), nil
	}, func() {
		out, err := NodeMetrics(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out != "metrics" {
			t.Fatalf("unexpected output %q", out)
		}
	})
}

func TestPrometheusQuery(t *testing.T) {
	expected := []string{"get", "--raw", "/api/v1/namespaces/openshift-monitoring/services/prometheus-k8s:9091/proxy/api/v1/query?query=up"}
	withRunMock(func(ctx context.Context, args ...string) ([]byte, error) {
		if fmt.Sprint(args) != fmt.Sprint(expected) {
			t.Fatalf("unexpected args %v", args)
		}
		return []byte("{\"status\":\"success\"}"), nil
	}, func() {
		out, err := PrometheusQuery(context.Background(), "up")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out != "{\"status\":\"success\"}" {
			t.Fatalf("unexpected output %q", out)
		}
	})
}
