import Group from './Group';
import { RegExpExecArray } from 'regexp-match-indices';
export default class GroupBuilder {
    source: string;
    capturing: boolean;
    private readonly groupBuilders;
    add(groupBuilder: GroupBuilder): void;
    build(match: RegExpExecArray, nextGroupIndex: () => number): Group;
    setNonCapturing(): void;
    get children(): GroupBuilder[];
    moveChildrenTo(groupBuilder: GroupBuilder): void;
}
//# sourceMappingURL=GroupBuilder.d.ts.map