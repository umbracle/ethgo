# mdast-util-mdxjs-esm

[![Build][build-badge]][build]
[![Coverage][coverage-badge]][coverage]
[![Downloads][downloads-badge]][downloads]
[![Size][size-badge]][size]
[![Sponsors][sponsors-badge]][collective]
[![Backers][backers-badge]][collective]
[![Chat][chat-badge]][chat]

Extension for [`mdast-util-from-markdown`][from-markdown] and/or
[`mdast-util-to-markdown`][to-markdown] to support MDX.js ESM import/exports.
When parsing (`from-markdown`), must be combined with
[`micromark-extension-mdxjs-esm`][extension].

This utility handles parsing and serializing.
See [`micromark-extension-mdxjs-esm`][extension] for how the syntax works.

## When to use this

Use [`mdast-util-mdx`][mdast-util-mdx] if you want all of MDX / MDX.js.
Use this otherwise.

## Install

This package is [ESM only](https://gist.github.com/sindresorhus/a39789f98801d908bbc7ff3ecc99d99c):
Node 12+ is needed to use it and it must be `import`ed instead of `require`d.

[npm][]:

```sh
npm install mdast-util-mdxjs-esm
```

## Use

Say we have an MDX.js file, `example.mdx`:

```mdx
import a from 'b'
export var c = ''

d
```

And our module, `example.js`, looks as follows:

```js
import fs from 'fs'
import * as acorn from 'acorn'
import {fromMarkdown} from 'mdast-util-from-markdown'
import {toMarkdown} from 'mdast-util-to-markdown'
import {mdxjsEsm} from 'micromark-extension-mdxjs-esm'
import {mdxjsEsmFromMarkdown, mdxjsEsmToMarkdown} from 'mdast-util-mdxjs-esm'

const doc = fs.readFileSync('example.mdx')

const tree = fromMarkdown(doc, {
  extensions: [mdxjsEsm({acorn, addResult: true})],
  mdastExtensions: [mdxjsEsmFromMarkdown]
})

console.log(tree)

const out = toMarkdown(tree, {extensions: [mdxjsEsmToMarkdown]})

console.log(out)
```

Now, running `node example` yields (positional info removed for brevity):

```js
{
  type: 'root',
  children: [
    {
      type: 'mdxjsEsm',
      value: "import a from 'b'\nexport var c = ''",
      data: {
        estree: {
          type: 'Program',
          body: [
            {
              type: 'ImportDeclaration',
              specifiers: [
                {
                  type: 'ImportDefaultSpecifier',
                  local: {type: 'Identifier', name: 'a'}
                }
              ],
              source: {type: 'Literal', value: 'b', raw: "'b'"}
            },
            {
              type: 'ExportNamedDeclaration',
              declaration: {
                type: 'VariableDeclaration',
                declarations: [
                  {
                    type: 'VariableDeclarator',
                    id: {type: 'Identifier', name: 'c'},
                    init: {type: 'Literal', value: '', raw: "''"}
                  }
                ],
                kind: 'var'
              },
              specifiers: [],
              source: null
            }
          ],
          sourceType: 'module'
        }
      }
    },
    {type: 'paragraph', children: [{type: 'text', value: 'd'}]}
  ]
}
```

```markdown
import a from 'b'
export var c = ''

d
```

## API

This package exports the following identifier: `mdxjsEsmFromMarkdown`,
`mdxjsEsmToMarkdown`.
There is no default export.

### `mdxjsEsmFromMarkdown`

### `mdxjsEsmToMarkdown`

Support MDX.js ESM import/exports.
The exports are extensions, respectively for
[`mdast-util-from-markdown`][from-markdown] and
[`mdast-util-to-markdown`][to-markdown].

When using the [syntax extension with `addResult`][extension], nodes will have
a `data.estree` field set to an [ESTree][]

## Syntax tree

The following interfaces are added to **[mdast][]** by this utility.

### Nodes

#### `MDXJSEsm`

```idl
interface MDXJSEsm <: Literal {
  type: "mdxjsEsm"
}
```

**MDXJSEsm** (**[Literal][dfn-literal]**) represents ESM import/exports embedded
in MDX.
It can be used where **[flow][dfn-flow-content]** content is expected.
Its content is represented by its `value` field.

For example, the following Markdown:

```markdown
import a from 'b'
```

Yields:

```js
{
  type: 'mdxjsEsm',
  value: 'import a from \'b\''
}
```

### Content model

#### `FlowContent` (MDX.js ESM)

```idl
type FlowContentMdxjsEsm = MDXJSEsm | FlowContent
```

Note that when ESM is present, it can only exist as top-level content: if it has
a *[parent][dfn-parent]*, that parent must be **[Root][dfn-root]**.

## Related

*   [`remarkjs/remark`][remark]
    — markdown processor powered by plugins
*   [`remarkjs/remark-mdx`][remark-mdx]
    — remark plugin to support MDX
*   [`syntax-tree/mdast-util-from-markdown`][from-markdown]
    — mdast parser using `micromark` to create mdast from markdown
*   [`syntax-tree/mdast-util-to-markdown`][to-markdown]
    — mdast serializer to create markdown from mdast
*   [`syntax-tree/mdast-util-mdx`][mdast-util-mdx]
    — mdast utility to support MDX
*   [`micromark/micromark`][micromark]
    — the smallest commonmark-compliant markdown parser that exists
*   [`micromark/micromark-extension-mdxjs-esm`][extension]
    — micromark extension to parse MDX.js ESM

## Contribute

See [`contributing.md` in `syntax-tree/.github`][contributing] for ways to get
started.
See [`support.md`][support] for ways to get help.

This project has a [code of conduct][coc].
By interacting with this repository, organization, or community you agree to
abide by its terms.

## License

[MIT][license] © [Titus Wormer][author]

<!-- Definitions -->

[build-badge]: https://github.com/syntax-tree/mdast-util-mdxjs-esm/workflows/main/badge.svg

[build]: https://github.com/syntax-tree/mdast-util-mdxjs-esm/actions

[coverage-badge]: https://img.shields.io/codecov/c/github/syntax-tree/mdast-util-mdxjs-esm.svg

[coverage]: https://codecov.io/github/syntax-tree/mdast-util-mdxjs-esm

[downloads-badge]: https://img.shields.io/npm/dm/mdast-util-mdxjs-esm.svg

[downloads]: https://www.npmjs.com/package/mdast-util-mdxjs-esm

[size-badge]: https://img.shields.io/bundlephobia/minzip/mdast-util-mdxjs-esm.svg

[size]: https://bundlephobia.com/result?p=mdast-util-mdxjs-esm

[sponsors-badge]: https://opencollective.com/unified/sponsors/badge.svg

[backers-badge]: https://opencollective.com/unified/backers/badge.svg

[collective]: https://opencollective.com/unified

[chat-badge]: https://img.shields.io/badge/chat-discussions-success.svg

[chat]: https://github.com/syntax-tree/unist/discussions

[npm]: https://docs.npmjs.com/cli/install

[license]: license

[author]: https://wooorm.com

[contributing]: https://github.com/syntax-tree/.github/blob/HEAD/contributing.md

[support]: https://github.com/syntax-tree/.github/blob/HEAD/support.md

[coc]: https://github.com/syntax-tree/.github/blob/HEAD/code-of-conduct.md

[mdast]: https://github.com/syntax-tree/mdast

[remark]: https://github.com/remarkjs/remark

[from-markdown]: https://github.com/syntax-tree/mdast-util-from-markdown

[to-markdown]: https://github.com/syntax-tree/mdast-util-to-markdown

[micromark]: https://github.com/micromark/micromark

[extension]: https://github.com/micromark/micromark-extension-mdxjs-esm

[mdast-util-mdx]: https://github.com/syntax-tree/mdast-util-mdx

[estree]: https://github.com/estree/estree

[dfn-literal]: https://github.com/syntax-tree/mdast#literal

[dfn-parent]: https://github.com/syntax-tree/unist#parent-1

[dfn-root]: https://github.com/syntax-tree/mdast#root

[dfn-flow-content]: #flowcontent-mdxjs-esm

[remark-mdx]: https://github.com/mdx-js/mdx/tree/next/packages/remark-mdx
