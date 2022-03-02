export declare type Provider = 'github' | 'gitlab' | 'bitbucket';
export interface Parsed {
    type: Provider;
    owner: string;
    name: string;
    branch: string;
    sha: string;
    subdir: string;
}
/**
 * A util to parse a Git URL. Supports GitHub, GitLab and Bitbucket.
 * @param url
 */
export default function parseGitUrl(url: string): Parsed | null;
