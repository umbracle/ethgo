# estree-util-attach-comments

[![Build][build-badge]][build]
[![Coverage][coverage-badge]][coverage]
[![Downloads][downloads-badge]][downloads]
[![Size][size-badge]][size]

Attach semistandard [estree][] comment nodes (such as from [espree][] or
[acorn][] with a couple lines of code) to the nodes in that tree.

This is useful because certain estree parsers give you an array (espree and
acorn) whereas other estree tools expect comments to be embedded on nodes in the
tree.

## Install

This package is ESM only: Node 12+ is needed to use it and it must be `import`ed
instead of `require`d.

[npm][]:

```sh
npm install estree-util-attach-comments
```

## Use

Say we have this weird `code`:

```js
/* 1 */ function /* 2 */ a /* 3 */ (/* 4 */b) /* 5 */ { /* 6 */ return /* 7 */ b + /* 8 */ 1 /* 9 */ }
```

And our script, `example.js`, looks as follows:

```js
import * as acorn from 'acorn'
import recast from 'recast'
import {attachComments} from 'estree-util-attach-comments'

var comments = []
var tree = acorn.parse(code, {ecmaVersion: 2020, onComment: comments})

attachComments(tree, comments)

console.log(recast.print(tree).code)
```

Yields:

```js
/* 1 */
function /* 2 */
a(
    /* 3 */
    /* 4 */
    b
) /* 5 */
{
    /* 6 */
    return (
        /* 7 */
        b + /* 8 */
        1
    );
}/* 9 */
```

Note that the lines are added by `recast` in this case.
And, some of these weird comments are off, but they’re pretty close.

## API

This package exports the following identifiers: `attachComment`.
There is no default export.

### `attachComment(tree, comments)`

Attach semistandard estree comment nodes to the tree.

This mutates the given [`tree`][estree] ([`Program`][program]).
It takes `comments`, walks the tree, and adds comments as close as possible
to where they originated.

Comment nodes are given two boolean fields: `leading` (`true` for `/* a */ b`)
and `trailing` (`true` for `a /* b */`).
Both fields are `false` for dangling comments: `[/* a */]`.
This is what `recast` uses too, and is somewhat similar to Babel, which is not
estree but instead uses `leadingComments`, `trailingComments`, and
`innerComments` arrays on nodes.

The algorithm checks any node: even recent (or future) proposals or nonstandard
syntax such as JSX, because it ducktypes to find nodes instead of having a list
of visitor keys.

The algorithm supports `loc` fields (line/column), `range` fields (offsets),
and direct `start` / `end` fields.

###### Returns

`Node` — The given `tree`.

## License

[MIT][license] © [Titus Wormer][author]

<!-- Definitions -->

[build-badge]: https://github.com/wooorm/estree-util-attach-comments/workflows/main/badge.svg

[build]: https://github.com/wooorm/estree-util-attach-comments/actions

[coverage-badge]: https://img.shields.io/codecov/c/github/wooorm/estree-util-attach-comments.svg

[coverage]: https://codecov.io/github/wooorm/estree-util-attach-comments

[downloads-badge]: https://img.shields.io/npm/dm/estree-util-attach-comments.svg

[downloads]: https://www.npmjs.com/package/estree-util-attach-comments

[size-badge]: https://img.shields.io/bundlephobia/minzip/estree-util-attach-comments.svg

[size]: https://bundlephobia.com/result?p=estree-util-attach-comments

[npm]: https://docs.npmjs.com/cli/install

[license]: license

[author]: https://wooorm.com

[acorn]: https://github.com/acornjs/acorn

[estree]: https://github.com/estree/estree

[espree]: https://github.com/eslint/espree

[program]: https://github.com/estree/estree/blob/master/es5.md#programs
