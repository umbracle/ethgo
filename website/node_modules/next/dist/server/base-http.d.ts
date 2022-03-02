/// <reference types="node" />
import type { ServerResponse, IncomingMessage, IncomingHttpHeaders } from 'http';
import type { Writable, Readable } from 'stream';
import { NextApiRequestCookies, SYMBOL_CLEARED_COOKIES } from './api-utils';
import { I18NConfig } from './config-shared';
import { NEXT_REQUEST_META, RequestMeta } from './request-meta';
export interface BaseNextRequestConfig {
    basePath: string | undefined;
    i18n?: I18NConfig;
    trailingSlash?: boolean | undefined;
}
export declare abstract class BaseNextRequest<Body = any> {
    method: string;
    url: string;
    body: Body;
    protected _cookies: NextApiRequestCookies | undefined;
    abstract headers: IncomingHttpHeaders;
    constructor(method: string, url: string, body: Body);
    abstract parseBody(limit: string | number): Promise<any>;
    get cookies(): NextApiRequestCookies;
}
export declare class NodeNextRequest extends BaseNextRequest<Readable> {
    private _req;
    headers: IncomingHttpHeaders;
    [NEXT_REQUEST_META]: RequestMeta;
    get originalRequest(): IncomingMessage & {
        cookies?: NextApiRequestCookies | undefined;
        [NEXT_REQUEST_META]?: RequestMeta | undefined;
    };
    constructor(_req: IncomingMessage & {
        [NEXT_REQUEST_META]?: RequestMeta;
        cookies?: NextApiRequestCookies;
    });
    parseBody(limit: string | number): Promise<any>;
}
export declare class WebNextRequest extends BaseNextRequest<ReadableStream | null> {
    request: Request;
    headers: IncomingHttpHeaders;
    constructor(request: Request);
    parseBody(_limit: string | number): Promise<any>;
}
export declare abstract class BaseNextResponse<Destination = any> {
    destination: Destination;
    abstract statusCode: number | undefined;
    abstract statusMessage: string | undefined;
    abstract get sent(): boolean;
    constructor(destination: Destination);
    /**
     * Sets a value for the header overwriting existing values
     */
    abstract setHeader(name: string, value: string | string[]): this;
    /**
     * Appends value for the given header name
     */
    abstract appendHeader(name: string, value: string): this;
    /**
     * Get all vaues for a header as an array or undefined if no value is present
     */
    abstract getHeaderValues(name: string): string[] | undefined;
    abstract hasHeader(name: string): boolean;
    /**
     * Get vaues for a header concatenated using `,` or undefined if no value is present
     */
    abstract getHeader(name: string): string | undefined;
    abstract body(value: string): this;
    abstract send(): void;
    redirect(destination: string, statusCode: number): this;
}
export declare class NodeNextResponse extends BaseNextResponse<Writable> {
    private _res;
    private textBody;
    [SYMBOL_CLEARED_COOKIES]?: boolean;
    get originalResponse(): ServerResponse & {
        [SYMBOL_CLEARED_COOKIES]?: boolean | undefined;
    };
    constructor(_res: ServerResponse & {
        [SYMBOL_CLEARED_COOKIES]?: boolean;
    });
    get sent(): boolean;
    get statusCode(): number;
    set statusCode(value: number);
    get statusMessage(): string;
    set statusMessage(value: string);
    setHeader(name: string, value: string | string[]): this;
    getHeaderValues(name: string): string[] | undefined;
    hasHeader(name: string): boolean;
    getHeader(name: string): string | undefined;
    appendHeader(name: string, value: string): this;
    body(value: string): this;
    send(): void;
}
export declare class WebNextResponse extends BaseNextResponse<WritableStream> {
    transformStream: TransformStream<any, any>;
    private headers;
    private textBody;
    private _sent;
    private sendPromise;
    private sendResolve?;
    private response;
    statusCode: number | undefined;
    statusMessage: string | undefined;
    get sent(): boolean;
    constructor(transformStream?: TransformStream<any, any>);
    setHeader(name: string, value: string | string[]): this;
    getHeaderValues(name: string): string[] | undefined;
    getHeader(name: string): string | undefined;
    hasHeader(name: string): boolean;
    appendHeader(name: string, value: string): this;
    body(value: string): this;
    send(): void;
    toResponse(): Promise<Response>;
}
