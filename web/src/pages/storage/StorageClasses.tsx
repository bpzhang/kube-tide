import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, message } from 'antd';
import { useTranslation } from 'react-i18next';
import { useClusterNamespace } from '@/hooks/useClusterNamespace';
import ClusterNamespaceToolbar from '@/components/k8s/common/ClusterNamespaceToolbar';
import { listStorageClasses, StorageClassInfo } from '@/api/storageclass';

const StorageClasses: React.FC = () => {
  const { t } = useTranslation();
  const { selectedCluster, setSelectedCluster, clusters, clustersLoading } = useClusterNamespace(t);
  const [items, setItems] = useState<StorageClassInfo[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchItems = async () => {
    if (!selectedCluster) return;
    setLoading(true);
    try {
      const response = await listStorageClasses(selectedCluster);
      if (response.data.code === 0) {
        setItems(response.data.data.storageclasses || []);
      } else {
        message.error(response.data.message || t('storageClasses.fetchFailed'));
        setItems([]);
      }
    } catch {
      message.error(t('storageClasses.fetchFailed'));
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
    { title: t('storageClasses.provisioner'), dataIndex: 'provisioner', key: 'provisioner' },
    { title: t('storageClasses.reclaimPolicy'), dataIndex: 'reclaimPolicy', key: 'reclaimPolicy', render: (v: string) => v || '-' },
    { title: t('storageClasses.volumeBindingMode'), dataIndex: 'volumeBindingMode', key: 'volumeBindingMode', render: (v: string) => v || '-' },
    {
      title: t('storageClasses.allowExpansion'),
      dataIndex: 'allowVolumeExpansion',
      key: 'allowVolumeExpansion',
      render: (v: boolean) => <Tag color={v ? 'success' : 'default'}>{v ? t('common.yes') : t('common.no')}</Tag>,
    },
  ];

  return (
    <Card
      title={t('storageClasses.management')}
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

export default StorageClasses;
