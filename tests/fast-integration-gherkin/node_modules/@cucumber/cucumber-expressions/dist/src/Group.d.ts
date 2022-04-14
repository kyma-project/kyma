export default class Group {
    readonly value: string | undefined;
    readonly start: number | undefined;
    readonly end: number | undefined;
    readonly children: readonly Group[];
    constructor(value: string | undefined, start: number | undefined, end: number | undefined, children: readonly Group[]);
    get values(): string[];
}
//# sourceMappingURL=Group.d.ts.map