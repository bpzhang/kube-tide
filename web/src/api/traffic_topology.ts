import api from './axios';

export interface TopologyNode {
  id: string;
  type: string;
  name: string;
  namespace: string;
  extra?: Record<string, unknown>;
}

export interface TopologyEdge {
  source: string;
  target: string;
  edgeType: string;
  port?: string;
  inferred?: boolean;
  evidence?: string;
}

export interface TrafficPath {
  namespace: string;
  ingressName?: string;
  ingressHost?: string;
  path?: string;
  serviceName: string;
  workloadType?: string;
  workloadName?: string;
  podCount: number;
}

export interface ClusterNetworkInfo {
  cni: string;
  terwayMode?: string;
  hubbleEnabled?: boolean;
  hubbleRelayReady?: boolean;
  hubbleMetricsSvc?: boolean;
  metricsSource?: string;
  message?: string;
}

export interface HubbleDropStat {
  reason: string;
  count: number;
}

export interface HubblePortStat {
  protocol: string;
  port: string;
  count: number;
}

export interface HubbleMetricsSummary {
  available: boolean;
  drops?: HubbleDropStat[];
  topPorts?: HubblePortStat[];
  message?: string;
}

export interface TrafficTopology {
  nodes: TopologyNode[];
  edges: TopologyEdge[];
  paths: TrafficPath[];
  network?: ClusterNetworkInfo;
  hubble?: HubbleMetricsSummary;
}

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export const getTrafficTopology = (clusterName: string, namespace?: string) => {
  const path = namespace
    ? `/clusters/${clusterName}/namespaces/${namespace}/traffic-topology`
    : `/clusters/${clusterName}/traffic-topology`;
  return api.get<ApiResponse<TrafficTopology>>(path);
};

export const parseNodeLabel = (id: string) => {
  const parts = id.split('/');
  if (parts.length >= 3) {
    return `${parts[0]}/${parts[2]}`;
  }
  return id;
};
