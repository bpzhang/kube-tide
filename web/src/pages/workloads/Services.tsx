import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, Select, Space, message, Button, Popconfirm } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { 
  getServicesByNamespace, 
  createService, 
  deleteService,
  updateService 
} from '../../api/service';
import { getClusterList } from '../../api/cluster';
import CreateServiceModal from '../../components/k8s/service/CreateServiceModal';
import { EditServiceModal } from '../../components/k8s/service/EditServiceModal';
import NamespaceSelector from '../../components/k8s/common/NamespaceSelector';

const { Option } = Select;

const Services: React.FC = () => {
  const [selectedCluster, setSelectedCluster] = useState<string>('');
  const [clusters, setClusters] = useState<string[]>([]);
  const [namespace, setNamespace] = useState<string>('default');
  const [services, setServices] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [currentService, setCurrentService] = useState<any>(null);

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

  const fetchServices = async () => {
    if (!selectedCluster) return;
    
    setLoading(true);
    try {
      const response = await getServicesByNamespace(selectedCluster, namespace);
      if (response.data.code === 0) {
        setServices(response.data.data.services || []);
      } else {
        message.error(response.data.message || '获取服务列表失败');
        setServices([]);
      }
    } catch (err) {
      message.error('获取服务列表失败');
      setServices([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchClusters();
  }, []);

  useEffect(() => {
    if (selectedCluster) {
      fetchServices();
    }
  }, [selectedCluster, namespace]);

  const handleDeleteService = async (name: string) => {
    try {
      await deleteService(selectedCluster, namespace, name);
      message.success('服务删除成功');
      fetchServices();
    } catch (err) {
      message.error('服务删除失败');
    }
  };

  const handleCreateService = async (values: any) => {
    try {
      await createService(selectedCluster, namespace, values);
      message.success('服务创建成功');
      setCreateModalVisible(false);
      fetchServices();
    } catch (err) {
      message.error('服务创建失败');
    }
  };

  const handleUpdateService = async (values: any) => {
    if (!currentService) return;
    
    try {
      await updateService(selectedCluster, namespace, currentService.metadata.name, values);
      message.success('服务更新成功');
      setEditModalVisible(false);
      fetchServices();
    } catch (err) {
      message.error('服务更新失败');
    }
  };

  const showEditModal = (service: any) => {
    setCurrentService(service);
    setEditModalVisible(true);
  };

  const columns = [
    {
      title: '名称',
      dataIndex: ['metadata', 'name'],
      key: 'name',
    },
    {
      title: '命名空间',
      dataIndex: ['metadata', 'namespace'],
      key: 'namespace',
    },
    {
      title: '类型',
      dataIndex: ['spec', 'type'],
      key: 'type',
      render: (type: string) => {
        let color = 'blue';
        if (type === 'ClusterIP') color = 'green';
        else if (type === 'NodePort') color = 'geekblue';
        else if (type === 'LoadBalancer') color = 'purple';
        
        return <Tag color={color}>{type}</Tag>;
      },
    },
    {
      title: '集群IP',
      dataIndex: ['spec', 'clusterIP'],
      key: 'clusterIP',
    },
    {
      title: '外部IP',
      dataIndex: ['spec'],
      key: 'externalIP',
      render: (spec: any) => (spec.externalIPs || []).join(', ') || '-',
    },
    {
      title: '端口',
      dataIndex: 'spec',
      key: 'ports',
      render: (spec: any) => (
        <div>
          {spec.ports?.map((port: any, index: number) => (
            <div key={index}>
              {port.port} → {port.targetPort}
              {port.protocol && ` (${port.protocol})`}
              {port.nodePort && ` NodePort: ${port.nodePort}`}
            </div>
          ))}
        </div>
      ),
    },
    {
      title: '选择器',
      dataIndex: 'spec',
      key: 'selector',
      render: (spec: any) => {
        if (!spec.selector || Object.keys(spec.selector).length === 0) {
          return <Tag color="red">无选择器</Tag>;
        }
        return (
          <div>
            {Object.entries(spec.selector).map(([key, value]: [string, any], index: number) => (
              <Tag color="blue" key={index}>{`${key}: ${value}`}</Tag>
            ))}
          </div>
        );
      },
    },
    {
      title: '操作',
      key: 'action',
      render: (text: string, record: any) => (
        <Space size="middle">
          <Button 
            type="link" 
            icon={<EditOutlined />} 
            onClick={() => showEditModal(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定要删除此服务吗?"
            onConfirm={() => handleDeleteService(record.metadata.name)}
            okText="是"
            cancelText="否"
          >
            <Button type="link" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Card
      title="Service管理"
      extra={
        <Space>
          <span>集群:</span>
          <Select
            value={selectedCluster}
            onChange={setSelectedCluster}
            style={{ width: 200 }}
            loading={loading}
          >
            {clusters.map(cluster => (
              <Option key={cluster} value={cluster}>{cluster}</Option>
            ))}
          </Select>
          <span>命名空间:</span>
          <NamespaceSelector
            clusterName={selectedCluster}
            value={namespace}
            onChange={setNamespace}
          />
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => setCreateModalVisible(true)}
          >
            创建服务
          </Button>
        </Space>
      }
    >
      <Table
        columns={columns}
        dataSource={services}
        rowKey={record => `${record.metadata.namespace}/${record.metadata.name}`}
        loading={loading}
        pagination={{ pageSize: 10 }}
      />

      <CreateServiceModal
        visible={createModalVisible}
        onClose={() => setCreateModalVisible(false)}
        onSubmit={handleCreateService}
        clusterName={selectedCluster}
        namespace={namespace}
      />

      {currentService && (
        <EditServiceModal
          visible={editModalVisible}
          onClose={() => {
            setEditModalVisible(false);
            setCurrentService(null);
          }}
          onSubmit={handleUpdateService}
          service={currentService}
          clusterName={selectedCluster}
          namespace={namespace}
        />
      )}
    </Card>
  );
};

export default Services;