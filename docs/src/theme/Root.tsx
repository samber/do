import React, { useEffect } from 'react';

declare const window: Window & { posthog?: { capture: (event: string, props?: Record<string, unknown>) => void } };

function getOutboundLinkType(href: string): string | null {
  if (href.includes('go.dev/play') || href.includes('play.golang.org')) return 'playground';
  if (href.includes('pkg.go.dev')) return 'godoc';
  if (href.includes('github.com/samber/do') && !href.includes('github.com/sponsors')) return 'source';
  return null;
}

export default function Root({ children }: { children: React.ReactNode }) {
  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      const anchor = (e.target as HTMLElement).closest('a');
      if (!anchor) return;
      const href = anchor.getAttribute('href') ?? '';
      const type = getOutboundLinkType(href);
      if (type) {
        window.posthog?.capture('link_clicked', { type, href });
      }
    };

    let searchTimeout: ReturnType<typeof setTimeout>;
    const handleInput = (e: Event) => {
      const target = e.target as HTMLInputElement;
      if (target.id !== 'docsearch-input') return;
      clearTimeout(searchTimeout);
      searchTimeout = setTimeout(() => {
        const query = target.value.trim();
        if (query) {
          window.posthog?.capture('search_queried', { query });
        }
      }, 500);
    };

    document.addEventListener('click', handleClick);
    document.addEventListener('input', handleInput);
    return () => {
      document.removeEventListener('click', handleClick);
      document.removeEventListener('input', handleInput);
      clearTimeout(searchTimeout);
    };
  }, []);

  return <>{children}</>;
}
