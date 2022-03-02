# estree-util-is-identifier-name

[![Build][build-badge]][build]
[![Coverage][coverage-badge]][coverage]
[![Downloads][downloads-badge]][downloads]
[![Size][size-badge]][size]

Check if something can be an identifier name.

## Install

This package is ESM only: Node 12+ is needed to use it and it must be `import`ed
instead of `require`d.

[npm][]:

```sh
npm install estree-util-is-identifier-name
```

## Use

```js
import {name, start, cont} from 'estree-util-is-identifier-name'

name('$something69') // => true
name('69') // => false
name('var') // => true (this does not handle keywords)

start(48) // => false (character code for `0`)
cont(48) // => true (character code for `0`)
```

## API

This package exports the following identifiers: `name`, `start`, and `cont`.
There is no default export.

### `name(value)`

Checks if the given string is a valid identifier name.

### `start(code)`

Checks if the given character code can start an identifier.

### `cont(code)`

Checks if the given character code can continue an identifier.

## Related

*   [`goto-bus-stop/estree-is-identifier`](https://github.com/goto-bus-stop/estree-is-identifier)
    — check if an AST node is an identifier

## License

[MIT][license] © [Titus Wormer][author]

<!-- Definitions -->

[build-badge]: https://github.com/wooorm/estree-util-is-identifier-name/workflows/main/badge.svg

[build]: https://github.com/wooorm/estree-util-is-identifier-name/actions

[coverage-badge]: https://img.shields.io/codecov/c/github/wooorm/estree-util-is-identifier-name.svg

[coverage]: https://codecov.io/github/wooorm/estree-util-is-identifier-name

[downloads-badge]: https://img.shields.io/npm/dm/estree-util-is-identifier-name.svg

[downloads]: https://www.npmjs.com/package/estree-util-is-identifier-name

[size-badge]: https://img.shields.io/bundlephobia/minzip/estree-util-is-identifier-name.svg

[size]: https://bundlephobia.com/result?p=estree-util-is-identifier-name

[npm]: https://docs.npmjs.com/cli/install

[license]: license

[author]: https://wooorm.com
