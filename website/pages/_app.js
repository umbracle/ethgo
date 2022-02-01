import 'nextra-theme-docs/style.css'
import Prism from 'prism-react-renderer/prism'

export default function Nextra({ Component, pageProps }) {
  return <Component {...pageProps} />
}

(typeof global !== "undefined" ? global : window).Prism = Prism
require("prismjs/components/prism-go")
