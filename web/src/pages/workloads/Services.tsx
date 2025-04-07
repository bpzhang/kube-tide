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
import { useTranslation } from 'react-i18next';

const { Option } = Select;

const Services: React.FC = () => {
  const { t } = useTranslation();
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
      message.error(t('clusters.fetchFailed'));
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
        message.error(response.data.message || t('services.fetchFailed'));
        setServices([]);
      }
    } catch (err) {
      message.error(t('services.fetchFailed'));
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
      message.success(t('services.deleteSuccess'));
      fetchServices();
    } catch (err) {
      message.error(t('services.deleteFailed'));
    }
  };

  const handleCreateService = async (values: any) => {
    try {
      await createService(selectedCluster, namespace, values);
      message.success(t('services.createSuccess'));
      setCreateModalVisible(false);
      fetchServices();
    } catch (err) {
      message.error(t('services.createFailed'));
    }
  };

  const handleUpdateService = async (values: any) => {
    if (!currentService) return;
    
    try {
      await updateService(selectedCluster, namespace, currentService.metadata.name, values);
      message.success(t('services.updateSuccess'));
      setEditModalVisible(false);
      fetchServices();
    } catch (err) {
      message.error(t('services.updateFailed'));
    }
  };

  const showEditModal = (service: any) => {
    setCurrentService(service);
    setEditModalVisible(true);
  };

  const columns = [
    {
      title: t('services.name'),
      dataIndex: ['metadata', 'name'],
      key: 'name',
    },
    {
      title: t('services.namespace'),
      dataIndex: ['metadata', 'namespace'],
      key: 'namespace',
    },
    {
      title: t('services.type'),
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
      title: t('services.clusterIp'),
      dataIndex: ['spec', 'clusterIP'],
      key: 'clusterIP',
    },
    {
      title: t('services.externalIp'),
      dataIndex: ['spec'],
      key: 'externalIP',
      render: (spec: any) => (spec.externalIPs || []).join(', ') || '-',
    },
    {
      title: t('services.ports'),
      dataIndex: 'spec',
      key: 'ports',
      render: (spec: any) => (
        <div>
          {spec.ports?.map((port: any, index: number) => (
            <div key={index}>
              {port.port} → {port.targetPort}
              {port.protocol && ` (${port.protocol})`}
              {port.nodePort && ` ${t('services.nodePort')}: ${port.nodePort}`}
            </div>
          ))}
        </div>
      ),
    },
    {
      title: t('services.selectors'),
      dataIndex: 'spec',
      key: 'selector',
      render: (spec: any) => {
        if (!spec.selector || Object.keys(spec.selector).length === 0) {
          return <Tag color="red">{t('services.noSelectors')}</Tag>;
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
      title: t('common.operations'),
      key: 'action',
      render: (text: string, record: any) => (
        <Space size="middle">
          <Button 
            type="link" 
            icon={<EditOutlined />} 
            onClick={() => showEditModal(record)}
          >
            {t('common.edit')}
          </Button>
          <Popconfirm
            title={t('services.deleteConfirm')}
            onConfirm={() => handleDeleteService(record.metadata.name)}
            okText={t('common.yes')}
            cancelText={t('common.no')}
          >
            <Button type="link" danger icon={<DeleteOutlined />}>
              {t('common.delete')}
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Card
      title={t('services.management')}
      extra={
        <Space>
          <span>{t('services.cluster')}</span>
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
          <span>{t('services.namespace')}</span>
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
            {t('services.createService')}
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