import {
  Bot,
  GitBranch,
  KeyRound,
  ListChecks,
  Search,
  Zap,
  type LucideIcon,
} from 'lucide-react';

export interface Feature {
  icon: LucideIcon;
  title: string;
  body: string;
}

export interface SiteConfig {
  /** Display name, e.g. "Acme CLI" */
  name: string;
  /** The binary invoked in examples, e.g. "acme" */
  binary: string;
  /** GitHub "owner/repo" */
  repo: string;
  /** One-line hero heading */
  tagline: string;
  /** Hero sub-paragraph */
  description: string;
  /** Small pill above the heading */
  badge: string;
  /** One-line install command shown in the hero */
  installCommand: string;
  /** Feature cards */
  features: Feature[];
  /** Title above the code block */
  exampleTitle: string;
  /** Shell example rendered in the terminal card */
  example: string;
  /** Optional: tech / query languages this CLI speaks (logo strip) */
  compatible?: string[];
}

export const site: SiteConfig = {
  name: 'ES CLI',
  binary: 'es',
  repo: 'piyush-gambhir/es-cli',
  tagline: 'Elasticsearch from your terminal',
  description:
    'A fast, scriptable CLI for Elasticsearch clusters. Inspect health, manage indices and documents, run searches, and operate pipelines and ILM policies — built for humans and coding agents alike.',
  badge: 'Open-source · Agent-friendly',
  installCommand:
    'curl -sSfL https://raw.githubusercontent.com/piyush-gambhir/es-cli/main/install.sh | sh',
  features: [
    {
      icon: Search,
      title: 'Search & documents',
      body: 'Run Query DSL, SQL, count, multi-search, and field-capability requests. Get, index, delete, bulk, and multi-get documents.',
    },
    {
      icon: KeyRound,
      title: 'Secure connections',
      body: 'Basic auth, Elasticsearch API keys, bearer tokens, named profiles, custom CA certificates, and TLS verification controls.',
    },
    {
      icon: Bot,
      title: 'Agent-friendly',
      body: '-o json|yaml for structured reads, --read-only safety mode, --no-input automation, idempotent flags, and quiet operation.',
    },
    {
      icon: GitBranch,
      title: 'Pipelines & lifecycle',
      body: 'Create, inspect, simulate, and delete ingest pipelines, then manage and explain Index Lifecycle Management policies.',
    },
    {
      icon: Zap,
      title: 'Fast & scriptable',
      body: 'A single cross-platform binary with file and stdin input, multiple profiles, shell completion, and self-update support.',
    },
    {
      icon: ListChecks,
      title: 'Cluster operations',
      body: 'Monitor health, stats, settings, pending tasks, allocation decisions, nodes, hot threads, indices, and shards.',
    },
  ],
  exampleTitle: 'An eight-line tour',
  example: `# Configure a cluster profile
es login
# Inspect health and indices as JSON
es cluster health -o json
es index list --pattern "logs-*" -o json
# Search and inspect lifecycle status
es search query logs-2026.07 --size 20 --sort timestamp:desc -o json
es ilm explain logs-2026.07 -o json`,
  compatible: [
    "Query DSL",
    "SQL",
    "Bulk API",
    "ILM",
    "Ingest pipelines",
    "API keys",
  ],
};
