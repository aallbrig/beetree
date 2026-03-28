package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "beetree",
	Short: "A behavior tree ecosystem CLI for authoring, validating, and generating engine code",
	Long: `BeeTree is a behavior tree ecosystem that makes it easy to author, share,
and deploy behavior trees to any game engine.

Define behavior trees in an engine-agnostic .beetree.yaml specification,
then generate native code for Unity (C#), Unreal (C++), and Godot (GDScript).

Quick start:
  beetree init                    Initialize a new project
  beetree new <name>              Create a behavior tree
  beetree builder [file]          Launch the interactive TUI editor
  beetree validate <file>         Validate a tree spec
  beetree generate unity <file>   Generate engine code (unity/unreal/godot)`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
