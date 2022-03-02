/**
 * @typedef {import('mdast-util-from-markdown').Extension} FromMarkdownExtension
 * @typedef {import('mdast-util-from-markdown').Handle} FromMarkdownHandle
 * @typedef {import('mdast-util-to-markdown').Options} ToMarkdownExtension
 * @typedef {import('mdast-util-to-markdown').Handle} ToMarkdownHandle
 * @typedef {import('estree-jsx').Program} Program
 * @typedef {import('./complex-types').MDXFlowExpression} MDXFlowExpression
 * @typedef {import('./complex-types').MDXTextExpression} MDXTextExpression
 */
/** @type {FromMarkdownExtension} */
export const mdxExpressionFromMarkdown: FromMarkdownExtension
/** @type {ToMarkdownExtension} */
export const mdxExpressionToMarkdown: ToMarkdownExtension
export type FromMarkdownExtension = import('mdast-util-from-markdown').Extension
export type FromMarkdownHandle = import('mdast-util-from-markdown').Handle
export type ToMarkdownExtension = import('mdast-util-to-markdown').Options
export type ToMarkdownHandle = import('mdast-util-to-markdown').Handle
export type Program = import('estree-jsx').Program
export type MDXFlowExpression = import('./complex-types').MDXFlowExpression
export type MDXTextExpression = import('./complex-types').MDXTextExpression
