import React, { useEffect, useState } from 'react';
import { Card, Table, Tag, Select, Button, Space, message } from 'antd';
import { ReloadOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useClusterNamespace } from '@/hooks/useClusterNamespace';
import ClusterNamespaceToolbar from '@/components/k8s/common/ClusterNamespaceToolbar';
import { getObservabilityAlerts, ObservabilityAlert } from '@/api/observability';

const Alerts: React.FC = () => {
  const { t } = useTranslation();
  const { selectedCluster, setSelectedCluster, clusters, clustersLoading } = useClusterNamespace(t);
  const [alerts, setAlerts] = useState<ObservabilityAlert[]>([]);
  const [loading, setLoading] = useState(false);
  const [levelFilter, setLevelFilter] = useState<string>('');
  const [categoryFilter, setCategoryFilter] = useState<string>('');

  const fetchAlerts = async () => {
    if (!selectedCluster) return;
    setLoading(true);
    try {
      const response = await getObservabilityAlerts(selectedCluster, 100);
      if (response.data.code === 0) {
        setAlerts(response.data.data.alerts || []);
      } else {
        message.error(response.data.message || t('observability.alerts.fetchFailed'));
      }
    } catch {
      message.error(t('observability.alerts.fetchFailed'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (selectedCluster) fetchAlerts();
  }, [selectedCluster]);

  const filtered = alerts.filter((a) => {
    if (levelFilter && a.level !== levelFilter) return false;
    if (categoryFilter && a.category !== categoryFilter) return false;
    return true;
  });

  const categories = [...new Set(alerts.map((a) => a.category))];

  const columns = [
    {
      title: t('observability.alerts.level'),
      dataIndex: 'level',
      key: 'level',
      width: 100,
      render: (level: string) => (
        <Tag color={level === 'error' ? 'red' : level === 'warning' ? 'orange' : 'blue'}>{level}</Tag>
      ),
    },
    {
      title: t('observability.alerts.category'),
      dataIndex: 'category',
      key: 'category',
      width: 120,
      render: (v: string) => <Tag>{v}</Tag>,
    },
    {
      title: t('common.namespace'),
      dataIndex: 'namespace',
      key: 'namespace',
      width: 140,
      render: (v: string) => v || '—',
    },
    {
      title: t('common.name'),
      dataIndex: 'name',
      key: 'name',
      width: 180,
    },
    {
      title: t('observability.alerts.message'),
      dataIndex: 'message',
      key: 'message',
      ellipsis: true,
    },
  ];

  return (
    <Card title={t('observability.alerts.title')}>
      <ClusterNamespaceToolbar
        clusters={clusters}
        selectedCluster={selectedCluster}
        onClusterChange={setSelectedCluster}
        namespace="all"
        onNamespaceChange={() => {}}
        loading={clustersLoading}
        showNamespace={false}
        extra={
          <Button icon={<ReloadOutlined />} onClick={fetchAlerts} loading={loading}>
            {t('common.refresh')}
          </Button>
        }
      />

      <Space style={{ margin: '16px 0' }}>
        <Select
          allowClear
          placeholder={t('observability.alerts.filterLevel')}
          style={{ width: 140 }}
          value={levelFilter || undefined}
          onChange={(v) => setLevelFilter(v || '')}
        >
          <Select.Option value="error">error</Select.Option>
          <Select.Option value="warning">warning</Select.Option>
          <Select.Option value="info">info</Select.Option>
        </Select>
        <Select
          allowClear
          placeholder={t('observability.alerts.filterCategory')}
          style={{ width: 160 }}
          value={categoryFilter || undefined}
          onChange={(v) => setCategoryFilter(v || '')}
        >
          {categories.map((c) => (
            <Select.Option key={c} value={c}>
              {c}
            </Select.Option>
          ))}
        </Select>
      </Space>

      <Table
        columns={columns}
        dataSource={filtered}
        rowKey={(row, idx) => `${row.category}-${row.namespace}-${row.name}-${idx}`}
        loading={loading}
        pagination={{ pageSize: 20, showSizeChanger: true }}
        locale={{ emptyText: t('observability.alerts.empty') }}
      />
    </Card>
  );
};

export default Alerts;
