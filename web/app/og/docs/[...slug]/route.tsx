import { getPageImage, source } from '@/lib/source';
import { notFound } from 'next/navigation';
import { ImageResponse } from 'next/og';
import { site } from '@/lib/site';

import { readFile } from 'node:fs/promises';
import { join } from 'node:path';

const fontBuffer = async (...fontPath: string[]) => {
  const data = await readFile(join(process.cwd(), 'node_modules', ...fontPath));
  return data.buffer.slice(data.byteOffset, data.byteOffset + data.byteLength) as ArrayBuffer;
};

export const revalidate = false;

const displayFont = fontBuffer('@fontsource', 'inter', 'files', 'inter-latin-400-normal.woff');

const monoFont = fontBuffer('@fontsource', 'jetbrains-mono', 'files', 'jetbrains-mono-latin-500-normal.woff');

export async function GET(_req: Request, { params }: RouteContext<'/og/docs/[...slug]'>) {
  const { slug } = await params;
  const page = source.getPage(slug.slice(0, -1));
  if (!page) notFound();

  const [inter, jetbrainsMono] = await Promise.all([displayFont, monoFont]);
  const accent = site.accentHex ?? '#36c6b8';

  return new ImageResponse(
    <div
      style={{
        width: '100%',
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'space-between',
        padding: '68px 76px',
        color: '#f3f4f1',
        background: '#131412',
        fontFamily: 'Inter',
      }}
    >
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          fontFamily: 'JetBrains Mono',
          fontSize: 24,
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <span style={{ color: accent }}>&gt;_</span>
          <span>{site.binary}</span>
        </div>
        <span style={{ color: '#7f827b' }}>docs</span>
      </div>

      <div
        style={{
          display: 'flex',
          flexDirection: 'column',
          maxWidth: 1000,
          padding: '44px 48px',
          borderRadius: 8,
          background: '#1e201b',
        }}
      >
        <h1
          style={{
            margin: 0,
            fontSize: 76,
            fontWeight: 400,
            lineHeight: 0.98,
            letterSpacing: '-0.05em',
          }}
        >
          {page.data.title}
        </h1>
        <p
          style={{
            margin: '28px 0 0',
            color: '#b6b8b3',
            fontFamily: 'JetBrains Mono',
            fontSize: 25,
            lineHeight: 1.35,
          }}
        >
          {page.data.description}
        </p>
      </div>

      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 14,
          color: '#7f827b',
          fontFamily: 'JetBrains Mono',
          fontSize: 20,
        }}
      >
        <span style={{ color: accent }}>$</span>
        <span>es --help</span>
      </div>
    </div>,
    {
      width: 1200,
      height: 630,
      fonts: [
        { name: 'Inter', data: inter, weight: 400, style: 'normal' },
        { name: 'JetBrains Mono', data: jetbrainsMono, weight: 500, style: 'normal' },
      ],
    },
  );
}

export function generateStaticParams() {
  return source.getPages().map((page) => ({
    lang: page.locale,
    slug: getPageImage(page).segments,
  }));
}
