import Formatter, { IFormatterOptions } from './';
export default class UsageJsonFormatter extends Formatter {
    constructor(options: IFormatterOptions);
    logUsage(): void;
    replacer(key: string, value: any): any;
}
