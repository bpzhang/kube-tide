import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, message, Button, Popconfirm, Modal, Form, Input, InputNumber, Select } from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useClusterNamespace } from '@/hooks/useClusterNamespace';
import ClusterNamespaceToolbar from '@/components/k8s/common/ClusterNamespaceToolbar';
import { listHPAs, createHPA, deleteHPA, HPAInfo } from '@/api/hpa';

const HPAs: React.FC = () => {
  const { t } = useTranslation();
  const { selectedCluster, setSelectedCluster, clusters, namespace, setNamespace, clustersLoading } =
    useClusterNamespace(t);
  const [items, setItems] = useState<HPAInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [form] = Form.useForm();

  const fetchItems = async () => {
    if (!selectedCluster) return;
    setLoading(true);
    try {
      const response = await listHPAs(selectedCluster, namespace);
      if (response.data.code === 0) {
        setItems(response.data.data.hpas || []);
      } else {
        message.error(response.data.message || t('hpas.fetchFailed'));
        setItems([]);
      }
    } catch {
      message.error(t('hpas.fetchFailed'));
      setItems([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (selectedCluster) fetchItems();
  }, [selectedCluster, namespace]);

  const handleCreate = async (values: Record<string, unknown>) => {
    try {
      await createHPA(selectedCluster, namespace, {
        name: values.name,
        minReplicas: values.minReplicas,
        maxReplicas: values.maxReplicas,
        targetRef: { kind: values.targetKind, name: values.targetName },
        metrics: [{
          type: 'Resource',
          resourceName: values.metricResource,
          targetUtilization: values.targetUtilization,
        }],
      });
      message.success(t('hpas.createSuccess'));
      setModalVisible(false);
      form.resetFields();
      fetchItems();
    } catch {
      message.error(t('hpas.createFailed'));
    }
  };

  const handleDelete = async (name: string) => {
    try {
      await deleteHPA(selectedCluster, namespace, name);
      message.success(t('hpas.deleteSuccess'));
      fetchItems();
    } catch {
      message.error(t('hpas.deleteFailed'));
    }
  };

  const columns = [
    { title: t('common.name'), dataIndex: 'name', key: 'name' },
    { title: t('common.namespace'), dataIndex: 'namespace', key: 'namespace' },
    {
      title: t('hpas.target'),
      key: 'targetRef',
      render: (_: unknown, r: HPAInfo) => `${r.targetRef?.kind}/${r.targetRef?.name}`,
    },
    {
      title: t('hpas.replicas'),
      key: 'replicas',
      render: (_: unknown, r: HPAInfo) =>
        `${r.currentReplicas}/${r.desiredReplicas} (${r.minReplicas ?? 1}-${r.maxReplicas})`,
    },
    {
      title: t('hpas.metrics'),
      dataIndex: 'metrics',
      key: 'metrics',
      render: (metrics: HPAInfo['metrics']) =>
        (metrics || []).map((m, i) => (
          <Tag key={i}>{m.type}{m.utilization != null ? `: ${m.utilization}%` : ''}</Tag>
        )),
    },
    {
      title: t('common.operations'),
      key: 'action',
      render: (_: unknown, record: HPAInfo) => (
        <Popconfirm title={t('hpas.deleteConfirm')} onConfirm={() => handleDelete(record.name)}>
          <Button type="link" danger icon={<DeleteOutlined />}>{t('common.delete')}</Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <Card
      title={t('hpas.management')}
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
              {t('hpas.create')}
            </Button>
          }
        />
      }
    >
      <Table columns={columns} dataSource={items} rowKey={(r) => `${r.namespace}/${r.name}`} loading={loading} pagination={{ pageSize: 10 }} />

      <Modal title={t('hpas.create')} open={modalVisible} onCancel={() => setModalVisible(false)} onOk={() => form.submit()} width={520}>
        <Form form={form} layout="vertical" onFinish={handleCreate} initialValues={{ targetKind: 'Deployment', metricResource: 'cpu', targetUtilization: 80, minReplicas: 1, maxReplicas: 10 }}>
          <Form.Item name="name" label={t('common.name')} rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="targetKind" label={t('hpas.targetKind')} rules={[{ required: true }]}>
            <Select options={[{ value: 'Deployment', label: 'Deployment' }, { value: 'StatefulSet', label: 'StatefulSet' }]} />
          </Form.Item>
          <Form.Item name="targetName" label={t('hpas.targetName')} rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="minReplicas" label={t('hpas.minReplicas')}>
            <InputNumber min={1} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="maxReplicas" label={t('hpas.maxReplicas')} rules={[{ required: true }]}>
            <InputNumber min={1} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="metricResource" label={t('hpas.metricResource')}>
            <Select options={[{ value: 'cpu', label: 'CPU' }, { value: 'memory', label: 'Memory' }]} />
          </Form.Item>
          <Form.Item name="targetUtilization" label={t('hpas.targetUtilization')}>
            <InputNumber min={1} max={100} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
};

export default HPAs;
