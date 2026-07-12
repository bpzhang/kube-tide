import React, { useCallback, useEffect, useState } from 'react';
import { Alert, Card, Select, Space, Spin } from 'antd';
import { useTranslation } from 'react-i18next';
import { queryPrometheusRange } from '@/api/prometheus';
import PrometheusChart from '@/components/observability/PrometheusChart';
import {
  buildRangeParams,
  deploymentCPUQuery,
  deploymentMemoryQuery,
  parsePrometheusRange,
  PrometheusRangeResponse,
} from '@/utils/prometheus';

interface DeploymentPrometheusTabProps {
  clusterName: string;
  namespace: string;
  deploymentName: string;
}

const DeploymentPrometheusTab: React.FC<DeploymentPrometheusTabProps> = ({
  clusterName,
  namespace,
  deploymentName,
}) => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [hours, setHours] = useState(1);
  const [cpuSeries, setCpuSeries] = useState<ReturnType<typeof parsePrometheusRange>>([]);
  const [memorySeries, setMemorySeries] = useState<ReturnType<typeof parsePrometheusRange>>([]);

  const fetchMetrics = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const cpuParams = buildRangeParams(deploymentCPUQuery(namespace, deploymentName), hours);
      const memParams = buildRangeParams(deploymentMemoryQuery(namespace, deploymentName), hours);

      const [cpuRes, memRes] = await Promise.all([
        queryPrometheusRange(clusterName, cpuParams),
        queryPrometheusRange(clusterName, memParams),
      ]);

      const cpuBody = cpuRes.data as PrometheusRangeResponse;
      const memBody = memRes.data as PrometheusRangeResponse;

      if (cpuBody.status !== 'success' && memBody.status !== 'success') {
        setError(cpuBody.error || memBody.error || t('prometheus.queryFailed'));
        setCpuSeries([]);
        setMemorySeries([]);
        return;
      }

      const cpu = parsePrometheusRange(cpuBody);
      const mem = parsePrometheusRange(memBody);
      setCpuSeries(cpu.length ? cpu : []);
      setMemorySeries(mem.length ? mem : []);

      if (!cpu.length && !mem.length) {
        setError(t('prometheus.noData'));
      }
    } catch {
      setError(t('prometheus.queryFailed'));
      setCpuSeries([]);
      setMemorySeries([]);
    } finally {
      setLoading(false);
    }
  }, [clusterName, namespace, deploymentName, hours, t]);

  useEffect(() => {
    fetchMetrics();
  }, [fetchMetrics]);

  if (error && !cpuSeries.length && !memorySeries.length) {
    return (
      <Alert type="info" showIcon message={t('prometheus.title')} description={error} />
    );
  }

  return (
    <Spin spinning={loading}>
      <Space style={{ marginBottom: 16 }}>
        <span>{t('prometheus.explorer.timeRange')}</span>
        <Select value={hours} onChange={setHours} style={{ width: 100 }}>
          <Select.Option value={1}>1h</Select.Option>
          <Select.Option value={6}>6h</Select.Option>
          <Select.Option value={24}>24h</Select.Option>
        </Select>
      </Space>

      <Card title={t('prometheus.cpuUsage')} style={{ marginBottom: 16 }}>
        {cpuSeries.length > 0 ? (
          <PrometheusChart series={cpuSeries} yLabel={t('prometheus.cpuCores')} />
        ) : (
          <Alert type="info" message={t('prometheus.noData')} />
        )}
      </Card>

      <Card title={t('prometheus.memoryUsage')}>
        {memorySeries.length > 0 ? (
          <PrometheusChart series={memorySeries} yLabel={t('prometheus.memoryBytes')} />
        ) : (
          <Alert type="info" message={t('prometheus.noData')} />
        )}
      </Card>
    </Spin>
  );
};

export default DeploymentPrometheusTab;
