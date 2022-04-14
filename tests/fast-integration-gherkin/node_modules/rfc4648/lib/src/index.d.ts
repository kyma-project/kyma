import { parse, ParseOptions, stringify, StringifyOptions } from './codec';
export declare const base16: {
    parse(string: string, opts?: ParseOptions | undefined): Uint8Array;
    stringify(data: ArrayLike<number>, opts?: StringifyOptions | undefined): string;
};
export declare const base32: {
    parse(string: string, opts?: ParseOptions): Uint8Array;
    stringify(data: ArrayLike<number>, opts?: StringifyOptions | undefined): string;
};
export declare const base32hex: {
    parse(string: string, opts?: ParseOptions | undefined): Uint8Array;
    stringify(data: ArrayLike<number>, opts?: StringifyOptions | undefined): string;
};
export declare const base64: {
    parse(string: string, opts?: ParseOptions | undefined): Uint8Array;
    stringify(data: ArrayLike<number>, opts?: StringifyOptions | undefined): string;
};
export declare const base64url: {
    parse(string: string, opts?: ParseOptions | undefined): Uint8Array;
    stringify(data: ArrayLike<number>, opts?: StringifyOptions | undefined): string;
};
export declare const codec: {
    parse: typeof parse;
    stringify: typeof stringify;
};
