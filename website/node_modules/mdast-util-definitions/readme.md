# mdast-util-definitions

[![Build][build-badge]][build]
[![Coverage][coverage-badge]][coverage]
[![Downloads][downloads-badge]][downloads]
[![Size][size-badge]][size]
[![Sponsors][sponsors-badge]][collective]
[![Backers][backers-badge]][collective]
[![Chat][chat-badge]][chat]

[**mdast**][mdast] utility to get definitions by `identifier`.

Supports funky keys, like `__proto__` or `toString`.

## Install

This package is [ESM only](https://gist.github.com/sindresorhus/a39789f98801d908bbc7ff3ecc99d99c):
Node 12+ is needed to use it and it must be `import`ed instead of `require`d.

[npm][]:

```sh
npm install mdast-util-definitions
```

## Use

```js
import {remark} from 'remark'
import {definitions} from 'mdast-util-definitions'

const tree = remark().parse('[example]: https://example.com "Example"')

const definition = definitions(tree)

definition('example')
// => {type: 'definition', 'title': 'Example', …}

definition('foo')
// => null
```

## API

This package exports the following identifiers: `definitions`.
There is no default export.

### `definitions(tree)`

Create a cache of all [definition][]s in [`tree`][node].

Uses CommonMark precedence: prefers the first definitions for duplicate
definitions.

###### Returns

[`Function`][fn-definition]

### `definition(identifier)`

###### Parameters

*   `identifier` (`string`) — [Identifier][] of [definition][].

###### Returns

[`Node?`][node] — [Definition][], if found.

## Security

Use of `mdast-util-definitions` does not involve [**hast**][hast] or user
content so there are no openings for [cross-site scripting (XSS)][xss] attacks.

Additionally, safe guards are in place to protect against prototype poisoning.

## Related

*   [`unist-util-index`](https://github.com/syntax-tree/unist-util-index)
    — index property values or computed keys to nodes

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

[build-badge]: https://github.com/syntax-tree/mdast-util-definitions/workflows/main/badge.svg

[build]: https://github.com/syntax-tree/mdast-util-definitions/actions

[coverage-badge]: https://img.shields.io/codecov/c/github/syntax-tree/mdast-util-definitions.svg

[coverage]: https://codecov.io/github/syntax-tree/mdast-util-definitions

[downloads-badge]: https://img.shields.io/npm/dm/mdast-util-definitions.svg

[downloads]: https://www.npmjs.com/package/mdast-util-definitions

[size-badge]: https://img.shields.io/bundlephobia/minzip/mdast-util-definitions.svg

[size]: https://bundlephobia.com/result?p=mdast-util-definitions

[sponsors-badge]: https://opencollective.com/unified/sponsors/badge.svg

[backers-badge]: https://opencollective.com/unified/backers/badge.svg

[collective]: https://opencollective.com/unified

[chat-badge]: https://img.shields.io/badge/chat-discussions-success.svg

[chat]: https://github.com/syntax-tree/unist/discussions

[license]: license

[author]: https://wooorm.com

[npm]: https://docs.npmjs.com/cli/install

[contributing]: https://github.com/syntax-tree/.github/blob/HEAD/contributing.md

[support]: https://github.com/syntax-tree/.github/blob/HEAD/support.md

[coc]: https://github.com/syntax-tree/.github/blob/HEAD/code-of-conduct.md

[mdast]: https://github.com/syntax-tree/mdast

[node]: https://github.com/syntax-tree/unist#node

[fn-definition]: #definitionidentifier

[definition]: https://github.com/syntax-tree/mdast#definition

[identifier]: https://github.com/syntax-tree/mdast#association

[xss]: https://en.wikipedia.org/wiki/Cross-site_scripting

[hast]: https://github.com/syntax-tree/hast
