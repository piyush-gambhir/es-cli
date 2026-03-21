---
name: es-cli
description: "Expert guide for the es (Elasticsearch) CLI — cluster health, indices, aliases, templates, component templates, documents, search, ingest pipelines, ILM, nodes, shards, and config profiles. Use this skill when the user mentions Elasticsearch operations from the terminal, es-cli, index lifecycle, reindex, rollover, allocation explain, or automating ES with a CLI. Trigger for coding agents that need exact commands, flags, auth env vars, or JSON output for scripts."
---

# Elasticsearch CLI (`es`) — agent skill

## Maintainer note

**When you change the CLI** (commands, flags, auth, config, output): update **`CLAUDE.md`** and any affected **`README.md` / `docs/`** in the same PR. If the change affects what this skill claims (scope, install, major workflows), update the YAML **`description`** above or the sections below. See **`CONTRIBUTING.md` → Documentation and agent materials**.

## Canonical guide

The full command reference, workflows, and examples live in **`CLAUDE.md`** in this repository. Read that file for complete coverage.

## Quick orientation

| Item | Value |
|------|--------|
| Binary | `es` |
| Config | `~/.config/es-cli/config.yaml` |
| Machine-readable output | `-o json` (prefer for agents) |
| Non-interactive | set env vars (`ES_URL`, …) and/or `--no-input` |

## Discovering commands

Cobra adds **`-h` / `--help`** on the root and every subcommand:

```bash
es --help
es index --help
es index list --help
```

Use this when the repo is unavailable or to confirm flags after upgrades.

## Install (reference)

```bash
go install github.com/piyush-gambhir/es-cli@latest
# or: clone repo && make install — see README.md
```
