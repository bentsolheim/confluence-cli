package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bentsolheim/confluence-cli/internal/auth"
	"github.com/bentsolheim/confluence-cli/internal/confluence"
)

var labelCmd = &cobra.Command{
	Use:   "label",
	Short: "Manage page labels",
	Long: `Add, remove, or list labels on Confluence pages.

Examples:
  confluence label list 12345
  confluence label add 12345 backend architecture
  confluence label remove https://wiki.sits.no/spaces/MUP/pages/12345/Title backend`,
}

var labelListCmd = &cobra.Command{
	Use:   "list [page ID or URL]",
	Short: "List labels on a page",
	Args:  cobra.ExactArgs(1),
	RunE:  runLabelList,
}

var labelAddCmd = &cobra.Command{
	Use:   "add [page ID or URL] [labels...]",
	Short: "Add labels to a page",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runLabelAdd,
}

var labelRemoveCmd = &cobra.Command{
	Use:   "remove [page ID or URL] [labels...]",
	Short: "Remove labels from a page",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runLabelRemove,
}

func init() {
	rootCmd.AddCommand(labelCmd)
	labelCmd.AddCommand(labelListCmd)
	labelCmd.AddCommand(labelAddCmd)
	labelCmd.AddCommand(labelRemoveCmd)
}

func resolvePageID(client *confluence.Client, input string) (string, error) {
	ref, err := parsePageInput(input)
	if err != nil {
		return "", err
	}
	if ref.ID != "" {
		return ref.ID, nil
	}
	page, err := client.GetPageByTitle(ref.SpaceKey, ref.Title)
	if err != nil {
		return "", err
	}
	return page.ID, nil
}

func newClientForLabels() (*confluence.Client, error) {
	token, err := auth.ResolveToken(confluenceURL)
	if err != nil {
		return nil, err
	}
	return confluence.NewClient(confluenceURL, token, verbose), nil
}

func runLabelList(cmd *cobra.Command, args []string) error {
	client, err := newClientForLabels()
	if err != nil {
		return err
	}
	pageID, err := resolvePageID(client, args[0])
	if err != nil {
		return err
	}

	labels, err := client.GetLabels(pageID)
	if err != nil {
		return err
	}

	if len(labels) == 0 {
		fmt.Fprintln(os.Stderr, "No labels")
		return nil
	}

	names := make([]string, len(labels))
	for i, l := range labels {
		names[i] = l.Name
	}
	fmt.Println(strings.Join(names, "\n"))
	return nil
}

func runLabelAdd(cmd *cobra.Command, args []string) error {
	client, err := newClientForLabels()
	if err != nil {
		return err
	}
	pageID, err := resolvePageID(client, args[0])
	if err != nil {
		return err
	}

	labels := args[1:]
	result, err := client.AddLabels(pageID, labels)
	if err != nil {
		return err
	}

	names := make([]string, len(result))
	for i, l := range result {
		names[i] = l.Name
	}
	fmt.Fprintf(os.Stderr, "✅ Labels: %s\n", strings.Join(names, ", "))
	return nil
}

func runLabelRemove(cmd *cobra.Command, args []string) error {
	client, err := newClientForLabels()
	if err != nil {
		return err
	}
	pageID, err := resolvePageID(client, args[0])
	if err != nil {
		return err
	}

	for _, label := range args[1:] {
		if err := client.RemoveLabel(pageID, label); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Removed: %s\n", label)
	}
	fmt.Fprintln(os.Stderr, "✅ Done")
	return nil
}
