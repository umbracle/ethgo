import type { NextConfig } from '../../../../server/config-shared';
import { NextRequest } from '../../../../server/web/spec-extension/request';
export declare function getRender({ Document, isServerComponent, config, }: {
    Document: any;
    isServerComponent: boolean;
    config: NextConfig;
}): (request: NextRequest) => Promise<Response>;
