# remark-mdx-code-meta

[![github actions][github actions badge]][github actions] [![npm][npm badge]][npm]
[![prettier][prettier badge]][prettier]

> A [remark][] MDX plugin for using markdown code block metadata

## Installation

```sh
npm install remark-mdx-code-meta
```

## Usage

This plugin interprets markdown code block metadata as JSX props.

For example, given a file named `example.mdx` with the following contents:

````markdown
```js copy filename="awesome.js" onUsage={props.beAwesome} {...props}
console.log('Everything is awesome!');
```
````

The following script:

```js
import { readFileSync } from 'fs';

import { remarkMdxCodeMeta } from 'remark-mdx-code-meta';
import { compileSync } from 'xdm';

const { contents } = compileSync(readFileSync('example.mdx'), {
  jsx: true,
  remarkPlugins: [remarkMdxCodeMeta],
});
console.log(contents);
```

Roughly yields:

```jsx
export default function MDXContent(props) {
  return (
    <pre copy filename="awesome.js" onUsage={props.beAwesome} {...props}>
      <code className="language-js">{"console.log('Everything is awesome!');\n"}</code>
    </pre>
  );
}
```

Of course the `<pre />` element doesn’t support those custom props. Use custom [components][] to
give the props meaning.

[components]: https://github.com/wooorm/xdm#components
[github actions badge]:
  https://github.com/remcohaszing/remark-mdx-code-meta/actions/workflows/ci.yml/badge.svg
[github actions]: https://github.com/remcohaszing/remark-mdx-code-meta/actions/workflows/ci.yml
[npm badge]: https://img.shields.io/npm/v/remark-mdx-code-meta
[npm]: https://www.npmjs.com/package/remark-mdx-code-meta
[prettier badge]: https://img.shields.io/badge/code_style-prettier-ff69b4.svg
[prettier]: https://prettier.io
[remark]: https://remark.js.org
