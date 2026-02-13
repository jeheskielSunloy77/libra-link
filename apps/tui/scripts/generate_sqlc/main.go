package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	cmd := exec.Command(
		"go",
		"run",
		"github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0",
		"generate",
		"-f",
		"sqlc.yaml",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "sqlc generation failed: %v\n", err)
		os.Exit(1)
	}
}
