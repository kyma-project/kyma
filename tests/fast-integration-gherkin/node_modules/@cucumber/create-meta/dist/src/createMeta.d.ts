import * as messages from '@cucumber/messages';
export declare type CiDict = Record<string, CiSystem>;
export declare type Env = Record<string, string | undefined>;
export interface CiSystem {
    url: string;
    git: {
        remote: string | undefined;
        branch: string | undefined;
        revision: string | undefined;
        tag: string | undefined;
    };
}
export default function createMeta(toolName: string, toolVersion: string, envDict: Env, ciDict?: CiDict): messages.Meta;
export declare function detectCI(ciDict: CiDict, envDict: Env): messages.Ci | undefined;
export declare function removeUserInfoFromUrl(value: string): string;
//# sourceMappingURL=createMeta.d.ts.map