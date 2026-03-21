# Elasticsearch CLI - Agent Guide

> **Contributors:** When you add or change commands, flags, auth, or output, update this file in the same PR as the code. See `CONTRIBUTING.md` → *Documentation and agent materials*, and keep `SKILL.md` aligned if the CLI’s described scope changes.

## Quick Reference

- **Binary:** `es`
- **Config file:** `~/.config/es-cli/config.yaml`
- **Env vars:** `ES_URL`, `ES_USERNAME`, `ES_PASSWORD`, `ES_API_KEY_ID`, `ES_API_KEY`, `ES_TOKEN`, `ES_CA_CERT`, `ES_INSECURE`, `ES_READ_ONLY`
- **Auth methods:** Basic auth (username/password), API key (id + secret), Bearer token
- **Config priority:** CLI flags > environment variables > profile config > defaults

## Setup

```bash
# Interactive login (prompts for URL, auth method, credentials, TLS, profile name)
es login

# Or set environment variables for non-interactive use
export ES_URL=https://elasticsearch.example.com:9200
export ES_USERNAME=elastic
export ES_PASSWORD=changeme
```

## Output Formats

All list/get commands support three output formats via `-o`:

- `-o table` (default) -- human-readable tabular output
- `-o json` -- JSON, ideal for programmatic parsing with jq
- `-o yaml` -- YAML, useful for config management

**For agents:** Always use `-o json` when you need to parse or process output programmatically.

## Common Workflows

### Cluster health and monitoring

```bash
# Check cluster health
es cluster health -o json

# Get cluster statistics (nodes, indices, shards, storage)
es cluster stats -o json

# View cluster settings (persistent and transient)
es cluster settings -o json

# Include default settings
es cluster settings --include-defaults -o json

# Check pending cluster tasks
es cluster pending-tasks -o json

# Explain why a shard is unassigned
es cluster allocation-explain -o json

# Explain a specific shard
es cluster allocation-explain --index my-index --shard 0 --primary -o json
```

### Index management

```bash
# List all indices
es index list -o json

# Filter by pattern
es index list --pattern "logs-*" -o json

# Filter by health (green, yellow, red)
es index list --health yellow -o json

# Filter by status (open, close)
es index list --status open -o json

# Get full index details (settings, mappings, aliases)
es index get my-index -o json

# Create an index with default settings
es index create my-index

# Create with settings and mappings from file
es index create my-index -f settings.json

# Create only if not already exists
es index create my-index -f settings.json --if-not-exists

# Delete an index (requires confirmation)
es index delete my-index

# Delete without confirmation prompt
es index delete my-index --confirm

# Delete only if it exists (no error on 404)
es index delete my-index --confirm --if-exists

# Open a closed index
es index open my-index

# Close an index
es index close my-index

# Get index settings
es index settings my-index -o json

# Update index settings
es index settings my-index --set number_of_replicas=2
es index settings my-index --set number_of_replicas=2 --set refresh_interval=30s

# Get index mappings
es index mappings my-index -o json

# Update mappings from file
es index mappings my-index -f mappings.json

# Get index statistics
es index stats my-index -o json

# Rollover an index alias
es index rollover my-alias
es index rollover my-alias -f conditions.json

# Reindex documents
es index reindex -f reindex.json
```

### Alias management

```bash
# List all aliases
es index alias list -o json

# Filter aliases by index
es index alias list --index my-index -o json

# Create an alias
es index alias create my-index my-alias

# Delete an alias (requires confirmation)
es index alias delete my-index my-alias

# Delete without confirmation
es index alias delete my-index my-alias --confirm
```

### Template management

```bash
# List all index templates
es index template list -o json

# Get a specific template
es index template get my-template -o json

# Create/update a template from file
es index template create my-template -f template.json

# Delete a template
es index template delete my-template --confirm

# Delete only if it exists
es index template delete my-template --confirm --if-exists

# List component templates
es index component-template list -o json

# Get a component template
es index component-template get my-component -o json

# Create a component template
es index component-template create my-component -f component.json

# Delete a component template
es index component-template delete my-component --confirm
```

### Search and query

```bash
# Search with Query DSL from file
es search query my-index -f query.json

# Search with size and offset
es search query my-index -f query.json --size 20 --from 40

# Search with sorting
es search query my-index -f query.json --sort timestamp:desc

# Search without a query file (match all)
es search query my-index -o json

# Run an SQL query inline
es search sql --query "SELECT * FROM my-index LIMIT 10"

# Run SQL from file
es search sql -f query.sql

# Count documents in an index
es search count my-index -o json

# Count with a query filter
es search count my-index -f query.json

# Multi-search from NDJSON file
es search msearch -f requests.ndjson

# Show field capabilities
es search field-caps my-index --fields "*"
es search field-caps my-index --fields "timestamp,message,level"
```

### Document operations

```bash
# Get a document by ID
es document get my-index abc123 -o json

# Index a document with auto-generated ID
es document index my-index -f doc.json

# Index with a specific ID
es document index my-index -f doc.json --id abc123

# Index from stdin
echo '{"name":"test"}' | es doc index my-index -f -

# Delete a document (requires confirmation)
es document delete my-index abc123

# Delete without confirmation
es document delete my-index abc123 --confirm

# Delete only if it exists (no error on 404)
es document delete my-index abc123 --confirm --if-exists

# Bulk index from NDJSON file
es document bulk -f bulk.ndjson

# Bulk index into a specific index
es document bulk -f bulk.ndjson --index my-index

# Multi-get documents
es document mget my-index -f ids.json
```

### Ingest pipeline management

```bash
# List all pipelines
es ingest list -o json

# Get a pipeline definition
es ingest get my-pipeline -o json

# Create a pipeline
es ingest create my-pipeline -f pipeline.json

# Create only if it doesn't exist
es ingest create my-pipeline -f pipeline.json --if-not-exists

# Delete a pipeline
es ingest delete my-pipeline --confirm

# Delete only if it exists
es ingest delete my-pipeline --confirm --if-exists

# Simulate a pipeline with sample documents
es ingest simulate my-pipeline -f docs.json
```

### ILM policy management

```bash
# List all ILM policies
es ilm list -o json

# Get an ILM policy
es ilm get my-policy -o json

# Create an ILM policy
es ilm create my-policy -f policy.json

# Create only if it doesn't exist
es ilm create my-policy -f policy.json --if-not-exists

# Delete a policy
es ilm delete my-policy --confirm

# Delete only if it exists
es ilm delete my-policy --confirm --if-exists

# Explain ILM status for an index
es ilm explain my-index -o json
```

### Shard viewing

```bash
# List all shards
es shard list -o json

# Filter shards by index
es shard list --index my-index -o json
```

### Node operations

```bash
# List all nodes
es node list -o json

# Get detailed node information
es node info node-1 -o json

# Get node statistics (all nodes)
es node stats -o json

# Get stats for a specific node
es node stats node-1 -o json

# Get only a specific metric (jvm, os, indices, etc.)
es node stats --metric jvm -o json

# Show hot threads
es node hot-threads
es node hot-threads node-1
```

### Configuration management

```bash
# View current configuration
es config view

# Set default output format
es config set defaults.output json

# List all profiles
es config list-profiles

# Switch to a different profile
es config use-profile staging
```

## Tips for Agents

- Always use `-o json` when you need to parse output programmatically.
- Use `--confirm` on destructive commands (delete) to skip interactive prompts.
- Use `--no-input` to disable all interactive prompts in CI/automation.
- Use `--read-only` as a safety guard to block all mutating operations.
- For bulk operations: list with `-o json`, parse with jq, then loop over results.
- Many create/update commands require a `-f` flag pointing to a JSON or YAML file. Prepare the file first, then pass it.
- Use `-f -` to pipe content from stdin into any command that accepts a file.
- Use `--if-not-exists` on create commands and `--if-exists` on delete commands for idempotent scripts.
- The `index` command has aliases: `indices`, `idx`. The `document` command has aliases: `doc`, `docs`. The `node` command has alias `nodes`. The `shard` command has alias `shards`. The `search` command has alias `query`. The `ingest` command has alias `pipeline`.
- Template subcommands also have aliases: `template` has `templates`, `tmpl`; `component-template` has `ctmpl`; `alias` has `aliases`.
- List subcommands often have an `ls` alias (e.g., `es index ls`, `es node ls`, `es shard ls`).
- Use `es config view` to check current connection settings and confirm which profile is active.
- The `es cluster allocation-explain` command is invaluable for debugging unassigned shards.
- The `es node hot-threads` output is always plain text regardless of `--output` setting.

## Complete Command Reference

### Top-level commands

| Command | Description |
|---------|-------------|
| `es login` | Interactively log in and save a connection profile |
| `es version` | Print CLI version, commit, and build date |
| `es update` | Check for and install CLI updates (--check for check only) |
| `es completion` | Generate shell completion scripts |

### `es config` -- Manage CLI configuration

| Command | Description |
|---------|-------------|
| `es config view` | Display the current configuration |
| `es config set <key> <value>` | Set a configuration value (defaults.output, current_profile) |
| `es config use-profile <name>` | Switch to a different profile |
| `es config list-profiles` | List all configured profiles |

### `es cluster` -- Manage the Elasticsearch cluster

| Command | Description |
|---------|-------------|
| `es cluster health` | Show cluster health status (name, status, nodes, shards) |
| `es cluster stats` | Show detailed cluster statistics |
| `es cluster settings` | Show cluster settings (--include-defaults for all) |
| `es cluster pending-tasks` | Show pending cluster-level tasks |
| `es cluster allocation-explain` | Explain shard allocation decisions (--index, --shard, --primary) |

### `es index` (aliases: `indices`, `idx`) -- Manage indices

| Command | Description |
|---------|-------------|
| `es index list` (alias: `ls`) | List indices (--pattern, --health, --status) |
| `es index create <index>` | Create an index (-f, --if-not-exists) |
| `es index get <index>` | Get index details (settings, mappings, aliases) |
| `es index delete <index>` | Delete an index (--confirm, --if-exists) |
| `es index open <index>` | Open a closed index |
| `es index close <index>` | Close an index |
| `es index settings <index>` | Get or update index settings (--set key=value) |
| `es index mappings <index>` | Get or update index mappings (-f) |
| `es index stats <index>` | Get index statistics |
| `es index rollover <alias>` | Rollover an index alias (-f for conditions) |
| `es index reindex` | Reindex documents from one index to another (-f required) |

#### `es index alias` (alias: `aliases`) -- Manage index aliases

| Command | Description |
|---------|-------------|
| `es index alias list` (alias: `ls`) | List aliases (--index to filter) |
| `es index alias create <index> <alias>` | Create an alias for an index |
| `es index alias delete <index> <alias>` | Delete an alias (--confirm) |

#### `es index template` (aliases: `templates`, `tmpl`) -- Manage index templates

| Command | Description |
|---------|-------------|
| `es index template list` (alias: `ls`) | List all index templates |
| `es index template get <name>` | Get an index template |
| `es index template create <name>` | Create/update a template (-f required) |
| `es index template delete <name>` | Delete a template (--confirm, --if-exists) |

#### `es index component-template` (alias: `ctmpl`) -- Manage component templates

| Command | Description |
|---------|-------------|
| `es index component-template list` (alias: `ls`) | List all component templates |
| `es index component-template get <name>` | Get a component template |
| `es index component-template create <name>` | Create/update a component template (-f required) |
| `es index component-template delete <name>` | Delete a component template (--confirm, --if-exists) |

### `es search` (alias: `query`) -- Search and query

| Command | Description |
|---------|-------------|
| `es search query <index>` | Run a Query DSL search (-f, --size, --from, --sort) |
| `es search sql` | Run an SQL query (--query for inline, -f for file) |
| `es search count <index>` | Count documents (-f for query filter) |
| `es search msearch` | Execute a multi-search request (-f NDJSON required) |
| `es search field-caps <index>` | Show field capabilities (--fields required) |

### `es document` (aliases: `doc`, `docs`) -- Manage documents

| Command | Description |
|---------|-------------|
| `es document get <index> <id>` | Get a document by ID |
| `es document index <index>` | Index a document (-f required, --id optional) |
| `es document delete <index> <id>` | Delete a document (--confirm, --if-exists) |
| `es document bulk` | Bulk index documents (-f NDJSON required, --index optional) |
| `es document mget <index>` | Multi-get documents (-f required) |

### `es node` (alias: `nodes`) -- Manage cluster nodes

| Command | Description |
|---------|-------------|
| `es node list` (alias: `ls`) | List all nodes (IP, heap, RAM, CPU, roles) |
| `es node info <node-id>` | Show detailed node information |
| `es node stats [node-id]` | Show node statistics (--metric for specific metric) |
| `es node hot-threads [node-id]` | Show hot threads (always plain text output) |

### `es shard` (alias: `shards`) -- View shard information

| Command | Description |
|---------|-------------|
| `es shard list` (alias: `ls`) | List all shards (--index to filter) |

### `es ingest` (alias: `pipeline`) -- Manage ingest pipelines

| Command | Description |
|---------|-------------|
| `es ingest list` | List all ingest pipelines |
| `es ingest get <name>` | Get a pipeline definition |
| `es ingest create <name>` | Create a pipeline (-f required, --if-not-exists) |
| `es ingest delete <name>` | Delete a pipeline (--confirm, --if-exists) |
| `es ingest simulate <name>` | Simulate a pipeline with sample documents (-f required) |

### `es ilm` -- Manage Index Lifecycle Management policies

| Command | Description |
|---------|-------------|
| `es ilm list` | List all ILM policies |
| `es ilm get <policy>` | Get an ILM policy definition |
| `es ilm create <name>` | Create an ILM policy (-f required, --if-not-exists) |
| `es ilm delete <name>` | Delete an ILM policy (--confirm, --if-exists) |
| `es ilm explain <index>` | Explain ILM lifecycle status for an index |

## Global Flags

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: table (default), json, yaml |
| `--profile <name>` | Configuration profile to use |
| `--url <url>` | Elasticsearch URL override |
| `-u, --username <user>` | Username for basic auth override |
| `-p, --password <pass>` | Password for basic auth override |
| `--api-key-id <id>` | API key ID override |
| `--api-key <secret>` | API key secret override |
| `--token <token>` | Bearer token override |
| `--ca-cert <path>` | Path to CA certificate for TLS |
| `-k, --insecure` | Skip TLS certificate verification |
| `--read-only` | Block write operations (safety mode for agents) |
| `--no-input` | Disable all interactive prompts (for CI/agent use) |
| `-q, --quiet` | Suppress informational output |
| `-v, --verbose` | Enable verbose HTTP logging |
