package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var (
	newEntryDir   string
	newEntryExt   string
	newEntryForce bool
)

// newEntryCmd represents the newEntry command
var newEntryCmd = &cobra.Command{
	Use:   "newEntry <label> <host> <ip[:port]>",
	Short: "Append or replace a reverse_proxy block for <host> inside a label file",
	Long: `Creates the label file if it does not exist, then appends a Caddy site block:
<host> {
  import common
  reverse_proxy <ip[:port]>
}
Fails if a block for <host> already exists unless --force is used.
Warns if the IP address is already assigned to another host in the same label.`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		label := args[0]
		host := args[1]
		upstream := args[2]

		// Validate label filename (forbid path separators/control chars).
		if !isSafeFilename(label) {
			return fmt.Errorf("invalid label %q (forbidden: path separators or control chars)", label)
		}

		// Conservative host validation (lowercase letters, digits, dots, dashes).
		hostRe := regexp.MustCompile(`^[a-z0-9.-]+$`)
		if !hostRe.MatchString(host) {
			return fmt.Errorf("invalid host %q (allowed: a-z, 0-9, '.', '-')", host)
		}

		if strings.TrimSpace(upstream) == "" {
			return fmt.Errorf("upstream is required (ex: 10.10.0.20 or 10.10.0.20:3002)")
		}

		// Normalize extension
		ext := newEntryExt
		if ext != "" && ext[0] != '.' {
			ext = "." + ext
		}
		if ext == "" {
			ext = ".caddy"
		}

		// Ensure directory exists
		if err := os.MkdirAll(newEntryDir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %q: %w", newEntryDir, err)
		}

		// Compute label file path
		filename := label
		if filepath.Ext(label) != ext {
			filename = label + ext
		}
		fullPath := filepath.Join(newEntryDir, filename)

		// Read or create file
		var content string
		if b, err := os.ReadFile(fullPath); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("failed to read file %s: %w", fullPath, err)
			}
			// Create an empty file if missing
			if err := os.WriteFile(fullPath, []byte(""), 0o644); err != nil {
				return fmt.Errorf("failed to create file %s: %w", fullPath, err)
			}
			content = ""
		} else {
			content = string(b)
		}

		// Warn if this IP already appears elsewhere in the label
		ipRe := regexp.MustCompile(`reverse_proxy\s+([^\s}]+)`)
		for _, match := range ipRe.FindAllStringSubmatch(content, -1) {
			if len(match) > 1 && match[1] == upstream {
				fmt.Printf("⚠️  Warning: upstream %s already used in this label file (%s)\n", upstream, fullPath)
				break
			}
		}

		// Desired block
		desired := fmt.Sprintf("%s {\n\timport common\n\treverse_proxy %s\n}\n", host, upstream)

		// Detect existing block for this host
		re := regexp.MustCompile("(?ms)^" + regexp.QuoteMeta(host) + `\s*\{.*?\n\}\n?`)
		loc := re.FindStringIndex(content)

		if loc != nil {
			if !newEntryForce {
				return fmt.Errorf("entry for host %q already exists in %s (use --force to replace it)", host, fullPath)
			}
			// Replace existing block
			newContent := content[:loc[0]] + desired + content[loc[1]:]
			if err := os.WriteFile(fullPath, []byte(newContent), 0o644); err != nil {
				return fmt.Errorf("failed to write file %s: %w", fullPath, err)
			}
			fmt.Printf("✅ Entry replaced: %s → %s in %s\n", host, upstream, fullPath)
			return nil
		}

		// Append with tidy spacing
		var out strings.Builder
		trimmed := strings.TrimRight(content, "\n")
		if trimmed == "" {
			out.WriteString(desired)
		} else {
			out.WriteString(trimmed)
			out.WriteString("\n\n")
			out.WriteString(desired)
		}

		if err := os.WriteFile(fullPath, []byte(out.String()), 0o644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", fullPath, err)
		}

		fmt.Printf("✅ Entry added: %s → %s in %s\n", host, upstream, fullPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(newEntryCmd)

	newEntryCmd.Flags().StringVarP(&newEntryDir, "dir", "d", "/srv/proxy/sites", "Directory where the label file resides")
	newEntryCmd.Flags().StringVar(&newEntryExt, "ext", ".caddy", "File extension for the label (default: .caddy)")
	newEntryCmd.Flags().BoolVarP(&newEntryForce, "force", "f", false, "Replace the entry if it already exists")
}
