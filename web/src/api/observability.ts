import api from './axios';

export interface PrometheusInfo {
  configured: boolean;
  url?: string;
  source: 'ack_managed_prometheus' | 'manual' | 'unconfigured' | string;
  healthy: boolean;
  message?: string;
}

export interface PodHealthSummary {
  total: number;
  running: number;
  pending: number;
  failed: number;
  crashLoop: number;
  notReady: number;
}

export interface ObservabilitySummary {
  prometheus: PrometheusInfo;
  health?: {
    overall: string;
    components: Array<{ name: string; status: string; message?: string }>;
  };
  pods: PodHealthSummary;
  events: { warningRecent: number };
}

export interface ObservabilityAlert {
  level: 'warning' | 'error' | 'info';
  category: string;
  namespace?: string;
  name: string;
  message: string;
}

export const getObservabilitySummary = (clusterName: string) =>
  api.get<{ code: number; message: string; data: { summary: ObservabilitySummary } }>(
    `/clusters/${clusterName}/observability/summary`,
  );

export const getObservabilityAlerts = (clusterName: string, limit = 50) =>
  api.get<{ code: number; message: string; data: { alerts: ObservabilityAlert[] } }>(
    `/clusters/${clusterName}/observability/alerts`,
    { params: { limit } },
  );

export const getPrometheusInfo = (clusterName: string) =>
  api.get<{ code: number; message: string; data: { prometheus: PrometheusInfo } }>(
    `/clusters/${clusterName}/prometheus/info`,
  );
