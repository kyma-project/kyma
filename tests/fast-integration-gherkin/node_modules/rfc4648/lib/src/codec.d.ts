export interface Encoding {
    bits: number;
    chars: string;
    codes?: {
        [char: string]: number;
    };
}
export interface ParseOptions {
    loose?: boolean;
    out?: new (size: number) => {
        [index: number]: number;
    };
}
export interface StringifyOptions {
    pad?: boolean;
}
export declare function parse(string: string, encoding: Encoding, opts?: ParseOptions): Uint8Array;
export declare function stringify(data: ArrayLike<number>, encoding: Encoding, opts?: StringifyOptions): string;
