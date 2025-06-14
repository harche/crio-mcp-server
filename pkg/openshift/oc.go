package openshift

import (
	"fmt"
	"os/exec"
	"strings"
)

// run executes the oc command with given arguments and returns combined output.
func run(args ...string) ([]byte, error) {
	cmd := exec.Command("oc", args...)
	return cmd.CombinedOutput()
}

// DebugNode runs `oc debug` for the given node and command.
func DebugNode(nodeName, command string) (string, error) {
	out, err := run("debug", fmt.Sprintf("node/%s", nodeName), "--", "chroot", "/host", "sh", "-c", command)
	if err != nil {
		return "", fmt.Errorf("oc debug failed: %w: %s", err, out)
	}
	return string(out), nil
}

// NodeLogs runs `oc adm node-logs` for the given node and since parameter.
func NodeLogs(nodeName, since string) (string, error) {
	args := []string{"adm", "node-logs", nodeName}
	if since != "" {
		args = append(args, "--since", since)
	}
	out, err := run(args...)
	if err != nil {
		return "", fmt.Errorf("oc adm node-logs failed: %w: %s", err, out)
	}
	return string(out), nil
}

// MustGather runs `oc adm must-gather` with optional destination directory and
// additional arguments.
func MustGather(destDir string, extra []string) (string, error) {
	args := []string{"adm", "must-gather"}
	if destDir != "" {
		args = append(args, fmt.Sprintf("--dest-dir=%s", destDir))
	}
	args = append(args, extra...)
	out, err := run(args...)
	if err != nil {
		return "", fmt.Errorf("oc adm must-gather failed: %w: %s", err, out)
	}
	return string(out), nil
}

// SosReport collects a sosreport from the specified node using toolbox.
// If caseID is non-empty, it is passed via --case-id.
func SosReport(nodeName, caseID string) (string, error) {
	args := []string{"debug", fmt.Sprintf("node/%s", nodeName), "--", "chroot", "/host", "toolbox", "--", "sosreport", "-k", "crio.all=on", "-k", "crio.logs=on", "--batch"}
	if caseID != "" {
		args = append(args, fmt.Sprintf("--case-id=%s", caseID))
	}
	out, err := run(args...)
	if err != nil {
		return "", fmt.Errorf("sosreport failed: %w: %s", err, out)
	}
	return string(out), nil
}

// Crictl runs `crictl` inside a debug pod on the specified node with the given arguments.
// The args slice corresponds to command-line arguments after "crictl".
func Crictl(nodeName string, args []string) (string, error) {
	cmd := fmt.Sprintf("crictl %s", strings.Join(args, " "))
	return DebugNode(nodeName, cmd)
}

// NetworkLogs runs the gather_network_logs must-gather addon.
// It accepts an optional destination directory where the results are written.
func NetworkLogs(destDir string) (string, error) {
	args := []string{"adm", "must-gather"}
	if destDir != "" {
		args = append(args, fmt.Sprintf("--dest-dir=%s", destDir))
	}
	args = append(args, "--", "/usr/bin/gather_network_logs")
	out, err := run(args...)
	if err != nil {
		return "", fmt.Errorf("gather_network_logs failed: %w: %s", err, out)
	}
	return string(out), nil
}

// ProfilingNode collects pprof dumps from kubelet and CRI-O using gather_profiling_node.
func ProfilingNode(destDir string) (string, error) {
	args := []string{"adm", "must-gather"}
	if destDir != "" {
		args = append(args, fmt.Sprintf("--dest-dir=%s", destDir))
	}
	args = append(args, "--", "/usr/bin/gather_profiling_node")
	out, err := run(args...)
	if err != nil {
		return "", fmt.Errorf("gather_profiling_node failed: %w: %s", err, out)
	}
	return string(out), nil
}

// Events retrieves recent cluster events across all namespaces.
func Events() (string, error) {
	out, err := run("get", "events", "-A")
	if err != nil {
		return "", fmt.Errorf("oc get events failed: %w: %s", err, out)
	}
	return string(out), nil
}

// PodLogs fetches logs from a specific pod and container.
// Namespace and pod name are required. Container and since are optional.
func PodLogs(namespace, pod, container, since string) (string, error) {
	args := []string{"logs", "-n", namespace, pod}
	if container != "" {
		args = append(args, "-c", container)
	}
	if since != "" {
		args = append(args, "--since", since)
	}
	out, err := run(args...)
	if err != nil {
		return "", fmt.Errorf("oc logs failed: %w: %s", err, out)
	}
	return string(out), nil
}

// NodeConfig gathers basic node configuration like kubelet and CRI-O settings.
func NodeConfig(nodeName string) (string, error) {
	cmd := "cat /etc/kubernetes/kubelet.conf && echo --- && cat /etc/crio/crio.conf"
	return DebugNode(nodeName, cmd)
}
