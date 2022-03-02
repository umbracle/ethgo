# parse-git-url

A lib for parsing the URL of GitHub, GitLab, and Bitbucket repositories.

## Usage

```js
import parseGitUrl from 'parse-git-url'

parseGitUrl('https://github.com/vercel/swr'))
// => {
//   type: 'github',
//   owner: 'vercel',
//   name: 'swr',
//   branch: '',
//   sha: '',
//   subdir: ''
// }

parseGitUrl('https://google.com')
// => null
```

It supports parsing various URL schemas including SSH, branch, sha, commit, subdirectories, subgroups (GitLab), etc.

## Author

Shu Ding ([@shuding_](https://twitter.com/shuding)) – [Vercel](https://vercel.com)
