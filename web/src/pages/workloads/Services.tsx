import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, Select, Space, message, Button, Popconfirm } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { 
  getServicesByNamespace, 
  createService, 
  deleteService 
} from '../../api/service';
import { getClusterList } from '../../api/cluster';
import CreateServiceModal from '../../components/k8s/service/CreateServiceModal';
import { EditServiceModal } from '../../components/k8s/service/EditServiceModal';

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
      // 每30秒刷新一次
      const timer = setInterval(fetchServices, 30000);
      return () => clearInterval(timer);
    }
  }, [selectedCluster, namespace]);

  const handleCreateService = async (values: any) => {
    try {
      await createService(selectedCluster, namespace, values);
      message.success('服务创建成功');
      fetchServices();
    } catch (error) {
      message.error('创建服务失败: ' + (error instanceof Error ? error.message : '未知错误'));
    }
  };

  const handleDeleteService = async (service: any) => {
    try {
      await deleteService(selectedCluster, service.metadata.namespace, service.metadata.name);
      message.success('服务删除成功');
      fetchServices();
    } catch (error) {
      message.error('删除服务失败: ' + (error instanceof Error ? error.message : '未知错误'));
    }
  };

  const columns = [
    {
      title: '名称',
      dataIndex: 'metadata',
      key: 'name',
      render: (metadata: any) => metadata.name,
    },
    {
      title: '命名空间',
      dataIndex: 'metadata',
      key: 'namespace',
      render: (metadata: any) => metadata.namespace,
    },
    {
      title: '类型',
      dataIndex: 'spec',
      key: 'type',
      render: (spec: any) => <Tag color="blue">{spec.type}</Tag>,
    },
    {
      title: 'Cluster IP',
      dataIndex: 'spec',
      key: 'clusterIP',
      render: (spec: any) => spec.clusterIP,
    },
    {
      title: '外部IP',
      dataIndex: 'spec',
      key: 'externalIPs',
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
              <Tag color="green" key={index}>
                {key}: {value}
              </Tag>
            ))}
          </div>
        );
      },
    },
    {
      title: '创建时间',
      dataIndex: 'metadata',
      key: 'creationTimestamp',
      render: (metadata: any) => new Date(metadata.creationTimestamp).toLocaleString(),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: any) => (
        <Space>
          <Button
            type="link"
            icon={<EditOutlined />}
            onClick={() => {
              setCurrentService(record);
              setEditModalVisible(true);
            }}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定要删除此服务吗？"
            onConfirm={() => handleDeleteService(record)}
            okText="确定"
            cancelText="取消"
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
          <Select 
            value={namespace} 
            onChange={setNamespace}
            style={{ width: 200 }}
          >
            <Option value="default">default</Option>
            <Option value="kube-system">kube-system</Option>
          </Select>
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
        namespace={namespace}
      />

      <EditServiceModal
        visible={editModalVisible}
        onClose={() => setEditModalVisible(false)}
        service={currentService}
        clusterName={selectedCluster}
        onSuccess={fetchServices}
      />
    </Card>
  );
};

export default Services;