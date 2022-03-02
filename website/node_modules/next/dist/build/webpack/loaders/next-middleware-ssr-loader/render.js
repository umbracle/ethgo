"use strict";
Object.defineProperty(exports, "__esModule", {
    value: true
});
exports.getRender = getRender;
var _utils = require("../../../../server/web/utils");
var _webServer = _interopRequireDefault(require("../../../../server/web-server"));
var _baseHttp = require("../../../../server/base-http");
function _interopRequireDefault(obj) {
    return obj && obj.__esModule ? obj : {
        default: obj
    };
}
const createHeaders = (args)=>({
        ...args,
        'x-middleware-ssr': '1',
        'Cache-Control': 'no-cache, no-store, max-age=0, must-revalidate'
    })
;
function sendError(req, error) {
    const defaultMessage = 'An error occurred while rendering ' + req.url + '.';
    return new Response(error && error.message || defaultMessage, {
        status: 500,
        headers: createHeaders()
    });
}
function getRender({ Document , isServerComponent , config  }) {
    // Polyfilled for `path-browserify`.
    process.cwd = ()=>''
    ;
    const server = new _webServer.default({
        conf: config,
        minimalMode: true
    });
    const requestHandler = server.getRequestHandler();
    return async function render(request) {
        const { nextUrl: url , cookies , headers  } = request;
        const { pathname , searchParams  } = url;
        const query = Object.fromEntries(searchParams);
        const req = {
            url: pathname,
            cookies,
            headers: (0, _utils).toNodeHeaders(headers)
        };
        // Preflight request
        if (request.method === 'HEAD') {
            return new Response(null, {
                headers: createHeaders()
            });
        }
        // @TODO: We should move this into server/render.
        if (Document.getInitialProps) {
            const err = new Error('`getInitialProps` in Document component is not supported with `concurrentFeatures` enabled.');
            return sendError(req, err);
        }
        const renderServerComponentData = isServerComponent ? query.__flight__ !== undefined : false;
        const serverComponentProps = isServerComponent && query.__props__ ? JSON.parse(query.__props__) : undefined;
        // Extend the context.
        Object.assign(self.__server_context, {
            renderServerComponentData,
            serverComponentProps
        });
        const extendedReq = new _baseHttp.WebNextRequest(request);
        const extendedRes = new _baseHttp.WebNextResponse();
        requestHandler(extendedReq, extendedRes);
        return await extendedRes.toResponse();
    };
}

//# sourceMappingURL=render.js.map