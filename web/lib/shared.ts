export const appName = 'ES CLI';
export const siteUrl = 'https://projects.piyushgambhir.com/es-cli';
export const docsRoute = '/docs';
export const docsImageRoute = '/og/docs';
export const docsContentRoute = '/llms.mdx/docs';

// The tracked file behind a docs page (path relative to content/docs).
export function sourcePath(pagePath: string): string {
  return `web/content/docs/${pagePath}`;
}

export const gitConfig = {
  user: 'piyush-gambhir',
  repo: 'es-cli',
  branch: 'main',
};
