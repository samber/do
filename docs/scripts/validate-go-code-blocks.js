#!/usr/bin/env node
// Extracts every ```go fenced code block from docs/docs/**/*.md,
// docs/static/llms.txt and the repository README.md, and checks that each one
// is syntactically valid Go.
//
// This is a SYNTAX check only (via `gofmt`), not a type check: fragments that
// reference intentionally undeclared placeholders (e.g. `predicate`, `stream1`,
// `userChannel` used for illustration) will not be flagged, since that requires
// full package resolution across the multi-module workspace. It still catches
// the class of bugs that recurs most often in hand-written examples: unbalanced
// brackets, missing commas before a closing paren on its own line, malformed
// composite literals.
//
// Full programs (blocks starting with `package `) are checked as-is. Fragments
// can be a single top-level declaration, a bare sequence of statements, or an
// `import (...)` block followed by statements — there's no reliable way to
// know which without a real parse, so each fragment is tried under several
// wrappings and only reported as broken if every one of them fails to parse.
//
// The docs also use `(...)` and `{ ... }` as a deliberate "elided" marker
// (e.g. `injector.Invoke(...)`, `func(err error) { ... }`) to keep an
// illustrative snippet short. That's not valid Go, so those two exact forms
// are normalized to `()` / `{}` before checking — anything else (a bare
// identifier like `predicate`, a real `...T` variadic parameter) is left
// untouched and still relies on gofmt's parser, not a type checker.
//
// Ported from samber/ro (docs/scripts/validate-go-code-blocks.js), with three
// additions for conventions used in this repo's docs that samber/ro's don't
// have:
//
// 1. "Spec" blocks that list API signatures as pseudocode rather than real
//    statements, e.g. `injector.HealthCheck() map[string]error` or
//    `do.HealthCheck[T any](do.Injector) error` (the latter is never valid
//    Go standalone: `[T any]` is constraint syntax, only legal in a `func`
//    declaration's own type-parameter list, not as an instantiation). These
//    are recognized by SIGNATURE_LINE_RE and skipped entirely rather than
//    run through gofmt.
// 2. Composite-literal field fragments shown in isolation for a before/after
//    diff, e.g. `HookAfterShutdown: func(...) {...},` on its own with no
//    enclosing literal. Recognized by FIELD_FRAGMENT_RE and skipped.
// 3. Migration-guide blocks that intentionally concatenate two complete
//    programs (a "before" and an "after", each with its own `package main`)
//    inside a single fence. splitMultiPackage() detects 2+ top-level
//    `package` lines and validates each program segment independently
//    instead of parsing the whole block as one file.
const fs = require('fs');
const path = require('path');
const os = require('os');
const {execFileSync} = require('child_process');

const repoRoot = path.resolve(__dirname, '..', '..');
const docsRoot = path.resolve(__dirname, '..');

function listFilesRecursive(dir, predicate) {
  const out = [];
  for (const entry of fs.readdirSync(dir, {withFileTypes: true})) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      out.push(...listFilesRecursive(full, predicate));
    } else if (predicate(full)) {
      out.push(full);
    }
  }
  return out;
}

function collectTargetFiles() {
  const files = [];
  files.push(
    ...listFilesRecursive(path.join(docsRoot, 'docs'), (f) =>
      f.endsWith('.md'),
    ),
  );
  files.push(path.join(docsRoot, 'static', 'llms.txt'));
  files.push(path.join(repoRoot, 'README.md'));
  return files.filter((f) => fs.existsSync(f));
}

function normalizePlaceholders(code) {
  return (
    code
      .replace(/\(\s*\.\.\.\s*\)/g, '()')
      .replace(/\{\s*\.\.\.\s*\}/g, '{}')
      // Same elision marker, kept alongside a real leading argument for
      // context (e.g. `do.Provide(nil, ...)`) instead of standing alone.
      .replace(/,\s*\.\.\.\s*\)/g, ')')
      .replace(/,\s*\.\.\.\s*\}/g, '}')
  );
}

// Extracts ```go ... ``` blocks, returning { code, line } for each (line = 1-indexed
// line number of the opening fence, for error reporting).
function extractGoBlocks(content) {
  const blocks = [];
  const lines = content.split(/\r?\n/);
  let inBlock = false;
  let current = [];
  let startLine = 0;
  lines.forEach((line, idx) => {
    if (!inBlock && /^```go\b/.test(line.trim())) {
      inBlock = true;
      current = [];
      startLine = idx + 1;
      return;
    }
    if (inBlock && line.trim() === '```') {
      inBlock = false;
      blocks.push({code: current.join('\n'), line: startLine});
      return;
    }
    if (inBlock) current.push(line);
  });
  return blocks;
}

// Doc snippets take several shapes: a full program, a single top-level
// declaration (func/type/var/const), a bare sequence of statements, or a
// leading `import (...)` block followed by statements. There is no reliable
// way to tell which shape a given fragment is without a real parse, so we
// try every plausible wrapping and accept the block as valid if ANY of them
// parses cleanly. A block is only reported as broken if all strategies fail
// — that keeps false positives low while still catching real syntax errors
// (unbalanced brackets, missing commas, stray tokens), which fail no matter
// how the fragment is wrapped.
// Strips string/rune literal and line-comment contents from a line (best
// effort, single-line only) so bracket counting below isn't confused by
// e.g. `fmt.Println("(details)")`. Multi-line block comments and raw
// strings aren't tracked — an acceptable gap since this is only used to
// find top-level chunk boundaries, and the real syntax validation is
// delegated to gofmt.
function stripStringsAndComments(line) {
  let out = '';
  let i = 0;
  let inStr = false;
  let inRune = false;
  let inRaw = false;
  while (i < line.length) {
    const c = line[i];
    if (inRaw) {
      if (c === '`') inRaw = false;
      i++;
      continue;
    }
    if (inStr || inRune) {
      if (c === '\\') {
        i += 2;
        continue;
      }
      if ((inStr && c === '"') || (inRune && c === "'")) {
        inStr = false;
        inRune = false;
      }
      i++;
      continue;
    }
    if (c === '"') {
      inStr = true;
      i++;
      continue;
    }
    if (c === "'") {
      inRune = true;
      i++;
      continue;
    }
    if (c === '`') {
      inRaw = true;
      i++;
      continue;
    }
    if (c === '/' && line[i + 1] === '/') break;
    out += c;
    i++;
  }
  return out;
}

function bracketDelta(line) {
  const stripped = stripStringsAndComments(line);
  let delta = 0;
  for (const ch of stripped) {
    if (ch === '{' || ch === '(' || ch === '[') delta++;
    else if (ch === '}' || ch === ')' || ch === ']') delta--;
  }
  return delta;
}

// Groups top-level lines into "declarations" (import/type/func/var/const)
// kept at the file top level, and everything else ("statements") gathered
// into a single synthetic `func _() { ... }` — this is what lets a mixed
// block (e.g. a `type User struct {...}` followed by `injector := do.New()`)
// parse as one file instead of needing to guess which shape it is.
const DECL_KEYWORDS = new Set([
  'import',
  'type',
  'func',
  'var',
  'const',
  'package',
]);

function smartWrap(code) {
  const lines = code.split('\n');
  let depth = 0;
  let mode = 'stmt';
  const declLines = [];
  const stmtLines = [];

  lines.forEach((line) => {
    const startDepth = depth;
    depth += bracketDelta(line);

    const trimmed = line.trim();
    if (startDepth === 0 && trimmed.length > 0) {
      const firstWord = trimmed.split(/[\s(]/)[0];
      mode = DECL_KEYWORDS.has(firstWord) ? 'decl' : 'stmt';
    }
    (mode === 'decl' ? declLines : stmtLines).push(line);
  });

  const declBlock = declLines.join('\n').trim();
  const stmtBlock = stmtLines.join('\n').trim();

  let out = 'package p\n\n';
  if (declBlock) out += `${declBlock}\n\n`;
  if (stmtBlock) out += `func _() {\n${stmtBlock}\n}\n`;
  return out;
}

function splitLeadingImports(code) {
  const m = code.match(/^\s*import\s*\(([\s\S]*?)\)\s*\n?/);
  if (m) {
    return {imports: `import (${m[1]})`, rest: code.slice(m[0].length)};
  }
  const m2 = code.match(/^\s*import\s+"[^"]+"\s*\n?/);
  if (m2) {
    return {imports: m2[0].trim(), rest: code.slice(m2[0].length)};
  }
  return null;
}

// A block is a full, standalone file if it declares `package X`, possibly
// preceded by comment lines (e.g. a `// main.go` filename hint).
const FULL_FILE_RE = /^(\s*\/\/[^\n]*\n|\s*\n)*\s*package\s+\w/;

// Returns the trimmed, top-level (bracket-depth 0) non-blank, non-comment
// lines of a block — the same depth bookkeeping smartWrap() uses to tell
// declarations from statements, reused here to classify a block before
// attempting to parse it at all.
function topLevelLines(code) {
  const lines = code.split('\n');
  let depth = 0;
  const result = [];
  for (const line of lines) {
    const startDepth = depth;
    depth += bracketDelta(line);
    const trimmed = line.trim();
    if (startDepth === 0 && trimmed.length > 0 && !trimmed.startsWith('//')) {
      result.push(trimmed);
    }
  }
  return result;
}

// Matches a pseudo-signature line like `injector.HealthCheck() map[string]error`
// or `do.HealthCheck[T any](do.Injector) (do.ExplainServiceOutput, bool)`: a
// dotted-identifier call/index head, a single non-nested paren group, then
// more content separated by real whitespace (a return type, possibly a
// tuple) that isn't a chained call (`.Foo()`). Never valid as a standalone
// statement or declaration.
const SIGNATURE_LINE_RE = /^[\w.]+(\[[^\]]*\])?\([^()]*\)\s+[^.\s]/;

// Matches an isolated composite-literal field (`Ident: value,`) — valid only
// inside an enclosing struct/map literal, which these before/after diff
// snippets deliberately omit.
const FIELD_FRAGMENT_RE = /^\w+:\s/;

// Matches a bare "type shape" opener (`do.InjectorOpts{`, nothing else on
// the line) used to sketch a struct's fields without the surrounding
// `type X struct` — these mix `Name Type` and `Name: Type` field syntax
// inconsistently and were never meant to compile.
const TYPE_SHAPE_OPEN_RE = /^[\w.]+\{$/;

function isSpecOrFragmentBlock(code) {
  return topLevelLines(code).some(
    (line) =>
      SIGNATURE_LINE_RE.test(line) ||
      FIELD_FRAGMENT_RE.test(line) ||
      TYPE_SHAPE_OPEN_RE.test(line),
  );
}

// Migration guides sometimes show a complete "before" and "after" program,
// each with its own `package main`, inside a single fence. Splits on every
// top-level `package` line so each program is validated independently
// instead of being parsed (and rejected) as one file with two packages.
function splitMultiPackage(code) {
  const lines = code.split('\n');
  let depth = 0;
  const boundaries = [];
  lines.forEach((line, idx) => {
    const startDepth = depth;
    depth += bracketDelta(line);
    if (startDepth === 0 && /^\s*package\s+\w/.test(line)) {
      boundaries.push(idx);
    }
  });
  if (boundaries.length < 2) return null;
  return boundaries.map((start, i) => {
    const end = i + 1 < boundaries.length ? boundaries[i + 1] : lines.length;
    return lines.slice(start, end).join('\n');
  });
}

function candidateWrappings(code) {
  if (FULL_FILE_RE.test(code)) return [code];
  const trimmed = code.trimStart();

  const candidates = [
    smartWrap(code), // mixed declarations + statements, imports included
    `package p\n\n${code}\n`, // top-level declarations only (func/type/var/const)
    `package p\n\nfunc _() {\n${code}\n}\n`, // bare statements only
  ];

  const split = splitLeadingImports(trimmed);
  if (split) {
    candidates.push(
      `package p\n\n${split.imports}\n\nfunc _() {\n${split.rest}\n}\n`,
    );
    candidates.push(`package p\n\n${split.imports}\n\n${split.rest}\n`);
  }

  return candidates;
}

function gofmtCheck(source, tmpDir, name) {
  const file = path.join(tmpDir, `${name}.go`);
  fs.writeFileSync(file, source);
  try {
    execFileSync('gofmt', ['-l', file], {stdio: 'pipe'});
    return null;
  } catch (err) {
    return (err.stderr ? err.stderr.toString() : err.message).trim();
  }
}

// Returns { skipped: true } for documentation pseudocode that was never
// meant to compile standalone, { error } for a syntax error, or null when
// the block (or, for a multi-package block, every segment) parses cleanly.
function checkBlock(rawCode, tmpDir, index) {
  const code = normalizePlaceholders(rawCode);

  if (isSpecOrFragmentBlock(code)) {
    return {skipped: true};
  }

  const segments = splitMultiPackage(code);
  if (segments) {
    const errors = segments
      .map((seg, i) => gofmtCheck(seg, tmpDir, `${index}_pkg${i}`))
      .filter(Boolean);
    return errors.length ? {error: errors.join('\n')} : null;
  }

  const candidates = candidateWrappings(code);
  const errors = [];
  for (let i = 0; i < candidates.length; i++) {
    const err = gofmtCheck(candidates[i], tmpDir, `${index}_${i}`);
    if (!err) return null; // at least one wrapping is valid Go syntax
    errors.push(err);
  }
  // every strategy failed: report the shortest error, usually the most relevant
  return {error: errors.sort((a, b) => a.length - b.length)[0]};
}

function main() {
  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'do-doc-examples-'));
  let hasError = false;
  let checked = 0;
  let skipped = 0;

  try {
    const files = collectTargetFiles();
    files.forEach((filePath) => {
      const content = fs.readFileSync(filePath, 'utf8');
      const blocks = extractGoBlocks(content);
      blocks.forEach((block, i) => {
        const result = checkBlock(
          block.code,
          tmpDir,
          `${path.basename(filePath)}_${i}`,
        );
        if (result && result.skipped) {
          skipped += 1;
          return;
        }
        checked += 1;
        if (result && result.error) {
          hasError = true;
          const rel = path.relative(repoRoot, filePath);
          console.error(
            `\nSyntax error in ${rel}:${block.line} (go block #${i + 1}):`,
          );
          console.error(result.error);
        }
      });
    });
  } finally {
    fs.rmSync(tmpDir, {recursive: true, force: true});
  }

  if (hasError) {
    console.error(
      `\nFAILED: one or more \`\`\`go blocks are not syntactically valid Go.`,
    );
    process.exit(1);
  }
  console.log(
    `OK: ${checked} \`\`\`go blocks are syntactically valid (${skipped} spec/fragment blocks skipped).`,
  );
}

main();
