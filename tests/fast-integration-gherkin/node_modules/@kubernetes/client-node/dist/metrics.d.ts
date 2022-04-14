import { KubeConfig } from './config';
export interface Usage {
    cpu: string;
    memory: string;
}
export interface ContainerMetric {
    name: string;
    usage: Usage;
}
export interface PodMetric {
    metadata: {
        name: string;
        namespace: string;
        selfLink: string;
        creationTimestamp: string;
    };
    timestamp: string;
    window: string;
    containers: ContainerMetric[];
}
export interface NodeMetric {
    metadata: {
        name: string;
        selfLink: string;
        creationTimestamp: string;
    };
    timestamp: string;
    window: string;
    usage: Usage;
}
export interface PodMetricsList {
    kind: 'PodMetricsList';
    apiVersion: 'metrics.k8s.io/v1beta1';
    metadata: {
        selfLink: string;
    };
    items: PodMetric[];
}
export interface NodeMetricsList {
    kind: 'NodeMetricsList';
    apiVersion: 'metrics.k8s.io/v1beta1';
    metadata: {
        selfLink: string;
    };
    items: NodeMetric[];
}
export declare class Metrics {
    private config;
    constructor(config: KubeConfig);
    getNodeMetrics(): Promise<NodeMetricsList>;
    getPodMetrics(namespace?: string): Promise<PodMetricsList>;
    private metricsApiRequest;
}
