import React, { useCallback, useEffect, useState } from 'react';
import {
  Alert,
  Card,
  Col,
  Descriptions,
  Progress,
  Row,
  Spin,
  Statistic,
  Table,
  Tag,
} from 'antd';
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
import { getDeploymentMetrics, WorkloadMetrics } from '@/api/deployment';
import { formatFileSize } from '@/utils/format';

interface DeploymentMonitoringProps {
  clusterName: string;
  namespace: string;
  deploymentName: string;
}

const healthColor: Record<string, string> = {
  healthy: 'success',
  warning: 'warning',
  critical: 'error',
  unknown: 'default',
};

const emptyMetrics = (): WorkloadMetrics => ({
  workloadType: 'deployment',
  name: '',
  namespace: '',
  summary: {
    replicas: 0,
    readyReplicas: 0,
    availableReplicas: 0,
    podCount: 0,
    runningPods: 0,
    metricsPodCount: 0,
    avgCpuUsage: 0,
    maxCpuUsage: 0,
    avgMemoryUsage: 0,
    maxMemoryUsage: 0,
    avgDiskUsed: '0 B',
    maxDiskUsed: '0 B',
    totalDiskUsed: '0 B',
    avgDiskUsedBytes: 0,
    maxDiskUsedBytes: 0,
    totalDiskUsedBytes: 0,
    cpuRequests: '0m',
    cpuLimits: '0m',
    memoryRequests: '0Mi',
    memoryLimits: '0Mi',
    diskRequests: '0Gi',
    diskLimits: '0Gi',
    healthStatus: 'unknown',
    alerts: [],
  },
  monitoringStrategy: {
    policy: 'container-group-aggregate',
    description: '',
    thresholds: {
      cpuWarning: 70,
      cpuCritical: 85,
      memoryWarning: 70,
      memoryCritical: 85,
    },
    podCoverage: '',
    recommendation: '',
  },
  pods: [],
  containerGroups: [],
  historicalData: {
    cpuUsage: [],
    memoryUsage: [],
    diskUsage: [],
  },
});

const normalizeMetrics = (raw: WorkloadMetrics | null | undefined): WorkloadMetrics => {
  const base = emptyMetrics();
  if (!raw) {
    return base;
  }
  return {
    ...base,
    ...raw,
    summary: { ...base.summary, ...(raw.summary || {}) },
    monitoringStrategy: {
      ...base.monitoringStrategy,
      ...(raw.monitoringStrategy || {}),
      thresholds: {
        ...base.monitoringStrategy.thresholds,
        ...(raw.monitoringStrategy?.thresholds || {}),
      },
    },
    pods: raw.pods || [],
    containerGroups: raw.containerGroups || [],
    historicalData: {
      cpuUsage: raw.historicalData?.cpuUsage || [],
      memoryUsage: raw.historicalData?.memoryUsage || [],
      diskUsage: raw.historicalData?.diskUsage || [],
    },
  };
};

const formatPercent = (value?: number) => (Number.isFinite(value) ? value!.toFixed(1) : '0.0');

const DeploymentMonitoring: React.FC<DeploymentMonitoringProps> = ({
  clusterName,
  namespace,
  deploymentName,
}) => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(true);
  const [metrics, setMetrics] = useState<WorkloadMetrics | null>(null);
  const [fetchError, setFetchError] = useState<string | null>(null);

  const fetchMetrics = useCallback(async () => {
    if (!clusterName || !namespace || !deploymentName) {
      setLoading(false);
      return;
    }
    try {
      setLoading(true);
      setFetchError(null);
      const response = await getDeploymentMetrics(clusterName, namespace, deploymentName);
      if (response.data.code === 0 && response.data.data?.metrics) {
        setMetrics(normalizeMetrics(response.data.data.metrics));
      } else {
        setMetrics(null);
        setFetchError(response.data.message || t('deployments.detail.monitoring.noData'));
      }
    } catch (error) {
      console.error('Failed to fetch deployment metrics:', error);
      setMetrics(null);
      setFetchError(t('deployments.detail.monitoring.fetchFailed'));
    } finally {
      setLoading(false);
    }
  }, [clusterName, namespace, deploymentName, t]);

  useEffect(() => {
    fetchMetrics();
    const intervalId = setInterval(fetchMetrics, 30000);
    return () => clearInterval(intervalId);
  }, [fetchMetrics]);

  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
    if (Number.isNaN(date.getTime())) {
      return timestamp;
    }
    return `${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`;
  };

  const renderHealthTag = (status?: string) => (
    <Tag color={healthColor[status || 'unknown'] || 'default'}>
      {t(`deployments.detail.monitoring.health.${status || 'unknown'}`, status || 'unknown')}
    </Tag>
  );

  if (loading && !metrics) {
    return (
      <div style={{ textAlign: 'center', padding: 24 }}>
        <Spin size="large" tip={t('deployments.detail.monitoring.loading')} />
      </div>
    );
  }

  if (!metrics) {
    return (
      <Alert
        type={fetchError ? 'warning' : 'info'}
        showIcon
        message={fetchError || t('deployments.detail.monitoring.noData')}
      />
    );
  }

  const { summary, monitoringStrategy, containerGroups, pods, historicalData } = metrics;
  const alerts = summary.alerts || [];

  const containerColumns = [
    {
      title: t('deployments.detail.monitoring.containerGroup'),
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: t('deployments.detail.monitoring.podInstances'),
      dataIndex: 'podCount',
      key: 'podCount',
    },
    {
      title: t('deployments.detail.monitoring.avgCpu'),
      dataIndex: 'avgCpuUsage',
      key: 'avgCpuUsage',
      render: (value: number) => `${formatPercent(value)}%`,
    },
    {
      title: t('deployments.detail.monitoring.maxCpu'),
      dataIndex: 'maxCpuUsage',
      key: 'maxCpuUsage',
      render: (value: number) => `${formatPercent(value)}%`,
    },
    {
      title: t('deployments.detail.monitoring.avgMemory'),
      dataIndex: 'avgMemoryUsage',
      key: 'avgMemoryUsage',
      render: (value: number) => `${formatPercent(value)}%`,
    },
    {
      title: t('deployments.detail.monitoring.maxMemory'),
      dataIndex: 'maxMemoryUsage',
      key: 'maxMemoryUsage',
      render: (value: number) => `${formatPercent(value)}%`,
    },
    {
      title: t('deployments.detail.monitoring.avgDisk'),
      dataIndex: 'avgDiskUsed',
      key: 'avgDiskUsed',
    },
    {
      title: t('deployments.detail.monitoring.maxDisk'),
      dataIndex: 'maxDiskUsed',
      key: 'maxDiskUsed',
    },
    {
      title: t('deployments.detail.monitoring.healthStatus'),
      dataIndex: 'healthStatus',
      key: 'healthStatus',
      render: (status: string) => renderHealthTag(status),
    },
  ];

  const podColumns = [
    {
      title: 'Pod',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: t('deployments.detail.monitoring.phase'),
      dataIndex: 'phase',
      key: 'phase',
    },
    {
      title: t('deployments.detail.monitoring.ready'),
      dataIndex: 'ready',
      key: 'ready',
      render: (ready: boolean) => (
        <Tag color={ready ? 'success' : 'default'}>
          {ready ? t('common.yes') : t('common.no')}
        </Tag>
      ),
    },
    {
      title: t('deployments.detail.monitoring.cpuUsage'),
      dataIndex: 'cpuUsage',
      key: 'cpuUsage',
      render: (value: number) => `${formatPercent(value)}%`,
    },
    {
      title: t('deployments.detail.monitoring.memoryUsage'),
      dataIndex: 'memoryUsage',
      key: 'memoryUsage',
      render: (value: number) => `${formatPercent(value)}%`,
    },
    {
      title: t('deployments.detail.monitoring.diskUsed'),
      dataIndex: 'diskUsed',
      key: 'diskUsed',
      render: (value: string, record: { diskUsedBytes?: number }) =>
        value || formatFileSize(record.diskUsedBytes || 0),
    },
    {
      title: t('deployments.detail.monitoring.restarts'),
      dataIndex: 'restarts',
      key: 'restarts',
    },
    {
      title: t('deployments.detail.monitoring.healthStatus'),
      dataIndex: 'healthStatus',
      key: 'healthStatus',
      render: (status: string) => renderHealthTag(status),
    },
  ];

  const cpuHistory = (historicalData?.cpuUsage || []).map((item) => ({
    name: formatTime(item.timestamp),
    value: item.value,
  }));
  const memoryHistory = (historicalData?.memoryUsage || []).map((item) => ({
    name: formatTime(item.timestamp),
    value: item.value,
  }));
  const diskHistory = (historicalData?.diskUsage || []).map((item) => ({
    name: formatTime(item.timestamp),
    value: item.value,
    label: formatFileSize(item.value),
  }));

  return (
    <div>
      {alerts.length > 0 && (
        <div style={{ marginBottom: 16 }}>
          {alerts.map((alert, index) => (
            <Alert
              key={`${alert.source}-${alert.metric}-${index}`}
              type={alert.level === 'critical' ? 'error' : 'warning'}
              showIcon
              message={alert.source}
              description={alert.message}
              style={{ marginBottom: 8 }}
            />
          ))}
        </div>
      )}

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title={t('deployments.detail.monitoring.appHealth')}
              value={t(
                `deployments.detail.monitoring.health.${summary.healthStatus || 'unknown'}`,
                summary.healthStatus || 'unknown',
              )}
              valueStyle={{
                color:
                  summary.healthStatus === 'critical'
                    ? '#cf1322'
                    : summary.healthStatus === 'warning'
                      ? '#d48806'
                      : '#3f8600',
                fontSize: 20,
              }}
            />
            <div style={{ marginTop: 8 }}>{renderHealthTag(summary.healthStatus)}</div>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title={t('deployments.detail.monitoring.replicaReady')}
              value={`${summary.readyReplicas}/${summary.replicas}`}
            />
            <Progress
              percent={
                summary.replicas > 0
                  ? Math.round((summary.readyReplicas / summary.replicas) * 100)
                  : 0
              }
              size="small"
              style={{ marginTop: 12 }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <Card title={t('deployments.detail.monitoring.cpuUsage')}>
            <Statistic
              value={formatPercent(summary.avgCpuUsage)}
              suffix={`% (${t('deployments.detail.monitoring.avg')})`}
            />
            <div style={{ marginTop: 8, color: '#666' }}>
              {t('deployments.detail.monitoring.max')}: {formatPercent(summary.maxCpuUsage)}%
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <Card title={t('deployments.detail.monitoring.memoryUsage')}>
            <Statistic
              value={formatPercent(summary.avgMemoryUsage)}
              suffix={`% (${t('deployments.detail.monitoring.avg')})`}
            />
            <div style={{ marginTop: 8, color: '#666' }}>
              {t('deployments.detail.monitoring.max')}: {formatPercent(summary.maxMemoryUsage)}%
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={4}>
          <Card title={t('deployments.detail.monitoring.diskUsed')}>
            <Statistic value={summary.avgDiskUsed || formatFileSize(summary.avgDiskUsedBytes || 0)} />
            <div style={{ marginTop: 8, color: '#666' }}>
              {t('deployments.detail.monitoring.max')}: {summary.maxDiskUsed || formatFileSize(summary.maxDiskUsedBytes || 0)}
            </div>
            <div style={{ marginTop: 4, color: '#666' }}>
              {t('deployments.detail.monitoring.total')}: {summary.totalDiskUsed || formatFileSize(summary.totalDiskUsedBytes || 0)}
            </div>
          </Card>
        </Col>
      </Row>

      <Card title={t('deployments.detail.monitoring.strategyTitle')} style={{ marginTop: 16 }} size="small">
        <Descriptions column={1} size="small">
          <Descriptions.Item label={t('deployments.detail.monitoring.policy')}>
            {monitoringStrategy.policy}
          </Descriptions.Item>
          <Descriptions.Item label={t('deployments.detail.monitoring.description')}>
            {monitoringStrategy.description}
          </Descriptions.Item>
          <Descriptions.Item label={t('deployments.detail.monitoring.coverage')}>
            {monitoringStrategy.podCoverage}
          </Descriptions.Item>
          <Descriptions.Item label={t('deployments.detail.monitoring.thresholds')}>
            CPU {monitoringStrategy.thresholds.cpuWarning}% / {monitoringStrategy.thresholds.cpuCritical}%,
            {' '}
            {t('deployments.detail.monitoring.memoryUsage')} {monitoringStrategy.thresholds.memoryWarning}% / {monitoringStrategy.thresholds.memoryCritical}%
          </Descriptions.Item>
          <Descriptions.Item label={t('deployments.detail.monitoring.recommendation')}>
            {monitoringStrategy.recommendation}
          </Descriptions.Item>
          <Descriptions.Item label={t('deployments.detail.monitoring.totalRequests')}>
            CPU {summary.cpuRequests} / {summary.cpuLimits},
            {' '}
            {t('deployments.detail.monitoring.memoryUsage')} {summary.memoryRequests} / {summary.memoryLimits},
            {' '}
            {t('deployments.detail.monitoring.diskQuota')} {summary.diskRequests} / {summary.diskLimits}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} md={8}>
          <Card title={t('deployments.detail.monitoring.cpuHistory')}>
            <div style={{ width: '100%', minWidth: 0, height: 260 }}>
              {cpuHistory.length > 0 ? (
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={cpuHistory} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="name" />
                    <YAxis unit="%" />
                    <Tooltip />
                    <Legend />
                    <Line
                      type="monotone"
                      dataKey="value"
                      name={t('deployments.detail.monitoring.avgCpu')}
                      stroke="#8884d8"
                    />
                  </LineChart>
                </ResponsiveContainer>
              ) : (
                <div style={{ padding: 24, textAlign: 'center', color: '#999' }}>
                  {t('deployments.detail.monitoring.noHistory')}
                </div>
              )}
            </div>
          </Card>
        </Col>
        <Col xs={24} md={8}>
          <Card title={t('deployments.detail.monitoring.memoryHistory')}>
            <div style={{ width: '100%', minWidth: 0, height: 260 }}>
              {memoryHistory.length > 0 ? (
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={memoryHistory} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="name" />
                    <YAxis unit="%" />
                    <Tooltip />
                    <Legend />
                    <Line
                      type="monotone"
                      dataKey="value"
                      name={t('deployments.detail.monitoring.avgMemory')}
                      stroke="#82ca9d"
                    />
                  </LineChart>
                </ResponsiveContainer>
              ) : (
                <div style={{ padding: 24, textAlign: 'center', color: '#999' }}>
                  {t('deployments.detail.monitoring.noHistory')}
                </div>
              )}
            </div>
          </Card>
        </Col>
        <Col xs={24} md={8}>
          <Card title={t('deployments.detail.monitoring.diskHistory')}>
            <div style={{ width: '100%', minWidth: 0, height: 260 }}>
              {diskHistory.length > 0 ? (
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={diskHistory} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="name" />
                    <YAxis tickFormatter={(value) => formatFileSize(value)} />
                    <Tooltip formatter={(value: number) => formatFileSize(value)} />
                    <Legend />
                    <Line
                      type="monotone"
                      dataKey="value"
                      name={t('deployments.detail.monitoring.avgDisk')}
                      stroke="#f5a442"
                    />
                  </LineChart>
                </ResponsiveContainer>
              ) : (
                <div style={{ padding: 24, textAlign: 'center', color: '#999' }}>
                  {t('deployments.detail.monitoring.noHistory')}
                </div>
              )}
            </div>
          </Card>
        </Col>
      </Row>

      <Card title={t('deployments.detail.monitoring.containerGroupSummary')} style={{ marginTop: 16 }}>
        <Table
          columns={containerColumns}
          dataSource={containerGroups}
          rowKey="name"
          pagination={false}
          scroll={{ x: 'max-content' }}
          locale={{ emptyText: t('deployments.detail.monitoring.noContainerGroups') }}
        />
      </Card>

      <Card title={t('deployments.detail.monitoring.podBreakdown')} style={{ marginTop: 16 }}>
        <Table
          columns={podColumns}
          dataSource={pods}
          rowKey="name"
          pagination={false}
          scroll={{ x: 'max-content' }}
          locale={{ emptyText: t('deployments.detail.monitoring.noPods') }}
        />
      </Card>
    </div>
  );
};

export default DeploymentMonitoring;
