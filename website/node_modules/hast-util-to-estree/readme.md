# hast-util-to-estree

[![Build][build-badge]][build]
[![Coverage][coverage-badge]][coverage]
[![Downloads][downloads-badge]][downloads]
[![Size][size-badge]][size]
[![Sponsors][sponsors-badge]][collective]
[![Backers][backers-badge]][collective]
[![Chat][chat-badge]][chat]

**[hast][]** utility to transform a *[tree][]* to **[estree][]** JSX.

## Install

This package is [ESM only](https://gist.github.com/sindresorhus/a39789f98801d908bbc7ff3ecc99d99c):
Node 12+ is needed to use it and it must be `import`ed instead of `require`d.

[npm][]:

```sh
npm install hast-util-to-estree
```

## Use

Say we have the following HTML file, `example.html`:

```html
<!doctype html>
<html lang=en>
<title>Hi!</title>
<link rel=stylesheet href=index.css>
<h1>Hello, world!</h1>
<a download style="width:1;height:10px"></a>
<!--commentz-->
<svg xmlns="http://www.w3.org/2000/svg">
  <title>SVG `&lt;ellipse&gt;` element</title>
  <ellipse
    cx="120"
    cy="70"
    rx="100"
    ry="50"
  />
</svg>
<script src="index.js"></script>
```

And our script, `example.js`, is:

```js
import fs from 'node:fs'
import parse5 from 'parse5'
import {fromParse5} from 'hast-util-from-parse5'
import {toEstree} from 'hast-util-to-estree'
import recast from 'recast'

const hast = fromParse5(parse5.parse(String(fs.readFileSync('example.html'))))

const estree = toEstree(hast)

estree.comments = null // `recast` doesn’t like comments on the root.
console.log(recast.prettyPrint(estree).code)
```

Now, `node example` (and prettier), yields:

```jsx
;<>
  <html lang="en">
    <head>
      <title>{'Hi!'}</title>
      {'\n'}
      <link rel="stylesheet" href="index.css" />
      {'\n'}
    </head>
    <body>
      <h1>{'Hello, world!'}</h1>
      {'\n'}
      <a
        download
        style={{
          width: '1',
          height: '10px'
        }}
      />
      {'\n'}
      {/*commentz*/}
      {'\n'}
      <svg xmlns="http://www.w3.org/2000/svg">
        {'\n  '}
        <title>{'SVG `<ellipse>` element'}</title>
        {'\n  '}
        <ellipse cx="120" cy="70" rx="100" ry="50" />
        {'\n'}
      </svg>
      {'\n'}
      <script src="index.js" />
      {'\n'}
    </body>
  </html>
</>
```

## API

This package exports the following identifiers: `toEstree`.
There is no default export.

### `toEstree(tree, options?)`

Transform a **[hast][]** *[tree][]* to an **[estree][]** (JSX).

##### `options`

*   `space` (enum, `'svg'` or `'html'`, default: `'html'`)
    — Whether node is in the `'html'` or `'svg'` space.
    If an `svg` element is found when inside the HTML space, `toEstree`
    automatically switches to the SVG space when entering the element, and
    switches back when exiting

###### Returns

**[estree][]** — a *[Program][]* node, whose last child in `body` is most
likely an `ExpressionStatement` whose expression is a `JSXFragment` or a
`JSXElement`.

Typically, there is only one node in `body`, however, this utility also supports
embedded MDX nodes in the HTML (when [`mdast-util-mdx`][mdast-util-mdx] is used
with mdast to parse markdown before passing its nodes through to hast).
When MDX ESM import/exports are used, those nodes are added before the fragment
or element in body.

###### Note

Comments are both attached to the tree in their neighbouring nodes (recast and
babel style), and added as a `comments` array on the program node (espree
style).
You may have to do `program.comments = null` for certain compilers.

There aren’t many great estree serializers out there that support JSX.
[recast][] does a fine job.
Or use [`estree-util-build-jsx`][build-jsx] to turn JSX into function
calls and then serialize with whatever (astring, escodegen).

## Security

You’re working with JavaScript.
It’s not safe.

## Related

*   [`hastscript`][hastscript]
    — Hyperscript compatible interface for creating nodes
*   [`hast-util-from-dom`](https://github.com/syntax-tree/hast-util-from-dom)
    — Transform a DOM tree to hast
*   [`unist-builder`](https://github.com/syntax-tree/unist-builder)
    — Create any unist tree
*   [`xastscript`](https://github.com/syntax-tree/xastscript)
    — Create a xast tree
*   [`estree-util-build-jsx`][build-jsx]
    — Transform JSX to function calls

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

[build-badge]: https://github.com/syntax-tree/hast-util-to-estree/workflows/main/badge.svg

[build]: https://github.com/syntax-tree/hast-util-to-estree/actions

[coverage-badge]: https://img.shields.io/codecov/c/github/syntax-tree/hast-util-to-estree.svg

[coverage]: https://codecov.io/github/syntax-tree/hast-util-to-estree

[downloads-badge]: https://img.shields.io/npm/dm/hast-util-to-estree.svg

[downloads]: https://www.npmjs.com/package/hast-util-to-estree

[size-badge]: https://img.shields.io/bundlephobia/minzip/hast-util-to-estree.svg

[size]: https://bundlephobia.com/result?p=hast-util-to-estree

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

[hastscript]: https://github.com/syntax-tree/hastscript

[tree]: https://github.com/syntax-tree/unist#tree

[hast]: https://github.com/syntax-tree/hast

[estree]: https://github.com/estree/estree

[program]: https://github.com/estree/estree/blob/master/es5.md#programs

[recast]: https://github.com/benjamn/recast

[mdast-util-mdx]: https://github.com/syntax-tree/mdast-util-mdx

[build-jsx]: https://github.com/wooorm/estree-util-build-jsx
