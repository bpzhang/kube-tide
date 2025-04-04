import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Statistic, Spin, Select, Button, Space, Progress, Tabs } from 'antd';
import {
  LineChart,
  Line,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer
} from 'recharts';
import { useTranslation } from 'react-i18next';
import { getClusterList, getClusterMetrics } from '../api/cluster';

// 定义颜色常量
const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884d8'];

const Dashboard: React.FC = () => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [clusterList, setClusterList] = useState<string[]>([]);
  const [selectedCluster, setSelectedCluster] = useState<string>('');
  const [metrics, setMetrics] = useState<any>(null);
  const [activeTabKey, setActiveTabKey] = useState('overview');

  // 获取集群列表
  const fetchClusterList = async () => {
    try {
      setLoading(true);
      const response = await getClusterList();
      if (response.data.code === 0) {
        const clusters = response.data.data.clusters;
        setClusterList(clusters);

        // 如果有集群，默认选择第一个
        if (clusters.length > 0 && !selectedCluster) {
          setSelectedCluster(clusters[0]);
        }
      }
    } catch (err) {
      console.error('获取集群列表失败', err);
    } finally {
      setLoading(false);
    }
  };

  // 获取集群监控指标
  const fetchClusterMetrics = async (clusterName: string) => {
    if (!clusterName) return;

    try {
      setLoading(true);
      const response = await getClusterMetrics(clusterName);
      if (response.data.code === 0) {
        setMetrics(response.data.data.metrics);
      }
    } catch (err) {
      console.error('获取集群监控指标失败', err);
    } finally {
      setLoading(false);
    }
  };

  // 集群切换事件处理
  const handleClusterChange = (value: string) => {
    setSelectedCluster(value);
  };

  // 刷新数据
  const handleRefresh = () => {
    if (selectedCluster) {
      fetchClusterMetrics(selectedCluster);
    }
  };

  // 格式化时间戳为小时:分钟
  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
    return `${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`;
  };

  // 选项卡内容变更
  const handleTabChange = (key: string) => {
    setActiveTabKey(key);
  };

  // 组件加载时获取集群列表
  useEffect(() => {
    fetchClusterList();
  }, []);

  // 当选中集群变化时获取监控数据
  useEffect(() => {
    if (selectedCluster) {
      fetchClusterMetrics(selectedCluster);

      // 设置定时刷新（每30秒）
      const timer = setInterval(() => {
        fetchClusterMetrics(selectedCluster);
      }, 30000);

      return () => clearInterval(timer);
    }
  }, [selectedCluster]);

  return (
    <div>
      <Card
        title={t('dashboard.clusterMonitoring')}
        extra={
          <Space>
            <Select
              placeholder={t('dashboard.selectClusterPlaceholder')}
              style={{ width: 200 }}
              value={selectedCluster}
              onChange={handleClusterChange}
              options={clusterList.map(cluster => ({ value: cluster, label: cluster }))}
            />
            <Button type="primary" onClick={handleRefresh}>
              {t('common.refresh')}
            </Button>
          </Space>
        }
      >
        {!selectedCluster ? (
          <div style={{ textAlign: 'center', padding: '20px' }}>
            {t('dashboard.selectClusterView')}
          </div>
        ) : loading ? (
          <div style={{ textAlign: 'center', padding: '20px' }}>
            <Spin spinning={true}>
              <div style={{ padding: '30px', textAlign: 'center' }}>{t('dashboard.loadingData')}</div>
            </Spin>
          </div>
        ) : !metrics ? (
          <div style={{ textAlign: 'center', padding: '20px' }}>
            {t('common.noData')}
          </div>
        ) : (
          <Tabs
            activeKey={activeTabKey}
            onChange={handleTabChange}
            items={[
              {
                key: 'overview',
                label: t('dashboard.overview'),
                children: (
                  <div>
                    <Row gutter={[16, 16]} style={{ marginBottom: 20 }}>
                      <Col span={6}>
                        <Card>
                          <Statistic
                            title={t('dashboard.cpuUsage')}
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
                            title={t('dashboard.memoryUsage')}
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
                            title={t('dashboard.podCount')}
                            value={metrics?.podCount || 0}
                          />
                        </Card>
                      </Col>
                      <Col span={6}>
                        <Card>
                          <Statistic
                            title={t('dashboard.deploymentAvailability')}
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
                        <Card title={t('dashboard.cpuUsageHistory')}>
                          <ResponsiveContainer width="100%" height={300}>
                            <LineChart
                              data={metrics?.historicalData?.cpuUsage?.map((item: any) => ({
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
                                name={t('dashboard.cpuUsage')}
                                stroke="#8884d8"
                                activeDot={{ r: 8 }}
                              />
                            </LineChart>
                          </ResponsiveContainer>
                        </Card>
                      </Col>
                      <Col span={12}>
                        <Card title={t('dashboard.memoryUsageHistory')}>
                          <ResponsiveContainer width="100%" height={300}>
                            <LineChart
                              data={metrics?.historicalData?.memoryUsage?.map((item: any) => ({
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
                                name={t('dashboard.memoryUsage')}
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
                label: t('dashboard.resources'),
                children: (
                  <div>
                    <Row gutter={[16, 16]}>
                      <Col span={12}>
                        <Card title={t('dashboard.cpuResourceAllocation')}>
                          <ResponsiveContainer width="100%" height={300}>
                            <BarChart
                              data={[
                                { name: t('dashboard.requested'), value: metrics?.cpuRequestsPercentage || 0 },
                                { name: t('dashboard.limited'), value: metrics?.cpuLimitsPercentage || 0 },
                                { name: t('dashboard.used'), value: metrics?.cpuUsage || 0 },
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
                        <Card title={t('dashboard.memoryResourceAllocation')}>
                          <ResponsiveContainer width="100%" height={300}>
                            <BarChart
                              data={[
                                { name: t('dashboard.requested'), value: metrics?.memoryRequestsPercentage || 0 },
                                { name: t('dashboard.limited'), value: metrics?.memoryLimitsPercentage || 0 },
                                { name: t('dashboard.used'), value: metrics?.memoryUsage || 0 },
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
                        <Card title={t('dashboard.nodeStatusDistribution')}>
                          <ResponsiveContainer width="100%" height={300}>
                            <PieChart>
                              <Pie
                                data={[
                                  { name: t('dashboard.healthyNodes'), value: metrics?.nodeCounts?.ready || 0 },
                                  { name: t('dashboard.unhealthyNodes'), value: metrics?.nodeCounts?.notReady || 0 },
                                ]}
                                cx="50%"
                                cy="50%"
                                labelLine={true}
                                outerRadius={80}
                                fill="#8884d8"
                                dataKey="value"
                                nameKey="name"
                                label={({ name, percent = 0 }: any) => `${name}: ${(percent * 100).toFixed(0)}%`}
                              >
                                {[
                                  { name: t('dashboard.healthyNodes'), value: metrics?.nodeCounts?.ready || 0 },
                                  { name: t('dashboard.unhealthyNodes'), value: metrics?.nodeCounts?.notReady || 0 },
                                ].map((_entry, index) => (
                                  <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                                ))}
                              </Pie>
                              <Tooltip formatter={(value) => [`${value}${t('dashboard.nodeCount')}`, t('common.details')]} />
                            </PieChart>
                          </ResponsiveContainer>
                        </Card>
                      </Col>
                      <Col span={12}>
                        <Card title={t('dashboard.podCountHistory')}>
                          <ResponsiveContainer width="100%" height={300}>
                            <LineChart
                              data={metrics?.historicalData?.podCount?.map((item: any) => ({
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
                                name={t('dashboard.podCount')}
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
        )}
      </Card>
    </div>
  );
};

export default Dashboard;
