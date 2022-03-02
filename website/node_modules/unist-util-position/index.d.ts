/**
 * Get the positional info of `node`.
 *
 * @param {NodeLike} [node]
 * @returns {Position}
 */
export function position(node?: NodeLike): Position
/**
 * Get the positional info of `node`.
 *
 * @param {NodeLike} [node]
 * @returns {Point}
 */
export function pointStart(node?: NodeLike): Point
/**
 * Get the positional info of `node`.
 *
 * @param {NodeLike} [node]
 * @returns {Point}
 */
export function pointEnd(node?: NodeLike): Point
export type Position = import('unist').Position
export type Point = import('unist').Point
export type PointLike = Partial<Point>
export type PositionLike = {
  start?: PointLike
  end?: PointLike
}
export type NodeLike = {
  position?: PositionLike
}
