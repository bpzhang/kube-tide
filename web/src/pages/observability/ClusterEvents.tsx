import React, { useEffect, useState } from 'react';
import { Card, Table, Tag, Button, Form, Input, Select, Space, message } from 'antd';
import { ReloadOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { formatDate } from '@/utils/format';
import { useClusterNamespace } from '@/hooks/useClusterNamespace';
import ClusterNamespaceToolbar from '@/components/k8s/common/ClusterNamespaceToolbar';
import { getClusterEvents } from '@/api/cluster';

const ClusterEventsPage: React.FC = () => {
  const { t } = useTranslation();
  const { selectedCluster, setSelectedCluster, clusters, namespace, setNamespace, clustersLoading } =
    useClusterNamespace(t);
  const [events, setEvents] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();

  const fetchEvents = async (values?: {
    type?: string;
    reason?: string;
    kind?: string;
    involvedObjectName?: string;
    limit?: number;
  }) => {
    if (!selectedCluster) return;
    setLoading(true);
    try {
      const response = await getClusterEvents(selectedCluster, {
        namespace: namespace === 'all' ? undefined : namespace,
        type: values?.type,
        reason: values?.reason,
        kind: values?.kind,
        involvedObjectName: values?.involvedObjectName,
        limit: values?.limit || 100,
      });
      if (response.data.code === 0) {
        setEvents(response.data.data.events || []);
      } else {
        message.error(response.data.message || t('clusterEvents.fetchFailed'));
        setEvents([]);
      }
    } catch {
      message.error(t('clusterEvents.fetchFailed'));
      setEvents([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (selectedCluster) {
      fetchEvents(form.getFieldsValue());
    }
  }, [selectedCluster, namespace]);

  const getEventTypeColor = (type: string) => (type === 'Warning' ? 'orange' : 'green');

  const columns = [
    {
      title: t('events.columns.type'),
      dataIndex: 'type',
      key: 'type',
      width: 100,
      render: (type: string) => <Tag color={getEventTypeColor(type)}>{type}</Tag>,
    },
    {
      title: t('events.columns.reason'),
      dataIndex: 'reason',
      key: 'reason',
      width: 140,
    },
    {
      title: t('clusterEvents.columns.object'),
      key: 'object',
      width: 200,
      render: (_: unknown, row: any) =>
        row.involvedObject ? `${row.involvedObject.kind}/${row.involvedObject.name}` : '-',
    },
    {
      title: t('events.columns.message'),
      dataIndex: 'message',
      key: 'message',
      ellipsis: true,
    },
    {
      title: t('events.columns.namespace'),
      dataIndex: ['metadata', 'namespace'],
      key: 'namespace',
      width: 120,
    },
    {
      title: t('events.columns.lastTimestamp'),
      dataIndex: 'lastTimestamp',
      key: 'lastTimestamp',
      width: 170,
      render: (time: string) => formatDate(time),
    },
    {
      title: t('events.columns.count'),
      dataIndex: 'count',
      key: 'count',
      width: 80,
    },
  ];

  return (
    <Card
      title={t('clusterEvents.title')}
      extra={
        <ClusterNamespaceToolbar
          selectedCluster={selectedCluster}
          clusters={clusters}
          namespace={namespace}
          onClusterChange={setSelectedCluster}
          onNamespaceChange={setNamespace}
          loading={clustersLoading}
        />
      }
    >
      <Form
        form={form}
        layout="inline"
        onFinish={fetchEvents}
        initialValues={{ limit: 100 }}
        style={{ marginBottom: 16 }}
      >
        <Form.Item name="type" label={t('events.columns.type')}>
          <Select allowClear style={{ width: 120 }} placeholder={t('clusterEvents.filters.type')}>
            <Select.Option value="Normal">Normal</Select.Option>
            <Select.Option value="Warning">Warning</Select.Option>
          </Select>
        </Form.Item>
        <Form.Item name="kind" label={t('clusterEvents.filters.kind')}>
          <Input placeholder="Pod" style={{ width: 120 }} allowClear />
        </Form.Item>
        <Form.Item name="reason" label={t('events.columns.reason')}>
          <Input placeholder="Failed" style={{ width: 140 }} allowClear />
        </Form.Item>
        <Form.Item name="involvedObjectName" label={t('clusterEvents.filters.objectName')}>
          <Input placeholder="my-pod" style={{ width: 140 }} allowClear />
        </Form.Item>
        <Form.Item>
          <Space>
            <Button type="primary" htmlType="submit" loading={loading}>
              {t('common.search')}
            </Button>
            <Button icon={<ReloadOutlined />} onClick={() => fetchEvents(form.getFieldsValue())} loading={loading}>
              {t('common.refresh')}
            </Button>
          </Space>
        </Form.Item>
      </Form>

      <Table
        columns={columns}
        dataSource={events}
        rowKey={(row) => `${row.metadata?.uid || ''}-${row.lastTimestamp || ''}`}
        loading={loading}
        pagination={{ pageSize: 20, showSizeChanger: true }}
        scroll={{ x: 'max-content' }}
      />
    </Card>
  );
};

export default ClusterEventsPage;
