/* eslint-disable @typescript-eslint/strict-boolean-expressions */

import {
  Encoding,
  parse,
  ParseOptions,
  stringify,
  StringifyOptions
} from './codec'

const base16Encoding: Encoding = {
  chars: '0123456789ABCDEF',
  bits: 4
}

const base32Encoding: Encoding = {
  chars: 'ABCDEFGHIJKLMNOPQRSTUVWXYZ234567',
  bits: 5
}

const base32HexEncoding: Encoding = {
  chars: '0123456789ABCDEFGHIJKLMNOPQRSTUV',
  bits: 5
}

const base64Encoding: Encoding = {
  chars: 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/',
  bits: 6
}

const base64UrlEncoding: Encoding = {
  chars: 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_',
  bits: 6
}

export const base16 = {
  parse(string: string, opts?: ParseOptions): Uint8Array {
    return parse(string.toUpperCase(), base16Encoding, opts)
  },

  stringify(data: ArrayLike<number>, opts?: StringifyOptions): string {
    return stringify(data, base16Encoding, opts)
  }
}

export const base32 = {
  parse(string: string, opts: ParseOptions = {}): Uint8Array {
    return parse(
      opts.loose
        ? string
            .toUpperCase()
            .replace(/0/g, 'O')
            .replace(/1/g, 'L')
            .replace(/8/g, 'B')
        : string,
      base32Encoding,
      opts
    )
  },

  stringify(data: ArrayLike<number>, opts?: StringifyOptions): string {
    return stringify(data, base32Encoding, opts)
  }
}

export const base32hex = {
  parse(string: string, opts?: ParseOptions): Uint8Array {
    return parse(string, base32HexEncoding, opts)
  },

  stringify(data: ArrayLike<number>, opts?: StringifyOptions): string {
    return stringify(data, base32HexEncoding, opts)
  }
}

export const base64 = {
  parse(string: string, opts?: ParseOptions): Uint8Array {
    return parse(string, base64Encoding, opts)
  },

  stringify(data: ArrayLike<number>, opts?: StringifyOptions): string {
    return stringify(data, base64Encoding, opts)
  }
}

export const base64url = {
  parse(string: string, opts?: ParseOptions): Uint8Array {
    return parse(string, base64UrlEncoding, opts)
  },

  stringify(data: ArrayLike<number>, opts?: StringifyOptions): string {
    return stringify(data, base64UrlEncoding, opts)
  }
}

export const codec = { parse, stringify }
