import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, message } from 'antd';
import { useTranslation } from 'react-i18next';
import { useClusterNamespace } from '@/hooks/useClusterNamespace';
import ClusterNamespaceToolbar from '@/components/k8s/common/ClusterNamespaceToolbar';
import { listPVs, PVInfo } from '@/api/pv';

const PVs: React.FC = () => {
  const { t } = useTranslation();
  const { selectedCluster, setSelectedCluster, clusters, clustersLoading } = useClusterNamespace(t);
  const [items, setItems] = useState<PVInfo[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchItems = async () => {
    if (!selectedCluster) return;
    setLoading(true);
    try {
      const response = await listPVs(selectedCluster);
      if (response.data.code === 0) {
        setItems(response.data.data.pvs || []);
      } else {
        message.error(response.data.message || t('pvs.fetchFailed'));
        setItems([]);
      }
    } catch {
      message.error(t('pvs.fetchFailed'));
      setItems([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (selectedCluster) fetchItems();
  }, [selectedCluster]);

  const columns = [
    { title: t('common.name'), dataIndex: 'name', key: 'name' },
    {
      title: t('common.status'),
      dataIndex: 'status',
      key: 'status',
      render: (s: string) => <Tag color={s === 'Available' ? 'success' : s === 'Bound' ? 'blue' : 'default'}>{s}</Tag>,
    },
    { title: t('pvs.capacity'), dataIndex: 'capacity', key: 'capacity', render: (v: string) => v || '-' },
    { title: t('pvs.storageClass'), dataIndex: 'storageClassName', key: 'storageClassName', render: (v: string) => v || '-' },
    { title: t('pvs.reclaimPolicy'), dataIndex: 'reclaimPolicy', key: 'reclaimPolicy', render: (v: string) => v || '-' },
    {
      title: t('pvs.claim'),
      dataIndex: 'claimRef',
      key: 'claimRef',
      render: (ref: PVInfo['claimRef']) => (ref ? `${ref.namespace}/${ref.name}` : '-'),
    },
  ];

  return (
    <Card
      title={t('pvs.management')}
      extra={
        <ClusterNamespaceToolbar
          selectedCluster={selectedCluster}
          clusters={clusters}
          namespace=""
          onClusterChange={setSelectedCluster}
          onNamespaceChange={() => {}}
          loading={clustersLoading}
          showNamespace={false}
        />
      }
    >
      <Table columns={columns} dataSource={items} rowKey="name" loading={loading} pagination={{ pageSize: 10 }} />
    </Card>
  );
};

export default PVs;
