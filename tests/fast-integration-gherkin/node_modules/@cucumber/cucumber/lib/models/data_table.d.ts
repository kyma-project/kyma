import * as messages from '@cucumber/messages';
export default class DataTable {
    private readonly rawTable;
    constructor(sourceTable: messages.PickleTable | string[][]);
    hashes(): any[];
    raw(): string[][];
    rows(): string[][];
    rowsHash(): Record<string, string>;
    transpose(): DataTable;
}
