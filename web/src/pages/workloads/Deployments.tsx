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

const { Title } = Typography;
const { Option } = Select;

interface DeploymentType {
  name: string;
  namespace: string;
  replicas: number;
  readyReplicas: number;
  strategy: string;
  creationTime: string;
  labels: { [key: string]: string };
  selector: { [key: string]: string };
  containerCount: number;
  images: string[];
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
        // 直接使用返回的集群列表，因为它已经是字符串数组
        const clusterList = response.data.data.clusters;
        setClusters(clusterList);
        if (clusterList.length > 0 && !selectedCluster) {
          setSelectedCluster(clusterList[0]);
        }
      }
    } catch (err) {
      console.error('获取集群列表失败:', err);
      message.error('获取集群列表失败');
    }
  };

  // 获取Deployments数据
  const fetchDeployments = async () => {
    if (!selectedCluster) return;
    
    setLoading(true);
    try {
      const response = await listDeploymentsByNamespace(selectedCluster, namespace);
      if (response.data.code === 0) {
        setDeployments(response.data.data.deployments || []);
        setFilteredDeployments(response.data.data.deployments || []);
      } else {
        message.error(response.data.message || '获取Deployments列表失败');
        setDeployments([]);
        setFilteredDeployments([]);
      }
    } catch (err) {
      console.error('获取Deployments列表失败:', err);
      message.error('获取Deployments列表失败');
      setDeployments([]);
      setFilteredDeployments([]);
    } finally {
      setLoading(false);
    }
  };

  // 初始化加载集群列表
  useEffect(() => {
    fetchClusters();
  }, []);

  // 当选择的集群或命名空间改变时，加载Deployments数据
  useEffect(() => {
    if (selectedCluster) {
      fetchDeployments();
      // 每30秒刷新一次
      const timer = setInterval(fetchDeployments, 30000);
      return () => clearInterval(timer);
    }
  }, [selectedCluster, namespace]);

  // 搜索过滤
  useEffect(() => {
    if (searchText) {
      const filtered = deployments.filter(
        (item) =>
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
        currentDeployment.namespace,
        currentDeployment.name,
        replicaCount
      );
      
      if (response.data.code === 0) {
        message.success(`已将 ${currentDeployment.name} 的副本数调整为 ${replicaCount}`);
        setScaleModalVisible(false);
        fetchDeployments();
      } else {
        message.error(response.data.message || '调整副本数失败');
      }
    } catch (error) {
      console.error('调整副本数失败:', error);
      message.error('调整副本数失败');
    }
  };

  // 重启Deployment
  const handleRestart = async (deployment: DeploymentType) => {
    try {
      console.log(`开始重启 ${deployment.name}...`);
      const response = await restartDeployment(selectedCluster, deployment.namespace, deployment.name);
      console.log(`重启响应:`, response.data);
      
      if (response.data.code === 0) {
        message.success(`已重启 ${deployment.name}`);
        fetchDeployments();
      } else {
        message.error(response.data.message || '重启失败');
        console.error('重启失败，错误信息:', response.data.message);
      }
    } catch (error) {
      console.error('重启失败:', error);
      message.error('重启失败');
    }
  };

  // 查看Deployment详情
  const viewDeploymentDetails = async (deployment: DeploymentType) => {
    setDetailVisible(true);
    setDetailLoading(true);
    try {
      console.log(`获取${deployment.name}详情...`);
      const response = await getDeploymentDetails(selectedCluster, deployment.namespace, deployment.name);
      console.log(`详情响应:`, response.data);
      
      if (response.data.code === 0) {
        const deploymentDetail = response.data.data.deployment;
        console.log(`处理后的详情数据:`, deploymentDetail);
        
        // 记录容器探针信息
        if (deploymentDetail.containers) {
          deploymentDetail.containers.forEach(container => {
            console.log(`容器 ${container.name} 的探针信息:`, {
              livenessProbe: container.livenessProbe,
              readinessProbe: container.readinessProbe,
              startupProbe: container.startupProbe
            });
          });
        }
        
        setCurrentDeploymentDetail(deploymentDetail);
      } else {
        message.error(response.data.message || '获取部署详情失败');
        console.error('获取部署详情失败，错误信息:', response.data.message);
      }
    } catch (error) {
      console.error('获取部署详情失败:', error);
      message.error('获取部署详情失败');
    } finally {
      setDetailLoading(false);
    }
  };

  // 关闭详情抽屉
  const handleDetailClose = () => {
    setDetailVisible(false);
    setCurrentDeploymentDetail(null);
  };

  // 打开创建Deployment模态框
  const showCreateModal = () => {
    if (!selectedCluster) {
      message.warning('请先选择集群');
      return;
    }
    setCreateModalVisible(true);
  };

  // 关闭创建Deployment模态框
  const handleCreateCancel = () => {
    setCreateModalVisible(false);
  };

  // 提交创建Deployment请求
  const handleCreateSubmit = async (deploymentData: CreateDeploymentRequest) => {
    try {
      const response = await createDeployment(selectedCluster, namespace, deploymentData);
      
      if (response.data.code === 0) {
        message.success(`Deployment ${deploymentData.name} 创建成功`);
        fetchDeployments();
        return Promise.resolve();
      } else {
        message.error(response.data.message || '创建Deployment失败');
        return Promise.reject(new Error(response.data.message));
      }
    } catch (error) {
      console.error('创建Deployment失败:', error);
      message.error('创建Deployment失败: ' + (error instanceof Error ? error.message : '未知错误'));
      return Promise.reject(error);
    }
  };

  // 表格列定义
  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      sorter: (a: DeploymentType, b: DeploymentType) => a.name.localeCompare(b.name),
      render: (text: string, record: DeploymentType) => (
        <a onClick={() => viewDeploymentDetails(record)}>{text}</a>
      ),
    },
    {
      title: '命名空间',
      dataIndex: 'namespace',
      key: 'namespace',
    },
    {
      title: '副本',
      key: 'replicas',
      render: (_: any, record: DeploymentType) => (
        <span>
          {record.readyReplicas || 0}/{record.replicas || 0}
        </span>
      ),
      sorter: (a: DeploymentType, b: DeploymentType) => (a.readyReplicas || 0) - (b.readyReplicas || 0),
    },
    {
      title: '状态',
      key: 'status',
      render: (_: any, record: DeploymentType) => {
        const ready = (record.readyReplicas || 0) === record.replicas;
        return (
          <Badge
            status={ready ? 'success' : 'processing'}
            text={ready ? '就绪' : '部署中'}
          />
        );
      },
    },
    {
      title: '镜像',
      key: 'images',
      render: (_: any, record: DeploymentType) => (
        <>
          {record.images && record.images.length > 0 ? (
            record.images.map((image, index) => (
              <div key={`${record.name}-image-${index}`} style={{ marginBottom: '4px' }}>
                <Typography.Text ellipsis={{ tooltip: image }} style={{ maxWidth: 200 }}>
                  {image}
                </Typography.Text>
              </div>
            ))
          ) : (
            <div>-</div>
          )}
        </>
      ),
    },
    {
      title: '策略',
      dataIndex: 'strategy',
      key: 'strategy',
      render: (text: string) => (
        <Tag color={text === 'RollingUpdate' ? 'blue' : 'orange'}>{text}</Tag>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'creationTime',
      key: 'creationTime',
      render: (text: string) => formatDate(text),
      sorter: (a: DeploymentType, b: DeploymentType) => new Date(a.creationTime).getTime() - new Date(b.creationTime).getTime(),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: DeploymentType) => (
        <Space size="small">
          <Tooltip title="查看详情">
            <Button
              icon={<EyeOutlined />}
              size="small"
              onClick={() => viewDeploymentDetails(record)}
            />
          </Tooltip>
          <Tooltip title="调整副本数">
            <Button
              icon={<ScissorOutlined />}
              size="small"
              onClick={() => showScaleModal(record)}
            />
          </Tooltip>
          <Tooltip title="重启">
            <Popconfirm
              title="确定要重启此Deployment吗?"
              onConfirm={() => handleRestart(record)}
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
          <Select
            // 确保值不为null
            value={namespace || 'default'}
            onChange={setNamespace}
            style={{ width: 180 }}
          >
            <Option key="default" value="default">default</Option>
            <Option key="kube-system" value="kube-system">kube-system</Option>
            {/* 可以通过API获取命名空间列表 */}
          </Select>
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

      {/* 调整副本数模态框 */}
      <Modal
        title={`调整 ${currentDeployment?.name || ''} 的副本数`}
        open={scaleModalVisible}
        onOk={handleScaleSubmit}
        onCancel={handleScaleCancel}
        okText="确定"
        cancelText="取消"
      >
        <div style={{ marginBottom: 16 }}>
          <Typography.Text>当前副本数: {currentDeployment?.replicas || 0}</Typography.Text>
        </div>
        <div>
          <Typography.Text>新副本数: </Typography.Text>
          <InputNumber
            min={0}
            max={100}
            value={replicaCount}
            onChange={(value) => setReplicaCount(value as number)}
            style={{ width: 120 }}
          />
        </div>
      </Modal>

      <Drawer
        title="Deployment详情"
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