/**
 * @param {Node|MDXJsxAttributeValueExpression|MDXJsxAttribute|MDXJsxExpressionAttribute|MDXJsxFlowElement|MDXJsxTextElement|MDXFlowExpression|MDXTextExpression} tree
 * @param {Options} options
 * @returns {EstreeProgram}
 */
export function toEstree(
  tree:
    | Node
    | MDXJsxAttributeValueExpression
    | MDXJsxAttribute
    | MDXJsxExpressionAttribute
    | MDXJsxFlowElement
    | MDXJsxTextElement
    | MDXFlowExpression
    | MDXTextExpression,
  options?: Options
): EstreeProgram
export type Root = import('hast').Root
export type Element = import('hast').Element
export type Text = import('hast').Text
export type Comment = import('hast').Comment
export type Properties = import('hast').Properties
export type Content = import('hast').Content
export type Node = Root | Content
export type Parent = Extract<Node, import('unist').Parent>
export type EstreeNode = import('estree-jsx').Node
export type EstreeProgram = import('estree-jsx').Program
export type EstreeJsxExpressionContainer =
  import('estree-jsx').JSXExpressionContainer
export type EstreeJsxElement = import('estree-jsx').JSXElement
export type EstreeJsxOpeningElement = import('estree-jsx').JSXOpeningElement
export type EstreeJsxFragment = import('estree-jsx').JSXFragment
export type EstreeJsxAttribute = import('estree-jsx').JSXAttribute
export type EstreeJsxSpreadAttribute = import('estree-jsx').JSXSpreadAttribute
export type EstreeComment = import('estree-jsx').Comment
export type EstreeDirective = import('estree-jsx').Directive
export type EstreeStatement = import('estree-jsx').Statement
export type EstreeModuleDeclaration = import('estree-jsx').ModuleDeclaration
export type EstreeExpression = import('estree-jsx').Expression
export type EstreeProperty = import('estree-jsx').Property
export type JSXIdentifier = import('estree-jsx').JSXIdentifier
export type JSXMemberExpression = import('estree-jsx').JSXMemberExpression
export type EstreeJsxElementName = EstreeJsxOpeningElement['name']
export type EstreeJsxAttributeName = EstreeJsxAttribute['name']
export type EstreeJsxChild = EstreeJsxElement['children'][number]
export type MDXJsxAttributeValueExpression =
  import('mdast-util-mdx-jsx').MDXJsxAttributeValueExpression
export type MDXJsxAttribute = import('mdast-util-mdx-jsx').MDXJsxAttribute
export type MDXJsxExpressionAttribute =
  import('mdast-util-mdx-jsx').MDXJsxExpressionAttribute
export type MDXJsxFlowElement = import('mdast-util-mdx-jsx').MDXJsxFlowElement
export type MDXJsxTextElement = import('mdast-util-mdx-jsx').MDXJsxTextElement
export type MDXFlowExpression =
  import('mdast-util-mdx-expression').MDXFlowExpression
export type MDXTextExpression =
  import('mdast-util-mdx-expression').MDXTextExpression
export type MDXJSEsm = import('mdast-util-mdxjs-esm').MDXJSEsm
export type Info = ReturnType<typeof find>
export type Space = 'html' | 'svg'
export type Handle = (node: any, context: Context) => EstreeJsxChild | null
export type Options = {
  space?: Space | undefined
  handlers?:
    | {
        [x: string]: Handle
      }
    | undefined
}
export type Context = {
  schema: typeof html
  comments: Array<EstreeComment>
  esm: Array<EstreeDirective | EstreeStatement | EstreeModuleDeclaration>
  handle: Handle
}
import {find} from 'property-information'
import {html} from 'property-information'
