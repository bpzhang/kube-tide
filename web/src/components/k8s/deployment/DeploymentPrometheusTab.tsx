import React, { useCallback, useEffect, useState } from 'react';
import { Alert, Card, Spin } from 'antd';
import {
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { useTranslation } from 'react-i18next';
import { queryPrometheusRange } from '@/api/prometheus';

interface DeploymentPrometheusTabProps {
  clusterName: string;
  namespace: string;
  deploymentName: string;
}

interface PrometheusRangeResponse {
  status: string;
  data?: {
    result?: Array<{ values?: [number, string][] }>;
  };
  error?: string;
  errorType?: string;
}

const DeploymentPrometheusTab: React.FC<DeploymentPrometheusTabProps> = ({
  clusterName,
  namespace,
  deploymentName,
}) => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [chartData, setChartData] = useState<{ time: string; value: number }[]>([]);

  const fetchMetrics = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const end = Math.floor(Date.now() / 1000);
      const start = end - 3600;
      const query = `sum(rate(container_cpu_usage_seconds_total{namespace="${namespace}",pod=~"${deploymentName}-.*"}[5m]))`;
      const res = await queryPrometheusRange(clusterName, {
        query,
        start: String(start),
        end: String(end),
        step: '60',
      });
      const body = res.data as PrometheusRangeResponse;
      if (body.status !== 'success') {
        setError(body.error || t('prometheus.queryFailed'));
        setChartData([]);
        return;
      }
      const result = body.data?.result;
      if (!result?.length) {
        setError(t('prometheus.noData'));
        setChartData([]);
        return;
      }
      const values = result[0]?.values || [];
      setChartData(
        values.map(([ts, val]: [number, string]) => ({
          time: new Date(ts * 1000).toLocaleTimeString(),
          value: parseFloat(val) || 0,
        }))
      );
    } catch {
      setError(t('prometheus.queryFailed'));
      setChartData([]);
    } finally {
      setLoading(false);
    }
  }, [clusterName, namespace, deploymentName, t]);

  useEffect(() => {
    fetchMetrics();
  }, [fetchMetrics]);

  if (error) {
    return (
      <Alert
        type="info"
        showIcon
        message={t('prometheus.title')}
        description={error}
      />
    );
  }

  return (
    <Card title={t('prometheus.cpuUsage')} loading={loading}>
      <Spin spinning={loading}>
        <ResponsiveContainer width="100%" height={300}>
          <LineChart data={chartData}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="time" />
            <YAxis />
            <Tooltip />
            <Legend />
            <Line type="monotone" dataKey="value" name={t('prometheus.cpuCores')} stroke="#1890ff" dot={false} />
          </LineChart>
        </ResponsiveContainer>
      </Spin>
    </Card>
  );
};

export default DeploymentPrometheusTab;
