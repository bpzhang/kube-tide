import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, Space, message, Button, Popconfirm, Modal, Form, Input } from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useClusterNamespace } from '@/hooks/useClusterNamespace';
import ClusterNamespaceToolbar from '@/components/k8s/common/ClusterNamespaceToolbar';
import { listJobs, createJob, deleteJob, JobInfo } from '@/api/job';

const Jobs: React.FC = () => {
  const { t } = useTranslation();
  const { selectedCluster, setSelectedCluster, clusters, namespace, setNamespace, clustersLoading } =
    useClusterNamespace(t);
  const [items, setItems] = useState<JobInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [form] = Form.useForm();

  const fetchItems = async () => {
    if (!selectedCluster) return;
    setLoading(true);
    try {
      const response = await listJobs(selectedCluster, namespace);
      if (response.data.code === 0) {
        setItems(response.data.data.jobs || []);
      } else {
        message.error(response.data.message || t('jobs.fetchFailed'));
        setItems([]);
      }
    } catch {
      message.error(t('jobs.fetchFailed'));
      setItems([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (selectedCluster) fetchItems();
  }, [selectedCluster, namespace]);

  const handleCreate = async (values: { name: string; image: string; command?: string }) => {
    try {
      await createJob(selectedCluster, namespace, {
        name: values.name,
        containers: [{
          name: values.name,
          image: values.image,
          command: values.command ? values.command.split(' ') : undefined,
        }],
      });
      message.success(t('jobs.createSuccess'));
      setModalVisible(false);
      form.resetFields();
      fetchItems();
    } catch {
      message.error(t('jobs.createFailed'));
    }
  };

  const handleDelete = async (name: string) => {
    try {
      await deleteJob(selectedCluster, namespace, name);
      message.success(t('jobs.deleteSuccess'));
      fetchItems();
    } catch {
      message.error(t('jobs.deleteFailed'));
    }
  };

  const columns = [
    { title: t('common.name'), dataIndex: 'name', key: 'name' },
    { title: t('common.namespace'), dataIndex: 'namespace', key: 'namespace' },
    {
      title: t('jobs.status'),
      key: 'status',
      render: (_: unknown, r: JobInfo) => (
        <Space>
          <Tag color="green">{t('jobs.succeeded')}: {r.succeeded}</Tag>
          <Tag color="red">{t('jobs.failed')}: {r.failed}</Tag>
          <Tag color="blue">{t('jobs.active')}: {r.active}</Tag>
        </Space>
      ),
    },
    {
      title: t('jobs.images'),
      dataIndex: 'images',
      key: 'images',
      render: (images: string[]) => (images || []).join(', ') || '-',
    },
    {
      title: t('common.operations'),
      key: 'action',
      render: (_: unknown, record: JobInfo) => (
        <Popconfirm title={t('jobs.deleteConfirm')} onConfirm={() => handleDelete(record.name)}>
          <Button type="link" danger icon={<DeleteOutlined />}>{t('common.delete')}</Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <Card
      title={t('jobs.management')}
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
              {t('jobs.create')}
            </Button>
          }
        />
      }
    >
      <Table columns={columns} dataSource={items} rowKey={(r) => `${r.namespace}/${r.name}`} loading={loading} pagination={{ pageSize: 10 }} />

      <Modal title={t('jobs.create')} open={modalVisible} onCancel={() => setModalVisible(false)} onOk={() => form.submit()}>
        <Form form={form} layout="vertical" onFinish={handleCreate}>
          <Form.Item name="name" label={t('common.name')} rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="image" label={t('jobs.image')} rules={[{ required: true }]}>
            <Input placeholder="nginx:latest" />
          </Form.Item>
          <Form.Item name="command" label={t('jobs.command')}>
            <Input placeholder="echo hello" />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
};

export default Jobs;
