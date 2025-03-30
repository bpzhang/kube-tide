import React, { useState, useEffect } from 'react';
import {
  Card,
  Table,
  Tag,
  Button,
  Space,
  Input,
  Popconfirm,
  message,
  Tooltip,
  Select,
  Badge,
  Modal,
  InputNumber,
  Typography,
  Row,
  Col,
  Drawer,
} from 'antd';
import {
  ReloadOutlined,
  SearchOutlined,
  EyeOutlined,
  ScissorOutlined,
  SyncOutlined,
  PlusOutlined,
} from '@ant-design/icons';
import { 
  listDeploymentsByNamespace, 
  restartDeployment, 
  scaleDeployment, 
  getDeploymentDetails,
  createDeployment,
  CreateDeploymentRequest
} from '../../api/deployment';
import { getClusterList } from '../../api/cluster';
import { formatDate } from '../../utils/format';
import DeploymentDetail from '../../components/k8s/deployment/DeploymentDetail';
import CreateDeploymentModal from '../../components/k8s/deployment/CreateDeploymentModal';
import NamespaceSelector from '../../components/k8s/common/NamespaceSelector';

const { Title } = Typography;
const { Option } = Select;

interface DeploymentType {
  name: string;
  namespace: string;
  replicas: number;
  readyReplicas: number;
  strategy: string;
  labels: { [key: string]: string };
  createdAt: string;
}

const DeploymentsPage: React.FC = () => {
  const [selectedCluster, setSelectedCluster] = useState<string>('');
  const [clusters, setClusters] = useState<string[]>([]);
  const [namespace, setNamespace] = useState<string>('default');
  const [deployments, setDeployments] = useState<DeploymentType[]>([]);
  const [filteredDeployments, setFilteredDeployments] = useState<DeploymentType[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [scaleModalVisible, setScaleModalVisible] = useState(false);
  const [currentDeployment, setCurrentDeployment] = useState<DeploymentType | null>(null);
  const [replicaCount, setReplicaCount] = useState<number>(1);
  const [detailVisible, setDetailVisible] = useState(false);
  const [currentDeploymentDetail, setCurrentDeploymentDetail] = useState<any>(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [createModalVisible, setCreateModalVisible] = useState(false);

  // 获取集群列表
  const fetchClusters = async () => {
    try {
      const response = await getClusterList();
      if (response.data.code === 0) {
        setClusters(response.data.data.clusters);
        if (response.data.data.clusters.length > 0 && !selectedCluster) {
          setSelectedCluster(response.data.data.clusters[0]);
        }
      }
    } catch (err) {
      message.error('获取集群列表失败');
    }
  };

  // 获取Deployment列表
  const fetchDeployments = async () => {
    if (!selectedCluster) return;
    
    setLoading(true);
    try {
      const response = await listDeploymentsByNamespace(selectedCluster, namespace);
      if (response.data.code === 0) {
        setDeployments(response.data.data.deployments || []);
      } else {
        message.error(response.data.message || '获取Deployment列表失败');
        setDeployments([]);
      }
    } catch (err) {
      message.error('获取Deployment列表失败');
      setDeployments([]);
    } finally {
      setLoading(false);
    }
  };

  // 获取Deployment详情
  const fetchDeploymentDetail = async (name: string) => {
    if (!selectedCluster) return;
    
    setDetailLoading(true);
    try {
      const response = await getDeploymentDetails(selectedCluster, namespace, name);
      if (response.data.code === 0) {
        setCurrentDeploymentDetail(response.data.data.deployment);
      } else {
        message.error(response.data.message || '获取Deployment详情失败');
        setCurrentDeploymentDetail(null);
      }
    } catch (err) {
      message.error('获取Deployment详情失败');
      setCurrentDeploymentDetail(null);
    } finally {
      setDetailLoading(false);
    }
  };

  // 初始化加载
  useEffect(() => {
    fetchClusters();
  }, []);

  // 当集群或命名空间变化时重新获取Deployment列表
  useEffect(() => {
    if (selectedCluster) {
      fetchDeployments();
    }
  }, [selectedCluster, namespace]);

  // 命名空间变化时重置搜索
  useEffect(() => {
    setSearchText('');
  }, [namespace]);

  // 搜索过滤
  useEffect(() => {
    if (searchText) {
      const filtered = deployments.filter(item => 
          item.name.toLowerCase().includes(searchText.toLowerCase()) ||
          Object.keys(item.labels || {}).some((key) =>
            `${key}:${item.labels[key]}`.toLowerCase().includes(searchText.toLowerCase())
          )
      );
      setFilteredDeployments(filtered);
    } else {
      setFilteredDeployments(deployments);
    }
  }, [searchText, deployments]);

  // 刷新数据
  const handleRefresh = () => {
    fetchDeployments();
  };

  // 打开缩放弹窗
  const showScaleModal = (deployment: DeploymentType) => {
    setCurrentDeployment(deployment);
    setReplicaCount(deployment.replicas);
    setScaleModalVisible(true);
  };

  // 关闭缩放弹窗
  const handleScaleCancel = () => {
    setScaleModalVisible(false);
    setCurrentDeployment(null);
  };

  // 提交缩放请求
  const handleScaleSubmit = async () => {
    if (!currentDeployment || replicaCount === undefined) return;

    try {
      const response = await scaleDeployment(
        selectedCluster,
        namespace,
        currentDeployment.name,
        replicaCount
      );
      
      if (response.data.code === 0) {
        message.success(`成功将 ${currentDeployment.name} 的副本数调整为 ${replicaCount}`);
        handleScaleCancel();
        fetchDeployments();
      } else {
        message.error(response.data.message || '调整副本数失败');
      }
    } catch (err) {
      message.error('调整副本数失败');
    }
  };

  // 重启Deployment
  const handleRestart = async (deploymentName: string) => {
    try {
      const response = await restartDeployment(selectedCluster, namespace, deploymentName);
      if (response.data.code === 0) {
        message.success(`成功重启 ${deploymentName}`);
        fetchDeployments();
      } else {
        message.error(response.data.message || '重启失败');
      }
    } catch (err) {
      message.error('重启失败');
    }
  };

  // 显示Deployment详情
  const handleShowDetail = (deploymentName: string) => {
    setDetailVisible(true);
    fetchDeploymentDetail(deploymentName);
  };

  // 关闭详情抽屉
  const handleDetailClose = () => {
    setDetailVisible(false);
    setTimeout(() => {
      setCurrentDeploymentDetail(null);
    }, 300);
  };

  // 显示创建Deployment模态框
  const showCreateModal = () => {
    setCreateModalVisible(true);
  };

  // 关闭创建Deployment模态框
  const handleCreateCancel = () => {
    setCreateModalVisible(false);
  };

  // 提交创建Deployment请求
  const handleCreateSubmit = async (values: CreateDeploymentRequest) => {
    try {
      const response = await createDeployment(selectedCluster, namespace, values);
      if (response.data.code === 0) {
        message.success('Deployment创建成功');
        handleCreateCancel();
        fetchDeployments();
      } else {
        message.error(response.data.message || 'Deployment创建失败');
      }
    } catch (err) {
      message.error('Deployment创建失败');
    }
  };

  // 表格列定义
  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string) => (
        <a onClick={() => handleShowDetail(text)}>{text}</a>
      ),
    },
    {
      title: '命名空间',
      dataIndex: 'namespace',
      key: 'namespace',
    },
    {
      title: '标签',
      dataIndex: 'labels',
      key: 'labels',
      render: (labels: { [key: string]: string }) => (
        <div style={{ maxWidth: 200, maxHeight: 60, overflow: 'auto' }}>
          {Object.entries(labels || {}).map(([key, value]) => (
            <Tag color="blue" key={key}>{`${key}: ${value}`}</Tag>
          ))}
        </div>
      ),
    },
    {
      title: '副本',
      key: 'replicas',
      render: (text: string, record: DeploymentType) => (
        <Space>
          <Badge status={record.readyReplicas === record.replicas ? 'success' : 'processing'} />
          <span>
            {record.readyReplicas}/{record.replicas}
          </span>
        </Space>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (time: string) => formatDate(time),
    },
    {
      title: '操作',
      key: 'action',
      render: (text: string, record: DeploymentType) => (
        <Space>
          <Tooltip title="查看详情">
            <Button
              type="link"
              icon={<EyeOutlined />}
              size="small"
              onClick={() => handleShowDetail(record.name)}
            />
          </Tooltip>
          <Tooltip title="调整副本数">
            <Button
              type="link"
              icon={<ScissorOutlined />}
              size="small"
              onClick={() => showScaleModal(record)}
            />
          </Tooltip>
          <Tooltip title="重启">
            <Popconfirm
              title="确定要重启这个Deployment吗?"
              onConfirm={() => handleRestart(record.name)}
              okText="确定"
              cancelText="取消"
            >
              <Button icon={<SyncOutlined />} size="small" />
            </Popconfirm>
          </Tooltip>
        </Space>
      ),
    },
  ];

  return (
    <Card
      title={
        <Row justify="space-between" align="middle">
          <Col>
            <Title level={4} style={{ marginBottom: 0 }}>
              Deployments 管理
            </Title>
          </Col>
          <Col>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={showCreateModal}
            >
              创建 Deployment
            </Button>
          </Col>
        </Row>
      }
      extra={
        <Space>
          <span>集群:</span>
          <Select
            // 确保值不为null
            value={selectedCluster || ''}
            onChange={setSelectedCluster}
            style={{ width: 180 }}
            loading={clusters.length === 0}
          >
            {clusters.map((cluster, index) => (
              <Option key={`cluster-${index}`} value={cluster}>{cluster}</Option>
            ))}
          </Select>
          <span>命名空间:</span>
          <NamespaceSelector
            clusterName={selectedCluster}
            value={namespace}
            onChange={setNamespace}
            width={180}
          />
          <Input
            placeholder="搜索Deployments"
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            prefix={<SearchOutlined />}
            allowClear
            style={{ width: 200 }}
          />
          <Tooltip title="刷新">
            <Button
              type="primary"
              icon={<ReloadOutlined />}
              onClick={handleRefresh}
              loading={loading}
            />
          </Tooltip>
        </Space>
      }
    >
      <Table
        columns={columns}
        dataSource={filteredDeployments}
        rowKey="name"
        loading={loading}
        pagination={{ pageSize: 10 }}
      />

      {/* 缩放模态框 */}
      <Modal
        title={`调整副本数: ${currentDeployment?.name}`}
        open={scaleModalVisible}
        onCancel={handleScaleCancel}
        onOk={handleScaleSubmit}
      >
        <div style={{ marginBottom: 16 }}>
          <p>当前副本数: {currentDeployment?.replicas}</p>
          <p>就绪副本数: {currentDeployment?.readyReplicas}</p>
        </div>
        <InputNumber
          min={0}
          value={replicaCount}
          onChange={(value) => setReplicaCount(value || 0)}
          style={{ width: '100%' }}
          placeholder="请输入目标副本数"
        />
      </Modal>

      {/* 详情抽屉 */}
      <Drawer
        title={`Deployment 详情: ${currentDeploymentDetail?.name}`}
        placement="right"
        width={800}
        onClose={handleDetailClose}
        open={detailVisible}
      >
        {detailLoading ? (
          <div style={{ textAlign: 'center', padding: '20px' }}>加载中...</div>
        ) : currentDeploymentDetail ? (
          <DeploymentDetail 
            deployment={currentDeploymentDetail} 
            clusterName={selectedCluster}
            onUpdate={() => {
              handleDetailClose();
              fetchDeployments();
            }}
          />
        ) : (
          <div style={{ textAlign: 'center', padding: '20px' }}>暂无数据</div>
        )}
      </Drawer>

      {/* 创建Deployment模态框 */}
      <CreateDeploymentModal
        visible={createModalVisible}
        onClose={handleCreateCancel}
        onSubmit={handleCreateSubmit}
        clusterName={selectedCluster}
        namespace={namespace}
      />
    </Card>
  );
};

export default DeploymentsPage;