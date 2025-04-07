import React, { useState, useEffect } from 'react';
import { Descriptions, Tag, Table, Card, Space, Typography, Button, message } from 'antd';
import { EditOutlined } from '@ant-design/icons';
import { formatDate } from '@/utils/format';
import EditDeploymentModal from './EditDeploymentModal';
import PodList from '../pod/PodList';
import { updateDeployment, UpdateDeploymentRequest } from '@/api/deployment';
import { getPodsBySelector } from '@/api/pod';
import DeploymentEvents from './DeploymentEvents';
import { useTranslation } from 'react-i18next';

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
  const { t } = useTranslation();
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
        message.error(response.data.message || t('pods.fetchFailed'));
      }
    } catch (error) {
      console.error(t('pods.fetchFailed') + ':', error);
      message.error(t('pods.fetchFailed'));
    } finally {
      setPodsLoading(false);
    }
  };

  // 从Pods中提取健康检查探针数据
  const extractProbesFromPods = (pods: any[]) => {
    console.log(t('deployments.detail.extractingProbes'), pods.length);
    
    if (!pods || pods.length === 0) return;
    
    // 创建容器名称到容器的映射
    const containerMap: { [key: string]: any } = {};
    deployment.containers.forEach(container => {
      containerMap[container.name] = container;
    });
    
    // 遍历所有Pod
    pods.forEach(pod => {
      console.log(t('deployments.detail.processingPod'), pod.metadata?.name);
      const containers = pod.spec?.containers || [];
      
      // 遍历Pod中的容器
      containers.forEach((podContainer: any) => {
        const containerName = podContainer.name;
        // 检查这个容器是否属于当前Deployment
        if (containerMap[containerName]) {
          console.log(t('deployments.detail.foundContainer', { name: containerName }));
          
          // 更新健康检查探针
          ['livenessProbe', 'readinessProbe', 'startupProbe'].forEach(probeType => {
            if (podContainer[probeType]) {
              console.log(t('deployments.detail.foundProbe', { container: containerName, type: probeType }));
              // 将探针数据添加到deployment的容器中
              containerMap[containerName][probeType] = podContainer[probeType];
            }
          });
        }
      });
    });
    
    console.log(t('deployments.detail.processedContainers'), deployment.containers);
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
      console.error(t('deployments.editFailed') + ':', error);
      message.error(t('deployments.editFailed'));
      return Promise.reject(error);
    }
  };

  // 容器列定义
  const containerColumns = [
    {
      title: t('deployments.detail.containerColumns.name'),
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: t('deployments.detail.containerColumns.image'),
      dataIndex: 'image',
      key: 'image',
    },
    {
      title: t('deployments.detail.containerColumns.ports'),
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
      title: t('deployments.detail.conditionColumns.type'),
      dataIndex: 'type',
      key: 'type',
    },
    {
      title: t('deployments.detail.conditionColumns.status'),
      dataIndex: 'status',
      key: 'status',
      render: (text: string) => (
        <Tag color={text === 'True' ? 'success' : 'error'}>{text}</Tag>
      ),
    },
    {
      title: t('deployments.detail.conditionColumns.lastUpdateTime'),
      dataIndex: 'lastUpdateTime',
      key: 'lastUpdateTime',
      render: (text: string) => formatDate(text),
    },
    {
      title: t('deployments.detail.conditionColumns.reason'),
      dataIndex: 'reason',
      key: 'reason',
    },
    {
      title: t('deployments.detail.conditionColumns.message'),
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
          {t('common.edit')}
        </Button>
      </div>
      
      <Card title={t('deployments.detail.basicInfo.title')}>
        <Descriptions column={2}>
          <Descriptions.Item label={t('deployments.namespace')}>{deployment.namespace}</Descriptions.Item>
          <Descriptions.Item label={t('deployments.createdAt')}>{formatDate(deployment.creationTime)}</Descriptions.Item>
          <Descriptions.Item label={t('deployments.detail.basicInfo.replicas')}>
            {deployment.readyReplicas || 0}/{deployment.replicas}
          </Descriptions.Item>
          <Descriptions.Item label={t('deployments.detail.basicInfo.strategy')}>{deployment.strategy}</Descriptions.Item>
        </Descriptions>
      </Card>

      <Card title={t('deployments.labels')}>
        <Space wrap>
          {Object.entries(deployment.labels || {}).map(([key, value], index) => (
            <Tag key={`label-${key}-${index}`}>{`${key}: ${value}`}</Tag>
          ))}
        </Space>
      </Card>

      <Card title={t('deployments.detail.selector')}>
        <Space wrap>
          {Object.entries(deployment.selector || {}).map(([key, value], index) => (
            <Tag key={`selector-${key}-${index}`}>{`${key}: ${value}`}</Tag>
          ))}
        </Space>
      </Card>

      <Card title={t('deployments.detail.containers')}>
        <Table
          columns={containerColumns}
          dataSource={deployment.containers}
          rowKey="name"
          loading={podsLoading}
          pagination={false}
        />
      </Card>

      <Card title={t('deployments.detail.conditions')}>
        <Table
          columns={conditionColumns}
          dataSource={deployment.conditions}
          rowKey="type"
          pagination={false}
        />
      </Card>

      <Card title={t('deployments.detail.pods')} loading={podsLoading}>
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

      <DeploymentEvents
        clusterName={clusterName}
        namespace={deployment.namespace}
        deploymentName={deployment.name}
      />
    </Space>
  );
};

export default DeploymentDetail;