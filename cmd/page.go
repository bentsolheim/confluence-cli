package cmd

import (
	"fmt"
	"os"

	"github.com/bentsolheim/confluence-cli/internal/confluence"
	"github.com/bentsolheim/confluence-cli/internal/formatter"
	"github.com/bentsolheim/confluence-cli/internal/keychain"
	"github.com/spf13/cobra"
)

var pageCmd = &cobra.Command{
	Use:   "page [ID]",
	Short: "Get details of a Confluence page",
	Long: `Fetch full details of a Confluence page by its ID.

Examples:
  confluence page 12345
  confluence page 12345 -o markdown
  confluence page 12345 -o json
  confluence page 12345 -o text`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		token, err := keychain.GetPAT(confluenceURL)
		if err != nil {
			return err
		}
		client := confluence.NewClient(confluenceURL, token, verbose)
		page, err := client.GetPage(id)
		if err != nil {
			return fmt.Errorf("failed to get page %s: %w", id, err)
		}
		f, err := formatter.New(outputFormat, confluenceURL)
		if err != nil {
			return err
		}
		return f.FormatPage(os.Stdout, page)
	},
}

func init() {
	rootCmd.AddCommand(pageCmd)
}
