"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const lodash_1 = __importDefault(require("lodash"));
class DataTable {
    constructor(sourceTable) {
        if (sourceTable instanceof Array) {
            this.rawTable = sourceTable;
        }
        else {
            this.rawTable = sourceTable.rows.map((row) => row.cells.map((cell) => cell.value));
        }
    }
    hashes() {
        const copy = this.raw();
        const keys = copy[0];
        const valuesArray = copy.slice(1);
        return valuesArray.map((values) => lodash_1.default.zipObject(keys, values));
    }
    raw() {
        return this.rawTable.slice(0);
    }
    rows() {
        const copy = this.raw();
        copy.shift();
        return copy;
    }
    rowsHash() {
        const rows = this.raw();
        const everyRowHasTwoColumns = lodash_1.default.every(rows, (row) => row.length === 2);
        if (!everyRowHasTwoColumns) {
            throw new Error('rowsHash can only be called on a data table where all rows have exactly two columns');
        }
        return lodash_1.default.fromPairs(rows);
    }
    transpose() {
        const transposed = this.rawTable[0].map((x, i) => this.rawTable.map((y) => y[i]));
        return new DataTable(transposed);
    }
}
exports.default = DataTable;
//# sourceMappingURL=data_table.js.map