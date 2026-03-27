package cmd

import (
	"sort"

	"github.com/aallbrig/beetree-cli/internal/model"
	"github.com/spf13/cobra"
)

var nodeFilter string

type nodeInfo struct {
	Name        string
	Category    string
	Description string
}

var nodeDescriptions = map[string]string{
	"action":           "Leaf node that executes a behavior and modifies world state",
	"condition":        "Leaf node that checks world state without modifying it",
	"sequence":         "Composite: executes children in order (AND logic)",
	"selector":         "Composite: tries children until one succeeds (OR logic)",
	"parallel":         "Composite: executes all children concurrently",
	"decorator":        "Wrapper: modifies single child behavior",
	"utility_selector": "Selects child based on utility scores (0.0-1.0)",
	"active_selector":  "Selector that re-evaluates higher-priority children each tick",
	"random_selector":  "Selector with randomized child evaluation order",
	"random_sequence":  "Sequence with randomized child execution order",
	"subtree":          "Reference to an external behavior tree file",
}

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Manage behavior tree node types",
}

var nodeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available node types",
	RunE: func(cmd *cobra.Command, args []string) error {
		var nodes []nodeInfo

		if nodeFilter == "" || nodeFilter == "core" {
			for name := range model.CoreNodeTypes() {
				nodes = append(nodes, nodeInfo{
					Name:        name,
					Category:    "CORE",
					Description: nodeDescriptions[name],
				})
			}
		}

		if nodeFilter == "" || nodeFilter == "extension" {
			for name := range model.ExtensionNodeTypes() {
				nodes = append(nodes, nodeInfo{
					Name:        name,
					Category:    "EXT",
					Description: nodeDescriptions[name],
				})
			}
		}

		sort.Slice(nodes, func(i, j int) bool {
			if nodes[i].Category != nodes[j].Category {
				return nodes[i].Category < nodes[j].Category
			}
			return nodes[i].Name < nodes[j].Name
		})

		cmd.Printf("%-20s %-6s %s\n", "NAME", "TYPE", "DESCRIPTION")
		cmd.Printf("%-20s %-6s %s\n", "----", "----", "-----------")
		for _, n := range nodes {
			cmd.Printf("%-20s %-6s %s\n", n.Name, n.Category, n.Description)
		}
		cmd.Printf("\n%d node type(s)\n", len(nodes))

		return nil
	},
}

func init() {
	nodeListCmd.Flags().StringVar(&nodeFilter, "filter", "", "Filter by category: core, extension")
	nodeCmd.AddCommand(nodeListCmd)
	rootCmd.AddCommand(nodeCmd)
}
