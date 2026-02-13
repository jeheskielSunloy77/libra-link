package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	specPath := filepath.Clean("../api/static/openapi.json")
	sanitizedPath, err := sanitizeSpec(specPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sanitize openapi spec failed: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(sanitizedPath)

	cmd := exec.Command(
		"go",
		"run",
		"github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1",
		"-config",
		"oapi-codegen.yaml",
		sanitizedPath,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "openapi client generation failed: %v\n", err)
		os.Exit(1)
	}
}

func sanitizeSpec(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var payload any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return "", err
	}

	normalizeExclusiveBounds(payload)

	tmpFile, err := os.CreateTemp("", "libra-link-openapi-*.json")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	encoder := json.NewEncoder(tmpFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(payload); err != nil {
		return "", err
	}
	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

func normalizeExclusiveBounds(node any) {
	switch value := node.(type) {
	case map[string]any:
		if exclusiveMin, ok := value["exclusiveMinimum"]; ok {
			if n, isNumber := exclusiveMin.(float64); isNumber {
				if _, hasMinimum := value["minimum"]; !hasMinimum {
					value["minimum"] = n
				}
				value["exclusiveMinimum"] = true
			}
		}
		if exclusiveMax, ok := value["exclusiveMaximum"]; ok {
			if n, isNumber := exclusiveMax.(float64); isNumber {
				if _, hasMaximum := value["maximum"]; !hasMaximum {
					value["maximum"] = n
				}
				value["exclusiveMaximum"] = true
			}
		}
		for _, child := range value {
			normalizeExclusiveBounds(child)
		}
	case []any:
		for _, child := range value {
			normalizeExclusiveBounds(child)
		}
	}
}
