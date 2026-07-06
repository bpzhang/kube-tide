import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, message, Button, Popconfirm, Switch } from 'antd';
import { DeleteOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useClusterNamespace } from '@/hooks/useClusterNamespace';
import ClusterNamespaceToolbar from '@/components/k8s/common/ClusterNamespaceToolbar';
import { listCronJobs, suspendCronJob, deleteCronJob, CronJobInfo } from '@/api/cronjob';

const CronJobs: React.FC = () => {
  const { t } = useTranslation();
  const { selectedCluster, setSelectedCluster, clusters, namespace, setNamespace, clustersLoading } =
    useClusterNamespace(t);
  const [items, setItems] = useState<CronJobInfo[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchItems = async () => {
    if (!selectedCluster) return;
    setLoading(true);
    try {
      const response = await listCronJobs(selectedCluster, namespace);
      if (response.data.code === 0) {
        setItems(response.data.data.cronjobs || []);
      } else {
        message.error(response.data.message || t('cronJobs.fetchFailed'));
        setItems([]);
      }
    } catch {
      message.error(t('cronJobs.fetchFailed'));
      setItems([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (selectedCluster) fetchItems();
  }, [selectedCluster, namespace]);

  const handleSuspend = async (name: string, suspend: boolean) => {
    try {
      await suspendCronJob(selectedCluster, namespace, name, suspend);
      message.success(t('cronJobs.suspendSuccess'));
      fetchItems();
    } catch {
      message.error(t('cronJobs.suspendFailed'));
    }
  };

  const handleDelete = async (name: string) => {
    try {
      await deleteCronJob(selectedCluster, namespace, name);
      message.success(t('cronJobs.deleteSuccess'));
      fetchItems();
    } catch {
      message.error(t('cronJobs.deleteFailed'));
    }
  };

  const columns = [
    { title: t('common.name'), dataIndex: 'name', key: 'name' },
    { title: t('common.namespace'), dataIndex: 'namespace', key: 'namespace' },
    { title: t('cronJobs.schedule'), dataIndex: 'schedule', key: 'schedule' },
    {
      title: t('cronJobs.suspend'),
      dataIndex: 'suspend',
      key: 'suspend',
      render: (suspend: boolean, record: CronJobInfo) => (
        <Switch checked={suspend} onChange={(v) => handleSuspend(record.name, v)} />
      ),
    },
    {
      title: t('cronJobs.activeJobs'),
      dataIndex: 'activeJobs',
      key: 'activeJobs',
      render: (n: number) => <Tag color={n > 0 ? 'processing' : 'default'}>{n}</Tag>,
    },
    {
      title: t('common.operations'),
      key: 'action',
      render: (_: unknown, record: CronJobInfo) => (
        <Popconfirm title={t('cronJobs.deleteConfirm')} onConfirm={() => handleDelete(record.name)}>
          <Button type="link" danger icon={<DeleteOutlined />}>{t('common.delete')}</Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <Card
      title={t('cronJobs.management')}
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
      <Table columns={columns} dataSource={items} rowKey={(r) => `${r.namespace}/${r.name}`} loading={loading} pagination={{ pageSize: 10 }} />
    </Card>
  );
};

export default CronJobs;
