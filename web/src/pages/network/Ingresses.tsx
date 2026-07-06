import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, Space, message, Button, Popconfirm, Modal, Form, Input } from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useClusterNamespace } from '@/hooks/useClusterNamespace';
import ClusterNamespaceToolbar from '@/components/k8s/common/ClusterNamespaceToolbar';
import { getIngressesByNamespace, createIngress, deleteIngress, IngressInfo } from '@/api/ingress';

const Ingresses: React.FC = () => {
  const { t } = useTranslation();
  const { selectedCluster, setSelectedCluster, clusters, namespace, setNamespace, clustersLoading } =
    useClusterNamespace(t);
  const [items, setItems] = useState<IngressInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [form] = Form.useForm();

  const fetchItems = async () => {
    if (!selectedCluster) return;
    setLoading(true);
    try {
      const response = await getIngressesByNamespace(selectedCluster, namespace);
      if (response.data.code === 0) {
        setItems(response.data.data.ingresses || []);
      } else {
        message.error(response.data.message || t('ingresses.fetchFailed'));
        setItems([]);
      }
    } catch {
      message.error(t('ingresses.fetchFailed'));
      setItems([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (selectedCluster) fetchItems();
  }, [selectedCluster, namespace]);

  const handleCreate = async (values: {
    name: string;
    host: string;
    path: string;
    serviceName: string;
    servicePort: string;
    ingressClassName?: string;
  }) => {
    try {
      await createIngress(selectedCluster, namespace, {
        name: values.name,
        ingressClassName: values.ingressClassName,
        rules: [{
          host: values.host,
          paths: [{
            path: values.path || '/',
            pathType: 'Prefix',
            backend: { serviceName: values.serviceName, servicePort: values.servicePort },
          }],
        }],
      });
      message.success(t('ingresses.createSuccess'));
      setModalVisible(false);
      form.resetFields();
      fetchItems();
    } catch {
      message.error(t('ingresses.createFailed'));
    }
  };

  const handleDelete = async (name: string) => {
    try {
      await deleteIngress(selectedCluster, namespace, name);
      message.success(t('ingresses.deleteSuccess'));
      fetchItems();
    } catch {
      message.error(t('ingresses.deleteFailed'));
    }
  };

  const columns = [
    { title: t('common.name'), dataIndex: 'name', key: 'name' },
    { title: t('common.namespace'), dataIndex: 'namespace', key: 'namespace' },
    { title: t('ingresses.class'), dataIndex: 'ingressClassName', key: 'ingressClassName', render: (v: string) => v || '-' },
    {
      title: t('ingresses.rules'),
      dataIndex: 'rules',
      key: 'rules',
      render: (rules: IngressInfo['rules']) => (
        <Space direction="vertical" size={0}>
          {(rules || []).flatMap((rule) =>
            (rule.paths || []).map((p, i) => (
              <span key={`${rule.host}-${i}`}>
                {rule.host || '*'}{p.path} → {p.backend?.serviceName}:{p.backend?.servicePort}
              </span>
            )),
          )}
        </Space>
      ),
    },
    {
      title: t('ingresses.tls'),
      dataIndex: 'tls',
      key: 'tls',
      render: (tls: IngressInfo['tls']) =>
        (tls || []).length > 0 ? <Tag color="green">{t('common.enabled')}</Tag> : <Tag>{t('common.disabled')}</Tag>,
    },
    {
      title: t('common.operations'),
      key: 'action',
      render: (_: unknown, record: IngressInfo) => (
        <Popconfirm title={t('ingresses.deleteConfirm')} onConfirm={() => handleDelete(record.name)}>
          <Button type="link" danger icon={<DeleteOutlined />}>{t('common.delete')}</Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <Card
      title={t('ingresses.management')}
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
              {t('ingresses.create')}
            </Button>
          }
        />
      }
    >
      <Table columns={columns} dataSource={items} rowKey={(r) => `${r.namespace}/${r.name}`} loading={loading} pagination={{ pageSize: 10 }} />

      <Modal title={t('ingresses.create')} open={modalVisible} onCancel={() => setModalVisible(false)} onOk={() => form.submit()} width={560}>
        <Form form={form} layout="vertical" onFinish={handleCreate} initialValues={{ path: '/', servicePort: '80' }}>
          <Form.Item name="name" label={t('common.name')} rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="ingressClassName" label={t('ingresses.class')}><Input placeholder="nginx" /></Form.Item>
          <Form.Item name="host" label={t('ingresses.host')} rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="path" label={t('ingresses.path')}><Input /></Form.Item>
          <Form.Item name="serviceName" label={t('ingresses.serviceName')} rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="servicePort" label={t('ingresses.servicePort')} rules={[{ required: true }]}><Input /></Form.Item>
        </Form>
      </Modal>
    </Card>
  );
};

export default Ingresses;
