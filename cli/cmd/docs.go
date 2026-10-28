package cmd

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var docsOut string

var docsCmd = &cobra.Command{
	Use:    "docs",
	Short:  "Write the CLI reference as Markdown",
	Hidden: true,
	Long: `Writes a single Markdown reference for every command, generated from the
command definitions so the published docs cannot drift from the binary.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		var w io.Writer = cmd.OutOrStdout()
		if docsOut != "" {
			f, err := os.Create(docsOut)
			if err != nil {
				return err
			}
			defer func() { _ = f.Close() }()
			w = f
		}
		return writeReference(w, cmd.Root())
	},
}

func init() {
	docsCmd.Flags().StringVarP(&docsOut, "out", "o", "", "write to this file instead of stdout")
	rootCmd.AddCommand(docsCmd)
}

// writeReference renders root and its visible subcommands as Hugo-ready Markdown.
func writeReference(w io.Writer, root *cobra.Command) error {
	var b strings.Builder
	b.WriteString("---\ntitle: \"CLI reference\"\ndescription: \"Every beetree command, generated from the binary.\"\nweight: 20\n---\n\n")
	b.WriteString("{{< generated >}}\n\n")
	var walk func(c *cobra.Command, level int)
	walk = func(c *cobra.Command, level int) {
		b.WriteString(fmt.Sprintf("%s `%s`\n\n", strings.Repeat("#", level), c.CommandPath()))
		if c.Long != "" {
			b.WriteString(proseBlock(c.Long))
		} else if c.Short != "" {
			b.WriteString(escapeMD(c.Short) + "\n\n")
		}
		b.WriteString("```\n" + c.UseLine() + "\n```\n\n")
		if ex := strings.TrimSpace(c.Example); ex != "" {
			b.WriteString("**Examples**\n\n```\n" + ex + "\n```\n\n")
		}
		if fl := strings.TrimSpace(c.NonInheritedFlags().FlagUsages()); fl != "" {
			b.WriteString("**Flags**\n\n```\n" + fl + "\n```\n\n")
		}
		subs := c.Commands()
		sort.Slice(subs, func(i, j int) bool { return subs[i].Name() < subs[j].Name() })
		for _, s := range subs {
			if s.Hidden || s.Name() == "help" || s.Name() == "completion" {
				continue
			}
			walk(s, min(level+1, 4))
		}
	}
	walk(root, 2)
	_, err := io.WriteString(w, strings.TrimRight(b.String(), "\n")+"\n")
	return err
}

// escapeMD stops placeholders like <file> from being read as raw HTML.
func escapeMD(s string) string {
	return strings.NewReplacer("<", "&lt;", ">", "&gt;").Replace(s)
}

// proseBlock renders a command's Long text. Indented blocks (usage listings
// in the help text) become fenced code so their layout survives; the rest is
// escaped prose.
func proseBlock(long string) string {
	var out strings.Builder
	var code []string
	flush := func() {
		if len(code) > 0 {
			out.WriteString("```\n" + strings.Join(code, "\n") + "\n```\n\n")
			code = nil
		}
	}
	var para []string
	flushPara := func() {
		if len(para) > 0 {
			out.WriteString(escapeMD(strings.Join(para, "\n")) + "\n\n")
			para = nil
		}
	}
	for _, line := range strings.Split(strings.TrimSpace(long), "\n") {
		switch {
		case strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t"):
			flushPara()
			code = append(code, strings.TrimRight(line, " "))
		case strings.TrimSpace(line) == "":
			flush()
			flushPara()
		default:
			flush()
			para = append(para, line)
		}
	}
	flush()
	flushPara()
	return out.String()
}
