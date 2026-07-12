export interface PrometheusRangeResponse {
  status: string;
  data?: {
    resultType?: string;
    result?: Array<{
      metric?: Record<string, string>;
      values?: [number, string][];
      value?: [number, string];
    }>;
  };
  error?: string;
  errorType?: string;
}

export interface PrometheusChartSeries {
  name: string;
  data: { time: string; value: number }[];
}

export const parsePrometheusRange = (body: PrometheusRangeResponse): PrometheusChartSeries[] => {
  if (body.status !== 'success' || !body.data?.result?.length) {
    return [];
  }
  return body.data.result.map((series, idx) => {
    const labelParts = Object.entries(series.metric || {})
      .filter(([k]) => k !== '__name__')
      .map(([k, v]) => `${k}=${v}`);
    const name = labelParts.length > 0 ? labelParts.join(', ') : `series-${idx + 1}`;
    const values = series.values || [];
    return {
      name,
      data: values.map(([ts, val]) => ({
        time: new Date(ts * 1000).toLocaleTimeString(),
        value: parseFloat(val) || 0,
      })),
    };
  });
};

export const parsePrometheusInstantTable = (
  body: PrometheusRangeResponse,
): Array<{ metric: Record<string, string>; value: string }> => {
  if (body.status !== 'success' || !body.data?.result?.length) {
    return [];
  }
  return body.data.result.map((row) => ({
    metric: row.metric || {},
    value: row.value?.[1] || '0',
  }));
};

export const buildRangeParams = (query: string, hours: number, step = '60') => {
  const end = Math.floor(Date.now() / 1000);
  const start = end - hours * 3600;
  return { query, start: String(start), end: String(end), step };
};

export const PROMQL_PRESETS = [
  {
    key: 'cluster_cpu',
    query: 'sum(rate(container_cpu_usage_seconds_total{container!=""}[5m]))',
  },
  {
    key: 'cluster_memory',
    query: 'sum(container_memory_working_set_bytes{container!=""})',
  },
  {
    key: 'pod_restarts',
    query: 'topk(10, increase(kube_pod_container_status_restarts_total[1h]))',
  },
  {
    key: 'node_cpu',
    query: '100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)',
  },
  {
    key: 'apiserver_requests',
    query: 'sum(rate(apiserver_request_total[5m])) by (code)',
  },
] as const;

export const deploymentCPUQuery = (namespace: string, deploymentName: string) =>
  `sum(rate(container_cpu_usage_seconds_total{namespace="${namespace}",pod=~"${deploymentName}-.*"}[5m]))`;

export const deploymentMemoryQuery = (namespace: string, deploymentName: string) =>
  `sum(container_memory_working_set_bytes{namespace="${namespace}",pod=~"${deploymentName}-.*"})`;
