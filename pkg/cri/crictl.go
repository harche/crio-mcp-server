package cri

import (
	"fmt"
	"os/exec"
)

// Crictl wraps the crictl CLI for simple interactions with a CRI implementation.
type Crictl struct {
	path string
}

// New returns a new Crictl helper. If path is empty, "crictl" is used.
func New(path string) *Crictl {
	if path == "" {
		path = "crictl"
	}
	return &Crictl{path: path}
}

func (c *Crictl) run(args ...string) ([]byte, error) {
	cmd := exec.Command(c.path, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, out)
	}
	return out, nil
}

// RuntimeStatus returns the JSON output of `crictl info`.
func (c *Crictl) RuntimeStatus() (string, error) {
	out, err := c.run("info")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// ListContainers returns the JSON output of `crictl ps -o json`.
func (c *Crictl) ListContainers() (string, error) {
	out, err := c.run("ps", "-o", "json")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// InspectContainer returns the JSON output of `crictl inspect -o json <id>`.
func (c *Crictl) InspectContainer(id string) (string, error) {
	out, err := c.run("inspect", "-o", "json", id)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
