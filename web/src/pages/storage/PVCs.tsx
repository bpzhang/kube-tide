import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, message, Button, Popconfirm, Modal, Form, Input, Select } from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useClusterNamespace } from '@/hooks/useClusterNamespace';
import ClusterNamespaceToolbar from '@/components/k8s/common/ClusterNamespaceToolbar';
import { listPVCs, createPVC, deletePVC, PVCInfo } from '@/api/pvc';

const PVCs: React.FC = () => {
  const { t } = useTranslation();
  const { selectedCluster, setSelectedCluster, clusters, namespace, setNamespace, clustersLoading } =
    useClusterNamespace(t);
  const [items, setItems] = useState<PVCInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [form] = Form.useForm();

  const fetchItems = async () => {
    if (!selectedCluster) return;
    setLoading(true);
    try {
      const response = await listPVCs(selectedCluster, namespace);
      if (response.data.code === 0) {
        setItems(response.data.data.pvcs || []);
      } else {
        message.error(response.data.message || t('pvcs.fetchFailed'));
        setItems([]);
      }
    } catch {
      message.error(t('pvcs.fetchFailed'));
      setItems([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (selectedCluster) fetchItems();
  }, [selectedCluster, namespace]);

  const handleCreate = async (values: { name: string; storage: string; storageClassName?: string; accessMode: string }) => {
    try {
      await createPVC(selectedCluster, namespace, {
        name: values.name,
        storage: values.storage,
        storageClassName: values.storageClassName,
        accessModes: [values.accessMode],
      });
      message.success(t('pvcs.createSuccess'));
      setModalVisible(false);
      form.resetFields();
      fetchItems();
    } catch {
      message.error(t('pvcs.createFailed'));
    }
  };

  const handleDelete = async (name: string) => {
    try {
      await deletePVC(selectedCluster, namespace, name);
      message.success(t('pvcs.deleteSuccess'));
      fetchItems();
    } catch {
      message.error(t('pvcs.deleteFailed'));
    }
  };

  const columns = [
    { title: t('common.name'), dataIndex: 'name', key: 'name' },
    { title: t('common.namespace'), dataIndex: 'namespace', key: 'namespace' },
    {
      title: t('common.status'),
      dataIndex: 'status',
      key: 'status',
      render: (s: string) => <Tag color={s === 'Bound' ? 'success' : 'processing'}>{s}</Tag>,
    },
    { title: t('pvcs.capacity'), dataIndex: 'capacity', key: 'capacity', render: (v: string) => v || '-' },
    { title: t('pvcs.storageClass'), dataIndex: 'storageClassName', key: 'storageClassName', render: (v: string) => v || '-' },
    { title: t('pvcs.volume'), dataIndex: 'volumeName', key: 'volumeName', render: (v: string) => v || '-' },
    {
      title: t('common.operations'),
      key: 'action',
      render: (_: unknown, record: PVCInfo) => (
        <Popconfirm title={t('pvcs.deleteConfirm')} onConfirm={() => handleDelete(record.name)}>
          <Button type="link" danger icon={<DeleteOutlined />}>{t('common.delete')}</Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <Card
      title={t('pvcs.management')}
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
              {t('pvcs.create')}
            </Button>
          }
        />
      }
    >
      <Table columns={columns} dataSource={items} rowKey={(r) => `${r.namespace}/${r.name}`} loading={loading} pagination={{ pageSize: 10 }} />

      <Modal title={t('pvcs.create')} open={modalVisible} onCancel={() => setModalVisible(false)} onOk={() => form.submit()}>
        <Form form={form} layout="vertical" onFinish={handleCreate} initialValues={{ accessMode: 'ReadWriteOnce', storage: '1Gi' }}>
          <Form.Item name="name" label={t('common.name')} rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="storage" label={t('pvcs.storage')} rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="storageClassName" label={t('pvcs.storageClass')}><Input /></Form.Item>
          <Form.Item name="accessMode" label={t('pvcs.accessMode')}>
            <Select options={[
              { value: 'ReadWriteOnce', label: 'ReadWriteOnce' },
              { value: 'ReadOnlyMany', label: 'ReadOnlyMany' },
              { value: 'ReadWriteMany', label: 'ReadWriteMany' },
            ]} />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
};

export default PVCs;
