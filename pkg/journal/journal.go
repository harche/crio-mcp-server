package journal

import (
	"bufio"
	"fmt"
	"os/exec"
	"time"
)

// Options controls how logs are retrieved.
type Options struct {
	Unit        string
	ContainerID string
	Since       time.Time
	Until       time.Time
	Priority    string
	Limit       int
	Follow      bool
}

// buildArgs converts Options into journalctl command arguments.
func buildArgs(opt Options) []string {
	args := []string{"-o", "cat"}
	if opt.Unit != "" {
		args = append(args, "-u", opt.Unit)
	}
	if opt.ContainerID != "" {
		args = append(args, fmt.Sprintf("CONTAINER_ID_FULL=%s", opt.ContainerID))
	}
	if !opt.Since.IsZero() {
		args = append(args, "--since", opt.Since.Format(time.RFC3339))
	}
	if !opt.Until.IsZero() {
		args = append(args, "--until", opt.Until.Format(time.RFC3339))
	}
	if opt.Priority != "" {
		args = append(args, "-p", opt.Priority)
	}
	if opt.Limit > 0 {
		args = append(args, "-n", fmt.Sprint(opt.Limit))
	}
	if opt.Follow {
		args = append(args, "-f")
	}
	return args
}

// ReadLogs fetches logs using journalctl and returns them as a single string.
func ReadLogs(opt Options) (string, error) {
	cmd := exec.Command("journalctl", buildArgs(opt)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, out)
	}
	return string(out), nil
}

// StreamLogs streams logs continuously using journalctl -f.
func StreamLogs(opt Options, fn func(string) error) error {
	opt.Follow = true
	cmd := exec.Command("journalctl", buildArgs(opt)...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		if err := fn(scanner.Text()); err != nil {
			cmd.Process.Kill()
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return cmd.Wait()
}
