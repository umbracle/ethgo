"use strict";
Object.defineProperty(exports, "__esModule", {
    value: true
});
exports.getRequestMeta = getRequestMeta;
exports.setRequestMeta = setRequestMeta;
exports.addRequestMeta = addRequestMeta;
exports.NEXT_REQUEST_META = void 0;
const NEXT_REQUEST_META = Symbol('NextRequestMeta');
exports.NEXT_REQUEST_META = NEXT_REQUEST_META;
function getRequestMeta(req, key) {
    const meta = req[NEXT_REQUEST_META] || {
    };
    return typeof key === 'string' ? meta[key] : meta;
}
function setRequestMeta(req, meta) {
    req[NEXT_REQUEST_META] = meta;
    return getRequestMeta(req);
}
function addRequestMeta(request, key, value) {
    const meta = getRequestMeta(request);
    meta[key] = value;
    return setRequestMeta(request, meta);
}

//# sourceMappingURL=request-meta.js.map