package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"

	"github.com/bentsolheim/confluence-cli/internal/auth"
	"github.com/bentsolheim/confluence-cli/internal/confluence"
	"github.com/spf13/cobra"
)

var (
	userByKey      bool
	userByUsername bool
)

var userCmd = &cobra.Command{
	Use:   "user [userkey or username]...",
	Short: "Look up Confluence users by user key or username",
	Long: `Look up one or more Confluence users by user key or username.

Use it to resolve @user(<userkey>) mentions from 'confluence page' into names,
or to find the user key of a username so you can write
<ri:user ri:userkey="..."/> in storage format.

An argument of 32 hex characters is treated as a user key, anything else as a
username. Use --key or --username to override the detection.

Examples:
  confluence user 1f58644b4d9ba951014dae08fed80006
  confluence user k34046
  confluence user k77319 k34046 -o json
  confluence user --username k77319 -o text`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if userByKey && userByUsername {
			return fmt.Errorf("--key and --username cannot be used together")
		}
		if !isUserOutputFormat(outputFormat) {
			return fmt.Errorf("unsupported output format for user: %q (use markdown, json, or text)", outputFormat)
		}

		token, err := auth.ResolveToken(confluenceURL)
		if err != nil {
			return err
		}
		client := confluence.NewClient(confluenceURL, token, verbose)

		var found []userResult
		failed := 0
		for _, arg := range args {
			byKey := lookupByKey(arg, userByKey, userByUsername)
			var u *confluence.User
			if byKey {
				u, err = client.GetUserByKey(arg)
			} else {
				u, err = client.GetUserByUsername(arg)
			}
			if err != nil {
				failed++
				kind := "username"
				if byKey {
					kind = "user key"
				}
				if strings.Contains(err.Error(), "HTTP 404") {
					fmt.Fprintf(os.Stderr, "user %q (%s): not found\n", arg, kind)
				} else {
					fmt.Fprintf(os.Stderr, "user %q (%s): %v\n", arg, kind, err)
				}
				continue
			}
			found = append(found, userResult{
				Query:       arg,
				Username:    u.Username,
				UserKey:     u.UserKey,
				DisplayName: u.DisplayName,
				Email:       u.Email,
			})
		}

		if err := writeUsers(os.Stdout, outputFormat, found); err != nil {
			return err
		}
		if failed > 0 {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			return fmt.Errorf("%d of %d user(s) could not be looked up", failed, len(args))
		}
		return nil
	},
}

// userResult is the output representation of a looked-up user.
type userResult struct {
	Query       string `json:"query"`
	Username    string `json:"username"`
	UserKey     string `json:"userKey"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email,omitempty"`
}

var userKeyRe = regexp.MustCompile(`^[0-9a-fA-F]{32}$`)

// isUserKey reports whether s looks like a Confluence user key (32 hex characters).
func isUserKey(s string) bool {
	return userKeyRe.MatchString(s)
}

// lookupByKey decides whether arg should be looked up as a user key.
func lookupByKey(arg string, forceKey, forceUsername bool) bool {
	switch {
	case forceKey:
		return true
	case forceUsername:
		return false
	default:
		return isUserKey(arg)
	}
}

func isUserOutputFormat(format string) bool {
	switch format {
	case "markdown", "md", "json", "text":
		return true
	}
	return false
}

func writeUsers(w io.Writer, format string, users []userResult) error {
	switch format {
	case "json":
		if users == nil {
			users = []userResult{}
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(users)
	case "text":
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		for _, u := range users {
			fmt.Fprintf(tw, "%s\t%s\t%s\n", u.Username, u.DisplayName, u.UserKey)
		}
		return tw.Flush()
	default:
		if len(users) == 0 {
			return nil
		}
		var b strings.Builder
		b.WriteString("| Username | Display name | User key |\n")
		b.WriteString("|----------|--------------|----------|\n")
		for _, u := range users {
			fmt.Fprintf(&b, "| %s | %s | %s |\n", escapeTableCell(u.Username), escapeTableCell(u.DisplayName), u.UserKey)
		}
		_, err := io.WriteString(w, b.String())
		return err
	}
}

// escapeTableCell makes a value safe for use in a Markdown table cell.
func escapeTableCell(s string) string {
	s = strings.ReplaceAll(s, "|", `\|`)
	return strings.ReplaceAll(s, "\n", " ")
}

func init() {
	userCmd.Flags().BoolVar(&userByKey, "key", false, "Treat all arguments as user keys")
	userCmd.Flags().BoolVar(&userByUsername, "username", false, "Treat all arguments as usernames")
	rootCmd.AddCommand(userCmd)
}
