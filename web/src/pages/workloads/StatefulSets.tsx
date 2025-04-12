import React, { useState, useEffect, useCallback } from 'react';
import { Card, Table, Tag, Space, Button, message, Popconfirm, Select, Spin, Tooltip } from 'antd';
import { 
  SyncOutlined, 
  DeleteOutlined, 
  PlusOutlined, 
  ReloadOutlined,
  ScissorOutlined,
  EyeOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { getClusterList } from '@/api/cluster';
import { 
  listStatefulSetsByNamespace,
  deleteStatefulSet,
  restartStatefulSet,
  StatefulSetInfo
} from '@/api/statefulset';
import { formatDate } from '@/utils/format';
import NamespaceSelector from '@/components/k8s/common/NamespaceSelector';
import CreateStatefulSetModal from '@/components/k8s/statefulset/CreateStatefulSetModal';
import ScaleStatefulSetModal from '@/components/k8s/statefulset/ScaleStatefulSetModal';

const { Option } = Select;

/**
 * StatefulSet管理页面
 */
const StatefulSets: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [statefulsets, setStatefulSets] = useState<StatefulSetInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [clusters, setClusters] = useState<string[]>([]);
  const [selectedCluster, setSelectedCluster] = useState<string>('');
  const [namespace, setNamespace] = useState<string>('default');
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [scaleModalVisible, setScaleModalVisible] = useState(false);
  const [currentStatefulSet, setCurrentStatefulSet] = useState<StatefulSetInfo | null>(null);

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
      message.error(t('clusters.fetchFailed'));
    }
  };

  // 获取StatefulSet列表
  const fetchStatefulSets = useCallback(async () => {
    if (!selectedCluster) return;
    
    setLoading(true);
    try {
      const response = await listStatefulSetsByNamespace(selectedCluster, namespace);
      if (response.data.code === 0) {
        setStatefulSets(response.data.data.statefulsets || []);
      } else {
        message.error(response.data.message || t('statefulsets.fetchFailed'));
        setStatefulSets([]);
      }
    } catch (err) {
      message.error(t('statefulsets.fetchFailed'));
      setStatefulSets([]);
    } finally {
      setLoading(false);
    }
  }, [selectedCluster, namespace, t]);

  // 初始化加载
  useEffect(() => {
    fetchClusters();
  }, []);

  // 当集群或命名空间变化时重新获取StatefulSet列表
  useEffect(() => {
    if (selectedCluster) {
      fetchStatefulSets();
      
      // 每30秒刷新一次
      const timer = setInterval(fetchStatefulSets, 30000);
      return () => clearInterval(timer);
    }
  }, [selectedCluster, namespace, fetchStatefulSets]);

  // 处理集群变化
  const handleClusterChange = (value: string) => {
    setSelectedCluster(value);
  };

  // 处理命名空间变化
  const handleNamespaceChange = (value: string) => {
    setNamespace(value);
  };

  // 查看StatefulSet详情
  const handleViewDetails = (record: StatefulSetInfo) => {
    navigate(`/workloads/statefulsets/detail/${selectedCluster}/${record.namespace}/${record.name}`);
  };

  // 处理删除StatefulSet
  const handleDelete = async (record: StatefulSetInfo) => {
    try {
      const response = await deleteStatefulSet(selectedCluster, record.namespace, record.name);
      if (response.data.code === 0) {
        message.success(t('statefulsets.deleteSuccess'));
        fetchStatefulSets();
      } else {
        message.error(response.data.message || t('statefulsets.deleteFailed'));
      }
    } catch (err) {
      message.error(t('statefulsets.deleteFailed'));
    }
  };

  // 重启StatefulSet
  const handleRestart = async (record: StatefulSetInfo) => {
    try {
      const response = await restartStatefulSet(selectedCluster, record.namespace, record.name);
      if (response.data.code === 0) {
        message.success(t('statefulsets.restartSuccess'));
        fetchStatefulSets();
      } else {
        message.error(response.data.message || t('statefulsets.restartFailed'));
      }
    } catch (err) {
      message.error(t('statefulsets.restartFailed'));
    }
  };

  // 扩缩容StatefulSet
  const handleScale = (record: StatefulSetInfo) => {
    setCurrentStatefulSet(record);
    setScaleModalVisible(true);
  };

  // 创建成功后的回调
  const handleCreateSuccess = () => {
    setCreateModalVisible(false);
    fetchStatefulSets();
  };

  // 扩缩容成功后的回调
  const handleScaleSuccess = () => {
    setScaleModalVisible(false);
    setCurrentStatefulSet(null);
    fetchStatefulSets();
  };

  // 表格列定义
  const columns = [
    {
      title: t('common.name'),
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: StatefulSetInfo) => (
        <a onClick={() => handleViewDetails(record)}>{text}</a>
      ),
    },
    {
      title: t('common.namespace'),
      dataIndex: 'namespace',
      key: 'namespace',
    },
    {
      title: t('common.status'),
      key: 'status',
      render: (_: unknown, record: StatefulSetInfo) => (
        <Tag color={record.replicas === record.readyReplicas ? 'green' : 'orange'}>
          {record.readyReplicas}/{record.replicas}
        </Tag>
      ),
    },
    {
      title: t('statefulsets.serviceName'),
      dataIndex: 'serviceName',
      key: 'serviceName',
    },
    {
      title: t('statefulsets.updateStrategy'),
      dataIndex: 'updateStrategy',
      key: 'updateStrategy',
    },
    {
      title: t('statefulsets.volumeClaimTemplates'),
      dataIndex: 'volumeClaimTemplates',
      key: 'volumeClaimTemplates',
      render: (pvcTemplates: string[]) => (
        <>
          {pvcTemplates && pvcTemplates.length > 0 ? (
            <Space direction="vertical">
              {pvcTemplates.map(pvc => (
                <Tag key={pvc} color="geekblue">{pvc}</Tag>
              ))}
            </Space>
          ) : (
            <span>-</span>
          )}
        </>
      ),
    },
    {
      title: t('common.createTime'),
      dataIndex: 'creationTime',
      key: 'creationTime',
      render: (time: string) => formatDate(time),
    },
    {
      title: t('common.operations'),
      key: 'action',
      render: (_: any, record: StatefulSetInfo) => (
        <Space size="small">
          <Tooltip title={t('common.view')}>
            <Button 
              type="text" 
              icon={<EyeOutlined />} 
              onClick={() => handleViewDetails(record)} 
            />
          </Tooltip>
          <Tooltip title={t('common.scale')}>
            <Button 
              type="text" 
              icon={<ScissorOutlined />} 
              onClick={() => handleScale(record)} 
            />
          </Tooltip>
          <Tooltip title={t('common.restart')}>
            <Popconfirm
              title={t('statefulsets.confirmRestart')}
              onConfirm={() => handleRestart(record)}
              okText={t('common.confirm')}
              cancelText={t('common.cancel')}
            >
              <Button type="text" icon={<ReloadOutlined />} />
            </Popconfirm>
          </Tooltip>
          <Tooltip title={t('common.delete')}>
            <Popconfirm
              title={t('statefulsets.confirmDelete')}
              onConfirm={() => handleDelete(record)}
              okText={t('common.confirm')}
              cancelText={t('common.cancel')}
            >
              <Button type="text" danger icon={<DeleteOutlined />} />
            </Popconfirm>
          </Tooltip>
        </Space>
      ),
    },
  ];

  return (
    <Card
      title={t('statefulsets.management')}
      extra={
        <Space>
          <span>{t('common.cluster')}</span>
          <Select
            value={selectedCluster}
            onChange={handleClusterChange}
            style={{ width: 200 }}
            loading={loading}
          >
            {clusters.map(cluster => (
              <Option key={cluster} value={cluster}>{cluster}</Option>
            ))}
          </Select>
          <span>{t('common.namespace')}:</span>
          <NamespaceSelector
            clusterName={selectedCluster}
            value={namespace}
            onChange={handleNamespaceChange}
          />
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => setCreateModalVisible(true)}
          >
            {t('statefulsets.create')}
          </Button>
          <Button
            icon={<SyncOutlined />}
            onClick={fetchStatefulSets}
          >
            {t('common.refresh')}
          </Button>
        </Space>
      }
    >
      {loading ? (
        <div style={{ textAlign: 'center', padding: '20px 0' }}>
          <Spin />
        </div>
      ) : (
        <Table 
          columns={columns} 
          dataSource={statefulsets} 
          rowKey="name" 
          pagination={{ pageSize: 10 }}
        />
      )}

      {/* 创建StatefulSet模态框 */}
      <CreateStatefulSetModal
        visible={createModalVisible}
        onCancel={() => setCreateModalVisible(false)}
        onSuccess={handleCreateSuccess}
        clusterName={selectedCluster}
        namespace={namespace}
      />

      {/* 扩缩容StatefulSet模态框 */}
      {currentStatefulSet && (
        <ScaleStatefulSetModal
          visible={scaleModalVisible}
          onCancel={() => {
            setScaleModalVisible(false);
            setCurrentStatefulSet(null);
          }}
          onSuccess={handleScaleSuccess}
          clusterName={selectedCluster}
          namespace={currentStatefulSet.namespace}
          statefulsetName={currentStatefulSet.name}
          currentReplicas={currentStatefulSet.replicas}
        />
      )}
    </Card>
  );
};

export default StatefulSets;
