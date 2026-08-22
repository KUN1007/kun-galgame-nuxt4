# Spoiler syntax

Two forms, one CSS contract. The renderer lives in
`apps/api/internal/infrastructure/markdown/spoiler.go` (a goldmark extension —
**not** a regex over the rendered HTML; that older approach leaked spoiler spans
into `<pre><code>` and `<code>`).

## Inline — `||…||`

Scoped to one block, so it spans soft and hard line breaks but stops at a blank
line:

```
||秘密||              -> hidden
||第一行              -> hidden, both lines (soft break)
第二行||
||第一行␣␣            -> hidden, both lines, <br> kept (hard break)
第二行||
||第一行              -> LITERAL. a blank line ends the paragraph;
                         use the block form below
第二行||
```

Pairing is per block and left-greedy, like emphasis. Delimiter runs must be
exactly two pipes: `||||` is literal text. CommonMark flanking is deliberately
not applied, so `|| 带空格 ||` still works.

## Block — `:::spoiler` … `:::`

```
:::spoiler
第一段

第二段

- 任意块级内容
:::
```

- Opener: 3+ colons, then `spoiler` **touching** the colons, then only
  whitespace. `::: spoiler` and `:::spoiler 带标签` are not fences — this matches
  remark-directive, so the editor and the server agree on what is a directive.
- Closer: a line of colons, at least as many as the opener. A longer opening
  fence therefore nests: `::::spoiler` wraps an inner `:::spoiler`.
- Unclosed fences run to the end of the document.
- Indented 4+ spaces it is an indented code block, per CommonMark.
- Works inside blockquotes and list items.

Renders to `<div class="kun-spoiler text-transparent kun-spoiler-hidden">`.

## Render contract (`@kungal/ui-vue`)

`KunContent` enhances any element carrying `.kun-spoiler.kun-spoiler-hidden`,
`<span>` or `<div>` alike — it stamps `role="button"`, `tabindex`,
`aria-expanded`, an `aria-label` (only when the host has not set one), and the
particle canvas; click or Enter reveals. The block styling
(`div.kun-spoiler-hidden { width: fit-content; display: flow-root }`, plus
`.kun-spoiler-hidden > :not(.kun-spoiler-canvas) { visibility: hidden }`) has
shipped since ui-vue 2.25.0. Nothing in KunUI needs to change.

## What kun-editor still owes

The forum renders both forms; the editor only produces the inline one.
`:::spoiler` survives a WYSIWYG round trip untouched (verified — the serializer
does not escape the colons), so hand-written blocks are safe today, but there is
no button and no WYSIWYG preview for them. Upstream needs a block node
(`remark-directive` container named `spoiler` + a ProseMirror block node + a
toolbar command) matching the rules above.
