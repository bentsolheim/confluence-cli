package confluence

import (
	"fmt"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
)

// ShowDiff displays the differences between two Markdown contents in unified diff format.
func ShowDiff(existingMarkdown, newMarkdown, filename string) error {
	diff := difflib.UnifiedDiff{
		A:        strings.Split(existingMarkdown, "\n"),
		B:        strings.Split(newMarkdown, "\n"),
		FromFile: "Confluence (existing)",
		ToFile:   filename,
		Context:  3,
	}

	result, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		return fmt.Errorf("generating diff: %w", err)
	}

	if result != "" {
		fmt.Printf("Differences found:\n%s\n", result)
		return nil
	}

	fmt.Println("No differences found.")
	return nil
}
