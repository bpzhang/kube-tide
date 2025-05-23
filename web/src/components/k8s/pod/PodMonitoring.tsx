import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Statistic, Progress, Spin, Descriptions, Space, Tabs, Radio, DatePicker } from 'antd';
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
import dayjs from 'dayjs';

interface PodMonitoringProps {
  clusterName: string;
  namespace: string;
  podName: string;
}

const PodMonitoring: React.FC<PodMonitoringProps> = ({ clusterName, namespace, podName }) => {
  const { t } = useTranslation();
  const { RangePicker } = DatePicker;
  const [loading, setLoading] = useState<boolean>(true);
  const [metrics, setMetrics] = useState<PodMetrics | null>(null);
  const [activeTab, setActiveTab] = useState<string>('overview');
  const [timeRange, setTimeRange] = useState<string>('1h'); // 默认显示最近1小时的数据
  const [customTimeRange, setCustomTimeRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(null);
  
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
  
  // 格式化时间戳，根据选择的时间范围显示不同级别的细节
  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
    
    // 根据时间范围选择合适的格式
    if (timeRange === '1h' || timeRange === '6h') {
      // 短时间范围仅显示时:分
      return `${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`;
    } else if (timeRange === '24h') {
      // 24小时显示日期和时间
      return `${(date.getMonth()+1).toString().padStart(2, '0')}-${date.getDate().toString().padStart(2, '0')} ${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`;
    } else {
      // 更长时间范围显示年-月-日
      return `${date.getFullYear()}-${(date.getMonth()+1).toString().padStart(2, '0')}-${date.getDate().toString().padStart(2, '0')}`;
    }
  };
  
  // 根据选择的时间范围过滤数据，并确保数据按时间顺序排序
  const filterDataByTimeRange = (data: { timestamp: string; value: number }[]) => {
    if (!data || data.length === 0) return [];
    
    const now = new Date();
    let startTime: Date;
    
    // 如果设置了自定义时间范围
    if (customTimeRange && customTimeRange[0] && customTimeRange[1]) {
      const filteredData = data.filter(item => {
        const itemTime = new Date(item.timestamp);
        return itemTime >= customTimeRange[0].toDate() && itemTime <= customTimeRange[1].toDate();
      });
      
      // 按时间戳排序
      return filteredData.sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());
    }
    
    // 根据预设时间范围过滤
    switch (timeRange) {
      case '1h':
        startTime = new Date(now.getTime() - 60 * 60 * 1000);
        break;
      case '6h':
        startTime = new Date(now.getTime() - 6 * 60 * 60 * 1000);
        break;
      case '24h':
        startTime = new Date(now.getTime() - 24 * 60 * 60 * 1000);
        break;
      case '7d':
        startTime = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
        break;
      case '30d':
        startTime = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
        break;
      default:
        startTime = new Date(now.getTime() - 60 * 60 * 1000); // 默认1小时
    }
    
    const filteredData = data.filter(item => new Date(item.timestamp) >= startTime);
    
    // 按时间戳排序
    return filteredData.sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());
  };
  
  // 处理时间范围变更
  const handleTimeRangeChange = (e: any) => {
    setTimeRange(e.target.value);
    setCustomTimeRange(null); // 清除自定义时间范围
  };
  
  // 处理自定义时间范围变更
  const handleCustomRangeChange = (dates: any) => {
    if (dates && dates.length === 2) {
      setCustomTimeRange([dates[0], dates[1]]);
      setTimeRange('custom'); // 设置为自定义模式
    } else {
      setCustomTimeRange(null);
    }
  };
  
  // 渲染Pod总体概览
  const renderOverview = () => {
    if (!metrics) return null;
    
    return (
      <div>
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={8}>
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
          
          <Col xs={24} sm={8}>
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
          
          <Col xs={24} sm={8}>
            <Card title={t('podDetail.monitoring.diskUsage')}>
              {metrics.diskUsage !== undefined ? (
                <>
                  <Statistic
                    value={metrics.diskUsage.toFixed(2)}
                    suffix="%"
                    valueStyle={{ color: metrics.diskUsage > 80 ? '#cf1322' : '#3f8600' }}
                  />
                  <Progress
                    percent={Math.min(100, parseFloat(metrics.diskUsage.toFixed(2)))}
                    status={metrics.diskUsage > 80 ? 'exception' : 'normal'}
                    showInfo={false}
                    strokeColor={{
                      '0%': '#108ee9',
                      '100%': metrics.diskUsage > 80 ? '#ff4d4f' : '#87d068',
                    }}
                  />
                </>
              ) : (
                <>
                  <Statistic
                    value={0}
                    suffix="%"
                    valueStyle={{ color: '#3f8600' }}
                  />
                  <Progress
                    percent={0}
                    status="normal"
                    showInfo={false}
                    strokeColor={{
                      '0%': '#108ee9',
                      '100%': '#87d068',
                    }}
                  />
                </>
              )}
              <Descriptions column={1} size="small" style={{ marginTop: 16 }}>
                <Descriptions.Item label={t('podDetail.monitoring.requests')}>{metrics.diskRequests || 'N/A'}</Descriptions.Item>
                <Descriptions.Item label={t('podDetail.monitoring.limits')}>{metrics.diskLimits || 'N/A'}</Descriptions.Item>
              </Descriptions>
            </Card>
          </Col>
        </Row>
        
        <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
          <Col xs={24} md={8}>
            <Card title={t('podDetail.monitoring.cpuHistory')}>
              <ResponsiveContainer width="100%" height={300}>
                <LineChart
                  data={filterDataByTimeRange(metrics.historicalData.cpuUsage).map(item => ({
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
          
          <Col xs={24} md={8}>
            <Card title={t('podDetail.monitoring.memoryHistory')}>
              <ResponsiveContainer width="100%" height={300}>
                <LineChart
                  data={filterDataByTimeRange(metrics.historicalData.memoryUsage).map(item => ({
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
          
          <Col xs={24} md={8}>
            <Card title={t('podDetail.monitoring.diskHistory')}>
              <ResponsiveContainer width="100%" height={300}>
                <LineChart
                  data={filterDataByTimeRange(metrics.historicalData.diskUsage || []).map(item => ({
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
                    name={t('podDetail.monitoring.diskUsage')}
                    stroke="#f5a442"
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
              <Col xs={24} sm={8}>
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
              
              <Col xs={24} sm={8}>
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
              
              <Col xs={24} sm={8}>
                <Card title={t('podDetail.monitoring.diskUsage')} size="small">
                  <Statistic
                    value={container.diskUsage.toFixed(2)}
                    suffix="%"
                    valueStyle={{ color: container.diskUsage > 80 ? '#cf1322' : '#3f8600' }}
                  />
                  <Progress
                    percent={Math.min(100, parseFloat(container.diskUsage.toFixed(2)))}
                    status={container.diskUsage > 80 ? 'exception' : 'normal'}
                    showInfo={false}
                  />
                  <Descriptions column={1} size="small" style={{ marginTop: 8 }}>
                    <Descriptions.Item label={t('podDetail.monitoring.requests')}>{container.diskRequests}</Descriptions.Item>
                    <Descriptions.Item label={t('podDetail.monitoring.limits')}>{container.diskLimits}</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
            </Row>
            
            {/* 容器历史数据图表 */}
            {container.historicalData && (
              <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
                <Col xs={24} md={8}>
                  <Card title={t('podDetail.monitoring.cpuHistory')} size="small">
                    <ResponsiveContainer width="100%" height={200}>
                      <LineChart
                        data={filterDataByTimeRange(container.historicalData.cpuUsage || []).map(item => ({
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
                
                <Col xs={24} md={8}>
                  <Card title={t('podDetail.monitoring.memoryHistory')} size="small">
                    <ResponsiveContainer width="100%" height={200}>
                      <LineChart
                        data={filterDataByTimeRange(container.historicalData.memoryUsage || []).map(item => ({
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
                
                <Col xs={24} md={8}>
                  <Card title={t('podDetail.monitoring.diskHistory')} size="small">
                    <ResponsiveContainer width="100%" height={200}>
                      <LineChart
                        data={filterDataByTimeRange(container.historicalData.diskUsage || []).map(item => ({
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
                          name={t('podDetail.monitoring.diskUsage')}
                          stroke="#f5a442"
                          activeDot={{ r: 8 }}
                        />
                      </LineChart>
                    </ResponsiveContainer>
                  </Card>
                </Col>
              </Row>
            )}
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
      <Radio.Group value={timeRange} onChange={handleTimeRangeChange} style={{ marginBottom: 16 }}>
        <Radio.Button value="1h">{t('timeRange.1h')}</Radio.Button>
        <Radio.Button value="6h">{t('timeRange.6h')}</Radio.Button>
        <Radio.Button value="24h">{t('timeRange.24h')}</Radio.Button>
        <Radio.Button value="7d">{t('timeRange.7d')}</Radio.Button>
        <Radio.Button value="30d">{t('timeRange.30d')}</Radio.Button>
        <Radio.Button value="custom">{t('timeRange.custom')}</Radio.Button>
      </Radio.Group>
      {timeRange === 'custom' && (
        <RangePicker
          showTime
          value={customTimeRange}
          onChange={handleCustomRangeChange}
          style={{ marginBottom: 16 }}
        />
      )}
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
