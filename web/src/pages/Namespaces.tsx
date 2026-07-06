import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, Space, message, Button, Popconfirm, Modal, Form, Input } from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useClusterNamespace } from '@/hooks/useClusterNamespace';
import ClusterNamespaceToolbar from '@/components/k8s/common/ClusterNamespaceToolbar';
import {
  getNamespaceList,
  createNamespace,
  deleteNamespace,
  NamespaceInfo,
} from '@/api/namespace';

const Namespaces: React.FC = () => {
  const { t } = useTranslation();
  const { selectedCluster, setSelectedCluster, clusters, clustersLoading } = useClusterNamespace(t);
  const [items, setItems] = useState<NamespaceInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [form] = Form.useForm();

  const fetchItems = async () => {
    if (!selectedCluster) return;
    setLoading(true);
    try {
      const response = await getNamespaceList(selectedCluster);
      if (response.data.code === 0) {
        setItems(response.data.data.items || response.data.data.namespaces.map((name) => ({
          name,
          status: 'Active',
          creationTime: '',
        })));
      } else {
        message.error(response.data.message || t('namespaces.fetchFailed'));
        setItems([]);
      }
    } catch {
      message.error(t('namespaces.fetchFailed'));
      setItems([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (selectedCluster) fetchItems();
  }, [selectedCluster]);

  const handleCreate = async (values: { name: string }) => {
    try {
      await createNamespace(selectedCluster, { name: values.name });
      message.success(t('namespaces.createSuccess'));
      setModalVisible(false);
      form.resetFields();
      fetchItems();
    } catch {
      message.error(t('namespaces.createFailed'));
    }
  };

  const handleDelete = async (name: string) => {
    try {
      await deleteNamespace(selectedCluster, name);
      message.success(t('namespaces.deleteSuccess'));
      fetchItems();
    } catch {
      message.error(t('namespaces.deleteFailed'));
    }
  };

  const columns = [
    { title: t('common.name'), dataIndex: 'name', key: 'name' },
    {
      title: t('common.status'),
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'Active' ? 'success' : 'default'}>{status || '-'}</Tag>
      ),
    },
    {
      title: t('common.labels'),
      dataIndex: 'labels',
      key: 'labels',
      render: (labels: Record<string, string>) =>
        labels && Object.keys(labels).length > 0 ? (
          <Space wrap>
            {Object.entries(labels).map(([k, v]) => (
              <Tag key={k}>{`${k}=${v}`}</Tag>
            ))}
          </Space>
        ) : '-',
    },
    {
      title: t('common.createTime'),
      dataIndex: 'creationTime',
      key: 'creationTime',
      render: (v: string) => (v ? new Date(v).toLocaleString() : '-'),
    },
    {
      title: t('common.operations'),
      key: 'action',
      render: (_: unknown, record: NamespaceInfo) => (
        <Popconfirm
          title={t('namespaces.deleteConfirm')}
          onConfirm={() => handleDelete(record.name)}
          disabled={['default', 'kube-system', 'kube-public'].includes(record.name)}
        >
          <Button
            type="link"
            danger
            icon={<DeleteOutlined />}
            disabled={['default', 'kube-system', 'kube-public'].includes(record.name)}
          >
            {t('common.delete')}
          </Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <Card
      title={t('namespaces.management')}
      extra={
        <ClusterNamespaceToolbar
          selectedCluster={selectedCluster}
          clusters={clusters}
          namespace=""
          onClusterChange={setSelectedCluster}
          onNamespaceChange={() => {}}
          loading={clustersLoading}
          showNamespace={false}
          extra={
            <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalVisible(true)}>
              {t('namespaces.create')}
            </Button>
          }
        />
      }
    >
      <Table columns={columns} dataSource={items} rowKey="name" loading={loading} pagination={{ pageSize: 10 }} />

      <Modal title={t('namespaces.create')} open={modalVisible} onCancel={() => setModalVisible(false)} onOk={() => form.submit()}>
        <Form form={form} layout="vertical" onFinish={handleCreate}>
          <Form.Item name="name" label={t('common.name')} rules={[{ required: true }]}>
            <Input placeholder="my-namespace" />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
};

export default Namespaces;
