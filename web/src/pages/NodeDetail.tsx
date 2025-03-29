import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Descriptions, Row, Col, Table, Tag, Button, Spin, Space, Progress, Tabs, Statistic } from 'antd';
import { ArrowLeftOutlined, CloudServerOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
import { getNodeDetails, getNodeMetrics, getNodePods, getNodeTaints, getNodeLabels } from '../api/node';

const { TabPane } = Tabs;

// 格式化内存大小
const formatMemorySize = (size: string): string => {
  if (!size || size === '0') return '0 GB';
  const value = parseInt(size.replace(/[^0-9]/g, ''));
  if (size.includes('Ki')) return `${(value / (1024 * 1024)).toFixed(2)} GB`;
  if (size.includes('Mi')) return `${(value / 1024).toFixed(2)} GB`;
  if (size.includes('Gi')) return `${value} GB`;
  return `${value} GB`;
};

// 格式化CPU
const formatCPU = (cpu: string): string => {
  if (!cpu || cpu === '0') return '0 Core';
  const value = parseInt(cpu.replace(/[^0-9]/g, ''));
  if (cpu.endsWith('m')) return `${(value / 1000).toFixed(2)} Core`;
  return `${value} Core`;
};

// 计算CPU使用百分比
const calculateCpuPercentage = (usage: string, capacity: string): number => {
  const usageValue = parseInt(usage.replace(/[^0-9]/g, ''));
  const capacityValue = parseInt(capacity.replace(/[^0-9]/g, ''));
  
  if (!capacityValue) return 0;
  
  // 如果是millicores (带m的)
  if (usage.endsWith('m') && !capacity.endsWith('m')) {
    return (usageValue / (capacityValue * 1000)) * 100;
  } else if (!usage.endsWith('m') && capacity.endsWith('m')) {
    return (usageValue * 1000 / capacityValue) * 100;
  } else {
    return (usageValue / capacityValue) * 100;
  }
};

// 计算内存使用百分比
const calculateMemoryPercentage = (usage: string, capacity: string): number => {
  let usageValue = parseInt(usage.replace(/[^0-9]/g, ''));
  let capacityValue = parseInt(capacity.replace(/[^0-9]/g, ''));
  
  // 转换为相同单位
  if (usage.includes('Mi') && capacity.includes('Ki')) {
    capacityValue = capacityValue / 1024;
  } else if (usage.includes('Ki') && capacity.includes('Mi')) {
    usageValue = usageValue / 1024;
  } else if (usage.includes('Gi') && capacity.includes('Mi')) {
    usageValue = usageValue * 1024;
  } else if (usage.includes('Mi') && capacity.includes('Gi')) {
    capacityValue = capacityValue * 1024;
  }
  
  if (!capacityValue) return 0;
  
  return (usageValue / capacityValue) * 100;
};

const NodeDetail: React.FC = () => {
  const { clusterName, nodeName } = useParams<{ clusterName: string; nodeName: string }>();
  const navigate = useNavigate();
  
  const [loading, setLoading] = useState<boolean>(true);
  const [node, setNode] = useState<any>(null);
  const [metrics, setMetrics] = useState<any>(null);
  const [pods, setPods] = useState<any[]>([]);
  const [taints, setTaints] = useState<any[]>([]);
  const [labels, setLabels] = useState<{[key: string]: string}>({});
  
  useEffect(() => {
    if (!clusterName || !nodeName) {
      navigate('/nodes');
      return;
    }
    
    // 获取节点详情
    const fetchNodeDetails = async () => {
      setLoading(true);
      try {
        // 并行请求所有数据
        const [nodeResponse, metricsResponse, podsResponse, taintsResponse, labelsResponse] = await Promise.all([
          getNodeDetails(clusterName, nodeName),
          getNodeMetrics(clusterName, nodeName),
          getNodePods(clusterName, nodeName),
          getNodeTaints(clusterName, nodeName),
          getNodeLabels(clusterName, nodeName)
        ]);
        
        // 处理节点详情
        if (nodeResponse.data.code === 0 && nodeResponse.data.data.node) {
          setNode(nodeResponse.data.data.node);
        }
        
        // 处理节点指标
        if (metricsResponse.data.code === 0) {
          setMetrics(metricsResponse.data.data.metrics);
        }
        
        // 处理Pod列表
        if (podsResponse.data.code === 0) {
          setPods(podsResponse.data.data.pods || []);
        }
        
        // 处理污点
        if (taintsResponse.data.code === 0) {
          setTaints(taintsResponse.data.data.taints || []);
        }
        
        // 处理标签
        if (labelsResponse.data.code === 0) {
          setLabels(labelsResponse.data.data.labels || {});
        }
      } catch (err) {
        console.error('获取节点详情失败:', err);
      } finally {
        setLoading(false);
      }
    };
    
    fetchNodeDetails();
  }, [clusterName, nodeName, navigate]);
  
  // Pod状态对应的颜色
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'Running':
        return 'green';
      case 'Pending':
        return 'gold';
      case 'Succeeded':
        return 'blue';
      case 'Failed':
        return 'red';
      default:
        return 'default';
    }
  };
  
  // Pod列表列定义
  const podColumns = [
    {
      title: '名称',
      dataIndex: 'metadata',
      key: 'name',
      render: (metadata: any) => metadata?.name || '-',
    },
    {
      title: '命名空间',
      dataIndex: 'metadata',
      key: 'namespace',
      render: (metadata: any) => metadata?.namespace || 'default',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'phase',
      render: (status: any) => (
        <Tag color={getStatusColor(status?.phase)}>{status?.phase || 'Unknown'}</Tag>
      ),
    },
    {
      title: 'Pod IP',
      dataIndex: 'status',
      key: 'podIP',
      render: (status: any) => status?.podIP || '-',
    },
    {
      title: '创建时间',
      dataIndex: 'metadata',
      key: 'creationTimestamp',
      render: (metadata: any) => metadata?.creationTimestamp
        ? new Date(metadata.creationTimestamp).toLocaleString()
        : '-',
      sorter: (a: any, b: any) => {
        const timeA = new Date(a.metadata?.creationTimestamp || 0).getTime();
        const timeB = new Date(b.metadata?.creationTimestamp || 0).getTime();
        return timeA - timeB;
      },
    },
    {
      title: '容器',
      dataIndex: 'spec',
      key: 'containers',
      render: (spec: any) => spec?.containers?.length || 0,
    },
  ];
  
  // 渲染污点标签
  const renderTaints = () => {
    if (!taints || taints.length === 0) {
      return <div style={{ color: '#999' }}>无污点</div>;
    }
    
    return (
      <Space wrap>
        {taints.map((taint, index) => {
          let color = 'blue';
          switch (taint.effect) {
            case 'NoSchedule':
              color = 'red';
              break;
            case 'PreferNoSchedule':
              color = 'orange';
              break;
            case 'NoExecute':
              color = 'volcano';
              break;
          }
          
          return (
            <Tag color={color} key={index}>
              {taint.key}{taint.value ? `=${taint.value}` : ''}:{taint.effect}
            </Tag>
          );
        })}
      </Space>
    );
  };
  
  // 渲染标签列表
  const renderLabels = () => {
    if (!labels || Object.keys(labels).length === 0) {
      return <div style={{ color: '#999' }}>无标签</div>;
    }
    
    return (
      <Space wrap>
        {Object.entries(labels).map(([key, value], index) => (
          <Tag color="blue" key={index}>{key}: {value}</Tag>
        ))}
      </Space>
    );
  };
  
  // 渲染资源使用情况
  const renderResourceUsage = () => {
    if (!metrics) {
      return <Spin />;
    }
    
    // 计算百分比
    const cpuUsagePercentage = calculateCpuPercentage(
      metrics.cpu_usage || '0',
      metrics.cpu_capacity || '1'
    );
    
    const memoryUsagePercentage = calculateMemoryPercentage(
      metrics.memory_usage || '0',
      metrics.memory_capacity || '1'
    );
    
    const cpuRequestPercentage = calculateCpuPercentage(
      metrics.cpu_requests || '0',
      metrics.cpu_capacity || '1'
    );
    
    const memoryRequestPercentage = calculateMemoryPercentage(
      metrics.memory_requests || '0',
      metrics.memory_capacity || '1'
    );
    
    return (
      <Row gutter={[24, 24]}>
        <Col xs={24} md={12}>
          <Card title="CPU使用情况" bordered={false}>
            <Statistic
              title="使用率"
              value={cpuUsagePercentage.toFixed(2)}
              suffix="%"
              valueStyle={{ color: cpuUsagePercentage > 80 ? '#cf1322' : '#3f8600' }}
            />
            <Descriptions column={1} size="small" style={{ marginTop: 16 }}>
              <Descriptions.Item label="总容量">{formatCPU(metrics.cpu_capacity || '0')}</Descriptions.Item>
              <Descriptions.Item label="已使用">{formatCPU(metrics.cpu_usage || '0')}</Descriptions.Item>
              <Descriptions.Item label="已请求">{formatCPU(metrics.cpu_requests || '0')} ({cpuRequestPercentage.toFixed(2)}%)</Descriptions.Item>
              <Descriptions.Item label="已限制">{formatCPU(metrics.cpu_limits || '0')}</Descriptions.Item>
            </Descriptions>
            <div style={{ marginTop: 16 }}>
              <div style={{ marginBottom: 8 }}>使用率</div>
              <Progress percent={Math.min(100, parseFloat(cpuUsagePercentage.toFixed(2)))} 
                        status={cpuUsagePercentage > 80 ? 'exception' : 'normal'} />
            </div>
            <div style={{ marginTop: 8 }}>
              <div style={{ marginBottom: 8 }}>请求率</div>
              <Progress percent={Math.min(100, parseFloat(cpuRequestPercentage.toFixed(2)))} 
                        strokeColor="#1890ff" />
            </div>
          </Card>
        </Col>
        <Col xs={24} md={12}>
          <Card title="内存使用情况" bordered={false}>
            <Statistic
              title="使用率"
              value={memoryUsagePercentage.toFixed(2)}
              suffix="%"
              valueStyle={{ color: memoryUsagePercentage > 80 ? '#cf1322' : '#3f8600' }}
            />
            <Descriptions column={1} size="small" style={{ marginTop: 16 }}>
              <Descriptions.Item label="总容量">{formatMemorySize(metrics.memory_capacity || '0')}</Descriptions.Item>
              <Descriptions.Item label="已使用">{formatMemorySize(metrics.memory_usage || '0')}</Descriptions.Item>
              <Descriptions.Item label="已请求">{formatMemorySize(metrics.memory_requests || '0')} ({memoryRequestPercentage.toFixed(2)}%)</Descriptions.Item>
              <Descriptions.Item label="已限制">{formatMemorySize(metrics.memory_limits || '0')}</Descriptions.Item>
            </Descriptions>
            <div style={{ marginTop: 16 }}>
              <div style={{ marginBottom: 8 }}>使用率</div>
              <Progress percent={Math.min(100, parseFloat(memoryUsagePercentage.toFixed(2)))} 
                        status={memoryUsagePercentage > 80 ? 'exception' : 'normal'} />
            </div>
            <div style={{ marginTop: 8 }}>
              <div style={{ marginBottom: 8 }}>请求率</div>
              <Progress percent={Math.min(100, parseFloat(memoryRequestPercentage.toFixed(2)))} 
                        strokeColor="#1890ff" />
            </div>
          </Card>
        </Col>
      </Row>
    );
  };
  
  if (loading) {
    return (
      <div style={{ padding: '24px', textAlign: 'center' }}>
        <Spin size="large" />
      </div>
    );
  }
  
  if (!node) {
    return (
      <div style={{ padding: '24px' }}>
        <Card>
          <div style={{ textAlign: 'center' }}>
            <ExclamationCircleOutlined style={{ fontSize: 48, color: '#ff4d4f' }} />
            <h2>未找到节点信息</h2>
            <Button type="primary" onClick={() => navigate('/nodes')}>
              返回节点列表
            </Button>
          </div>
        </Card>
      </div>
    );
  }
  
  // 获取节点状态
  const nodeStatus = node.status?.conditions?.find((c: any) => c.type === 'Ready')?.status === 'True' ? 'Ready' : 'NotReady';
  // 获取IP地址
  const nodeIP = node.status?.addresses?.find((addr: any) => addr.type === 'InternalIP')?.address || '-';
  // 节点池
  const nodePool = labels?.['k8s.io/pool-name'] || '未分配';
  
  const tabItems = [
    {
      key: 'info',
      label: '节点信息',
      children: (
        <Row gutter={[24, 24]}>
          <Col span={24}>
            <Card title="基本信息" bordered={false}>
              <Descriptions bordered column={{ xs: 1, sm: 2, md: 3 }}>
                <Descriptions.Item label="名称">{nodeName}</Descriptions.Item>
                <Descriptions.Item label="IP地址">{nodeIP}</Descriptions.Item>
                <Descriptions.Item label="状态">
                  <Space>
                    <span style={{ color: nodeStatus === 'Ready' ? '#52c41a' : '#ff4d4f' }}>
                      {nodeStatus}
                    </span>
                    {node.spec?.unschedulable && (
                      <Tag color="orange">不可调度</Tag>
                    )}
                  </Space>
                </Descriptions.Item>
                <Descriptions.Item label="操作系统">{node.status?.nodeInfo?.osImage}</Descriptions.Item>
                <Descriptions.Item label="内核版本">{node.status?.nodeInfo?.kernelVersion}</Descriptions.Item>
                <Descriptions.Item label="容器运行时">{node.status?.nodeInfo?.containerRuntimeVersion}</Descriptions.Item>
                <Descriptions.Item label="Kubelet版本">{node.status?.nodeInfo?.kubeletVersion}</Descriptions.Item>
                <Descriptions.Item label="创建时间">{node.metadata?.creationTimestamp ? new Date(node.metadata.creationTimestamp).toLocaleString() : '-'}</Descriptions.Item>
              </Descriptions>
            </Card>
          </Col>
          
          <Col span={24}>
            <Card title="资源使用情况" bordered={false}>
              {renderResourceUsage()}
            </Card>
          </Col>
          
          <Col span={24}>
            <Card title="污点" bordered={false}>
              {renderTaints()}
            </Card>
          </Col>
          
          <Col span={24}>
            <Card title="标签" bordered={false}>
              {renderLabels()}
            </Card>
          </Col>
        </Row>
      )
    },
    {
      key: 'pods',
      label: `Pod列表 (${pods.length})`,
      children: (
        <Card bordered={false}>
          <Table 
            dataSource={pods}
            columns={podColumns}
            rowKey={(record) => record.metadata?.uid || Math.random().toString()}
            pagination={{ pageSize: 10 }}
          />
        </Card>
      )
    }
  ];

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ marginBottom: 16 }}>
        <Button 
          icon={<ArrowLeftOutlined />} 
          onClick={() => navigate('/nodes')}
        >
          返回节点列表
        </Button>
      </div>
      
      <Card
        title={
          <Space>
            <CloudServerOutlined />
            <span style={{ fontSize: 18 }}>{nodeName}</span>
            <Tag color={nodeStatus === 'Ready' ? 'green' : 'red'}>{nodeStatus}</Tag>
            {node.spec?.unschedulable && <Tag color="orange">不可调度</Tag>}
          </Space>
        }
      >
        <Tabs defaultActiveKey="info" items={tabItems} />
      </Card>
    </div>
  );
};

export default NodeDetail;