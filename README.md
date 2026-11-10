# confluence-cli

A command-line tool for reading and publishing content on your internal Confluence installation.

## Features

### Read
- **Page details** — Fetch complete page data including body, ancestors, children
- **CQL search** — Full Confluence Query Language support
- **HTML → Markdown conversion** — Confluence storage format automatically converted to clean Markdown
- **Agent-friendly output** — Markdown (default), JSON, and plain text formats
- **Flattened structure** — Output is deliberately simplified for LLM context windows
- **User lookup** — Resolve user keys (`@user(<userkey>)`) to names, and usernames to user keys
- **Rendered view & link check** — Fetch the rendered HTML of a page and flag links to pages that do not exist

### Write
- **Publish** — Convert Markdown files to Confluence storage format and update existing pages
- **Create** — Create new pages from Markdown in any space
- **Diff** — Compare local Markdown against existing Confluence pages
- **Image upload** — Local images uploaded as attachments with Confluence macros
- **Layout preservation** — Existing two-column sidebar layouts are preserved when publishing Markdown. Content in storage format is published as is, with its own layout

### Auth
- **macOS Keychain integration** — PAT stored securely, no config files
- **Environment variable fallback** — Set `CONFLUENCE_PAT` for CI/CD or non-macOS use

## Installation

```bash
go install github.com/bentsolheim/confluence-cli@latest
```

Or build from source:

```bash
git clone https://github.com/bentsolheim/confluence-cli.git
cd confluence-cli
go build -o confluence-cli .
```

## Setup

### 1. Store your PAT

Generate a Personal Access Token in Confluence (Profile → Personal Access Tokens), then:

```bash
confluence auth store
```

Or set the environment variable:

```bash
export CONFLUENCE_PAT=your-token-here
```

### 2. Verify authentication

```bash
confluence auth test
```

## Usage

### Get page details

```bash
# Markdown (default) — body is converted from HTML to Markdown
confluence page 12345

# JSON
confluence page 12345 -o json

# From a full URL
confluence page https://wiki.sits.no/spaces/DEV/pages/12345/My+Page

# Show @user mentions as names instead of user keys
confluence page 12345 --resolve-users

# Rendered HTML (body.view), as shown in the browser
confluence page 12345 -o view
```

User mentions are shown as `@user(<userkey>)` by default. With `--resolve-users`,
each unique user on the page is looked up once and shown as
`@user(Solheim, Bent Andre / k77319)` in the markdown, json and text formats.

### Look up users

```bash
# User key → name and username (e.g. from an @user(<userkey>) mention)
confluence user 1f58644b4d9ba951014dae08fed80006

# Username → user key (e.g. to write <ri:user ri:userkey="..."/> in storage format)
confluence user k34046

# Several at once, as JSON
confluence user k77319 k34046 -o json
```

An argument of 32 hex characters is treated as a user key, anything else as a
username. Use `--key` or `--username` to override. Output formats are
`markdown` (table, default), `json` and `text` (`username<TAB>display name<TAB>user key`).
The command exits with a non-zero status if any user could not be found.

### Check links on a page

```bash
# List all links in the rendered page and flag links to missing pages
confluence links 12345

# Only the links to pages that do not exist
confluence links 12345 --missing-only

# JSON, e.g. for agents
confluence links 12345 -o json
```

`links` reads the rendered page (`body.view`). Confluence renders links to pages
that do not exist (for example an `<ac:link><ri:page ri:content-title="..."/></ac:link>`
with a misspelled title) as "create page" links, and links into unknown spaces as
unresolved. These are reported as `MISSING`, and the command exits with a non-zero
status. Other links are classified as `page`, `attachment`, `user`, `anchor`,
`internal` or `external`. Use it after `publish` to verify that page links resolve.

### Search for content

```bash
# Text search
confluence search "deployment pipeline"

# Scoped to specific spaces
confluence search "deployment pipeline" --spaces MUP,DEV

# Raw CQL
confluence search --cql "space = DEV AND type = page"

# JSON output with limited results
confluence search --cql "label = backend" --max-results 10 -o json
```

### Publish Markdown to Confluence

Create a Markdown file with frontmatter:

```markdown
---
confluence:
  url: https://wiki.sits.no
  pageId: "12345"
  title: "My Page Title"
---

# My Page

Content goes here...
```

Then publish:

```bash
# Preview what would change
confluence publish page.md --dry-run

# Publish (with confirmation prompt)
confluence publish page.md

# Publish without confirmation
confluence publish page.md --force
```

### Create a new page

```bash
confluence create page.md --space "~username" --title "New Page"

# With a parent page
confluence create page.md --space DEV --title "New Page" --parent-id 12345
```

After creation, add the printed page ID to your frontmatter for future publish/diff.

### Diff local vs remote

```bash
confluence diff page.md
```

### Image handling

Local images referenced in Markdown are automatically uploaded as Confluence attachments.
External URLs are converted to `<ri:url>` macros. Configure image behavior in frontmatter:

```yaml
confluence:
  assetsBase: "./images"           # Base directory for relative image paths
  overwriteAttachments: true       # Overwrite existing attachments (default: true)
  failOnMissingImages: true        # Fail if referenced image is missing (default: true)
```

## Output Formats (read commands)

| Format | Flag | Best for |
|--------|------|----------|
| Markdown | `-o markdown` | Default. Human-readable, LLM context windows |
| JSON | `-o json` | AI agents, piping to `jq`, programmatic use |
| Text | `-o text` | Human terminal use |
| Storage | `-o storage` | Lossless round-trip editing (page only) |
| View | `-o view` | Rendered HTML as shown in the browser (page only) |

## Authentication

```bash
confluence auth store    # Store/update PAT in macOS Keychain
confluence auth test     # Verify PAT works
confluence auth delete   # Remove PAT from Keychain
```

The PAT is stored in macOS Keychain under service name `confluence-cli` with the Confluence URL as the account identifier. As a fallback, the `CONFLUENCE_PAT` environment variable is also checked.

All commands, including `user` and `links`, use this token internally and never print it.
There is no need to extract the PAT and call the REST API with `curl`.
