import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, Space, message, Button, Popconfirm, Modal, Form, Input } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, EyeOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useClusterNamespace } from '@/hooks/useClusterNamespace';
import ClusterNamespaceToolbar from '@/components/k8s/common/ClusterNamespaceToolbar';
import {
  listConfigMapsByNamespace,
  createConfigMap,
  updateConfigMap,
  deleteConfigMap,
  getConfigMap,
  ConfigMapInfo,
} from '@/api/configmap';

const ConfigMaps: React.FC = () => {
  const { t } = useTranslation();
  const { selectedCluster, setSelectedCluster, clusters, namespace, setNamespace, clustersLoading } =
    useClusterNamespace(t);
  const [items, setItems] = useState<ConfigMapInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [viewModalVisible, setViewModalVisible] = useState(false);
  const [editing, setEditing] = useState<ConfigMapInfo | null>(null);
  const [viewData, setViewData] = useState<Record<string, string>>({});
  const [form] = Form.useForm();

  const fetchItems = async () => {
    if (!selectedCluster) return;
    setLoading(true);
    try {
      const response = await listConfigMapsByNamespace(selectedCluster, namespace);
      if (response.data.code === 0) {
        setItems(response.data.data.configmaps || []);
      } else {
        message.error(response.data.message || t('configMaps.fetchFailed'));
        setItems([]);
      }
    } catch {
      message.error(t('configMaps.fetchFailed'));
      setItems([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (selectedCluster) fetchItems();
  }, [selectedCluster, namespace]);

  const openCreate = () => {
    setEditing(null);
    form.resetFields();
    form.setFieldsValue({ dataJson: 'key=value' });
    setModalVisible(true);
  };

  const openEdit = async (record: ConfigMapInfo) => {
    try {
      const response = await getConfigMap(selectedCluster, namespace, record.name);
      if (response.data.code !== 0) {
        message.error(response.data.message || t('configMaps.fetchFailed'));
        return;
      }
      const detail = response.data.data.configmap;
      const data = detail?.data || {};
      setEditing(record);
      form.setFieldsValue({
        name: record.name,
        dataJson: Object.entries(data).map(([k, v]) => `${k}=${v}`).join('\n'),
      });
      setModalVisible(true);
    } catch {
      message.error(t('configMaps.fetchFailed'));
    }
  };

  const openView = async (record: ConfigMapInfo) => {
    try {
      const response = await getConfigMap(selectedCluster, namespace, record.name);
      if (response.data.code !== 0) {
        message.error(response.data.message || t('configMaps.fetchFailed'));
        return;
      }
      setViewData(response.data.data.configmap?.data || {});
      setViewModalVisible(true);
    } catch {
      message.error(t('configMaps.fetchFailed'));
    }
  };

  const parseDataJson = (text: string): Record<string, string> => {
    const result: Record<string, string> = {};
    text.split('\n').forEach((line) => {
      const trimmed = line.trim();
      if (!trimmed) return;
      const idx = trimmed.indexOf('=');
      if (idx > 0) {
        result[trimmed.slice(0, idx)] = trimmed.slice(idx + 1);
      }
    });
    return result;
  };

  const handleSubmit = async (values: { name: string; dataJson: string }) => {
    const data = parseDataJson(values.dataJson);
    try {
      if (editing) {
        await updateConfigMap(selectedCluster, namespace, editing.name, { data });
        message.success(t('configMaps.updateSuccess'));
      } else {
        await createConfigMap(selectedCluster, namespace, { name: values.name, data });
        message.success(t('configMaps.createSuccess'));
      }
      setModalVisible(false);
      fetchItems();
    } catch {
      message.error(editing ? t('configMaps.updateFailed') : t('configMaps.createFailed'));
    }
  };

  const handleDelete = async (name: string) => {
    try {
      await deleteConfigMap(selectedCluster, namespace, name);
      message.success(t('configMaps.deleteSuccess'));
      fetchItems();
    } catch {
      message.error(t('configMaps.deleteFailed'));
    }
  };

  const columns = [
    { title: t('common.name'), dataIndex: 'name', key: 'name' },
    { title: t('common.namespace'), dataIndex: 'namespace', key: 'namespace' },
    {
      title: t('configMaps.dataKeys'),
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
      render: (_: unknown, record: ConfigMapInfo) => (
        <Space>
          <Button type="link" icon={<EyeOutlined />} onClick={() => openView(record)}>
            {t('common.view')}
          </Button>
          <Button type="link" icon={<EditOutlined />} onClick={() => openEdit(record)}>
            {t('common.edit')}
          </Button>
          <Popconfirm title={t('configMaps.deleteConfirm')} onConfirm={() => handleDelete(record.name)}>
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
      title={t('configMaps.management')}
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
              {t('configMaps.create')}
            </Button>
          }
        />
      }
    >
      <Table
        columns={columns}
        dataSource={items}
        rowKey={(r) => `${r.namespace}/${r.name}`}
        loading={loading}
        pagination={{ pageSize: 10 }}
      />

      <Modal
        title={editing ? t('configMaps.edit') : t('configMaps.create')}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => form.submit()}
        width={600}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="name" label={t('common.name')} rules={[{ required: !editing }]}>
            <Input disabled={!!editing} />
          </Form.Item>
          <Form.Item
            name="dataJson"
            label={t('configMaps.data')}
            rules={[{ required: true }]}
            extra={t('configMaps.dataHint')}
          >
            <Input.TextArea rows={8} style={{ fontFamily: 'monospace' }} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={t('configMaps.viewData')}
        open={viewModalVisible}
        onCancel={() => setViewModalVisible(false)}
        footer={null}
        width={700}
      >
        <pre style={{ maxHeight: 400, overflow: 'auto' }}>{JSON.stringify(viewData, null, 2)}</pre>
      </Modal>
    </Card>
  );
};

export default ConfigMaps;
