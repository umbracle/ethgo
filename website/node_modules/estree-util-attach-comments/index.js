var push = [].push

/**
 * @typedef {import('estree').BaseNode} EstreeNode
 * @typedef {import('estree').Comment} EstreeComment
 *
 * @typedef State
 * @property {EstreeComment[]} comments
 * @property {number} index
 *
 * @typedef Fields
 * @property {boolean} leading
 * @property {boolean} trailing
 */

/**
 * Attach semistandard estree comment nodes to the tree.
 *
 * @param {EstreeNode} tree
 * @param {EstreeComment[]} [comments]
 */
export function attachComments(tree, comments) {
  var list = (comments || []).concat().sort(compare)
  if (list.length) walk(tree, {comments: list, index: 0})
  return tree
}

/**
 * Attach semistandard estree comment nodes to the tree.
 *
 * @param {EstreeNode} node
 * @param {State} state
 */
function walk(node, state) {
  /** @type {EstreeNode[]} */
  var children = []
  /** @type {EstreeComment[]} */
  var comments = []
  /** @type {string} */
  var key
  /** @type {EstreeNode|EstreeNode[]} */
  var value
  /** @type {number} */
  var index

  // Done, we can quit.
  if (state.index === state.comments.length) {
    return
  }

  // Find all children of `node`
  for (key in node) {
    value = node[key]

    // Ignore comments.
    if (value && typeof value === 'object' && key !== 'comments') {
      if (Array.isArray(value)) {
        index = -1

        while (++index < value.length) {
          if (value[index] && typeof value[index].type === 'string') {
            children.push(value[index])
          }
        }
      } else if (typeof value.type === 'string') {
        children.push(value)
      }
    }
  }

  // Sort the children.
  children.sort(compare)

  // Initial comments.
  push.apply(
    comments,
    slice(state, node, false, {leading: true, trailing: false})
  )

  index = -1

  while (++index < children.length) {
    walk(children[index], state)
  }

  // Dangling or trailing comments.
  push.apply(
    comments,
    slice(state, node, true, {
      leading: false,
      trailing: Boolean(children.length)
    })
  )

  if (comments.length) {
    // @ts-ignore, yes, because they’re nonstandard.
    node.comments = comments
  }
}

/**
 * @param {State} state
 * @param {EstreeNode} node
 * @param {boolean} compareEnd
 * @param {Fields} fields
 */
function slice(state, node, compareEnd, fields) {
  /** @type {EstreeComment[]} */
  var result = []

  while (
    state.comments[state.index] &&
    compare(state.comments[state.index], node, compareEnd) < 1
  ) {
    result.push(Object.assign({}, state.comments[state.index++], fields))
  }

  return result
}

/**
 * @param {EstreeNode|EstreeComment} left
 * @param {EstreeNode|EstreeComment} right
 * @param {boolean} [compareEnd]
 * @returns {number}
 */
function compare(left, right, compareEnd) {
  var field = compareEnd ? 'end' : 'start'

  // Offsets.
  if (left.range && right.range) {
    return left.range[0] - right.range[compareEnd ? 1 : 0]
  }

  // Points.
  if (left.loc && left.loc.start && right.loc && right.loc[field]) {
    return (
      left.loc.start.line - right.loc[field].line ||
      left.loc.start.column - right.loc[field].column
    )
  }

  // Just `start` (and `end`) on nodes.
  // Default in most parsers.
  if ('start' in left && field in right) {
    // @ts-ignore Added by Acorn
    return left.start - right[field]
  }

  return NaN
}
