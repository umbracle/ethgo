/**
 * @typedef {import('mdast').Text} Text
 * @typedef {import('../index.js').Handler} Handler
 */

import {u} from 'unist-builder'

/**
 * @type {Handler}
 * @param {Text} node
 */
export function text(h, node) {
  return h.augment(
    node,
    u('text', String(node.value).replace(/[ \t]*(\r?\n|\r)[ \t]*/g, '$1'))
  )
}
