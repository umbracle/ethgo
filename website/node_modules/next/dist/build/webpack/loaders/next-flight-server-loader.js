"use strict";
Object.defineProperty(exports, "__esModule", {
    value: true
});
exports.isImageImport = isImageImport;
exports.default = transformSource;
var acorn = _interopRequireWildcard(require("next/dist/compiled/acorn"));
var _utils = require("../../utils");
function _interopRequireWildcard(obj) {
    if (obj && obj.__esModule) {
        return obj;
    } else {
        var newObj = {
        };
        if (obj != null) {
            for(var key in obj){
                if (Object.prototype.hasOwnProperty.call(obj, key)) {
                    var desc = Object.defineProperty && Object.getOwnPropertyDescriptor ? Object.getOwnPropertyDescriptor(obj, key) : {
                    };
                    if (desc.get || desc.set) {
                        Object.defineProperty(newObj, key, desc);
                    } else {
                        newObj[key] = obj[key];
                    }
                }
            }
        }
        newObj.default = obj;
        return newObj;
    }
}
function isClientComponent(importSource, pageExtensions) {
    return new RegExp(`\\.client(\\.(${pageExtensions.join('|')}))?`).test(importSource);
}
function isServerComponent(importSource, pageExtensions) {
    return new RegExp(`\\.server(\\.(${pageExtensions.join('|')}))?`).test(importSource);
}
function isNextComponent(importSource) {
    return importSource.includes('next/link') || importSource.includes('next/image');
}
function isImageImport(importSource) {
    // TODO: share extension with next/image
    // TODO: add other static assets, jpeg -> jpg
    return [
        'jpg',
        'jpeg',
        'png',
        'webp',
        'avif'
    ].some((imageExt)=>importSource.endsWith('.' + imageExt)
    );
}
async function parseImportsInfo(source, imports, isClientCompilation, pageExtensions) {
    const { body  } = acorn.parse(source, {
        ecmaVersion: 11,
        sourceType: 'module'
    });
    let transformedSource = '';
    let lastIndex = 0;
    let defaultExportName = 'RSCComponent';
    for(let i = 0; i < body.length; i++){
        const node = body[i];
        switch(node.type){
            case 'ImportDeclaration':
                {
                    const importSource = node.source.value;
                    if (!isClientCompilation) {
                        if (!(isClientComponent(importSource, pageExtensions) || isNextComponent(importSource) || isImageImport(importSource))) {
                            continue;
                        }
                        transformedSource += source.substring(lastIndex, node.source.start - 1);
                        transformedSource += JSON.stringify(`${node.source.value}?flight`);
                    } else {
                        // For the client compilation, we skip all modules imports but
                        // always keep client components in the bundle. All client components
                        // have to be imported from either server or client components.
                        if (!(isClientComponent(importSource, pageExtensions) || isServerComponent(importSource, pageExtensions) || // Special cases for Next.js APIs that are considered as client
                        // components:
                        isNextComponent(importSource) || isImageImport(importSource))) {
                            continue;
                        }
                    }
                    lastIndex = node.source.end;
                    imports.push(`require(${JSON.stringify(importSource)})`);
                    continue;
                }
            case 'ExportDefaultDeclaration':
                {
                    const def = node.declaration;
                    if (def.type === 'Identifier') {
                        defaultExportName = def.name;
                    } else if (def.type === 'FunctionDeclaration') {
                        defaultExportName = def.id.name;
                    }
                    break;
                }
            default:
                break;
        }
    }
    if (!isClientCompilation) {
        transformedSource += source.substring(lastIndex);
    }
    return {
        source: transformedSource,
        defaultExportName
    };
}
async function transformSource(source) {
    const { client: isClientCompilation , pageExtensions: pageExtensionsJson  } = this.getOptions();
    const { resourcePath  } = this;
    const pageExtensions = JSON.parse(pageExtensionsJson);
    if (typeof source !== 'string') {
        throw new Error('Expected source to have been transformed to a string.');
    }
    if (resourcePath.includes('/node_modules/')) {
        return source;
    }
    const imports = [];
    const { source: transformedSource , defaultExportName  } = await parseImportsInfo(source, imports, isClientCompilation, (0, _utils).getRawPageExtensions(pageExtensions));
    /**
   * Server side component module output:
   *
   * export default function ServerComponent() { ... }
   * + export const __rsc_noop__=()=>{ ... }
   * + ServerComponent.__next_rsc__=1;
   *
   * Client side component module output:
   *
   * The function body of ServerComponent will be removed
   */ const noop = `export const __rsc_noop__=()=>{${imports.join(';')}}`;
    const defaultExportNoop = isClientCompilation ? `export default function ${defaultExportName}(){}\n${defaultExportName}.__next_rsc__=1;` : `${defaultExportName}.__next_rsc__=1;`;
    const transformed = transformedSource + '\n' + noop + '\n' + defaultExportNoop;
    return transformed;
}

//# sourceMappingURL=next-flight-server-loader.js.map