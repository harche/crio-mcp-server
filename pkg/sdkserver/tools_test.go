package sdkserver

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/harche/crio-mcp-server/pkg/openshift"
	"github.com/harche/crio-mcp-server/pkg/redhat"
	mcp "github.com/mark3labs/mcp-go/mcp"
)

func withRunMock(t *testing.T, expected []string, output string, err error, f func()) {
	orig := openshift.Run
	openshift.Run = func(args ...string) ([]byte, error) {
		if fmt.Sprint(args) != fmt.Sprint(expected) {
			t.Fatalf("unexpected args %v", args)
		}
		return []byte(output), err
	}
	defer func() { openshift.Run = orig }()
	f()
}

func text(result *mcp.CallToolResult) string {
	if len(result.Content) == 0 {
		return ""
	}
	return result.Content[0].(mcp.TextContent).Text
}

func TestHandleDebugNode(t *testing.T) {
	expectedArgs := []string{"debug", "node/test", "--", "chroot", "/host", "sh", "-c", "echo hi"}
	withRunMock(t, expectedArgs, "hi", nil, func() {
		req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
			"node_name": "test",
			"commands":  []any{"echo hi"},
		}}}
		res, err := handleDebugNode(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.IsError {
			t.Fatalf("unexpected error result: %v", text(res))
		}
		if text(res) != "hi" {
			t.Fatalf("unexpected result %q", text(res))
		}
	})
}

func TestHandleDebugNodeMissingArg(t *testing.T) {
	req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{}}}
	res, err := handleDebugNode(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected error result")
	}
}

func TestHandleNodeLogs(t *testing.T) {
	args := []string{"adm", "node-logs", "node1", "--since", "1h"}
	withRunMock(t, args, "logs", nil, func() {
		req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
			"node_name": "node1",
			"since":     "1h",
		}}}
		res, err := handleNodeLogs(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.IsError || text(res) != "logs" {
			t.Fatalf("unexpected result: %v", text(res))
		}
	})
}

func TestHandlePprofArgsRequired(t *testing.T) {
	req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{}}}
	res, err := handlePprof(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected error result")
	}
}

func TestHandlePprof(t *testing.T) {
	req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
		"args": []any{"-h"},
	}}}
	res, err := handlePprof(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error result: %v", text(res))
	}
	if text(res) == "" {
		t.Fatalf("expected output")
	}
}

func TestHandleMustGather(t *testing.T) {
	args := []string{"adm", "must-gather", "--dest-dir=/tmp", "--foo"}
	withRunMock(t, args, "out", nil, func() {
		req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
			"dest_dir":   "/tmp",
			"extra_args": []any{"--foo"},
		}}}
		res, err := handleMustGather(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.IsError || text(res) != "out" {
			t.Fatalf("unexpected result: %v", text(res))
		}
	})
}

func TestHandleCrictl(t *testing.T) {
	args := []string{"debug", "node/n1", "--", "chroot", "/host", "sh", "-c", "crictl ps"}
	withRunMock(t, args, "ok", nil, func() {
		req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
			"node_name": "n1",
		}}}
		res, err := handleCrictl(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.IsError || text(res) != "ok" {
			t.Fatalf("unexpected result: %v", text(res))
		}
	})
}

func TestHandleTraverseCgroupfs(t *testing.T) {
	script := "find /sys/fs/cgroup/kubepods.slice -name memory.current | xargs grep -H ."
	args := []string{"debug", "node/n1", "--", "chroot", "/host", "sh", "-c", script}
	withRunMock(t, args, "data", nil, func() {
		req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
			"node_name": "n1",
		}}}
		res, err := handleTraverseCgroupfs(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.IsError || text(res) != "data" {
			t.Fatalf("unexpected result: %v", text(res))
		}
	})
}

func TestHandleSosReport(t *testing.T) {
	args := []string{"debug", "node/n1", "--", "chroot", "/host", "toolbox", "--", "sosreport", "-k", "crio.all=on", "-k", "crio.logs=on", "--batch"}
	withRunMock(t, args, "sos", nil, func() {
		req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
			"node_name": "n1",
		}}}
		res, err := handleSosReport(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.IsError || text(res) != "sos" {
			t.Fatalf("unexpected result: %v", text(res))
		}
	})
}

func TestHandleNetworkLogs(t *testing.T) {
	args := []string{"adm", "must-gather", "--", "/usr/bin/gather_network_logs"}
	withRunMock(t, args, "net", nil, func() {
		req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{}}}
		res, err := handleNetworkLogs(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.IsError || text(res) != "net" {
			t.Fatalf("unexpected result: %v", text(res))
		}
	})
}

func TestHandleProfilingNode(t *testing.T) {
	args := []string{"adm", "must-gather", "--dest-dir=/tmp", "--", "/usr/bin/gather_profiling_node"}
	withRunMock(t, args, "prof", nil, func() {
		req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
			"dest_dir": "/tmp",
		}}}
		res, err := handleProfilingNode(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.IsError || text(res) != "prof" {
			t.Fatalf("unexpected result: %v", text(res))
		}
	})
}

func TestHandleEvents(t *testing.T) {
	args := []string{"get", "events", "-A"}
	withRunMock(t, args, "ev", nil, func() {
		req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{}}}
		res, err := handleEvents(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.IsError || text(res) != "ev" {
			t.Fatalf("unexpected result: %v", text(res))
		}
	})
}

func TestHandleNodeMetrics(t *testing.T) {
	args := []string{"adm", "top", "nodes"}
	withRunMock(t, args, "metrics", nil, func() {
		req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{}}}
		res, err := handleNodeMetrics(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.IsError || text(res) != "metrics" {
			t.Fatalf("unexpected result: %v", text(res))
		}
	})
}

func TestHandlePrometheusQuery(t *testing.T) {
	args := []string{"get", "--raw", "/api/v1/namespaces/openshift-monitoring/services/prometheus-k8s:9091/proxy/api/v1/query?query=up"}
	withRunMock(t, args, "{\"status\":\"success\"}", nil, func() {
		req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
			"query": "up",
		}}}
		res, err := handlePrometheusQuery(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.IsError || text(res) != "{\"status\":\"success\"}" {
			t.Fatalf("unexpected result: %v", text(res))
		}
	})
}

func TestHandlePodLogs(t *testing.T) {
	args := []string{"logs", "-n", "ns", "pod", "-c", "ctr", "--since", "1m"}
	withRunMock(t, args, "p", nil, func() {
		req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
			"namespace": "ns",
			"pod_name":  "pod",
			"container": "ctr",
			"since":     "1m",
		}}}
		res, err := handlePodLogs(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.IsError || text(res) != "p" {
			t.Fatalf("unexpected result: %v", text(res))
		}
	})
}

func TestHandleNodeConfig(t *testing.T) {
	args := []string{"debug", "node/n1", "--", "chroot", "/host", "sh", "-c", "cat /etc/kubernetes/kubelet.conf && echo --- && cat /etc/crio/crio.conf"}
	withRunMock(t, args, "cfg", nil, func() {
		req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
			"node_name": "n1",
		}}}
		res, err := handleNodeConfig(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.IsError || text(res) != "cfg" {
			t.Fatalf("unexpected result: %v", text(res))
		}
	})
}

func TestHandleSearchKCS(t *testing.T) {
	calls := 0
	redhat.Do = func(req *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"access_token":"tok"}`))}, nil
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("result"))}, nil
	}
	defer func() { redhat.Do = http.DefaultClient.Do }()

	req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
		"query":         "bug",
		"offline_token": "off",
	}}}
	res, err := handleSearchKCS(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.IsError || text(res) != "result" {
		t.Fatalf("unexpected result: %v", text(res))
	}
}

func TestHandleCVEInfo(t *testing.T) {
	redhat.Do = func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("info"))}, nil
	}
	defer func() { redhat.Do = http.DefaultClient.Do }()
	req := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
		"cve_id": "CVE-1234",
	}}}
	res, err := handleCVEInfo(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.IsError || text(res) != "info" {
		t.Fatalf("unexpected result: %v", text(res))
	}
}
