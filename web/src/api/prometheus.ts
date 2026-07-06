import api from './axios';

export interface QueryRangeParams {
  query: string;
  start: string;
  end: string;
  step: string;
}

export const queryPrometheusRange = (clusterName: string, params: QueryRangeParams) =>
  api.get(`/clusters/${clusterName}/prometheus/query_range`, { params, timeout: 60000 });

export const queryPrometheusRangePost = (clusterName: string, params: QueryRangeParams) =>
  api.post(`/clusters/${clusterName}/prometheus/query_range`, params, { timeout: 60000 });
