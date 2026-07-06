import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, message, Button, Popconfirm } from 'antd';
import { DeleteOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useClusterNamespace } from '@/hooks/useClusterNamespace';
import ClusterNamespaceToolbar from '@/components/k8s/common/ClusterNamespaceToolbar';
import { listLimitRanges, deleteLimitRange, LimitRangeInfo } from '@/api/limitrange';

const LimitRanges: React.FC = () => {
  const { t } = useTranslation();
  const { selectedCluster, setSelectedCluster, clusters, namespace, setNamespace, clustersLoading } =
    useClusterNamespace(t);
  const [items, setItems] = useState<LimitRangeInfo[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchItems = async () => {
    if (!selectedCluster) return;
    setLoading(true);
    try {
      const response = await listLimitRanges(selectedCluster, namespace);
      if (response.data.code === 0) {
        setItems(response.data.data.limitranges || []);
      } else {
        message.error(response.data.message || t('limitRanges.fetchFailed'));
        setItems([]);
      }
    } catch {
      message.error(t('limitRanges.fetchFailed'));
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
      await deleteLimitRange(selectedCluster, namespace, name);
      message.success(t('limitRanges.deleteSuccess'));
      fetchItems();
    } catch {
      message.error(t('limitRanges.deleteFailed'));
    }
  };

  const columns = [
    { title: t('common.name'), dataIndex: 'name', key: 'name' },
    { title: t('common.namespace'), dataIndex: 'namespace', key: 'namespace' },
    {
      title: t('limitRanges.limits'),
      dataIndex: 'limits',
      key: 'limits',
      render: (limits: LimitRangeInfo['limits']) =>
        (limits || []).map((l, i) => <Tag key={i}>{l.type}</Tag>),
    },
    {
      title: t('common.operations'),
      key: 'action',
      render: (_: unknown, record: LimitRangeInfo) => (
        <Popconfirm title={t('limitRanges.deleteConfirm')} onConfirm={() => handleDelete(record.name)}>
          <Button type="link" danger icon={<DeleteOutlined />}>{t('common.delete')}</Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <Card
      title={t('limitRanges.management')}
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

export default LimitRanges;
