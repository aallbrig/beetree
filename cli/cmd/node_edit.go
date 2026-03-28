package cmd

import (
	"fmt"

	"github.com/aallbrig/beetree-cli/internal/model"
	"github.com/aallbrig/beetree-cli/internal/renderer"
	"github.com/aallbrig/beetree-cli/internal/spec"
	"github.com/aallbrig/beetree-cli/internal/treeedit"
	"github.com/spf13/cobra"
)

var (
	nodeAddType      string
	nodeAddNode      string
	nodeAddDecorator string
)

var nodeAddCmd = &cobra.Command{
	Use:   "add <file> <parent-name> <node-name>",
	Short: "Add a node to a behavior tree",
	Long: `Add a new child node under the specified parent.

Example:
  beetree node add trees/enemy.beetree.yaml combat reload --type action --node Reload
  beetree node add trees/enemy.beetree.yaml root flee_check --type condition --node FleeCheck`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		parentName := args[1]
		nodeName := args[2]

		tree, err := spec.ParseFile(filePath)
		if err != nil {
			return fmt.Errorf("parse: %w", err)
		}

		newNode := model.NodeSpec{
			Type:      nodeAddType,
			Name:      nodeName,
			Node:      nodeAddNode,
			Decorator: nodeAddDecorator,
		}

		if err := treeedit.AddNode(&tree.Tree, parentName, newNode); err != nil {
			return err
		}

		if err := treeedit.SaveSpec(tree, filePath); err != nil {
			return err
		}

		cmd.Printf("✓ Added [%s] %s under %s\n\n", nodeAddType, nodeName, parentName)
		cmd.Print(renderer.RenderSpecASCII(&tree.Tree))
		return nil
	},
}

var nodeRemoveCmd = &cobra.Command{
	Use:   "remove <file> <node-name>",
	Short: "Remove a node from a behavior tree",
	Long: `Remove a node and all its children from the tree.

Example:
  beetree node remove trees/enemy.beetree.yaml patrol`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		nodeName := args[1]

		tree, err := spec.ParseFile(filePath)
		if err != nil {
			return fmt.Errorf("parse: %w", err)
		}

		if err := treeedit.RemoveNode(&tree.Tree, nodeName); err != nil {
			return err
		}

		if err := treeedit.SaveSpec(tree, filePath); err != nil {
			return err
		}

		cmd.Printf("✓ Removed %s\n\n", nodeName)
		cmd.Print(renderer.RenderSpecASCII(&tree.Tree))
		return nil
	},
}

var nodeMoveDest string

var nodeMoveCmd = &cobra.Command{
	Use:   "move <file> <node-name> --to <parent-name>",
	Short: "Move a node to a new parent",
	Long: `Detach a node (and its subtree) and reattach under a new parent.

Example:
  beetree node move trees/enemy.beetree.yaml patrol --to combat`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		nodeName := args[1]

		if nodeMoveDest == "" {
			return fmt.Errorf("--to flag is required")
		}

		tree, err := spec.ParseFile(filePath)
		if err != nil {
			return fmt.Errorf("parse: %w", err)
		}

		if err := treeedit.MoveNode(&tree.Tree, nodeName, nodeMoveDest); err != nil {
			return err
		}

		if err := treeedit.SaveSpec(tree, filePath); err != nil {
			return err
		}

		cmd.Printf("✓ Moved %s → %s\n\n", nodeName, nodeMoveDest)
		cmd.Print(renderer.RenderSpecASCII(&tree.Tree))
		return nil
	},
}

func init() {
	nodeAddCmd.Flags().StringVar(&nodeAddType, "type", "action", "Node type (action, condition, sequence, selector, parallel, decorator)")
	nodeAddCmd.Flags().StringVar(&nodeAddNode, "node", "", "Node class name (e.g., Patrol, HasTarget)")
	nodeAddCmd.Flags().StringVar(&nodeAddDecorator, "decorator", "", "Decorator type (for decorator nodes)")
	nodeCmd.AddCommand(nodeAddCmd)

	nodeCmd.AddCommand(nodeRemoveCmd)

	nodeMoveCmd.Flags().StringVar(&nodeMoveDest, "to", "", "Destination parent node name")
	nodeCmd.AddCommand(nodeMoveCmd)
}
