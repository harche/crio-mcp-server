package openshift

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

// run executes the oc command with given arguments and returns combined output.
//
// It is defined as a variable to allow tests to substitute a fake implementation
// without spawning external processes.
// Run is used by helper functions to execute the oc command. Tests may override
// this variable to avoid running external commands.
var Run = func(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "oc", args...)
	return cmd.CombinedOutput()
}

// DebugNode runs `oc debug` for the given node and command.
func DebugNode(ctx context.Context, nodeName, command string) (string, error) {
	out, err := Run(ctx, "debug", fmt.Sprintf("node/%s", nodeName), "--", "chroot", "/host", "sh", "-c", command)
	if err != nil {
		return "", fmt.Errorf("oc debug failed: %w: %s", err, out)
	}
	return string(out), nil
}

// NodeLogs runs `oc adm node-logs` for the given node and since parameter.
func NodeLogs(ctx context.Context, nodeName, since string) (string, error) {
	args := []string{"adm", "node-logs", nodeName}
	if since != "" {
		args = append(args, "--since", since)
	}
	out, err := Run(ctx, args...)
	if err != nil {
		return "", fmt.Errorf("oc adm node-logs failed: %w: %s", err, out)
	}
	return string(out), nil
}

// MustGather runs `oc adm must-gather` with optional destination directory and
// additional arguments.
func MustGather(ctx context.Context, destDir string, extra []string) (string, error) {
	args := []string{"adm", "must-gather"}
	if destDir != "" {
		args = append(args, fmt.Sprintf("--dest-dir=%s", destDir))
	}
	args = append(args, extra...)
	out, err := Run(ctx, args...)
	if err != nil {
		return "", fmt.Errorf("oc adm must-gather failed: %w: %s", err, out)
	}
	return string(out), nil
}

// SosReport collects a sosreport from the specified node using toolbox.
// If caseID is non-empty, it is passed via --case-id.
func SosReport(ctx context.Context, nodeName, caseID string) (string, error) {
	args := []string{"debug", fmt.Sprintf("node/%s", nodeName), "--", "chroot", "/host", "toolbox", "--", "sosreport", "-k", "crio.all=on", "-k", "crio.logs=on", "--batch"}
	if caseID != "" {
		args = append(args, fmt.Sprintf("--case-id=%s", caseID))
	}
	out, err := Run(ctx, args...)
	if err != nil {
		return "", fmt.Errorf("sosreport failed: %w: %s", err, out)
	}
	return string(out), nil
}

// Crictl runs `crictl` inside a debug pod on the specified node with the given arguments.
// The args slice corresponds to command-line arguments after "crictl".
func Crictl(ctx context.Context, nodeName string, args []string) (string, error) {
	cmd := fmt.Sprintf("crictl %s", strings.Join(args, " "))
	return DebugNode(ctx, nodeName, cmd)
}

// NetworkLogs runs the gather_network_logs must-gather addon.
// It accepts an optional destination directory where the results are written.
func NetworkLogs(ctx context.Context, destDir string) (string, error) {
	args := []string{"adm", "must-gather"}
	if destDir != "" {
		args = append(args, fmt.Sprintf("--dest-dir=%s", destDir))
	}
	args = append(args, "--", "/usr/bin/gather_network_logs")
	out, err := Run(ctx, args...)
	if err != nil {
		return "", fmt.Errorf("gather_network_logs failed: %w: %s", err, out)
	}
	return string(out), nil
}

// ProfilingNode collects pprof dumps from kubelet and CRI-O using gather_profiling_node.
func ProfilingNode(ctx context.Context, destDir string) (string, error) {
	args := []string{"adm", "must-gather"}
	if destDir != "" {
		args = append(args, fmt.Sprintf("--dest-dir=%s", destDir))
	}
	args = append(args, "--", "/usr/bin/gather_profiling_node")
	out, err := Run(ctx, args...)
	if err != nil {
		return "", fmt.Errorf("gather_profiling_node failed: %w: %s", err, out)
	}
	return string(out), nil
}

// Events retrieves recent cluster events across all namespaces.
func Events(ctx context.Context) (string, error) {
	out, err := Run(ctx, "get", "events", "-A")
	if err != nil {
		return "", fmt.Errorf("oc get events failed: %w: %s", err, out)
	}
	return string(out), nil
}

// PodLogs fetches logs from a specific pod and container.
// Namespace and pod name are required. Container and since are optional.
func PodLogs(ctx context.Context, namespace, pod, container, since string) (string, error) {
	args := []string{"logs", "-n", namespace, pod}
	if container != "" {
		args = append(args, "-c", container)
	}
	if since != "" {
		args = append(args, "--since", since)
	}
	out, err := Run(ctx, args...)
	if err != nil {
		return "", fmt.Errorf("oc logs failed: %w: %s", err, out)
	}
	return string(out), nil
}

// NodeConfig gathers basic node configuration like kubelet and CRI-O settings.
func NodeConfig(ctx context.Context, nodeName string) (string, error) {
	cmd := "cat /etc/kubernetes/kubelet.conf && echo --- && cat /etc/crio/crio.conf"
	return DebugNode(ctx, nodeName, cmd)
}

// CopyFilesFromNode retrieves the specified files or directories from the node
// and returns them as a gzip-compressed tar archive.
func CopyFilesFromNode(ctx context.Context, nodeName string, paths []string) ([]byte, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("no paths specified")
	}
	args := []string{"debug", fmt.Sprintf("node/%s", nodeName), "--", "chroot", "/host", "tar", "czf", "-", "--ignore-failed-read"}
	args = append(args, paths...)
	cmd := exec.CommandContext(ctx, "oc", args...)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("oc debug failed: %w: %s", err, string(ee.Stderr))
		}
		return nil, fmt.Errorf("oc debug failed: %w", err)
	}
	return out, nil
}

// NodeMetrics retrieves CPU and memory usage for all nodes using
// `oc adm top nodes`.
func NodeMetrics(ctx context.Context) (string, error) {
	out, err := Run(ctx, "adm", "top", "nodes")
	if err != nil {
		return "", fmt.Errorf("oc adm top nodes failed: %w: %s", err, out)
	}
	return string(out), nil
}

// PrometheusQuery executes a PromQL query against the in-cluster Prometheus
// service using the apiserver proxy.
func PrometheusQuery(ctx context.Context, query string) (string, error) {
	if query == "" {
		return "", fmt.Errorf("query required")
	}
	path := fmt.Sprintf(
		"/api/v1/namespaces/openshift-monitoring/services/prometheus-k8s:9091/proxy/api/v1/query?query=%s",
		url.QueryEscape(query),
	)
	out, err := Run(ctx, "get", "--raw", path)
	if err != nil {
		return "", fmt.Errorf("oc get --raw failed: %w: %s", err, out)
	}
	return string(out), nil
}
