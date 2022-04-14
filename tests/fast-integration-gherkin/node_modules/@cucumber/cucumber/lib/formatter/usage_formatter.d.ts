import Formatter, { IFormatterOptions } from './';
export default class UsageFormatter extends Formatter {
    constructor(options: IFormatterOptions);
    logUsage(): void;
}
