package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	reloadContainer string
	reloadConfig    string
	reloadTTY       bool
	reloadQuiet     bool
)

// reloadCmd represents the reload command
var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Format, validate and reload Caddy config inside Docker",
	Long: `Runs the following commands in the target container:
1) caddy fmt --overwrite <Caddyfile>
2) caddy validate --config <Caddyfile>
3) caddy reload   --config <Caddyfile>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Build the common docker exec prefix
		baseArgs := []string{"exec", "-i"}
		if reloadTTY {
			baseArgs = append(baseArgs, "-t")
		}
		baseArgs = append(baseArgs, reloadContainer)

		// 1) caddy fmt --overwrite
		fmtArgs := append([]string{}, baseArgs...)
		fmtArgs = append(fmtArgs, "caddy", "fmt", "--overwrite", reloadConfig)
		if !reloadQuiet {
			fmt.Printf("→ docker %v\n", fmtArgs)
		}
		if err := run("docker", fmtArgs...); err != nil {
			return fmt.Errorf("failed to run caddy fmt: %w", err)
		}

		// 2) caddy validate --config
		validateArgs := append([]string{}, baseArgs...)
		validateArgs = append(validateArgs, "caddy", "validate", "--config", reloadConfig)
		if !reloadQuiet {
			fmt.Printf("→ docker %v\n", validateArgs)
		}
		if err := run("docker", validateArgs...); err != nil {
			return fmt.Errorf("failed to run caddy validate: %w", err)
		}

		// 3) caddy reload --config
		reloadArgs := append([]string{}, baseArgs...)
		reloadArgs = append(reloadArgs, "caddy", "reload", "--config", reloadConfig)
		if !reloadQuiet {
			fmt.Printf("→ docker %v\n", reloadArgs)
		}
		if err := run("docker", reloadArgs...); err != nil {
			return fmt.Errorf("failed to run caddy reload: %w", err)
		}

		if !reloadQuiet {
			fmt.Println("✅ Caddy configuration reloaded successfully.")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(reloadCmd)

	reloadCmd.Flags().StringVarP(&reloadContainer, "container", "c", "caddy", "Docker container name running Caddy")
	reloadCmd.Flags().StringVarP(&reloadConfig, "config", "f", "/etc/caddy/Caddyfile", "Path to Caddyfile inside the container")
	reloadCmd.Flags().BoolVar(&reloadTTY, "tty", true, "Attach a TTY (-t) in addition to -i for docker exec")
	reloadCmd.Flags().BoolVarP(&reloadQuiet, "quiet", "q", false, "Reduce output verbosity (only errors)")
}
