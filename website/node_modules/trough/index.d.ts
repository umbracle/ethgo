/**
 * @typedef {(error?: Error|null|undefined, ...output: any[]) => void} Callback
 * @typedef {(...input: any[]) => any} Middleware
 *
 * @typedef {(...input: any[]) => void} Run Call all middleware.
 * @typedef {(fn: Middleware) => Pipeline} Use Add `fn` (middleware) to the list.
 * @typedef {{run: Run, use: Use}} Pipeline
 */
/**
 * Create new middleware.
 *
 * @returns {Pipeline}
 */
export function trough(): Pipeline
/**
 * Wrap `middleware`.
 * Can be sync or async; return a promise, receive a callback, or return new
 * values and errors.
 *
 * @param {Middleware} middleware
 * @param {Callback} callback
 */
export function wrap(
  middleware: Middleware,
  callback: Callback
): (...parameters: any[]) => void
export type Callback = (
  error?: Error | null | undefined,
  ...output: any[]
) => void
export type Middleware = (...input: any[]) => any
/**
 * Call all middleware.
 */
export type Run = (...input: any[]) => void
/**
 * Add `fn` (middleware) to the list.
 */
export type Use = (fn: Middleware) => Pipeline
export type Pipeline = {
  run: Run
  use: Use
}
