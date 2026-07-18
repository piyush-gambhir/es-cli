import type { Metadata } from 'next';
import '@fontsource-variable/inter';
import '@fontsource-variable/jetbrains-mono';
import '@fontsource/instrument-serif';
import type { CSSProperties } from 'react';
import { Provider } from '@/components/provider';
import { siteMetadataDescription } from '@/lib/metadata';
import { site } from '@/lib/site';
import { siteUrl } from '@/lib/shared';
import './global.css';

const socialDescription =
  'Agent-ready, harness-agnostic Elasticsearch CLI for terminal automation, with JSON/YAML output, read-only safety, and no-input flags.';

export const metadata: Metadata = {
  metadataBase: new URL(siteUrl),
  applicationName: site.name,
  title: {
    default: `${site.name}: ${site.tagline}`,
    template: `%s · ${site.name}`,
  },
  description: siteMetadataDescription,
  authors: [{ name: 'Piyush Gambhir', url: 'https://github.com/piyush-gambhir' }],
  creator: 'Piyush Gambhir',
  publisher: 'Piyush Gambhir',
  alternates: {
    canonical: siteUrl,
  },
  icons: {
    icon: [{ url: '/es-cli/favicon.svg', type: 'image/svg+xml' }],
  },
  openGraph: {
    type: 'website',
    url: siteUrl,
    siteName: site.name,
    locale: 'en_US',
    title: `${site.name}: ${site.tagline}`,
    description: socialDescription,
    images: [
      {
        url: `${siteUrl}/og/docs/image.png`,
        width: 1200,
        height: 630,
        alt: `${site.name}: ${site.tagline}`,
      },
    ],
  },
  twitter: {
    card: 'summary_large_image',
    title: `${site.name}: ${site.tagline}`,
    description: socialDescription,
    images: [
      {
        url: `${siteUrl}/og/docs/image.png`,
        alt: `${site.name}: ${site.tagline}`,
      },
    ],
  },
};

export default function Layout({ children }: LayoutProps<'/'>) {
  const rootStyle = {
    ...(site.accent ? { '--site-accent': site.accent } : {}),
  } as CSSProperties;

  return (
    <html
      lang="en"
      data-accent={site.accentName}
      style={rootStyle}
      suppressHydrationWarning
    >
      <head>
        <script
          dangerouslySetInnerHTML={{
            __html: "document.documentElement.classList.add('js')",
          }}
        />
      </head>
      <body className="flex flex-col min-h-screen">
        <Provider>{children}</Provider>
      </body>
    </html>
  );
}
