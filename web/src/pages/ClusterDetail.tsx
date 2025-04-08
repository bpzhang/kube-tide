import React, { useState, useEffect } from 'react';
import { Card, Descriptions, Space, Button, message, Spin, Table, Tag, Tabs, Progress, Row, Col, Statistic } from 'antd';
import { useParams, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  testClusterConnection,
  getClusterDetails,
  getClusterMetrics,
  getClusterEvents
} from '../api/cluster';
import type { ClusterDetail, ClusterMetrics } from '../api/cluster';
import K8sEvents from '../components/k8s/common/K8sEvents';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell
} from 'recharts';

// 定义颜色常量
const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884d8'];

// 自定义组件：集群事件列表
const ClusterEvents: React.FC<{ clusterName: string }> = ({ clusterName }) => {
  // 包装获取集群事件的函数，使其符合K8sEvents组件的fetchEvents参数格式
  const fetchClusterEvents = async (clusterName: string) => {
    return getClusterEvents(clusterName);
  };

  return (
    <K8sEvents 
      clusterName={clusterName}
      namespace="all"
      resourceName={clusterName}
      resourceKind="Cluster"
      fetchEvents={fetchClusterEvents}
    />
  );
};

const ClusterDetailPage: React.FC = () => {
  const { t } = useTranslation();
  const { clusterName } = useParams<{ clusterName: string }>();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [metricsLoading, setMetricsLoading] = useState(false);
  const [connectionStatus, setConnectionStatus] = useState<'unknown' | 'connected' | 'failed'>('unknown');
  const [clusterInfo, setClusterInfo] = useState<ClusterDetail | null>(null);
  const [metrics, setMetrics] = useState<ClusterMetrics | null>(null);
  const [activeTabKey, setActiveTabKey] = useState('overview');

  const fetchClusterDetails = async () => {
    if (!clusterName) return;

    try {
      setLoading(true);
      const response = await getClusterDetails(clusterName);
      if (response.data.code === 0) {
        setClusterInfo(response.data.data.cluster);
      } else {
        message.error(response.data.message || t('clusterDetail.fetchDetailFailed'));
      }
    } catch (err) {
      message.error(t('clusterDetail.fetchDetailFailed'));
    } finally {
      setLoading(false);
    }
  };

  const handleTestConnection = async () => {
    if (!clusterName) return;

    try {
      setLoading(true);
      const response = await testClusterConnection(clusterName);
      if (response.data.code === 0) {
        setConnectionStatus('connected');
        fetchClusterDetails();
      } else {
        setConnectionStatus('failed');
        message.error(response.data.message || t('clusterDetail.actions.testFailed'));
      }
    } catch (err) {
      setConnectionStatus('failed');
      message.error(t('clusterDetail.actions.testFailed'));
    } finally {
      setLoading(false);
    }
  };

  // 获取集群监控指标
  const fetchClusterMetrics = async () => {
    if (!clusterName || connectionStatus !== 'connected') return;

    try {
      setMetricsLoading(true);
      const response = await getClusterMetrics(clusterName);
      if (response.data.code === 0) {
        setMetrics(response.data.data.metrics);
      } else {
        message.error(response.data.message || t('clusterDetail.fetchMetricsFailed'));
      }
    } catch (err) {
      message.error(t('clusterDetail.fetchMetricsFailed'));
    } finally {
      setMetricsLoading(false);
    }
  };

  // 当集群连接状态变更时，获取监控数据
  useEffect(() => {
    if (connectionStatus === 'connected') {
      fetchClusterMetrics();

      // 设置定时刷新（每30秒）
      const timer = setInterval(fetchClusterMetrics, 30000);
      return () => clearInterval(timer);
    }
  }, [connectionStatus, clusterName]);

  useEffect(() => {
    if (clusterName) {
      handleTestConnection();
    }
  }, [clusterName]);

  if (!clusterName) {
    return <div>{t('clusterDetail.noClusterName')}</div>;
  }

  // 节点状态分布的Pie图表数据
  const getNodeStatusData = (metrics: ClusterMetrics | null) => {
    if (!metrics) return [];
    return [
      { name: t('clusterDetail.monitoring.metrics.healthyNodes'), value: metrics.nodeCounts?.ready || 0 },
      { name: t('clusterDetail.monitoring.metrics.unhealthyNodes'), value: metrics.nodeCounts?.notReady || 0 }
    ];
  };

  // 命名空间列表列定义
  const namespaceColumns = [
    {
      title: t('clusterDetail.namespaces.columns.name'),
      dataIndex: 'metadata',
      key: 'name',
      render: (metadata: any) => metadata.name,
    },
    {
      title: t('clusterDetail.namespaces.columns.status'),
      dataIndex: 'status',
      key: 'status',
      render: (status: any) => (
        <Tag color={status.phase === 'Active' ? 'green' : 'red'}>
          {status.phase}
        </Tag>
      ),
    },
    {
      title: t('clusterDetail.namespaces.columns.creationTime'),
      dataIndex: 'metadata',
      key: 'creationTimestamp',
      render: (metadata: any) => new Date(metadata.creationTimestamp).toLocaleString(),
    },
  ];

  // 格式化时间戳为小时:分钟
  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
    return `${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`;
  };

  // 选项卡内容变更
  const handleTabChange = (key: string) => {
    setActiveTabKey(key);
  };

  return (
    <Spin spinning={loading}>
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <Card
          title={t('clusterDetail.title')}
          extra={
            <Space>
              <Button onClick={() => navigate('/clusters')}>{t('clusterDetail.backToList')}</Button>
              <Button type="primary" onClick={handleTestConnection}>
                {t('clusterDetail.actions.test')}
              </Button>
            </Space>
          }
        >
          <Descriptions bordered column={2}>
            <Descriptions.Item label={t('clusterDetail.clusterName')} span={1}>
              {clusterInfo?.name}
            </Descriptions.Item>
            <Descriptions.Item label="Status" span={1}>
              {connectionStatus === 'connected' && <span style={{ color: '#52c41a' }}>{t('clusterDetail.connectionStatus.connected')}</span>}
              {connectionStatus === 'failed' && <span style={{ color: '#ff4d4f' }}>{t('clusterDetail.connectionStatus.failed')}</span>}
              {connectionStatus === 'unknown' && <span style={{ color: '#faad14' }}>{t('clusterDetail.connectionStatus.unknown')}</span>}
            </Descriptions.Item>
            <Descriptions.Item label={t('clusterDetail.basicInfo.version')}>
              {clusterInfo?.version}
            </Descriptions.Item>
            <Descriptions.Item label={t('clusterDetail.basicInfo.platform')}>
              {clusterInfo?.platform}
            </Descriptions.Item>
            <Descriptions.Item label={t('clusterDetail.basicInfo.nodeCount')}>
              {clusterInfo?.totalNodes}
            </Descriptions.Item>
            <Descriptions.Item label={t('clusterDetail.basicInfo.namespaceCount')}>
              {clusterInfo?.totalNamespaces}
            </Descriptions.Item>
            <Descriptions.Item label={t('clusterDetail.basicInfo.totalCPU')}>
              {clusterInfo?.totalCPU} Core
            </Descriptions.Item>
            <Descriptions.Item label={t('clusterDetail.basicInfo.totalMemory')}>
              {clusterInfo?.totalMemory}
            </Descriptions.Item>
            <Descriptions.Item label={t('clusterDetail.basicInfo.addType') || '添加方式'}>
              {clusterInfo?.addType === 'path' && (t('clusters.addTypeFile') || '通过文件路径')}
              {clusterInfo?.addType === 'content' && (t('clusters.addTypeContent') || '通过内容填写')}
              {(!clusterInfo?.addType || clusterInfo?.addType === 'unknown') && (t('clusters.addTypeUnknown') || '未知方式')}
            </Descriptions.Item>
          </Descriptions>
        </Card>

        {/* 监控仪表板 */}
        {connectionStatus === 'connected' && (
          <Card
            title={t('clusterDetail.monitoring.title')}
            loading={metricsLoading}
          >
            <Tabs
              activeKey={activeTabKey}
              onChange={handleTabChange}
              items={[
                {
                  key: 'overview',
                  label: t('clusterDetail.monitoring.overview'),
                  children: (
                    <div>
                      <Row gutter={[16, 16]} style={{ marginBottom: 20 }}>
                        <Col span={6}>
                          <Card>
                            <Statistic
                              title={t('clusterDetail.monitoring.metrics.cpu')}
                              value={metrics?.cpuUsage || 0}
                              suffix="%"
                              precision={1}
                            />
                            <Progress
                              percent={metrics?.cpuUsage || 0}
                              status={metrics?.cpuUsage && metrics.cpuUsage > 80 ? 'exception' : 'normal'}
                              showInfo={false}
                              strokeColor={{
                                '0%': '#108ee9',
                                '100%': metrics?.cpuUsage && metrics.cpuUsage > 80 ? '#ff4d4f' : '#87d068',
                              }}
                            />
                          </Card>
                        </Col>
                        <Col span={6}>
                          <Card>
                            <Statistic
                              title={t('clusterDetail.monitoring.memoryUsage')}
                              value={metrics?.memoryUsage || 0}
                              suffix="%"
                              precision={1}
                            />
                            <Progress
                              percent={metrics?.memoryUsage || 0}
                              status={metrics?.memoryUsage && metrics.memoryUsage > 80 ? 'exception' : 'normal'}
                              showInfo={false}
                              strokeColor={{
                                '0%': '#108ee9',
                                '100%': metrics?.memoryUsage && metrics.memoryUsage > 80 ? '#ff4d4f' : '#87d068',
                              }}
                            />
                          </Card>
                        </Col>
                        <Col span={6}>
                          <Card>
                            <Statistic
                              title={t('clusterDetail.monitoring.metrics.podCount')}
                              value={metrics?.podCount || 0}
                            />
                          </Card>
                        </Col>
                        <Col span={6}>
                          <Card>
                            <Statistic
                              title={t('clusterDetail.monitoring.metrics.deploymentAvailability')}
                              value={metrics?.deploymentReadiness?.available || 0}
                              suffix={`/${metrics?.deploymentReadiness?.total || 0}`}
                            />
                            {metrics?.deploymentReadiness?.total && (
                              <Progress
                                percent={(metrics.deploymentReadiness.available / metrics.deploymentReadiness.total) * 100}
                                showInfo={false}
                              />
                            )}
                          </Card>
                        </Col>
                      </Row>

                      <Row gutter={[16, 16]}>
                        <Col span={12}>
                          <Card title={t('clusterDetail.monitoring.metrics.cpuTrend')}>
                            <ResponsiveContainer width="100%" height={300}>
                              <LineChart
                                data={metrics?.historicalData?.cpuUsage?.map(item => ({
                                  name: formatTime(item.timestamp),
                                  value: item.value
                                })) || []}
                                margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
                              >
                                <CartesianGrid strokeDasharray="3 3" />
                                <XAxis dataKey="name" />
                                <YAxis unit="%" />
                                <Tooltip />
                                <Legend />
                                <Line
                                  type="monotone"
                                  dataKey="value"
                                  name={t('clusterDetail.monitoring.metrics.cpu')}
                                  stroke="#8884d8"
                                  activeDot={{ r: 8 }}
                                />
                              </LineChart>
                            </ResponsiveContainer>
                          </Card>
                        </Col>
                        <Col span={12}>
                          <Card title={t('clusterDetail.monitoring.metrics.memoryTrend')}>
                            <ResponsiveContainer width="100%" height={300}>
                              <LineChart
                                data={metrics?.historicalData?.memoryUsage?.map(item => ({
                                  name: formatTime(item.timestamp),
                                  value: item.value
                                })) || []}
                                margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
                              >
                                <CartesianGrid strokeDasharray="3 3" />
                                <XAxis dataKey="name" />
                                <YAxis unit="%" />
                                <Tooltip />
                                <Legend />
                                <Line
                                  type="monotone"
                                  dataKey="value"
                                  name={t('clusterDetail.monitoring.metrics.memory')}
                                  stroke="#82ca9d"
                                  activeDot={{ r: 8 }}
                                />
                              </LineChart>
                            </ResponsiveContainer>
                          </Card>
                        </Col>
                      </Row>
                    </div>
                  ),
                },
                {
                  key: 'resources',
                  label: t('clusterDetail.monitoring.resources'),
                  children: (
                    <div>
                      <Row gutter={[16, 16]}>
                        <Col span={12}>
                          <Card title={t('clusterDetail.monitoring.metrics.cpuAllocation')}>
                            <ResponsiveContainer width="100%" height={300}>
                              <BarChart
                                data={[
                                  { name: t('clusterDetail.monitoring.metrics.requested'), value: metrics?.cpuRequestsPercentage || 0 },
                                  { name: t('clusterDetail.monitoring.metrics.limited'), value: metrics?.cpuLimitsPercentage || 0 },
                                  { name: t('clusterDetail.monitoring.metrics.used'), value: metrics?.cpuUsage || 0 },
                                ]}
                                margin={{ top: 20, right: 30, left: 20, bottom: 5 }}
                              >
                                <CartesianGrid strokeDasharray="3 3" />
                                <XAxis dataKey="name" />
                                <YAxis unit="%" />
                                <Tooltip />
                                <Legend />
                                <Bar dataKey="value" name={t('common.percentage')} fill="#8884d8" />
                              </BarChart>
                            </ResponsiveContainer>
                          </Card>
                        </Col>
                        <Col span={12}>
                          <Card title={t('clusterDetail.monitoring.metrics.memoryAllocation')}>
                            <ResponsiveContainer width="100%" height={300}>
                              <BarChart
                                data={[
                                  { name: t('clusterDetail.monitoring.metrics.requested'), value: metrics?.memoryRequestsPercentage || 0 },
                                  { name: t('clusterDetail.monitoring.metrics.limited'), value: metrics?.memoryLimitsPercentage || 0 },
                                  { name: t('clusterDetail.monitoring.metrics.used'), value: metrics?.memoryUsage || 0 },
                                ]}
                                margin={{ top: 20, right: 30, left: 20, bottom: 5 }}
                              >
                                <CartesianGrid strokeDasharray="3 3" />
                                <XAxis dataKey="name" />
                                <YAxis unit="%" />
                                <Tooltip />
                                <Legend />
                                <Bar dataKey="value" name={t('common.percentage')} fill="#82ca9d" />
                              </BarChart>
                            </ResponsiveContainer>
                          </Card>
                        </Col>
                      </Row>
                      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
                        <Col span={12}>
                          <Card title={t('clusterDetail.monitoring.metrics.nodeStatus')}>
                            <ResponsiveContainer width="100%" height={300}>
                              <PieChart>
                                <Pie
                                  data={getNodeStatusData(metrics)}
                                  cx="50%"
                                  cy="50%"
                                  labelLine={true}
                                  outerRadius={80}
                                  fill="#8884d8"
                                  dataKey="value"
                                  nameKey="name"
                                  label={({ name, percent = 0 }) => `${name}: ${(percent * 100).toFixed(0)}%`}
                                >
                                  {getNodeStatusData(metrics).map((_entry, index) => (
                                    <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                                  ))}
                                </Pie>
                                <Tooltip formatter={(value) => [`${value}${t('nodes.nodeCount')}`, t('common.count')]} />
                              </PieChart>
                            </ResponsiveContainer>
                          </Card>
                        </Col>
                        <Col span={12}>
                          <Card title={t('clusterDetail.monitoring.metrics.podCountTrend')}>
                            <ResponsiveContainer width="100%" height={300}>
                              <LineChart
                                data={metrics?.historicalData?.podCount?.map(item => ({
                                  name: formatTime(item.timestamp),
                                  value: item.value
                                })) || []}
                                margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
                              >
                                <CartesianGrid strokeDasharray="3 3" />
                                <XAxis dataKey="name" />
                                <YAxis />
                                <Tooltip />
                                <Legend />
                                <Line
                                  type="monotone"
                                  dataKey="value"
                                  name={t('clusterDetail.monitoring.metrics.podCount')}
                                  stroke="#FF8042"
                                  activeDot={{ r: 8 }}
                                />
                              </LineChart>
                            </ResponsiveContainer>
                          </Card>
                        </Col>
                      </Row>
                    </div>
                  ),
                },
              ]}
            />
          </Card>
        )}

        {/* 命名空间列表卡片 */}
        <Card title={t('clusterDetail.namespaces.title')}>
          <Table
            dataSource={clusterInfo?.namespaces || []}
            columns={namespaceColumns}
            rowKey={(record) => record.metadata.name}
            pagination={{ pageSize: 10 }}
          />
        </Card>

        {/* 集群事件 */}
        {connectionStatus === 'connected' && (
          <Card title={t('clusterDetail.events.title')}>
            <ClusterEvents clusterName={clusterName} />
          </Card>
        )}
      </Space>
    </Spin>
  );
};

export default ClusterDetailPage;
