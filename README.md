# confluence-cli

A command-line tool for querying your internal Confluence installation, optimized for AI/KI agent consumption.

## Features

- **macOS Keychain integration** — PAT stored securely, no config files
- **CQL search** — Full Confluence Query Language support
- **Page details** — Fetch complete page data including body, ancestors, children
- **HTML → Markdown conversion** — Confluence storage format automatically converted to clean Markdown
- **Agent-friendly output** — Markdown (default), JSON, and plain text formats
- **Flattened structure** — Output is deliberately simplified for LLM context windows

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
confluence-cli auth store
```

### 2. Verify authentication

```bash
confluence-cli auth test
```

## Usage

### Get page details

```bash
# Markdown (default) — body is converted from HTML to Markdown
confluence-cli page 12345

# JSON
confluence-cli page 12345 -o json

# Plain text
confluence-cli page 12345 -o text
```

### Search for content

```bash
# Default markdown output
confluence-cli search "space = DEV AND type = page"

# Search by title
confluence-cli search "title ~ 'architecture'"

# JSON output with limited results
confluence-cli search "label = backend" --max-results 10 -o json
```

### Use with a different Confluence instance

```bash
confluence-cli --url https://other-wiki.example.com page 12345
```

## Output Formats

| Format | Flag | Best for |
|--------|------|----------|
| Markdown | `-o markdown` | Default. Human-readable, LLM context windows |
| JSON | `-o json` | AI agents, piping to `jq`, programmatic use |
| Text | `-o text` | Human terminal use |

## Keychain Management

```bash
confluence-cli auth store    # Store/update PAT
confluence-cli auth test     # Verify PAT works
confluence-cli auth delete   # Remove PAT from Keychain
```

The PAT is stored in macOS Keychain under service name `confluence-cli` with the Confluence URL as the account identifier.
