import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, message, Button, Popconfirm, Modal, Form, Input } from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useClusterNamespace } from '@/hooks/useClusterNamespace';
import ClusterNamespaceToolbar from '@/components/k8s/common/ClusterNamespaceToolbar';
import { listResourceQuotas, createResourceQuota, deleteResourceQuota, ResourceQuotaInfo } from '@/api/resourcequota';

const ResourceQuotas: React.FC = () => {
  const { t } = useTranslation();
  const { selectedCluster, setSelectedCluster, clusters, namespace, setNamespace, clustersLoading } =
    useClusterNamespace(t);
  const [items, setItems] = useState<ResourceQuotaInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [form] = Form.useForm();

  const fetchItems = async () => {
    if (!selectedCluster) return;
    setLoading(true);
    try {
      const response = await listResourceQuotas(selectedCluster, namespace);
      if (response.data.code === 0) {
        setItems(response.data.data.resourcequotas || []);
      } else {
        message.error(response.data.message || t('resourceQuotas.fetchFailed'));
        setItems([]);
      }
    } catch {
      message.error(t('resourceQuotas.fetchFailed'));
      setItems([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (selectedCluster) fetchItems();
  }, [selectedCluster, namespace]);

  const parseHard = (text: string): Record<string, string> => {
    const result: Record<string, string> = {};
    text.split('\n').forEach((line) => {
      const trimmed = line.trim();
      if (!trimmed) return;
      const idx = trimmed.indexOf('=');
      if (idx > 0) result[trimmed.slice(0, idx)] = trimmed.slice(idx + 1);
    });
    return result;
  };

  const handleCreate = async (values: { name: string; hard: string }) => {
    try {
      await createResourceQuota(selectedCluster, namespace, { name: values.name, hard: parseHard(values.hard) });
      message.success(t('resourceQuotas.createSuccess'));
      setModalVisible(false);
      form.resetFields();
      fetchItems();
    } catch {
      message.error(t('resourceQuotas.createFailed'));
    }
  };

  const handleDelete = async (name: string) => {
    try {
      await deleteResourceQuota(selectedCluster, namespace, name);
      message.success(t('resourceQuotas.deleteSuccess'));
      fetchItems();
    } catch {
      message.error(t('resourceQuotas.deleteFailed'));
    }
  };

  const columns = [
    { title: t('common.name'), dataIndex: 'name', key: 'name' },
    { title: t('common.namespace'), dataIndex: 'namespace', key: 'namespace' },
    {
      title: t('resourceQuotas.hard'),
      dataIndex: 'hard',
      key: 'hard',
      render: (hard: Record<string, string>) =>
        hard ? Object.entries(hard).map(([k, v]) => <Tag key={k}>{`${k}=${v}`}</Tag>) : '-',
    },
    {
      title: t('resourceQuotas.used'),
      dataIndex: 'used',
      key: 'used',
      render: (used: Record<string, string>) =>
        used ? Object.entries(used).map(([k, v]) => <Tag key={k} color="blue">{`${k}=${v}`}</Tag>) : '-',
    },
    {
      title: t('common.operations'),
      key: 'action',
      render: (_: unknown, record: ResourceQuotaInfo) => (
        <Popconfirm title={t('resourceQuotas.deleteConfirm')} onConfirm={() => handleDelete(record.name)}>
          <Button type="link" danger icon={<DeleteOutlined />}>{t('common.delete')}</Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <Card
      title={t('resourceQuotas.management')}
      extra={
        <ClusterNamespaceToolbar
          selectedCluster={selectedCluster}
          clusters={clusters}
          namespace={namespace}
          onClusterChange={setSelectedCluster}
          onNamespaceChange={setNamespace}
          loading={clustersLoading}
          extra={
            <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalVisible(true)}>
              {t('resourceQuotas.create')}
            </Button>
          }
        />
      }
    >
      <Table columns={columns} dataSource={items} rowKey={(r) => `${r.namespace}/${r.name}`} loading={loading} pagination={{ pageSize: 10 }} />

      <Modal title={t('resourceQuotas.create')} open={modalVisible} onCancel={() => setModalVisible(false)} onOk={() => form.submit()}>
        <Form form={form} layout="vertical" onFinish={handleCreate} initialValues={{ hard: 'pods=10\ncpu=4\nmemory=8Gi' }}>
          <Form.Item name="name" label={t('common.name')} rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="hard" label={t('resourceQuotas.hard')} rules={[{ required: true }]} extra={t('resourceQuotas.hardHint')}>
            <Input.TextArea rows={5} style={{ fontFamily: 'monospace' }} />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
};

export default ResourceQuotas;
