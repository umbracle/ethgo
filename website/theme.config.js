export default {
    projectLink: 'https://github.com/umbracle/ethgo', // GitHub link in the navbar
    docsRepositoryBase: 'https://github.com/umbracle/ethgo/tree/master/website/pages', // base URL for the docs repository
    titleSuffix: ' – Ethgo',
    nextLinks: true,
    prevLinks: true,
    search: false,
    customSearch: null, // customizable, you can use algolia for example
    darkMode: true,
    footer: true,
    footerText: (
      <>
        Powered by <a href="https://umbracle.io">Umbracle</a>
      </>
    ),
    footerEditLink: `Edit this page on GitHub`,
    floatTOC: true,
    logo: (
        <>
          <span className="mr-2 font-extrabold hidden md:inline">Ethgo</span>
          <span className="text-gray-600 font-normal hidden md:inline">
            Go Ethereum SDK
          </span>
        </>
    ),
    head: (
        <>
            <meta name="viewport" content="width=device-width, initial-scale=1.0" />
            <meta name="description" content="Ethgo: lightweight Go Ethereum SDK" />
            <meta name="og:title" content="Ethgo: lightweight Go Ethereum SDK" />
        </>
    ),
}