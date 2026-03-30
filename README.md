# confluence-cli

A command-line tool for reading and publishing content on your internal Confluence installation.

## Features

### Read
- **Page details** — Fetch complete page data including body, ancestors, children
- **CQL search** — Full Confluence Query Language support
- **HTML → Markdown conversion** — Confluence storage format automatically converted to clean Markdown
- **Agent-friendly output** — Markdown (default), JSON, and plain text formats
- **Flattened structure** — Output is deliberately simplified for LLM context windows

### Write
- **Publish** — Convert Markdown files to Confluence storage format and update existing pages
- **Create** — Create new pages from Markdown in any space
- **Diff** — Compare local Markdown against existing Confluence pages
- **Image upload** — Local images uploaded as attachments with Confluence macros
- **Layout preservation** — Existing two-column sidebar layouts are preserved on update

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
```

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

## Authentication

```bash
confluence auth store    # Store/update PAT in macOS Keychain
confluence auth test     # Verify PAT works
confluence auth delete   # Remove PAT from Keychain
```

The PAT is stored in macOS Keychain under service name `confluence-cli` with the Confluence URL as the account identifier. As a fallback, the `CONFLUENCE_PAT` environment variable is also checked.
