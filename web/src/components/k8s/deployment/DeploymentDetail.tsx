import React, { useState, useEffect } from 'react';
import { Descriptions, Tag, Table, Card, Space, Typography, Button, message } from 'antd';
import { EditOutlined } from '@ant-design/icons';
import { formatDate } from '@/utils/format';
import EditDeploymentModal from './EditDeploymentModal';
import PodList from '../pod/PodList';
import { updateDeployment, UpdateDeploymentRequest, getDeploymentEvents } from '@/api/deployment';
import { getPodsBySelector } from '@/api/pod';
import K8sEvents from '../common/K8sEvents';

const { Title } = Typography;

interface Container {
  name: string;
  image: string;
  ports: any[];
  env: any[];
  resources: {
    limits?: { cpu?: string; memory?: string };
    requests?: { cpu?: string; memory?: string };
  };
}

interface DeploymentDetailProps {
  deployment: {
    name: string;
    namespace: string;
    replicas: number;
    readyReplicas: number;
    strategy: string;
    creationTime: string;
    labels: { [key: string]: string };
    selector: { [key: string]: string };
    annotations: { [key: string]: string };
    containers: Container[];
    conditions: Array<{
      type: string;
      status: string;
      lastUpdateTime: string;
      reason: string;
      message: string;
    }>;
    minReadySeconds?: number;
    revisionHistoryLimit?: number;
    paused?: boolean;
  };
  clusterName: string;
  onUpdate?: () => void;
}

const DeploymentDetail: React.FC<DeploymentDetailProps> = ({ 
  deployment, 
  clusterName, 
  onUpdate 
}) => {
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [pods, setPods] = useState<any[]>([]);
  const [podsLoading, setPodsLoading] = useState(false);

  // 获取相关的Pod列表
  const fetchPods = async () => {
    if (!deployment.selector) return;

    setPodsLoading(true);
    try {
      const response = await getPodsBySelector(
        clusterName,
        deployment.namespace,
        deployment.selector
      );
      if (response.data.code === 0) {
        // 设置Pods数据
        setPods(response.data.data.pods);
        
        // 提取Pod中的健康检查探针数据
        extractProbesFromPods(response.data.data.pods);
      } else {
        message.error(response.data.message || '获取Pod列表失败');
      }
    } catch (error) {
      console.error('获取Pod列表失败:', error);
      message.error('获取Pod列表失败');
    } finally {
      setPodsLoading(false);
    }
  };

  // 从Pods中提取健康检查探针数据
  const extractProbesFromPods = (pods: any[]) => {
    console.log('从Pods提取探针数据，Pods数量:', pods.length);
    
    if (!pods || pods.length === 0) return;
    
    // 创建容器名称到容器的映射
    const containerMap: { [key: string]: any } = {};
    deployment.containers.forEach(container => {
      containerMap[container.name] = container;
    });
    
    // 遍历所有Pod
    pods.forEach(pod => {
      console.log('处理Pod:', pod.metadata?.name);
      const containers = pod.spec?.containers || [];
      
      // 遍历Pod中的容器
      containers.forEach((podContainer: any) => {
        const containerName = podContainer.name;
        // 检查这个容器是否属于当前Deployment
        if (containerMap[containerName]) {
          console.log(`在Pod中找到Deployment的容器: ${containerName}`);
          
          // 更新健康检查探针
          ['livenessProbe', 'readinessProbe', 'startupProbe'].forEach(probeType => {
            if (podContainer[probeType]) {
              console.log(`容器 ${containerName} 存在 ${probeType}:`, podContainer[probeType]);
              // 将探针数据添加到deployment的容器中
              containerMap[containerName][probeType] = podContainer[probeType];
            }
          });
        }
      });
    });
    
    console.log('处理后的Deployment容器数据:', deployment.containers);
  };

  useEffect(() => {
    fetchPods();
    // 每30秒刷新一次Pod列表
    const timer = setInterval(fetchPods, 30000);
    return () => clearInterval(timer);
  }, [deployment.selector, deployment.namespace, clusterName]);

  // 处理编辑模态框的显示和隐藏
  const showEditModal = () => {
    setEditModalVisible(true);
  };

  const hideEditModal = () => {
    setEditModalVisible(false);
  };

  // 处理更新提交
  const handleUpdateSubmit = async (updateData: UpdateDeploymentRequest) => {
    try {
      await updateDeployment(
        clusterName,
        deployment.namespace,
        deployment.name,
        updateData
      );
      
      // 如果有更新回调，则调用
      if (onUpdate) {
        onUpdate();
      }
      
      return Promise.resolve();
    } catch (error) {
      console.error('更新Deployment失败:', error);
      message.error('更新Deployment失败');
      return Promise.reject(error);
    }
  };

  // 容器列定义
  const containerColumns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '镜像',
      dataIndex: 'image',
      key: 'image',
    },
    {
      title: '端口',
      key: 'ports',
      render: (record: Container) => (
        <>
          {record.ports?.map((port, index) => (
            <Tag key={index}>
              {port.containerPort}/{port.protocol}
            </Tag>
          ))}
        </>
      ),
    },
  ];

  // 状态条件列定义
  const conditionColumns = [
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (text: string) => (
        <Tag color={text === 'True' ? 'success' : 'error'}>{text}</Tag>
      ),
    },
    {
      title: '最后更新',
      dataIndex: 'lastUpdateTime',
      key: 'lastUpdateTime',
      render: (text: string) => formatDate(text),
    },
    {
      title: '原因',
      dataIndex: 'reason',
      key: 'reason',
    },
    {
      title: '消息',
      dataIndex: 'message',
      key: 'message',
    },
  ];

  return (
    <Space direction="vertical" size="middle" style={{ width: '100%' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Title level={4}>{deployment.name}</Title>
        <Button 
          type="primary" 
          icon={<EditOutlined />} 
          onClick={showEditModal}
        >
          编辑
        </Button>
      </div>
      
      <Card title="基本信息">
        <Descriptions column={2}>
          <Descriptions.Item label="命名空间">{deployment.namespace}</Descriptions.Item>
          <Descriptions.Item label="创建时间">{formatDate(deployment.creationTime)}</Descriptions.Item>
          <Descriptions.Item label="副本数">
            {deployment.readyReplicas || 0}/{deployment.replicas}
          </Descriptions.Item>
          <Descriptions.Item label="更新策略">{deployment.strategy}</Descriptions.Item>
        </Descriptions>
      </Card>

      <Card title="标签">
        <Space wrap>
          {Object.entries(deployment.labels || {}).map(([key, value], index) => (
            <Tag key={`label-${key}-${index}`}>{`${key}: ${value}`}</Tag>
          ))}
        </Space>
      </Card>

      <Card title="选择器">
        <Space wrap>
          {Object.entries(deployment.selector || {}).map(([key, value], index) => (
            <Tag key={`selector-${key}-${index}`}>{`${key}: ${value}`}</Tag>
          ))}
        </Space>
      </Card>

      <Card title="容器">
        <Table
          columns={containerColumns}
          dataSource={deployment.containers}
          rowKey="name"
          loading={podsLoading}
          pagination={false}
        />
      </Card>

      <Card title="状态条件">
        <Table
          columns={conditionColumns}
          dataSource={deployment.conditions}
          rowKey="type"
          pagination={false}
        />
      </Card>

      <Card title="容器组" loading={podsLoading}>
        <PodList
          clusterName={clusterName}
          namespace={deployment.namespace}
          pods={pods}
          onRefresh={fetchPods}
        />
      </Card>

      <EditDeploymentModal
        visible={editModalVisible}
        onClose={hideEditModal}
        onSubmit={handleUpdateSubmit}
        deployment={deployment}
      />

      <K8sEvents
        clusterName={clusterName}
        namespace={deployment.namespace}
        resourceName={deployment.name}
        resourceKind="Deployment"
        fetchEvents={getDeploymentEvents}
      />
    </Space>
  );
};

export default DeploymentDetail;