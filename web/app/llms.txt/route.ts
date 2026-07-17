import { source } from '@/lib/source';
import { llms } from 'fumadocs-core/source';
import { site } from '@/lib/site';
import { siteUrl } from '@/lib/shared';
import { getOtherSuiteProjects } from '@/lib/suite';

export const revalidate = false;

export function GET() {
  const index = llms(source)
    .index()
    .replace(/\]\((\/[^)]+)\)/g, (_match, path: string) => `](${siteUrl}${path})`);
  const relatedSites = getOtherSuiteProjects(site.repo)
    .map(({ name, href }) => `- ${name}: ${href}`)
    .join('\n');
  const intro =
    'ES CLI is an independent, unofficial Elasticsearch CLI that is agent-ready and harness-agnostic. Any coding agent or agent harness that can run shell commands can use structured JSON/YAML output, read-only safety mode, and no-input automation flags to manage clusters, indices, and searches from the terminal.';
  return new Response(
    `${intro}\n\n${index}\n\n## Related CLI sites\n\n${relatedSites}\n`,
  );
}
