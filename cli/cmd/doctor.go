package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/aallbrig/beetree-cli/internal/spec"
	"github.com/aallbrig/beetree-cli/internal/validator"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check project health and configuration",
	Long:  "Run diagnostics on the current BeeTree project.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Println("BeeTree Doctor")
		cmd.Println("══════════════")

		// Go version
		cmd.Printf("  Go version: %s\n", runtime.Version())
		cmd.Printf("  OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)

		// Check for trees directory
		issues := 0
		treesDir := "trees"
		if _, err := os.Stat(treesDir); os.IsNotExist(err) {
			cmd.Println("  ⚠ trees/ directory not found (run 'beetree init')")
			issues++
		} else {
			specFiles, _ := filepath.Glob(filepath.Join(treesDir, "*.beetree.yaml"))
			cmd.Printf("  ✓ trees/ directory: %d spec(s) found\n", len(specFiles))

			// Validate each spec
			for _, sf := range specFiles {
				tree, err := spec.ParseFile(sf)
				if err != nil {
					cmd.Printf("  ✗ %s: parse error: %s\n", filepath.Base(sf), err)
					issues++
					continue
				}
				errs := validator.Validate(tree)
				if len(errs) > 0 {
					cmd.Printf("  ✗ %s: %d validation error(s)\n", filepath.Base(sf), len(errs))
					for _, e := range errs {
						cmd.Printf("      %s\n", e)
					}
					issues++
				} else {
					cmd.Printf("  ✓ %s: valid\n", filepath.Base(sf))
				}
			}
		}

		// Check for generated directory
		genDir := "generated"
		if _, err := os.Stat(genDir); os.IsNotExist(err) {
			cmd.Println("  · generated/ directory not found (no code generated yet)")
		} else {
			engines := []string{"unity", "unreal", "godot"}
			for _, eng := range engines {
				engDir := filepath.Join(genDir, eng)
				if _, err := os.Stat(engDir); err == nil {
					files, _ := filepath.Glob(filepath.Join(engDir, "*"))
					cmd.Printf("  ✓ generated/%s/: %d file(s)\n", eng, len(files))
				}
			}
		}

		cmd.Println()
		if issues > 0 {
			return fmt.Errorf("%d issue(s) found", issues)
		}
		cmd.Println("  All checks passed!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
