"use strict";
Object.defineProperty(exports, "__esModule", {
    value: true
});
var _constants = require("../shared/lib/constants");
var _apiUtils = require("./api-utils");
var _requestMeta = require("./request-meta");
class BaseNextRequest {
    constructor(method, url, body){
        this.method = method;
        this.url = url;
        this.body = body;
    }
    // Utils implemented using the abstract methods above
    get cookies() {
        if (this._cookies) return this._cookies;
        return this._cookies = (0, _apiUtils).getCookieParser(this.headers)();
    }
}
exports.BaseNextRequest = BaseNextRequest;
class NodeNextRequest extends BaseNextRequest {
    get originalRequest() {
        // Need to mimic these changes to the original req object for places where we use it:
        // render.tsx, api/ssg requests
        this._req[_requestMeta.NEXT_REQUEST_META] = this[_requestMeta.NEXT_REQUEST_META];
        this._req.url = this.url;
        this._req.cookies = this.cookies;
        return this._req;
    }
    constructor(_req){
        super(_req.method.toUpperCase(), _req.url, _req);
        this._req = _req;
        this.headers = this._req.headers;
    }
    async parseBody(limit) {
        return (0, _apiUtils).parseBody(this._req, limit);
    }
}
exports.NodeNextRequest = NodeNextRequest;
class WebNextRequest extends BaseNextRequest {
    constructor(request){
        const url = new URL(request.url);
        super(request.method, url.href.slice(url.origin.length), request.clone().body);
        this.request = request;
        this.headers = {
        };
        for (const [name, value] of request.headers.entries()){
            this.headers[name] = value;
        }
    }
    async parseBody(_limit) {
        throw new Error('parseBody is not implemented in the web runtime');
    }
}
exports.WebNextRequest = WebNextRequest;
class BaseNextResponse {
    constructor(destination){
        this.destination = destination;
    }
    // Utils implemented using the abstract methods above
    redirect(destination, statusCode) {
        this.setHeader('Location', destination);
        this.statusCode = statusCode;
        // Since IE11 doesn't support the 308 header add backwards
        // compatibility using refresh header
        if (statusCode === _constants.PERMANENT_REDIRECT_STATUS) {
            this.setHeader('Refresh', `0;url=${destination}`);
        }
        return this;
    }
}
exports.BaseNextResponse = BaseNextResponse;
class NodeNextResponse extends BaseNextResponse {
    get originalResponse() {
        if (_apiUtils.SYMBOL_CLEARED_COOKIES in this) {
            this._res[_apiUtils.SYMBOL_CLEARED_COOKIES] = this[_apiUtils.SYMBOL_CLEARED_COOKIES];
        }
        return this._res;
    }
    constructor(_res){
        super(_res);
        this._res = _res;
        this.textBody = undefined;
    }
    get sent() {
        return this._res.finished || this._res.headersSent;
    }
    get statusCode() {
        return this._res.statusCode;
    }
    set statusCode(value) {
        this._res.statusCode = value;
    }
    get statusMessage() {
        return this._res.statusMessage;
    }
    set statusMessage(value) {
        this._res.statusMessage = value;
    }
    setHeader(name, value) {
        this._res.setHeader(name, value);
        return this;
    }
    getHeaderValues(name) {
        const values = this._res.getHeader(name);
        if (values === undefined) return undefined;
        return (Array.isArray(values) ? values : [
            values
        ]).map((value)=>value.toString()
        );
    }
    hasHeader(name) {
        return this._res.hasHeader(name);
    }
    getHeader(name) {
        const values = this.getHeaderValues(name);
        return Array.isArray(values) ? values.join(',') : undefined;
    }
    appendHeader(name, value) {
        var ref;
        const currentValues = (ref = this.getHeaderValues(name)) !== null && ref !== void 0 ? ref : [];
        if (!currentValues.includes(value)) {
            this._res.setHeader(name, [
                ...currentValues,
                value
            ]);
        }
        return this;
    }
    body(value) {
        this.textBody = value;
        return this;
    }
    send() {
        this._res.end(this.textBody);
    }
}
exports.NodeNextResponse = NodeNextResponse;
class WebNextResponse extends BaseNextResponse {
    get sent() {
        return this._sent;
    }
    constructor(transformStream = new TransformStream()){
        super(transformStream.writable);
        this.transformStream = transformStream;
        this.headers = new Headers();
        this.textBody = undefined;
        this._sent = false;
        this.sendPromise = new Promise((resolve)=>{
            this.sendResolve = resolve;
        });
        this.response = this.sendPromise.then(()=>{
            var _textBody;
            return new Response((_textBody = this.textBody) !== null && _textBody !== void 0 ? _textBody : this.transformStream.readable, {
                headers: this.headers,
                status: this.statusCode,
                statusText: this.statusMessage
            });
        });
    }
    setHeader(name, value) {
        this.headers.delete(name);
        for (const val of Array.isArray(value) ? value : [
            value
        ]){
            this.headers.append(name, val);
        }
        return this;
    }
    getHeaderValues(name) {
        var ref;
        // https://developer.mozilla.org/en-US/docs/Web/API/Headers/get#example
        return (ref = this.getHeader(name)) === null || ref === void 0 ? void 0 : ref.split(',').map((v)=>v.trimStart()
        );
    }
    getHeader(name) {
        var ref;
        return (ref = this.headers.get(name)) !== null && ref !== void 0 ? ref : undefined;
    }
    hasHeader(name) {
        return this.headers.has(name);
    }
    appendHeader(name, value) {
        this.headers.append(name, value);
        return this;
    }
    body(value) {
        this.textBody = value;
        return this;
    }
    send() {
        var _obj, ref;
        (ref = (_obj = this).sendResolve) === null || ref === void 0 ? void 0 : ref.call(_obj);
        this._sent = true;
    }
    toResponse() {
        return this.response;
    }
}
exports.WebNextResponse = WebNextResponse;

//# sourceMappingURL=base-http.js.map