import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, Space, message, Button, Popconfirm, Modal, Form, Input, Select } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, EyeOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useClusterNamespace } from '@/hooks/useClusterNamespace';
import ClusterNamespaceToolbar from '@/components/k8s/common/ClusterNamespaceToolbar';
import {
  listSecretsByNamespace,
  createSecret,
  updateSecret,
  deleteSecret,
  getSecret,
  SecretInfo,
} from '@/api/secret';

const Secrets: React.FC = () => {
  const { t } = useTranslation();
  const { selectedCluster, setSelectedCluster, clusters, namespace, setNamespace, clustersLoading } =
    useClusterNamespace(t);
  const [items, setItems] = useState<SecretInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [viewModalVisible, setViewModalVisible] = useState(false);
  const [editing, setEditing] = useState<SecretInfo | null>(null);
  const [viewData, setViewData] = useState<Record<string, string>>({});
  const [form] = Form.useForm();

  const fetchItems = async () => {
    if (!selectedCluster) return;
    setLoading(true);
    try {
      const response = await listSecretsByNamespace(selectedCluster, namespace);
      if (response.data.code === 0) {
        setItems(response.data.data.secrets || []);
      } else {
        message.error(response.data.message || t('secrets.fetchFailed'));
        setItems([]);
      }
    } catch {
      message.error(t('secrets.fetchFailed'));
      setItems([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (selectedCluster) fetchItems();
  }, [selectedCluster, namespace]);

  const parseDataJson = (text: string): Record<string, string> => {
    const result: Record<string, string> = {};
    text.split('\n').forEach((line) => {
      const trimmed = line.trim();
      if (!trimmed) return;
      const idx = trimmed.indexOf('=');
      if (idx > 0) result[trimmed.slice(0, idx)] = trimmed.slice(idx + 1);
    });
    return result;
  };

  const openCreate = () => {
    setEditing(null);
    form.resetFields();
    form.setFieldsValue({ type: 'Opaque', dataJson: 'key=value' });
    setModalVisible(true);
  };

  const openEdit = async (record: SecretInfo) => {
    try {
      const response = await getSecret(selectedCluster, namespace, record.name);
      if (response.data.code !== 0) {
        message.error(response.data.message || t('secrets.fetchFailed'));
        return;
      }
      const detail = response.data.data.secret;
      setEditing(record);
      form.setFieldsValue({
        name: record.name,
        type: record.type,
        dataJson: Object.entries(detail?.data || {}).map(([k, v]) => `${k}=${v}`).join('\n'),
      });
      setModalVisible(true);
    } catch {
      message.error(t('secrets.fetchFailed'));
    }
  };

  const openView = async (record: SecretInfo) => {
    try {
      const response = await getSecret(selectedCluster, namespace, record.name);
      if (response.data.code !== 0) {
        message.error(response.data.message || t('secrets.fetchFailed'));
        return;
      }
      setViewData(response.data.data.secret?.data || {});
      setViewModalVisible(true);
    } catch {
      message.error(t('secrets.fetchFailed'));
    }
  };

  const handleSubmit = async (values: { name: string; type: string; dataJson: string }) => {
    const stringData = parseDataJson(values.dataJson);
    try {
      if (editing) {
        await updateSecret(selectedCluster, namespace, editing.name, { stringData });
        message.success(t('secrets.updateSuccess'));
      } else {
        await createSecret(selectedCluster, namespace, {
          name: values.name,
          type: values.type,
          stringData,
        });
        message.success(t('secrets.createSuccess'));
      }
      setModalVisible(false);
      fetchItems();
    } catch {
      message.error(editing ? t('secrets.updateFailed') : t('secrets.createFailed'));
    }
  };

  const handleDelete = async (name: string) => {
    try {
      await deleteSecret(selectedCluster, namespace, name);
      message.success(t('secrets.deleteSuccess'));
      fetchItems();
    } catch {
      message.error(t('secrets.deleteFailed'));
    }
  };

  const columns = [
    { title: t('common.name'), dataIndex: 'name', key: 'name' },
    { title: t('common.namespace'), dataIndex: 'namespace', key: 'namespace' },
    {
      title: t('secrets.type'),
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => <Tag color="purple">{type}</Tag>,
    },
    {
      title: t('secrets.dataKeys'),
      dataIndex: 'dataKeys',
      key: 'dataKeys',
      render: (keys: string[]) => (
        <Space wrap>{(keys || []).map((k) => <Tag key={k}>{k}</Tag>)}</Space>
      ),
    },
    { title: t('common.createTime'), dataIndex: 'creationTime', key: 'creationTime' },
    {
      title: t('common.operations'),
      key: 'action',
      render: (_: unknown, record: SecretInfo) => (
        <Space>
          <Button type="link" icon={<EyeOutlined />} onClick={() => openView(record)}>
            {t('common.view')}
          </Button>
          <Button type="link" icon={<EditOutlined />} onClick={() => openEdit(record)}>
            {t('common.edit')}
          </Button>
          <Popconfirm title={t('secrets.deleteConfirm')} onConfirm={() => handleDelete(record.name)}>
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
      title={t('secrets.management')}
      extra={
        <ClusterNamespaceToolbar
          selectedCluster={selectedCluster}
          clusters={clusters}
          namespace={namespace}
          onClusterChange={setSelectedCluster}
          onNamespaceChange={setNamespace}
          loading={clustersLoading}
          extra={
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
              {t('secrets.create')}
            </Button>
          }
        />
      }
    >
      <Table columns={columns} dataSource={items} rowKey={(r) => `${r.namespace}/${r.name}`} loading={loading} pagination={{ pageSize: 10 }} />

      <Modal title={editing ? t('secrets.edit') : t('secrets.create')} open={modalVisible} onCancel={() => setModalVisible(false)} onOk={() => form.submit()} width={600}>
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="name" label={t('common.name')} rules={[{ required: !editing }]}>
            <Input disabled={!!editing} />
          </Form.Item>
          <Form.Item name="type" label={t('secrets.type')}>
            <Select options={[
              { value: 'Opaque', label: 'Opaque' },
              { value: 'kubernetes.io/tls', label: 'TLS' },
              { value: 'kubernetes.io/dockerconfigjson', label: 'Docker Config' },
            ]} />
          </Form.Item>
          <Form.Item name="dataJson" label={t('secrets.data')} rules={[{ required: true }]} extra={t('secrets.dataHint')}>
            <Input.TextArea rows={8} style={{ fontFamily: 'monospace' }} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal title={t('secrets.viewData')} open={viewModalVisible} onCancel={() => setViewModalVisible(false)} footer={null} width={700}>
        <pre style={{ maxHeight: 400, overflow: 'auto' }}>{JSON.stringify(viewData, null, 2)}</pre>
      </Modal>
    </Card>
  );
};

export default Secrets;
