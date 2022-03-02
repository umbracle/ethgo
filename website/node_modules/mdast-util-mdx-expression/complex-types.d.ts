import {Literal} from 'mdast'
import {Program} from 'estree-jsx'

export interface MDXFlowExpression extends Literal {
  type: 'mdxFlowExpression'
  data?: {
    estree?: Program
  } & Literal['data']
}

export interface MDXTextExpression extends Literal {
  type: 'mdxTextExpression'
  data?: {
    estree?: Program
  } & Literal['data']
}

declare module 'mdast' {
  interface StaticPhrasingContentMap {
    mdxTextExpression: MDXTextExpression
  }

  interface BlockContentMap {
    mdxFlowExpression: MDXFlowExpression
  }
}
