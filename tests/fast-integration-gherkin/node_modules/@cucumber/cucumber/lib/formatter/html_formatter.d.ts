import Formatter, { IFormatterOptions } from '.';
export default class HtmlFormatter extends Formatter {
    private readonly _finished;
    constructor(options: IFormatterOptions);
    finished(): Promise<void>;
}
