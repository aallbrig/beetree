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
	generateAll       bool
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate engine-specific code from a .beetree.yaml spec",
	Long: `Generate native game engine code from a .beetree.yaml specification.

Supported engines:
  unity   - Unity C# (MonoBehaviour, ScriptableObject)
  unreal  - Unreal Engine C++ (BTTaskNode, BTDecorator)
  godot   - Godot GDScript (Node-based, Godot 4.x)

First run generates everything including stubs. Subsequent runs regenerate
tree definitions but skip existing stubs unless --overwrite is passed.

Use --all to generate for all .beetree.yaml files found in the trees/ directory.`,
}

func addGenerateFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&generateOutput, "output", "", "Output directory for generated code")
	cmd.Flags().BoolVar(&generateDryRun, "dry-run", false, "Preview what would be generated")
	cmd.Flags().BoolVar(&generateOverwrite, "overwrite", false, "Overwrite existing stub files")
	cmd.Flags().BoolVar(&generateAll, "all", false, "Generate for all .beetree.yaml files in trees/")
}

func makeEngineCmd(engineName, engineDesc string, newGen func() codegen.Generator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s <file>", engineName),
		Short: fmt.Sprintf("Generate %s code from a .beetree.yaml spec", engineDesc),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			gen := newGen()

			if generateAll {
				return generateAllSpecs(cmd, gen, engineName)
			}

			if len(args) < 1 {
				return fmt.Errorf("specify a .beetree.yaml file or use --all")
			}

			return generateSingleSpec(cmd, gen, args[0], engineName)
		},
	}
	addGenerateFlags(cmd)
	return cmd
}

func init() {
	generateCmd.AddCommand(makeEngineCmd("unity", "Unity C#", func() codegen.Generator { return codegen.NewUnityGenerator() }))
	generateCmd.AddCommand(makeEngineCmd("unreal", "Unreal C++", func() codegen.Generator { return codegen.NewUnrealGenerator() }))
	generateCmd.AddCommand(makeEngineCmd("godot", "Godot GDScript", func() codegen.Generator { return codegen.NewGodotGenerator() }))
	rootCmd.AddCommand(generateCmd)
}

func generateSingleSpec(cmd *cobra.Command, gen codegen.Generator, specFile, engineName string) error {
	tree, err := spec.ParseFile(specFile)
	if err != nil {
		return fmt.Errorf("parse spec: %w", err)
	}

	files, err := gen.Generate(tree)
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}

	return writeGeneratedFiles(cmd, files, engineName)
}

func generateAllSpecs(cmd *cobra.Command, gen codegen.Generator, engineName string) error {
	specFiles, err := filepath.Glob("trees/*.beetree.yaml")
	if err != nil {
		return fmt.Errorf("glob trees: %w", err)
	}
	if len(specFiles) == 0 {
		return fmt.Errorf("no .beetree.yaml files found in trees/")
	}

	for _, sf := range specFiles {
		if err := generateSingleSpec(cmd, gen, sf, engineName); err != nil {
			return fmt.Errorf("%s: %w", sf, err)
		}
	}
	return nil
}

func writeGeneratedFiles(cmd *cobra.Command, files []codegen.GeneratedFile, engineName string) error {
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
}
