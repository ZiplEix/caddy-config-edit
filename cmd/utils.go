package cmd

import (
	"os"
	"os/exec"
	"unicode"
)

// isSafeFilename ensures no path separators or control chars are present.
func isSafeFilename(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r == '/' || r == '\\' || r == 0 || unicode.IsControl(r) {
			return false
		}
	}
	return true
}

// run executes a command streaming stdout/stderr to the current process.
func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
