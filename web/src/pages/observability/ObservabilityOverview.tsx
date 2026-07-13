import React, { useEffect, useState } from 'react';
import { Card, Row, Col, Statistic, Tag, Alert, List, Button, Space, Spin, message } from 'antd';
import {
  AlertOutlined,
  ApiOutlined,
  ClusterOutlined,
  LineChartOutlined,
  ApartmentOutlined,
  ThunderboltOutlined,
  FileSearchOutlined,
} from '@ant-design/icons';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useClusterNamespace } from '@/hooks/useClusterNamespace';
import ClusterNamespaceToolbar from '@/components/k8s/common/ClusterNamespaceToolbar';
import { getObservabilitySummary, getObservabilityAlerts, ObservabilityAlert } from '@/api/observability';

const statusColor = (status: string) => {
  if (status === 'healthy' || status === 'ok') return 'green';
  if (status === 'degraded') return 'orange';
  return 'red';
};

const ObservabilityOverview: React.FC = () => {
  const { t } = useTranslation();
  const { selectedCluster, setSelectedCluster, clusters, clustersLoading } = useClusterNamespace(t);
  const [loading, setLoading] = useState(false);
  const [summary, setSummary] = useState<Awaited<ReturnType<typeof getObservabilitySummary>>['data']['data']['summary'] | null>(null);
  const [alerts, setAlerts] = useState<ObservabilityAlert[]>([]);

  const fetchData = async () => {
    if (!selectedCluster) return;
    setLoading(true);
    try {
      const [summaryRes, alertsRes] = await Promise.all([
        getObservabilitySummary(selectedCluster),
        getObservabilityAlerts(selectedCluster, 10),
      ]);
      if (summaryRes.data.code === 0) {
        setSummary(summaryRes.data.data.summary);
      }
      if (alertsRes.data.code === 0) {
        setAlerts(alertsRes.data.data.alerts || []);
      }
    } catch {
      message.error(t('observability.fetchFailed'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (selectedCluster) fetchData();
  }, [selectedCluster]);

  const quickLinks = [
    { path: '/observability/prometheus', icon: <LineChartOutlined />, label: t('navigation.prometheus') },
    { path: '/observability/alerts', icon: <AlertOutlined />, label: t('navigation.alerts') },
    { path: '/observability/cluster-events', icon: <ThunderboltOutlined />, label: t('navigation.clusterEvents') },
    { path: '/observability/call-chain', icon: <ApartmentOutlined />, label: t('navigation.callChain') },
    { path: '/observability/service-topology', icon: <ApartmentOutlined />, label: t('navigation.serviceTopology') },
    { path: '/observability/label-logs', icon: <FileSearchOutlined />, label: t('navigation.labelLogs') },
  ];

  return (
    <Card title={t('observability.overview.title')}>
      <ClusterNamespaceToolbar
        clusters={clusters}
        selectedCluster={selectedCluster}
        onClusterChange={setSelectedCluster}
        namespace="all"
        onNamespaceChange={() => {}}
        loading={clustersLoading}
        showNamespace={false}
        extra={
          <Button onClick={fetchData} loading={loading}>
            {t('common.refresh')}
          </Button>
        }
      />

      {!selectedCluster && (
        <Alert type="warning" showIcon style={{ marginBottom: 16 }} message={t('observability.selectCluster')} />
      )}

      <Spin spinning={loading}>
        {summary && (
          <>
            <Row gutter={16} style={{ marginBottom: 16 }}>
              <Col span={6}>
                <Card size="small">
                  <Statistic
                    title={t('observability.overview.prometheus')}
                    value={summary.prometheus.healthy ? t('observability.overview.connected') : t('observability.overview.disconnected')}
                    valueStyle={{ color: summary.prometheus.healthy ? '#52c41a' : '#faad14', fontSize: 18 }}
                    prefix={<ApiOutlined />}
                  />
                  <div style={{ marginTop: 8 }}>
                    <Tag>
                      {t(`observability.overview.sources.${summary.prometheus.source}`, {
                        defaultValue: summary.prometheus.source,
                      })}
                    </Tag>
                  </div>
                </Card>
              </Col>
              <Col span={6}>
                <Card size="small">
                  <Statistic
                    title={t('observability.overview.clusterHealth')}
                    value={summary.health?.overall || 'unknown'}
                    valueStyle={{ fontSize: 18 }}
                    prefix={<ClusterOutlined />}
                  />
                </Card>
              </Col>
              <Col span={6}>
                <Card size="small">
                  <Statistic
                    title={t('observability.overview.problemPods')}
                    value={(summary.pods.crashLoop || 0) + (summary.pods.failed || 0) + (summary.pods.notReady || 0)}
                    valueStyle={{ color: '#ff4d4f', fontSize: 18 }}
                  />
                  <div style={{ marginTop: 8, fontSize: 12, color: '#888' }}>
                    {t('observability.overview.running')}: {summary.pods.running}/{summary.pods.total}
                  </div>
                </Card>
              </Col>
              <Col span={6}>
                <Card size="small">
                  <Statistic
                    title={t('observability.overview.warningEvents')}
                    value={summary.events.warningRecent}
                    valueStyle={{ fontSize: 18 }}
                  />
                </Card>
              </Col>
            </Row>

            {summary.health?.components && (
              <Card type="inner" title={t('observability.overview.components')} style={{ marginBottom: 16 }}>
                <Space wrap>
                  {summary.health.components.map((c) => (
                    <Tag key={c.name} color={statusColor(c.status)}>
                      {c.name}: {c.status}
                    </Tag>
                  ))}
                </Space>
              </Card>
            )}
          </>
        )}

        <Row gutter={16}>
          <Col span={12}>
            <Card type="inner" title={t('observability.overview.quickLinks')}>
              <Space direction="vertical" style={{ width: '100%' }}>
                {quickLinks.map((link) => (
                  <Link key={link.path} to={link.path}>
                    <Button block icon={link.icon} style={{ textAlign: 'left' }}>
                      {link.label}
                    </Button>
                  </Link>
                ))}
              </Space>
            </Card>
          </Col>
          <Col span={12}>
            <Card
              type="inner"
              title={t('observability.overview.recentAlerts')}
              extra={
                <Link to="/observability/alerts">{t('observability.overview.viewAll')}</Link>
              }
            >
              <List
                size="small"
                dataSource={alerts}
                locale={{ emptyText: t('observability.alerts.empty') }}
                renderItem={(item) => (
                  <List.Item>
                    <Space>
                      <Tag color={item.level === 'error' ? 'red' : 'orange'}>{item.level}</Tag>
                      <Tag>{item.category}</Tag>
                      <span>
                        {item.namespace ? `${item.namespace}/` : ''}
                        {item.name}
                      </span>
                      <span style={{ color: '#888' }}>{item.message}</span>
                    </Space>
                  </List.Item>
                )}
              />
            </Card>
          </Col>
        </Row>
      </Spin>
    </Card>
  );
};

export default ObservabilityOverview;
