import Formatter, { IFormatterOptions } from './';
export default class RerunFormatter extends Formatter {
    private readonly separator;
    constructor(options: IFormatterOptions);
    logFailedTestCases(): void;
}
