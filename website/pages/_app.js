import 'nextra-theme-docs/style.css'

import Prism from 'prism-react-renderer/prism'
(typeof global !== "undefined" ? global : window).Prism = Prism
require("prismjs/components/prism-go")

import Head from 'next/head';

const prod = process.env.NODE_ENV === 'production'

export default function Nextra({ Component, pageProps }) {
  return (
    <>
      <Head>
        {prod &&
          <script defer data-domain="ethgoproject.io" src="https://plausible.io/js/script.js"></script>
        }
      </Head>
      <Component {...pageProps} />
    </>
  )
}
