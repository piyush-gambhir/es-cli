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
  docsLink?: {
    label: string;
    href: string;
  };
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
  /** Optional: features section heading (default: "Everything, from one binary") */
  featuresTitle?: string;
  /** Optional: features section subheading */
  featuresSubtitle?: string;
  /** Optional: CTA band body (default mentions installing the binary) */
  ctaBody?: string;
  /** Optional: per-site accent expressed as an OKLCH color */
  accent?: string;
  /** Optional: human-readable accent name */
  accentName?: string;
  /** Optional: sRGB equivalent used by generated images and static assets */
  accentHex?: string;
}

export const site: SiteConfig = {
  name: 'ES CLI',
  binary: 'es',
  repo: 'piyush-gambhir/es-cli',
  tagline: 'Elasticsearch from your terminal',
  description:
    'ES CLI is an independent, unofficial open-source CLI for Elasticsearch clusters. Inspect health, manage indices and documents, run searches, and operate pipelines and ILM policies from a fast, scriptable tool built for humans and coding agents alike.',
  badge: 'Open-source · Agent-friendly',
  accent: 'oklch(0.75 0.12 185)',
  accentName: 'teal',
  accentHex: '#36c6b8',
  installCommand:
    'curl -sSfL https://raw.githubusercontent.com/piyush-gambhir/es-cli/main/install.sh | sh',
  features: [
    {
      icon: Search,
      title: 'Search & documents',
      body: 'Run Query DSL, SQL, count, multi-search, and field-capability requests. Get, index, delete, bulk, and multi-get documents.',
      docsLink: {
        label: 'Query DSL',
        href: '/docs/commands/documents-search',
      },
    },
    {
      icon: KeyRound,
      title: 'Secure connections',
      body: 'Basic auth, Elasticsearch API keys, bearer tokens, named profiles, custom CA certificates, and TLS verification controls.',
      docsLink: {
        label: 'Basic auth',
        href: '/docs/authentication',
      },
    },
    {
      icon: Bot,
      title: 'Agent-friendly',
      body: '-o json|yaml for structured reads, --read-only safety mode, --no-input automation, idempotent flags, and quiet operation.',
      docsLink: {
        label: 'structured reads',
        href: '/docs/agents',
      },
    },
    {
      icon: GitBranch,
      title: 'Pipelines & lifecycle',
      body: 'Create, inspect, simulate, and delete ingest pipelines, then manage and explain Index Lifecycle Management policies.',
      docsLink: {
        label: 'ingest pipelines',
        href: '/docs/commands/ingest-ilm',
      },
    },
    {
      icon: Zap,
      title: 'Fast & scriptable',
      body: 'A single cross-platform binary with file and stdin input, multiple profiles, shell completion, and self-update support.',
      docsLink: {
        label: 'single cross-platform binary',
        href: '/docs/installation',
      },
    },
    {
      icon: ListChecks,
      title: 'Cluster operations',
      body: 'Monitor health, stats, settings, pending tasks, allocation decisions, nodes, hot threads, indices, and shards.',
      docsLink: {
        label: 'Monitor health',
        href: '/docs/commands/cluster-monitoring',
      },
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
