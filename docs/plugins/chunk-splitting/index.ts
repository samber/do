import type {LoadContext, Plugin} from '@docusaurus/types';

/**
 * mermaid's parser (@mermaid-js/parser, which bundles a Langium/Chevrotain
 * grammar) and its rendering deps (d3, cytoscape, dagre-d3-es, katex, elkjs)
 * are only used by @docusaurus/theme-mermaid, which lazy-loads mermaid itself
 * via a dynamic import(). But Docusaurus's own webpack `common` cacheGroup
 * (priority 40, minChunks based on page count) hoists shared modules into a
 * chunk requested on every route, and these packages end up there even
 * though only 4 pages out of ~30 render a diagram. That shipped ~140KB of
 * 100%-unused JS on every non-diagram page (confirmed via Lighthouse
 * unused-javascript audits and by grepping the built chunk for Langium
 * grammar tokens like RuleCall/TerminalRule/ParserRule).
 *
 * This cacheGroup runs at a higher priority so these modules are pulled out
 * into their own async chunk instead, only fetched by pages that actually
 * call the mermaid dynamic import.
 */
export default function chunkSplittingPlugin(_context: LoadContext): Plugin {
  return {
    name: 'chunk-splitting-plugin',
    configureWebpack(_config, isServer) {
      if (isServer) {
        return {};
      }
      return {
        optimization: {
          splitChunks: {
            cacheGroups: {
              mermaid: {
                test: /[\\/]node_modules[\\/](mermaid|@mermaid-js|cytoscape|dagre-d3-es|khroma|katex|elkjs|d3|d3-[a-z-]+)[\\/]/,
                name: 'mermaid',
                chunks: 'async',
                priority: 60,
                enforce: true,
                reuseExistingChunk: true,
              },
            },
          },
        },
      };
    },
  };
}
