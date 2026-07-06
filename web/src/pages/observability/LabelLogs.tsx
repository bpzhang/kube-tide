import React, { useState } from 'react';
import { Card, Form, Input, InputNumber, Button, Collapse, Typography, Alert, Space, message } from 'antd';
import { SearchOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useClusterNamespace } from '@/hooks/useClusterNamespace';
import ClusterNamespaceToolbar from '@/components/k8s/common/ClusterNamespaceToolbar';
import { getLogsByLabelSelector, PodLogEntry } from '@/api/pod';

const { Panel } = Collapse;
const { Text } = Typography;

const LabelLogs: React.FC = () => {
  const { t } = useTranslation();
  const { selectedCluster, setSelectedCluster, clusters, namespace, setNamespace, clustersLoading } =
    useClusterNamespace(t);
  const [logs, setLogs] = useState<PodLogEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();

  const handleSearch = async (values: {
    labelSelector: string;
    container?: string;
    tailLines?: number;
    concurrencyLimit?: number;
  }) => {
    if (!selectedCluster) return;
    setLoading(true);
    try {
      const response = await getLogsByLabelSelector(selectedCluster, namespace, {
        labelSelector: values.labelSelector,
        container: values.container,
        tailLines: values.tailLines || 100,
        concurrencyLimit: values.concurrencyLimit || 5,
      });
      if (response.data.code === 0) {
        setLogs(response.data.data.logs || []);
        if ((response.data.data.logs || []).length === 0) {
          message.info(t('labelLogs.noPods'));
        }
      } else {
        message.error(response.data.message || t('labelLogs.fetchFailed'));
        setLogs([]);
      }
    } catch {
      message.error(t('labelLogs.fetchFailed'));
      setLogs([]);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card
      title={t('labelLogs.management')}
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
        onFinish={handleSearch}
        initialValues={{ tailLines: 100, concurrencyLimit: 5 }}
        style={{ marginBottom: 16 }}
      >
        <Form.Item name="labelSelector" label={t('labelLogs.labelSelector')} rules={[{ required: true }]}>
          <Input placeholder="app=nginx" style={{ width: 220 }} />
        </Form.Item>
        <Form.Item name="container" label={t('labelLogs.container')}>
          <Input placeholder={t('labelLogs.containerOptional')} style={{ width: 160 }} />
        </Form.Item>
        <Form.Item name="tailLines" label={t('labelLogs.tailLines')}>
          <InputNumber min={10} max={5000} />
        </Form.Item>
        <Form.Item name="concurrencyLimit" label={t('labelLogs.concurrency')}>
          <InputNumber min={1} max={20} />
        </Form.Item>
        <Form.Item>
          <Button type="primary" htmlType="submit" icon={<SearchOutlined />} loading={loading}>
            {t('labelLogs.fetch')}
          </Button>
        </Form.Item>
      </Form>

      <Alert type="info" showIcon message={t('labelLogs.hint')} style={{ marginBottom: 16 }} />

      {logs.length > 0 ? (
        <Collapse defaultActiveKey={logs.map((l) => l.podName)}>
          {logs.map((entry) => (
            <Panel
              key={entry.podName}
              header={
                <Space>
                  <Text strong>{entry.podName}</Text>
                  {entry.container && <Text type="secondary">({entry.container})</Text>}
                  {entry.error && <Text type="danger">{entry.error}</Text>}
                </Space>
              }
            >
              {entry.error ? (
                <Alert type="error" message={entry.error} />
              ) : (
                <pre style={{ maxHeight: 400, overflow: 'auto', fontSize: 12, margin: 0 }}>
                  {entry.logs || t('common.noData')}
                </pre>
              )}
            </Panel>
          ))}
        </Collapse>
      ) : (
        !loading && <div style={{ textAlign: 'center', padding: 24 }}>{t('labelLogs.empty')}</div>
      )}
    </Card>
  );
};

export default LabelLogs;
