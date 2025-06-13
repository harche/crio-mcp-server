package cri

import (
	"fmt"
	"os"
	"path/filepath"
)

// ReadContainerConfig attempts to read the container's config.json from
// well-known CRI-O overlay storage locations.
func ReadContainerConfig(id string) (string, error) {
	cleanID := filepath.Base(id)
	if cleanID != id {
		return "", fmt.Errorf("invalid container id")
	}

	dirs := []string{
		"/run/containers/storage",
		"/var/lib/containers/storage",
	}
	if runtimeDir := os.Getenv("XDG_RUNTIME_DIR"); runtimeDir != "" {
		dirs = append([]string{filepath.Join(runtimeDir, "containers/storage")}, dirs...)
	}

	for _, base := range dirs {
		path := filepath.Join(base, "overlay-containers", cleanID, "userdata", "config.json")
		data, err := os.ReadFile(path)
		if err == nil {
			return string(data), nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
	}
	return "", fmt.Errorf("container config not found")
}
