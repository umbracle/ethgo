import 'nextra-theme-docs/style.css'

import Prism from 'prism-react-renderer/prism'
(typeof global !== "undefined" ? global : window).Prism = Prism
require("prismjs/components/prism-go")

import Head from 'next/head';

export default function Nextra({ Component, pageProps }) {
  return (
    <>
      <Head>
        <script defer data-domain="ethgo.dev" src="https://plausible.io/js/plausible.js"></script>
      </Head>
      <Component {...pageProps} />
    </>
  )
}
