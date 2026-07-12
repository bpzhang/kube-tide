import React, { useEffect, useMemo, useState } from 'react';
import {
  Card,
  Table,
  Tag,
  Space,
  Typography,
  Alert,
  Statistic,
  Row,
  Col,
  message,
  Tabs,
  Button,
  Tooltip,
} from 'antd';
import { ArrowRightOutlined, ReloadOutlined, LinkOutlined, EyeOutlined, QuestionCircleOutlined } from '@ant-design/icons';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useClusterNamespace } from '@/hooks/useClusterNamespace';
import ClusterNamespaceToolbar from '@/components/k8s/common/ClusterNamespaceToolbar';
import TrafficTopologyGraph from '@/components/observability/TrafficTopologyGraph';
import {
  getTrafficTopology,
  parseNodeLabel,
  CallChainPath,
  CallFlowStat,
  TrafficTopology,
} from '@/api/traffic_topology';

const { Text } = Typography;

const formatRate = (v?: number) => (v && v > 0 ? `${v.toFixed(2)}/s` : '—');

const CallChain: React.FC = () => {
  const { t } = useTranslation();
  const { selectedCluster, setSelectedCluster, clusters, namespace, setNamespace, clustersLoading } =
    useClusterNamespace(t);
  const [topology, setTopology] = useState<TrafficTopology | null>(null);
  const [loading, setLoading] = useState(false);

  const fetchData = async () => {
    if (!selectedCluster) return;
    setLoading(true);
    try {
      const response = await getTrafficTopology(selectedCluster, namespace);
      if (response.data.code === 0) {
        setTopology(response.data.data);
      } else {
        message.error(response.data.message || t('callChain.fetchFailed'));
        setTopology(null);
      }
    } catch {
      message.error(t('callChain.fetchFailed'));
      setTopology(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (selectedCluster) fetchData();
  }, [selectedCluster, namespace]);

  const stats = useMemo(() => {
    const flows = topology?.callFlows || [];
    const chains = topology?.callChains || [];
    const observedFlows = flows.filter((f) => f.observed);
    const observedChains = chains.filter((c) => c.observed);
    const totalQPS = observedFlows.reduce((sum, f) => sum + (f.flowsPerSec || 0), 0);
    const totalDrops = observedFlows.reduce((sum, f) => sum + (f.drops || 0), 0);
    return {
      flows: flows.length,
      observedFlows: observedFlows.length,
      chains: chains.length,
      observedChains: observedChains.length,
      totalQPS,
      totalDrops,
    };
  }, [topology]);

  const flowColumns = [
    {
      title: t('callChain.columns.source'),
      key: 'source',
      render: (_: unknown, row: CallFlowStat) => (
        <Tag color="purple">
          {row.sourceNamespace}/{row.source}
          {row.sourceType && <Text type="secondary"> ({row.sourceType})</Text>}
        </Tag>
      ),
    },
    {
      title: '',
      key: 'arrow',
      width: 40,
      render: () => <ArrowRightOutlined />,
    },
    {
      title: t('callChain.columns.target'),
      key: 'target',
      render: (_: unknown, row: CallFlowStat) => (
        <Tag color="blue">
          {row.targetNamespace}/{row.target}
          {row.targetType && <Text type="secondary"> ({row.targetType})</Text>}
        </Tag>
      ),
    },
    {
      title: t('callChain.columns.flows'),
      dataIndex: 'flowsPerSec',
      key: 'flows',
      render: (v: number) => <Text strong>{formatRate(v)}</Text>,
      sorter: (a: CallFlowStat, b: CallFlowStat) => a.flowsPerSec - b.flowsPerSec,
    },
    {
      title: t('callChain.columns.drops'),
      dataIndex: 'drops',
      key: 'drops',
      render: (v: number) => (v > 0 ? <Tag color="red">{formatRate(v)}</Tag> : '—'),
    },
    {
      title: t('callChain.columns.port'),
      key: 'port',
      render: (_: unknown, row: CallFlowStat) => row.port || row.protocol || '—',
    },
    {
      title: t('callChain.columns.type'),
      key: 'observed',
      width: 100,
      render: (_: unknown, row: CallFlowStat) =>
        row.observed ? (
          <Tag color="green" icon={<EyeOutlined />}>
            {t('callChain.observed')}
          </Tag>
        ) : (
          <Tag icon={<QuestionCircleOutlined />}>{t('callChain.inferred')}</Tag>
        ),
    },
  ];

  const renderChain = (chain: CallChainPath) => (
    <Space wrap key={chain.id} style={{ marginBottom: 8 }}>
      {chain.entry && (
        <Tag color="geekblue">{chain.entry}</Tag>
      )}
      {chain.hops.map((hop, idx) => (
        <React.Fragment key={hop.nodeId}>
          {idx > 0 && <ArrowRightOutlined style={{ color: '#8c8c8c' }} />}
          <Tag color={hop.type === 'ingress' ? 'purple' : hop.type === 'service' ? 'blue' : 'green'}>
            {hop.type}/{hop.namespace}/{hop.name}
          </Tag>
        </React.Fragment>
      ))}
      {chain.flowsPerSec ? (
        <Tag color="orange">{formatRate(chain.flowsPerSec)}</Tag>
      ) : null}
      {chain.observed && <Tag color="green">{t('callChain.observed')}</Tag>}
    </Space>
  );

  const callEdges = (topology?.edges || []).filter((e) => e.edgeType === 'calls');

  return (
    <Card title={t('callChain.title')}>
      <ClusterNamespaceToolbar
        clusters={clusters}
        selectedCluster={selectedCluster}
        onClusterChange={setSelectedCluster}
        namespace={namespace}
        onNamespaceChange={setNamespace}
        loading={clustersLoading}
        extra={
          <Space>
            <Link to="/observability/service-topology">
              <Button icon={<LinkOutlined />}>{t('callChain.viewTopology')}</Button>
            </Link>
            <Button icon={<ReloadOutlined />} onClick={fetchData} loading={loading}>
              {t('common.refresh')}
            </Button>
          </Space>
        }
      />

      <Alert
        type="info"
        showIcon
        style={{ margin: '16px 0' }}
        message={t('callChain.hintTitle')}
        description={t('callChain.hint')}
      />

      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={4}>
          <Statistic title={t('callChain.stats.flows')} value={stats.flows} />
        </Col>
        <Col span={4}>
          <Statistic
            title={t('callChain.stats.observedFlows')}
            value={stats.observedFlows}
            valueStyle={{ color: stats.observedFlows > 0 ? '#52c41a' : undefined }}
          />
        </Col>
        <Col span={4}>
          <Statistic title={t('callChain.stats.chains')} value={stats.chains} />
        </Col>
        <Col span={4}>
          <Statistic title={t('callChain.stats.totalQPS')} value={stats.totalQPS.toFixed(2)} />
        </Col>
        <Col span={4}>
          <Statistic
            title={t('callChain.stats.drops')}
            value={stats.totalDrops.toFixed(2)}
            valueStyle={{ color: stats.totalDrops > 0 ? '#ff4d4f' : undefined }}
          />
        </Col>
        <Col span={4}>
          <Tooltip title={t('callChain.stats.observedChainsTip')}>
            <Statistic title={t('callChain.stats.observedChains')} value={stats.observedChains} />
          </Tooltip>
        </Col>
      </Row>

      <Tabs
        items={[
          {
            key: 'chains',
            label: t('callChain.tabs.chains'),
            children: (
              <Card type="inner" loading={loading}>
                {(topology?.callChains || []).length > 0 ? (
                  (topology?.callChains || []).map(renderChain)
                ) : (
                  <Text type="secondary">{t('callChain.emptyChains')}</Text>
                )}
              </Card>
            ),
          },
          {
            key: 'flows',
            label: t('callChain.tabs.flows'),
            children: (
              <Table
                rowKey={(r, idx) => `${r.source}-${r.target}-${idx}`}
                loading={loading}
                dataSource={topology?.callFlows || []}
                columns={flowColumns}
                pagination={{ pageSize: 15 }}
                locale={{ emptyText: t('callChain.emptyFlows') }}
              />
            ),
          },
          {
            key: 'calls',
            label: t('callChain.tabs.inferredCalls'),
            children: (
              <Table
                rowKey={(r) => `${r.source}-${r.target}-${r.evidence}`}
                loading={loading}
                dataSource={callEdges}
                pagination={{ pageSize: 15 }}
                locale={{ emptyText: t('trafficTopology.emptyCalls') }}
                columns={[
                  {
                    title: t('trafficTopology.columns.caller'),
                    dataIndex: 'source',
                    render: parseNodeLabel,
                  },
                  { title: '', width: 40, render: () => <ArrowRightOutlined /> },
                  {
                    title: t('trafficTopology.columns.callee'),
                    dataIndex: 'target',
                    render: parseNodeLabel,
                  },
                  {
                    title: t('callChain.columns.flows'),
                    key: 'flows',
                    render: (_: unknown, row: (typeof callEdges)[0]) =>
                      row.metrics?.flowsPerSec ? formatRate(row.metrics.flowsPerSec) : '—',
                  },
                  {
                    title: t('callChain.columns.type'),
                    key: 'type',
                    render: (_: unknown, row: (typeof callEdges)[0]) =>
                      row.metrics?.observed ? (
                        <Tag color="green">{t('callChain.observed')}</Tag>
                      ) : (
                        <Tag color="orange">{t('callChain.inferred')}</Tag>
                      ),
                  },
                  {
                    title: t('trafficTopology.columns.evidence'),
                    dataIndex: 'evidence',
                    render: (v: string) => <Text code>{v}</Text>,
                  },
                ]}
              />
            ),
          },
          {
            key: 'graph',
            label: t('callChain.tabs.graph'),
            children: (
              <Card type="inner" title={t('callChain.graphTitle')}>
                <TrafficTopologyGraph topology={topology} loading={loading} highlightCalls />
              </Card>
            ),
          },
        ]}
      />
    </Card>
  );
};

export default CallChain;
