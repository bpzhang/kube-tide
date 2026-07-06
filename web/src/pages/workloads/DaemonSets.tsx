import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, Space, message, Button, Popconfirm } from 'antd';
import { DeleteOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useClusterNamespace } from '@/hooks/useClusterNamespace';
import ClusterNamespaceToolbar from '@/components/k8s/common/ClusterNamespaceToolbar';
import { listDaemonSets, deleteDaemonSet, DaemonSetInfo } from '@/api/daemonset';

const DaemonSets: React.FC = () => {
  const { t } = useTranslation();
  const { selectedCluster, setSelectedCluster, clusters, namespace, setNamespace, clustersLoading } =
    useClusterNamespace(t);
  const [items, setItems] = useState<DaemonSetInfo[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchItems = async () => {
    if (!selectedCluster) return;
    setLoading(true);
    try {
      const response = await listDaemonSets(selectedCluster, namespace);
      if (response.data.code === 0) {
        setItems(response.data.data.daemonsets || []);
      } else {
        message.error(response.data.message || t('daemonSets.fetchFailed'));
        setItems([]);
      }
    } catch {
      message.error(t('daemonSets.fetchFailed'));
      setItems([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (selectedCluster) fetchItems();
  }, [selectedCluster, namespace]);

  const handleDelete = async (name: string) => {
    try {
      await deleteDaemonSet(selectedCluster, namespace, name);
      message.success(t('daemonSets.deleteSuccess'));
      fetchItems();
    } catch {
      message.error(t('daemonSets.deleteFailed'));
    }
  };

  const columns = [
    { title: t('common.name'), dataIndex: 'name', key: 'name' },
    { title: t('common.namespace'), dataIndex: 'namespace', key: 'namespace' },
    {
      title: t('daemonSets.ready'),
      key: 'ready',
      render: (_: unknown, r: DaemonSetInfo) => `${r.numberReady}/${r.desiredNumberScheduled}`,
    },
    { title: t('daemonSets.updateStrategy'), dataIndex: 'updateStrategy', key: 'updateStrategy' },
    {
      title: t('daemonSets.images'),
      dataIndex: 'images',
      key: 'images',
      render: (images: string[]) => (
        <Space direction="vertical" size={0}>
          {(images || []).map((img) => <Tag key={img}>{img}</Tag>)}
        </Space>
      ),
    },
    {
      title: t('common.operations'),
      key: 'action',
      render: (_: unknown, record: DaemonSetInfo) => (
        <Popconfirm title={t('daemonSets.deleteConfirm')} onConfirm={() => handleDelete(record.name)}>
          <Button type="link" danger icon={<DeleteOutlined />}>{t('common.delete')}</Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <Card
      title={t('daemonSets.management')}
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

export default DaemonSets;
