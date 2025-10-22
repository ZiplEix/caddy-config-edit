package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	labelDir   string
	labelExt   string
	labelForce bool
)

// labelCmd represents the label command
var labelCmd = &cobra.Command{
	Use:   "label <name>",
	Short: "Create an (empty) label file (.caddy) to group multiple entries",
	Long: `Creates an (empty) file named after the provided label in the specified directory.
By default: directory "/srv/proxy/sites", extension ".caddy".
Does not overwrite an existing file unless --force is specified.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Relaxed validation: allow anything that is a single filename (no slashes, no control chars)
		if !isSafeFilename(name) {
			return fmt.Errorf("invalid label name %q (forbidden: path separators or control chars)", name)
		}

		ext := labelExt
		if ext != "" && ext[0] != '.' {
			ext = "." + ext
		}
		if ext == "" {
			ext = ".caddy"
		}

		filename := name
		if filepath.Ext(name) != ext {
			filename = name + ext
		}

		if err := os.MkdirAll(labelDir, 0o755); err != nil {
			return fmt.Errorf("unable to create directory %q: %w", labelDir, err)
		}

		fullPath := filepath.Join(labelDir, filename)

		if _, err := os.Stat(fullPath); err == nil && !labelForce {
			return fmt.Errorf("file already exists: %s (use --force to overwrite)", fullPath)
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("error checking file %s: %w", fullPath, err)
		}

		f, err := os.OpenFile(fullPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("unable to create file %s: %w", fullPath, err)
		}
		defer f.Close()

		fmt.Printf("✅ File created: %s\n", fullPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(labelCmd)

	labelCmd.Flags().StringVarP(&labelDir, "dir", "d", "/srv/proxy/sites", "Directory where the label file will be created")
	labelCmd.Flags().StringVar(&labelExt, "ext", ".caddy", "File extension (default \".caddy\"). Empty for no extension")
	labelCmd.Flags().BoolVarP(&labelForce, "force", "f", false, "Overwrite the file if it already exists")
}
