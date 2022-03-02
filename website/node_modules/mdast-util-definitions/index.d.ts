/**
 *
 * @param {Node} node
 */
export function definitions(
  node: Node
): (identifier: string) => Definition | null
export type Node = import('mdast').Root | import('mdast').Content
export type Definition = import('mdast').Definition
export type DefinitionVisitor = import('unist-util-visit').Visitor<Definition>
