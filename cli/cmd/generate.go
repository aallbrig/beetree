package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aallbrig/beetree-cli/internal/codegen"
	"github.com/aallbrig/beetree-cli/internal/spec"
	"github.com/spf13/cobra"
)

var (
	generateOutput    string
	generateDryRun    bool
	generateOverwrite bool
)

var generateCmd = &cobra.Command{
	Use:   "generate <engine> <file>",
	Short: "Generate engine-specific code from a .beetree.yaml spec",
	Long: `Generate native game engine code from a .beetree.yaml specification.

Supported engines:
  unity   - Unity C# (MonoBehaviour, ScriptableObject)
  unreal  - Unreal Engine C++ (BTTaskNode, BTDecorator)
  godot   - Godot GDScript (Node-based, Godot 4.x)

First run generates everything including stubs. Subsequent runs regenerate
tree definitions but skip existing stubs unless --overwrite is passed.`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		engineName := args[0]
		specFile := args[1]

		gen, err := getGenerator(engineName)
		if err != nil {
			return err
		}

		tree, err := spec.ParseFile(specFile)
		if err != nil {
			return fmt.Errorf("parse spec: %w", err)
		}

		files, err := gen.Generate(tree)
		if err != nil {
			return fmt.Errorf("generate: %w", err)
		}

		outDir := generateOutput
		if outDir == "" {
			outDir = filepath.Join(".", "generated", engineName)
		}

		if generateDryRun {
			cmd.Printf("Dry run — would generate %d files in %s:\n", len(files), outDir)
			for _, f := range files {
				stubLabel := ""
				if f.IsStub {
					stubLabel = " (stub)"
				}
				cmd.Printf("  %s%s\n", f.Path, stubLabel)
			}
			return nil
		}

		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("create output directory: %w", err)
		}

		written := 0
		skipped := 0
		for _, f := range files {
			outPath := filepath.Join(outDir, f.Path)

			if f.IsStub && !generateOverwrite {
				if _, err := os.Stat(outPath); err == nil {
					skipped++
					continue
				}
			}

			if err := os.WriteFile(outPath, []byte(f.Content), 0644); err != nil {
				return fmt.Errorf("write %s: %w", f.Path, err)
			}
			written++
		}

		cmd.Printf("Generated %d files in %s", written, outDir)
		if skipped > 0 {
			cmd.Printf(" (%d stubs skipped, use --overwrite to replace)", skipped)
		}
		cmd.Println()

		return nil
	},
}

func getGenerator(engine string) (codegen.Generator, error) {
	switch engine {
	case "unity":
		return codegen.NewUnityGenerator(), nil
	case "unreal":
		return codegen.NewUnrealGenerator(), nil
	case "godot":
		return codegen.NewGodotGenerator(), nil
	default:
		return nil, fmt.Errorf("unknown engine %q — supported: unity, unreal, godot", engine)
	}
}

func init() {
	generateCmd.Flags().StringVar(&generateOutput, "output", "", "Output directory for generated code")
	generateCmd.Flags().BoolVar(&generateDryRun, "dry-run", false, "Preview what would be generated")
	generateCmd.Flags().BoolVar(&generateOverwrite, "overwrite", false, "Overwrite existing stub files")
	rootCmd.AddCommand(generateCmd)
}
