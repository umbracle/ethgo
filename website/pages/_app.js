import 'nextra-theme-docs/style.css'

import Prism from 'prism-react-renderer/prism'
(typeof global !== "undefined" ? global : window).Prism = Prism
require("prismjs/components/prism-kotlin")
require("prismjs/components/prism-csharp")
require("prismjs/components/prism-go")

export default function Nextra({ Component, pageProps }) {
  return <Component {...pageProps} />
}
