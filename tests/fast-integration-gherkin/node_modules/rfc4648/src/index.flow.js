// @flow

type Encoding = {
  bits: number,
  chars: string,
  codes?: { [char: string]: number }
}

type ParseOptions = {
  loose?: boolean,
  out?: any
}

type StringifyOptions = {
  pad?: boolean
}

type ArrayLike<T> = {
  +length: number,
  +[n: number]: T
}

function parse(string: string, opts?: ParseOptions): Uint8Array {
  return new Uint8Array(0)
}

function stringify(data: ArrayLike<number>, opts?: StringifyOptions): string {
  return ''
}

export const base16 = { parse, stringify }
export const base32 = { parse, stringify }
export const base32hex = { parse, stringify }
export const base64 = { parse, stringify }
export const base64url = { parse, stringify }

export const codec = {
  parse(string: string, encoding: Encoding, opts?: ParseOptions): Uint8Array {
    return new Uint8Array(0)
  },

  stringify(
    data: ArrayLike<number>,
    encoding: Encoding,
    opts?: StringifyOptions
  ): string {
    return ''
  }
}
