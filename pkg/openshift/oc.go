package openshift

import (
	"fmt"
	"os/exec"
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
