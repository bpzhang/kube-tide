import React, { useEffect, useMemo, useState } from 'react';
import { Card, Table, Tag, Space, Typography, Alert, Statistic, Row, Col, message, Tabs, Descriptions, Button, Spin } from 'antd';
import TrafficTopologyGraph from '@/components/observability/TrafficTopologyGraph';
import { ArrowRightOutlined, ApiOutlined, ClusterOutlined, DeploymentUnitOutlined, ReloadOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useClusterNamespace } from '@/hooks/useClusterNamespace';
import ClusterNamespaceToolbar from '@/components/k8s/common/ClusterNamespaceToolbar';
import {
  getTrafficTopology,
  parseNodeLabel,
  TopologyEdge,
  TrafficPath,
  TrafficTopology,
} from '@/api/traffic_topology';

const { Text } = Typography;

const ServiceTopology: React.FC = () => {
  const { t } = useTranslation();
  const { selectedCluster, setSelectedCluster, clusters, namespace, setNamespace, clustersLoading } =
    useClusterNamespace(t);
  const [topology, setTopology] = useState<TrafficTopology | null>(null);
  const [loading, setLoading] = useState(false);

  const fetchTopology = async () => {
    if (!selectedCluster) return;
    setLoading(true);
    try {
      const response = await getTrafficTopology(selectedCluster, namespace);
      if (response.data.code === 0) {
        setTopology(response.data.data);
      } else {
        message.error(response.data.message || t('trafficTopology.fetchFailed'));
        setTopology(null);
      }
    } catch {
      message.error(t('trafficTopology.fetchFailed'));
      setTopology(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (selectedCluster) fetchTopology();
  }, [selectedCluster, namespace]);

  const stats = useMemo(() => {
    if (!topology) return { ingress: 0, service: 0, routes: 0, calls: 0 };
    const ingress = topology.nodes.filter((n) => n.type === 'ingress').length;
    const service = topology.nodes.filter((n) => n.type === 'service').length;
    const routes = topology.edges.filter((e) => e.edgeType === 'routes').length;
    const calls = topology.edges.filter((e) => e.edgeType === 'calls').length;
    return { ingress, service, routes, calls };
  }, [topology]);

  const callEdges = useMemo(
    () => (topology?.edges || []).filter((e) => e.edgeType === 'calls'),
    [topology],
  );

  const pathColumns = [
    {
      title: t('trafficTopology.columns.ingress'),
      key: 'ingress',
      render: (_: unknown, row: TrafficPath) =>
        row.ingressName ? (
          <Space direction="vertical" size={0}>
            <Text strong>{row.ingressName}</Text>
            {row.ingressHost && <Text type="secondary">{row.ingressHost}</Text>}
            {row.path && <Tag>{row.path}</Tag>}
          </Space>
        ) : (
          <Text type="secondary">—</Text>
        ),
    },
    {
      title: t('trafficTopology.columns.service'),
      dataIndex: 'serviceName',
      key: 'service',
      render: (name: string, row: TrafficPath) => (
        <Tag color="blue">
          {row.namespace}/{name}
        </Tag>
      ),
    },
    {
      title: t('trafficTopology.columns.workload'),
      key: 'workload',
      render: (_: unknown, row: TrafficPath) =>
        row.workloadName ? (
          <Tag color="green">
            {row.workloadType}/{row.workloadName}
          </Tag>
        ) : (
          <Text type="secondary">—</Text>
        ),
    },
    { title: t('trafficTopology.columns.pods'), dataIndex: 'podCount', key: 'podCount', width: 80 },
  ];

  const callColumns = [
    {
      title: t('trafficTopology.columns.caller'),
      dataIndex: 'source',
      key: 'source',
      render: (id: string) => <Tag color="purple">{parseNodeLabel(id)}</Tag>,
    },
    {
      title: '',
      key: 'arrow',
      width: 40,
      render: () => <ArrowRightOutlined />,
    },
    {
      title: t('trafficTopology.columns.callee'),
      dataIndex: 'target',
      key: 'target',
      render: (id: string) => <Tag color="blue">{parseNodeLabel(id)}</Tag>,
    },
    {
      title: t('trafficTopology.columns.evidence'),
      dataIndex: 'evidence',
      key: 'evidence',
      render: (v: string) => <Text code>{v}</Text>,
    },
    {
      title: t('trafficTopology.columns.inferred'),
      dataIndex: 'inferred',
      key: 'inferred',
      width: 100,
      render: (v: boolean) =>
        v ? <Tag color="orange">{t('trafficTopology.inferred')}</Tag> : <Tag>{t('trafficTopology.declared')}</Tag>,
    },
  ];

  const routeEdges = (topology?.edges || []).filter((e: TopologyEdge) => e.edgeType === 'routes');

  const networkMessage = topology?.network?.message
    ? t(`trafficTopology.network.messages.${topology.network.message}`, { defaultValue: topology.network.message })
    : '';

  const hubbleMessage = topology?.hubble?.message
    ? t(`trafficTopology.hubble.messages.${topology.hubble.message}`, { defaultValue: topology.hubble.message })
    : '';

  const policyEdges = useMemo(
    () => (topology?.edges || []).filter((e) => e.edgeType === 'policy_allow'),
    [topology],
  );

  return (
    <Card title={t('trafficTopology.title')}>
      <ClusterNamespaceToolbar
        clusters={clusters}
        selectedCluster={selectedCluster}
        onClusterChange={setSelectedCluster}
        namespace={namespace}
        onNamespaceChange={setNamespace}
        loading={clustersLoading}
        extra={
          <Button icon={<ReloadOutlined />} onClick={fetchTopology} loading={loading}>
            {t('common.refresh')}
          </Button>
        }
      />

      {!selectedCluster && (
        <Alert type="warning" showIcon style={{ marginBottom: 16 }} message={t('trafficTopology.selectCluster')} />
      )}

      <Spin spinning={loading}>
      {topology?.network && (
        <Card type="inner" title={t('trafficTopology.network.title')} style={{ marginBottom: 16 }}>
          <Descriptions size="small" column={2} bordered>
            <Descriptions.Item label={t('trafficTopology.network.cni')}>
              <Tag color={topology.network.cni === 'terway' ? 'blue' : 'default'}>
                {topology.network.cni === 'terway' ? 'Terway' : topology.network.cni}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label={t('trafficTopology.network.mode')}>
              {topology.network.terwayMode || '—'}
            </Descriptions.Item>
            <Descriptions.Item label="Hubble Metrics">
              {topology.network.hubbleMetricsSvc ? t('trafficTopology.network.ready') : t('trafficTopology.network.notReady')}
            </Descriptions.Item>
            <Descriptions.Item label={t('trafficTopology.network.hubbleRelay')}>
              {topology.network.hubbleRelayReady ? t('trafficTopology.network.ready') : t('trafficTopology.network.notReady')}
            </Descriptions.Item>
            <Descriptions.Item label={t('trafficTopology.network.dataSource')} span={2}>
              {topology.network.metricsSource
                ? t(`trafficTopology.hubble.messages.${topology.network.metricsSource}`, {
                    defaultValue: topology.network.metricsSource,
                  })
                : networkMessage || '—'}
            </Descriptions.Item>
          </Descriptions>
        </Card>
      )}

      {topology?.hubble?.available && (
        <Card type="inner" title={t('trafficTopology.hubble.title')} style={{ marginBottom: 16 }}>
          <Row gutter={16}>
            <Col span={12}>
              <Table
                size="small"
                pagination={false}
                rowKey="reason"
                dataSource={topology.hubble.drops || []}
                locale={{ emptyText: t('trafficTopology.hubble.emptyDrops') }}
                columns={[
                  { title: t('trafficTopology.hubble.dropReason'), dataIndex: 'reason' },
                  { title: t('trafficTopology.hubble.count'), dataIndex: 'count', render: (v: number) => Math.round(v) },
                ]}
              />
            </Col>
            <Col span={12}>
              <Table
                size="small"
                pagination={false}
                rowKey={(r) => `${r.protocol}-${r.port}`}
                dataSource={topology.hubble.topPorts || []}
                locale={{ emptyText: t('trafficTopology.hubble.emptyPorts') }}
                columns={[
                  { title: t('trafficTopology.columns.port'), dataIndex: 'port' },
                  { title: t('trafficTopology.hubble.protocol'), dataIndex: 'protocol' },
                  { title: t('trafficTopology.hubble.count'), dataIndex: 'count', render: (v: number) => Math.round(v) },
                ]}
              />
            </Col>
          </Row>
        </Card>
      )}

      {topology?.hubble && !topology.hubble.available && topology.hubble.message && (
        <Alert type="info" showIcon style={{ marginBottom: 16 }} message={t('trafficTopology.hubble.title')} description={hubbleMessage} />
      )}

      <Alert
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
        message={t('trafficTopology.hintTitle')}
        description={t('trafficTopology.hint')}
      />

      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={6}>
          <Statistic title={t('trafficTopology.stats.ingress')} value={stats.ingress} prefix={<ApiOutlined />} />
        </Col>
        <Col span={6}>
          <Statistic title={t('trafficTopology.stats.services')} value={stats.service} prefix={<ClusterOutlined />} />
        </Col>
        <Col span={6}>
          <Statistic title={t('trafficTopology.stats.routes')} value={stats.routes} prefix={<ArrowRightOutlined />} />
        </Col>
        <Col span={6}>
          <Statistic title={t('trafficTopology.stats.calls')} value={stats.calls} prefix={<DeploymentUnitOutlined />} />
        </Col>
      </Row>

      <Card type="inner" title={t('trafficTopology.graphTitle')} style={{ marginBottom: 16 }}>
        <TrafficTopologyGraph topology={topology} loading={loading} />
      </Card>

      <Tabs
        style={{ marginBottom: 16 }}
        items={[
          {
            key: 'paths',
            label: t('trafficTopology.pathsTitle'),
            children: (
              <Table
                rowKey={(r) => `${r.namespace}-${r.ingressName}-${r.serviceName}-${r.path}`}
                loading={loading}
                dataSource={topology?.paths || []}
                columns={pathColumns}
                pagination={{ pageSize: 10 }}
                locale={{ emptyText: t('trafficTopology.emptyPaths') }}
              />
            ),
          },
          {
            key: 'routes',
            label: t('trafficTopology.routesTitle'),
            children: (
              <Table
                rowKey={(r) => `${r.source}-${r.target}-${r.port}`}
                loading={loading}
                dataSource={routeEdges}
                pagination={false}
                locale={{ emptyText: t('trafficTopology.emptyRoutes') }}
                columns={[
                  { title: t('trafficTopology.columns.ingress'), dataIndex: 'source', render: parseNodeLabel },
                  { title: '', width: 40, render: () => <ArrowRightOutlined /> },
                  { title: t('trafficTopology.columns.service'), dataIndex: 'target', render: parseNodeLabel },
                  { title: t('trafficTopology.columns.port'), dataIndex: 'port' },
                ]}
              />
            ),
          },
          {
            key: 'calls',
            label: t('trafficTopology.callsTitle'),
            children: (
              <Table
                rowKey={(r) => `${r.source}-${r.target}-${r.evidence}`}
                loading={loading}
                dataSource={callEdges}
                columns={callColumns}
                pagination={{ pageSize: 10 }}
                locale={{ emptyText: t('trafficTopology.emptyCalls') }}
              />
            ),
          },
          {
            key: 'policy',
            label: t('trafficTopology.policyTitle'),
            children: (
              <Table
                rowKey={(r) => `${r.source}-${r.target}-${r.evidence}`}
                loading={loading}
                dataSource={policyEdges}
                pagination={{ pageSize: 10 }}
                locale={{ emptyText: t('trafficTopology.emptyPolicy') }}
                columns={[
                  { title: t('trafficTopology.columns.caller'), dataIndex: 'source', render: parseNodeLabel },
                  { title: '', width: 40, render: () => <ArrowRightOutlined /> },
                  { title: t('trafficTopology.columns.callee'), dataIndex: 'target', render: parseNodeLabel },
                  { title: t('trafficTopology.columns.port'), dataIndex: 'port' },
                  { title: t('trafficTopology.columns.evidence'), dataIndex: 'evidence', render: (v: string) => <Text code>{v}</Text> },
                ]}
              />
            ),
          },
        ]}
      />
      </Spin>

    </Card>
  );
};

export default ServiceTopology;
