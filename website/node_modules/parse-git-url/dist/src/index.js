"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const url_1 = require("url");
function isChecksum(str) {
    return /^[a-f0-9]{40}$/i.test(str);
}
/**
 * A util to parse a Git URL. Supports GitHub, GitLab and Bitbucket.
 * @param url
 */
function parseGitUrl(url) {
    if (typeof url !== 'string' || !url.length) {
        return null;
    }
    let type;
    let slug;
    if (url.startsWith('git@')) {
        // handle SSH
        switch (true) {
            case url.startsWith('git@github.com:'):
                type = 'github';
                slug = url.replace('git@github.com:', '');
                break;
            case url.startsWith('git@gitlab.com:'):
                type = 'gitlab';
                slug = url.replace('git@gitlab.com:', '');
                break;
            case url.startsWith('git@bitbucket.org:'):
                type = 'bitbucket';
                slug = url.replace('git@bitbucket.org:', '');
                break;
            default:
                // failed to parse
                return null;
        }
    }
    else {
        // handle HTTPS
        const obj = url_1.parse(url);
        if (!obj.pathname) {
            return null;
        }
        switch (obj.hostname) {
            case 'github.com':
            case 'www.github.com':
                type = 'github';
                break;
            case 'gitlab.com':
            case 'www.gitlab.com':
                type = 'gitlab';
                break;
            case 'bitbucket.org':
            case 'www.bitbucket.org':
                type = 'bitbucket';
                break;
            default:
                // failed to parse
                return null;
        }
        // remove leading and trailing `/`s
        slug = obj.pathname.replace(/(^\/|\/$)/g, '');
    }
    // remove trailing `.git`
    slug = slug.replace(/\.git$/, '');
    const seg = slug.split('/').filter(Boolean);
    if (seg.length < 2) {
        return null;
    }
    if (seg.length === 2) {
        return {
            type,
            owner: seg[0],
            name: seg[1],
            branch: '',
            sha: '',
            subdir: ''
        };
    }
    let branch = '';
    let sha = '';
    let subdir = '';
    let owner = seg[0];
    let name = seg[1];
    if (type === 'github') {
        if (seg[2] === 'blob' || seg[2] === 'tree' || seg[2] === 'commit') {
            // https://github.com/zeit/front/tree/fix-bug
            if (isChecksum(seg[3])) {
                sha = seg[3];
            }
            else {
                branch = seg[3];
            }
            subdir = seg.slice(4).join('/');
        }
        else {
            // failed to parse
            return null;
        }
    }
    else if (type === 'gitlab') {
        if (seg[2] === '-') {
            if (seg[3] === 'blob' || seg[3] === 'tree' || seg[3] === 'commit') {
                // https://gitlab.com/zeit/front/-/tree/master
                if (isChecksum(seg[4])) {
                    sha = seg[4];
                }
                else {
                    branch = seg[4];
                }
                subdir = seg.slice(5).join('/');
            }
        }
        else {
            // gitlab subgroups
            // https://gitlab.com/shu-paco/subgroup-1/my-awesome-project/-/tree/master
            //                    seg[0]   seg[1]     seg[2]             3 4    5
            const idx = seg.indexOf('-');
            if (idx === -1) {
                name = seg.slice(1).join('/');
            }
            else {
                name = seg.slice(1, idx).join('/');
                if (seg[idx + 1] === 'blob' || seg[idx + 1] === 'tree' || seg[idx + 1] === 'commit') {
                    if (isChecksum(seg[idx + 2])) {
                        sha = seg[idx + 2];
                    }
                    else {
                        branch = seg[idx + 2];
                    }
                    subdir = seg.slice(idx + 3).join('/');
                }
            }
        }
    }
    else if (type === 'bitbucket') {
        if (seg[2] === 'src') {
            // https://bitbucket.org/shudin/test/src/bug-fix
            if (isChecksum(seg[3])) {
                sha = seg[3];
            }
            else {
                branch = seg[3];
            }
            subdir = seg.slice(4).join('/');
        }
        else {
            // failed to parse
            return null;
        }
    }
    return {
        type,
        owner,
        name,
        branch,
        sha,
        subdir
    };
}
exports.default = parseGitUrl;
