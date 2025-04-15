import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Statistic, Progress, Spin, Descriptions, Space, Tabs } from 'antd';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer
} from 'recharts';
import { useTranslation } from 'react-i18next';
import { getPodMetrics } from '@/api/pod_metrics';
import type { PodMetrics, ContainerMetrics } from '@/api/pod_metrics';

interface PodMonitoringProps {
  clusterName: string;
  namespace: string;
  podName: string;
}

const PodMonitoring: React.FC<PodMonitoringProps> = ({ clusterName, namespace, podName }) => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState<boolean>(true);
  const [metrics, setMetrics] = useState<PodMetrics | null>(null);
  const [activeTab, setActiveTab] = useState<string>('overview');
  
  // 获取Pod指标数据
  const fetchPodMetrics = async () => {
    if (!clusterName || !namespace || !podName) return;
    
    try {
      setLoading(true);
      const response = await getPodMetrics(clusterName, namespace, podName);
      if (response.data.code === 0) {
        setMetrics(response.data.data.metrics);
      }
    } catch (error) {
      console.error('获取Pod指标数据失败:', error);
    } finally {
      setLoading(false);
    }
  };
  
  // 组件加载时获取Pod指标数据，并设置定时刷新
  useEffect(() => {
    fetchPodMetrics();
    
    // 设置定时刷新，每30秒刷新一次
    const intervalId = setInterval(() => {
      fetchPodMetrics();
    }, 30000);
    
    // 组件卸载时清除定时器
    return () => clearInterval(intervalId);
  }, [clusterName, namespace, podName]);
  
  // 格式化时间戳为小时:分钟
  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
    return `${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`;
  };
  
  // 渲染Pod总体概览
  const renderOverview = () => {
    if (!metrics) return null;
    
    return (
      <div>
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={12}>
            <Card title={t('podDetail.monitoring.cpuUsage')}>
              <Statistic
                value={metrics.cpuUsage.toFixed(2)}
                suffix="%"
                valueStyle={{ color: metrics.cpuUsage > 80 ? '#cf1322' : '#3f8600' }}
              />
              <Progress
                percent={Math.min(100, parseFloat(metrics.cpuUsage.toFixed(2)))}
                status={metrics.cpuUsage > 80 ? 'exception' : 'normal'}
                showInfo={false}
                strokeColor={{
                  '0%': '#108ee9',
                  '100%': metrics.cpuUsage > 80 ? '#ff4d4f' : '#87d068',
                }}
              />
              <Descriptions column={1} size="small" style={{ marginTop: 16 }}>
                <Descriptions.Item label={t('podDetail.monitoring.requests')}>{metrics.cpuRequests}</Descriptions.Item>
                <Descriptions.Item label={t('podDetail.monitoring.limits')}>{metrics.cpuLimits}</Descriptions.Item>
              </Descriptions>
            </Card>
          </Col>
          
          <Col xs={24} sm={12}>
            <Card title={t('podDetail.monitoring.memoryUsage')}>
              <Statistic
                value={metrics.memoryUsage.toFixed(2)}
                suffix="%"
                valueStyle={{ color: metrics.memoryUsage > 80 ? '#cf1322' : '#3f8600' }}
              />
              <Progress
                percent={Math.min(100, parseFloat(metrics.memoryUsage.toFixed(2)))}
                status={metrics.memoryUsage > 80 ? 'exception' : 'normal'}
                showInfo={false}
                strokeColor={{
                  '0%': '#108ee9',
                  '100%': metrics.memoryUsage > 80 ? '#ff4d4f' : '#87d068',
                }}
              />
              <Descriptions column={1} size="small" style={{ marginTop: 16 }}>
                <Descriptions.Item label={t('podDetail.monitoring.requests')}>{metrics.memoryRequests}</Descriptions.Item>
                <Descriptions.Item label={t('podDetail.monitoring.limits')}>{metrics.memoryLimits}</Descriptions.Item>
              </Descriptions>
            </Card>
          </Col>
        </Row>
        
        <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
          <Col xs={24} md={12}>
            <Card title={t('podDetail.monitoring.cpuHistory')}>
              <ResponsiveContainer width="100%" height={300}>
                <LineChart
                  data={metrics.historicalData.cpuUsage.map(item => ({
                    name: formatTime(item.timestamp),
                    value: item.value
                  }))}
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
                    name={t('podDetail.monitoring.cpuUsage')}
                    stroke="#8884d8"
                    activeDot={{ r: 8 }}
                  />
                </LineChart>
              </ResponsiveContainer>
            </Card>
          </Col>
          
          <Col xs={24} md={12}>
            <Card title={t('podDetail.monitoring.memoryHistory')}>
              <ResponsiveContainer width="100%" height={300}>
                <LineChart
                  data={metrics.historicalData.memoryUsage.map(item => ({
                    name: formatTime(item.timestamp),
                    value: item.value
                  }))}
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
                    name={t('podDetail.monitoring.memoryUsage')}
                    stroke="#82ca9d"
                    activeDot={{ r: 8 }}
                  />
                </LineChart>
              </ResponsiveContainer>
            </Card>
          </Col>
        </Row>
      </div>
    );
  };
  
  // 渲染容器指标
  const renderContainerMetrics = () => {
    if (!metrics || !metrics.containers || metrics.containers.length === 0) {
      return <div>{t('podDetail.monitoring.noContainerMetrics')}</div>;
    }
    
    return (
      <Space direction="vertical" style={{ width: '100%' }}>
        {metrics.containers.map(container => (
          <Card 
            key={container.name} 
            title={`${t('podDetail.container')}: ${container.name}`}
            style={{ marginBottom: 16 }}
          >
            <Row gutter={[16, 16]}>
              <Col xs={24} sm={12}>
                <Card title={t('podDetail.monitoring.cpuUsage')} size="small">
                  <Statistic
                    value={container.cpuUsage.toFixed(2)}
                    suffix="%"
                    valueStyle={{ color: container.cpuUsage > 80 ? '#cf1322' : '#3f8600' }}
                  />
                  <Progress
                    percent={Math.min(100, parseFloat(container.cpuUsage.toFixed(2)))}
                    status={container.cpuUsage > 80 ? 'exception' : 'normal'}
                    showInfo={false}
                  />
                  <Descriptions column={1} size="small" style={{ marginTop: 8 }}>
                    <Descriptions.Item label={t('podDetail.monitoring.requests')}>{container.cpuRequests}</Descriptions.Item>
                    <Descriptions.Item label={t('podDetail.monitoring.limits')}>{container.cpuLimits}</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
              
              <Col xs={24} sm={12}>
                <Card title={t('podDetail.monitoring.memoryUsage')} size="small">
                  <Statistic
                    value={container.memoryUsage.toFixed(2)}
                    suffix="%"
                    valueStyle={{ color: container.memoryUsage > 80 ? '#cf1322' : '#3f8600' }}
                  />
                  <Progress
                    percent={Math.min(100, parseFloat(container.memoryUsage.toFixed(2)))}
                    status={container.memoryUsage > 80 ? 'exception' : 'normal'}
                    showInfo={false}
                  />
                  <Descriptions column={1} size="small" style={{ marginTop: 8 }}>
                    <Descriptions.Item label={t('podDetail.monitoring.requests')}>{container.memoryRequests}</Descriptions.Item>
                    <Descriptions.Item label={t('podDetail.monitoring.limits')}>{container.memoryLimits}</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
            </Row>
          </Card>
        ))}
      </Space>
    );
  };
  
  if (loading && !metrics) {
    return (
      <div style={{ textAlign: 'center', padding: '20px' }}>
        <Spin size="large" />
      </div>
    );
  }
  
  if (!metrics) {
    return (
      <div style={{ textAlign: 'center', padding: '20px' }}>
        {t('podDetail.monitoring.noMetricsData')}
      </div>
    );
  }
  
  return (
    <div className="pod-monitoring">
      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        items={[
          {
            key: 'overview',
            label: t('podDetail.monitoring.overview'),
            children: renderOverview()
          },
          {
            key: 'containers',
            label: t('podDetail.monitoring.containers'),
            children: renderContainerMetrics()
          }
        ]}
      />
    </div>
  );
};

export default PodMonitoring;
