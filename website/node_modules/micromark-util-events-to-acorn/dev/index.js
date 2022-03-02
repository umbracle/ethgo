/**
 * @typedef {import('micromark-util-types').Event} Event
 * @typedef {import('micromark-util-types').Point} Point
 * @typedef {import('acorn').Options} AcornOptions
 * @typedef {import('acorn').Comment} Comment
 * @typedef {import('acorn').Node} Node
 * @typedef {import('estree').Program} Program
 */

/**
 * @typedef {{parse: import('acorn').parse, parseExpressionAt: import('acorn').parseExpressionAt}} Acorn
 *
 * @typedef {Error & {raisedAt: number, pos: number, loc: {line: number, column: number}}} AcornError
 */

/**
 * @typedef Options
 * @property {Acorn} acorn
 * @property {AcornOptions} [acornOptions]
 * @property {Point} [start]
 * @property {string} [prefix='']
 * @property {string} [suffix='']
 * @property {boolean} [expression=false]
 * @property {boolean} [allowEmpty=false]
 */

import {ok as assert} from 'uvu/assert'
import {visit} from 'estree-util-visit'
import {VFileMessage} from 'vfile-message'

const own = {}.hasOwnProperty

/**
 * Parse a list of micromark events with acorn.
 *
 * @param {Event[]} events
 * @param {Options} options
 * @returns {{estree: Program|undefined, error: Error|undefined, swallow: boolean}}
 */
export function eventsToAcorn(events, options) {
  const {prefix = '', suffix = ''} = options
  /** @type {Array.<Comment>} */
  const comments = []
  const acornConfig = Object.assign({}, options.acornOptions, {
    onComment: comments,
    preserveParens: true
  })
  /** @type {Array.<string>} */
  const chunks = []
  /** @type {Record<string, Point>} */
  const lines = {}
  let index = -1
  let swallow = false
  /** @type {Node|undefined} */
  let estree
  /** @type {Error|undefined} */
  let exception
  /** @type {number|undefined} */
  let mdStartOffset

  if (options.start) {
    mdStartOffset = options.start.offset
    lines[options.start.line] = options.start
  }

  // Assume only void events (and `enter` followed immediately by an `exit`).
  while (++index < events.length) {
    const token = events[index][1]

    if (events[index][0] === 'exit') {
      chunks.push(events[index][2].sliceSerialize(token))

      // Not passed by `micromark-extension-mdxjs-esm`
      /* c8 ignore next 3 */
      if (mdStartOffset === undefined) {
        mdStartOffset = events[index][1].start.offset
      }

      if (
        !(token.start.line in lines) ||
        lines[token.start.line].offset > token.start.offset
      ) {
        lines[token.start.line] = token.start
      }
    }
  }

  const source = chunks.join('')
  const value = prefix + source + suffix
  const isEmptyExpression = options.expression && empty(source)

  if (isEmptyExpression && !options.allowEmpty) {
    throw new VFileMessage(
      'Unexpected empty expression',
      parseOffsetToUnistPoint(0),
      'micromark-extension-mdx-expression:unexpected-empty-expression'
    )
  }

  try {
    estree =
      options.expression && !isEmptyExpression
        ? options.acorn.parseExpressionAt(value, 0, acornConfig)
        : options.acorn.parse(value, acornConfig)
  } catch (error_) {
    const error = /** @type {AcornError} */ (error_)
    const point = parseOffsetToUnistPoint(error.pos)
    error.message = String(error.message).replace(/ \(\d+:\d+\)$/, '')
    error.pos = point.offset
    error.loc = {line: point.line, column: point.column - 1}
    exception = error
    swallow =
      error.raisedAt >= prefix.length + source.length ||
      // Broken comments are raised at their start, not their end.
      error.message === 'Unterminated comment'
  }

  if (estree && options.expression && !isEmptyExpression) {
    if (empty(value.slice(estree.end, value.length - suffix.length))) {
      estree = {
        type: 'Program',
        start: 0,
        end: prefix.length + source.length,
        // @ts-expect-error: It’s good.
        body: [
          {
            type: 'ExpressionStatement',
            expression: estree,
            start: 0,
            end: prefix.length + source.length
          }
        ],
        sourceType: 'module',
        comments: []
      }
    } else {
      const point = parseOffsetToUnistPoint(estree.end)
      exception = new Error('Unexpected content after expression')
      // @ts-expect-error: acorn exception.
      exception.pos = point.offset
      // @ts-expect-error: acorn exception.
      exception.loc = {line: point.line, column: point.column - 1}
      estree = undefined
    }
  }

  if (estree) {
    // @ts-expect-error: acorn *does* allow comments
    estree.comments = comments

    visit(estree, (esnode, field, index, parents) => {
      let context = /** @type {Node|Node[]} */ (parents[parents.length - 1])
      /** @type {string|number|null} */
      let prop = field

      // Remove non-standard `ParenthesizedExpression`.
      if (esnode.type === 'ParenthesizedExpression' && context && prop) {
        /* c8 ignore next 5 */
        if (typeof index === 'number') {
          // @ts-expect-error: indexable.
          context = context[prop]
          prop = index
        }

        // @ts-expect-error: indexable.
        context[prop] = esnode.expression
      }

      assert('start' in esnode, 'expected `start` in node from acorn')
      assert('end' in esnode, 'expected `end` in node from acorn')
      // @ts-expect-error: acorn has positions.
      const pointStart = parseOffsetToUnistPoint(esnode.start)
      // @ts-expect-error: acorn has positions.
      const pointEnd = parseOffsetToUnistPoint(esnode.end)
      // @ts-expect-error: acorn has positions.
      esnode.start = pointStart.offset
      // @ts-expect-error: acorn has positions.
      esnode.end = pointEnd.offset
      // @ts-expect-error: acorn has positions.
      esnode.loc = {
        start: {line: pointStart.line, column: pointStart.column - 1},
        end: {line: pointEnd.line, column: pointEnd.column - 1}
      }
      // @ts-expect-error: acorn has positions.
      esnode.range = [esnode.start, esnode.end]
    })
  }

  // @ts-expect-error: It’s a program now.
  return {estree, error: exception, swallow}

  /**
   * @param {number} offset
   * @returs {Point}
   */
  function parseOffsetToUnistPoint(offset) {
    let srcOffset = offset - prefix.length
    /** @type {string} */
    let line
    /** @type {Point|undefined} */
    let lineStart

    if (srcOffset < 0) {
      srcOffset = 0
    } else if (srcOffset > source.length) {
      srcOffset = source.length
    }

    assert(mdStartOffset !== undefined, 'expected `mdStartOffset` to be found')
    srcOffset += mdStartOffset

    // Then, update it.
    for (line in lines) {
      if (own.call(lines, line)) {
        // First line we find.
        if (!lineStart) {
          lineStart = lines[line]
        }

        if (lines[line].offset > offset) {
          break
        }

        lineStart = lines[line]
      }
    }

    assert(lineStart, 'expected `lineStart` to be defined')
    return {
      line: lineStart.line,
      column: lineStart.column + (srcOffset - lineStart.offset),
      offset: srcOffset
    }
  }
}

/**
 * @param {string} value
 * @returns {boolean}
 */
function empty(value) {
  return /^\s*$/.test(
    value
      // Multiline comments.
      .replace(/\/\*[\s\S]*?\*\//g, '')
      // Line comments.
      // EOF instead of EOL is specifically not allowed, because that would
      // mean the closing brace is on the commented-out line
      .replace(/\/\/[^\r\n]*(\r\n|\n|\r)/g, '')
  )
}
