import type { NodeHeaders } from './types';
export declare function streamToIterator<T>(readable: ReadableStream<T>): AsyncIterableIterator<T>;
export declare function readableStreamTee<T = any>(readable: ReadableStream<T>): [ReadableStream<T>, ReadableStream<T>];
export declare function notImplemented(name: string, method: string): any;
export declare function fromNodeHeaders(object: NodeHeaders): Headers;
export declare function toNodeHeaders(headers?: Headers): NodeHeaders;
export declare function splitCookiesString(cookiesString: string): string[];
/**
 * We will be soon deprecating the usage of relative URLs in Middleware introducing
 * URL validation. This helper puts the future code in place and prints a warning
 * for cases where it will break. Meanwhile we preserve the previous behavior.
 */
export declare function validateURL(url: string | URL): string;
