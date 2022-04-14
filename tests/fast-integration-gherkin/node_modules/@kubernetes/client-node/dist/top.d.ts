import { CoreV1Api, V1Node, V1Pod } from './gen/api';
import { Metrics } from './metrics';
export declare class ResourceUsage {
    readonly Capacity: number | BigInt;
    readonly RequestTotal: number | BigInt;
    readonly LimitTotal: number | BigInt;
    constructor(Capacity: number | BigInt, RequestTotal: number | BigInt, LimitTotal: number | BigInt);
}
export declare class CurrentResourceUsage {
    readonly CurrentUsage: number | BigInt;
    readonly RequestTotal: number | BigInt;
    readonly LimitTotal: number | BigInt;
    constructor(CurrentUsage: number | BigInt, RequestTotal: number | BigInt, LimitTotal: number | BigInt);
}
export declare class NodeStatus {
    readonly Node: V1Node;
    readonly CPU: ResourceUsage;
    readonly Memory: ResourceUsage;
    constructor(Node: V1Node, CPU: ResourceUsage, Memory: ResourceUsage);
}
export declare class ContainerStatus {
    readonly Container: string;
    readonly CPUUsage: CurrentResourceUsage;
    readonly MemoryUsage: CurrentResourceUsage;
    constructor(Container: string, CPUUsage: CurrentResourceUsage, MemoryUsage: CurrentResourceUsage);
}
export declare class PodStatus {
    readonly Pod: V1Pod;
    readonly CPU: CurrentResourceUsage;
    readonly Memory: CurrentResourceUsage;
    readonly Containers: ContainerStatus[];
    constructor(Pod: V1Pod, CPU: CurrentResourceUsage, Memory: CurrentResourceUsage, Containers: ContainerStatus[]);
}
export declare function topNodes(api: CoreV1Api): Promise<NodeStatus[]>;
export declare function topPods(api: CoreV1Api, metrics: Metrics, namespace?: string): Promise<PodStatus[]>;
