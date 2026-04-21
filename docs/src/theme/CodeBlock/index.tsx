import React, { useRef } from 'react';
import OriginalCodeBlock from '@theme-original/CodeBlock';
import type CodeBlockType from '@theme/CodeBlock';
import type { WrapperProps } from '@docusaurus/types';

type Props = WrapperProps<typeof CodeBlockType>;

declare const window: Window & { posthog?: { capture: (event: string, props?: Record<string, unknown>) => void } };

export default function CodeBlockWrapper(props: Props) {
  const ref = useRef<HTMLDivElement>(null);

  const handleClick = (e: React.MouseEvent<HTMLDivElement>) => {
    const target = e.target as HTMLElement;
    const button = target.closest('button');
    if (!button) return;

    const isCopyButton =
      button.getAttribute('aria-label')?.toLowerCase().includes('copy') ||
      button.className?.includes('copyButton') ||
      button.className?.includes('copy');

    if (isCopyButton) {
      window.posthog?.capture('code_example_copied', {
        language: typeof props.language === 'string' ? props.language : (props as { className?: string }).className?.replace('language-', '') ?? 'unknown',
        title: typeof props.title === 'string' ? props.title : undefined,
      });
    }
  };

  return (
    <div ref={ref} onClick={handleClick}>
      <OriginalCodeBlock {...props} />
    </div>
  );
}
