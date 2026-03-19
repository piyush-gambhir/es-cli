# Elasticsearch CLI

A command-line interface for managing Elasticsearch clusters -- indices, documents, aliases, templates, pipelines, ILM policies, nodes, shards, and more.

Designed for both human operators and coding agents (LLMs). All commands support `--help` for detailed usage, and `-o json` / `-o yaml` for machine-readable output.

[![Go Version](https://img.shields.io/github/go-mod/go-version/piyush-gambhir/es-cli)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/piyush-gambhir/es-cli)](https://github.com/piyush-gambhir/es-cli/releases)
[![License](https://img.shields.io/github/license/piyush-gambhir/es-cli)](LICENSE)

## Features

- Full API coverage -- cluster, indices, documents, search, aliases, templates, pipelines, ILM, nodes, shards
- Multiple output formats -- table, JSON, YAML (`-o json`)
- Profile management -- multiple clusters with `--profile`
- Three auth methods -- basic auth, API key, bearer token
- TLS support -- custom CA certificates and insecure mode
- Safety features -- read-only mode, confirmation prompts, idempotent flags
- Auto-update -- checks for new versions, `es update` to self-update
- Agent-friendly -- comprehensive help text, structured output for LLM coding agents
- Cross-platform -- macOS, Linux, Windows (amd64 and arm64)

## Installation

```bash
# Go
go install github.com/piyush-gambhir/es-cli@latest

# From releases
# Download the appropriate binary from https://github.com/piyush-gambhir/es-cli/releases

# From source
git clone https://github.com/piyush-gambhir/es-cli.git
cd es-cli && make install
```

## Quick Start

```bash
# Authenticate
es login

# Check cluster health
es cluster health

# List indices
es index list

# Get index details as JSON
es index get my-index -o json
```

## Authentication

```bash
# Interactive login (saves profile to ~/.config/es-cli/config.yaml)
es login

# Environment variables
export ES_URL=https://elasticsearch.example.com:9200
export ES_USERNAME=elastic
export ES_PASSWORD=changeme
```

### Auth Methods

**Basic auth** (username + password):

```bash
export ES_URL=https://elasticsearch.example.com:9200
export ES_USERNAME=elastic
export ES_PASSWORD=changeme
```

**API key** (id + secret):

```bash
export ES_URL=https://elasticsearch.example.com:9200
export ES_API_KEY_ID=my-key-id
export ES_API_KEY=my-api-key-secret
```

**Bearer token**:

```bash
export ES_URL=https://elasticsearch.example.com:9200
export ES_TOKEN=my-bearer-token
```

### Auth Priority

Configuration is resolved in this order (first match wins):

1. CLI flags (`--url`, `--username`, `--password`, `--api-key-id`, `--api-key`, `--token`)
2. Environment variables (`ES_URL`, `ES_USERNAME`, `ES_PASSWORD`, `ES_API_KEY_ID`, `ES_API_KEY`, `ES_TOKEN`)
3. Config file profile (`~/.config/es-cli/config.yaml`)

## TLS Configuration

```bash
# Use a custom CA certificate
es cluster health --ca-cert /path/to/ca.pem

# Skip TLS verification (not recommended for production)
es cluster health --insecure

# Or via environment variables
export ES_CA_CERT=/path/to/ca.pem
export ES_INSECURE=true
```

## Configuration

### Profiles

```bash
# Interactive login creates a profile
es login

# List profiles
es config list-profiles

# Switch profiles
es config use-profile prod

# Use a profile for a single command
es cluster health --profile staging
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `ES_URL` | Elasticsearch URL |
| `ES_USERNAME` | Username for basic auth |
| `ES_PASSWORD` | Password for basic auth |
| `ES_API_KEY_ID` | API key ID |
| `ES_API_KEY` | API key secret |
| `ES_TOKEN` | Bearer token |
| `ES_CA_CERT` | Path to CA certificate |
| `ES_INSECURE` | Skip TLS verification (`true`/`1`) |
| `ES_READ_ONLY` | Block write operations (`true`/`1`) |
| `ES_NO_INPUT` | Disable interactive prompts |
| `ES_QUIET` | Suppress informational output |
| `ES_VERBOSE` | Enable verbose HTTP logging |

## Commands

| Group | Description | Aliases |
|-------|-------------|---------|
| `es cluster` | Cluster health, stats, settings, pending tasks, allocation | |
| `es index` | Manage indices, aliases, templates, component templates | `indices`, `idx` |
| `es search` | Query DSL, SQL, count, multi-search, field capabilities | `query` |
| `es document` | Get, index, delete, bulk, multi-get documents | `doc`, `docs` |
| `es node` | List nodes, info, stats, hot threads | `nodes` |
| `es shard` | View shard allocation and status | `shards` |
| `es ingest` | Manage ingest pipelines | `pipeline` |
| `es ilm` | Manage ILM policies | |
| `es config` | View/set configuration, manage profiles | |
| `es login` | Interactive authentication setup | |
| `es version` | Print CLI version | |
| `es update` | Self-update to latest version | |
| `es completion` | Generate shell completions | |

## Output Formats

All list and get commands support three output formats:

```bash
es index list                  # table (default, human-readable)
es index list -o json          # JSON (machine-readable)
es index list -o yaml          # YAML
```

## Global Flags

These flags are available on every command:

| Flag | Description |
|------|-------------|
| `--output`, `-o` | Output format: `table`, `json`, `yaml` |
| `--profile` | Configuration profile to use |
| `--url` | Elasticsearch URL |
| `--username`, `-u` | Username for basic auth |
| `--password`, `-p` | Password for basic auth |
| `--api-key-id` | API key ID |
| `--api-key` | API key secret |
| `--token` | Bearer token |
| `--ca-cert` | Path to CA certificate for TLS |
| `--insecure`, `-k` | Skip TLS certificate verification |
| `--read-only` | Block write operations |
| `--no-input` | Disable interactive prompts |
| `--quiet`, `-q` | Suppress informational output |
| `--verbose`, `-v` | Enable verbose HTTP logging |

## Safety Features

### Read-only mode

Block all write operations to prevent accidental changes:

```bash
es index delete my-index --read-only    # blocked
export ES_READ_ONLY=true                # block all writes globally
```

### No-input mode

Disable all interactive prompts for CI/automation:

```bash
es --no-input index delete my-index --confirm
export ES_NO_INPUT=true
```

### Confirmation prompts

Destructive commands require confirmation:

```bash
es index delete my-index              # prompts for confirmation
es index delete my-index --confirm    # skips the prompt
```

### Idempotent operations

```bash
es index create my-index --if-not-exists     # no error if already exists
es index delete my-index --confirm --if-exists  # no error if not found
```

## File Input Format

Commands that accept `--file/-f` support:
- JSON files (`.json`)
- YAML files (`.yaml`, `.yml`)
- Stdin (use `-f -` and pipe input)

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT
