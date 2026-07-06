import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, message, Button, Popconfirm } from 'antd';
import { DeleteOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useClusterNamespace } from '@/hooks/useClusterNamespace';
import ClusterNamespaceToolbar from '@/components/k8s/common/ClusterNamespaceToolbar';
import { listPDBs, deletePDB, PDBInfo } from '@/api/pdb';

const PDBs: React.FC = () => {
  const { t } = useTranslation();
  const { selectedCluster, setSelectedCluster, clusters, namespace, setNamespace, clustersLoading } =
    useClusterNamespace(t);
  const [items, setItems] = useState<PDBInfo[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchItems = async () => {
    if (!selectedCluster) return;
    setLoading(true);
    try {
      const response = await listPDBs(selectedCluster, namespace);
      if (response.data.code === 0) {
        setItems(response.data.data.pdbs || []);
      } else {
        message.error(response.data.message || t('pdbs.fetchFailed'));
        setItems([]);
      }
    } catch {
      message.error(t('pdbs.fetchFailed'));
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
      await deletePDB(selectedCluster, namespace, name);
      message.success(t('pdbs.deleteSuccess'));
      fetchItems();
    } catch {
      message.error(t('pdbs.deleteFailed'));
    }
  };

  const columns = [
    { title: t('common.name'), dataIndex: 'name', key: 'name' },
    { title: t('common.namespace'), dataIndex: 'namespace', key: 'namespace' },
    {
      title: t('pdbs.minAvailable'),
      dataIndex: 'minAvailable',
      key: 'minAvailable',
      render: (v: string, r: PDBInfo) => v || r.maxUnavailable || '-',
    },
    {
      title: t('pdbs.healthy'),
      key: 'healthy',
      render: (_: unknown, r: PDBInfo) => `${r.currentHealthy}/${r.desiredHealthy}`,
    },
    {
      title: t('pdbs.disruptionsAllowed'),
      dataIndex: 'disruptionsAllowed',
      key: 'disruptionsAllowed',
      render: (n: number) => <Tag color={n > 0 ? 'success' : 'warning'}>{n}</Tag>,
    },
    {
      title: t('pdbs.selector'),
      dataIndex: 'selector',
      key: 'selector',
      render: (sel: Record<string, string>) =>
        sel ? Object.entries(sel).map(([k, v]) => <Tag key={k}>{`${k}=${v}`}</Tag>) : '-',
    },
    {
      title: t('common.operations'),
      key: 'action',
      render: (_: unknown, record: PDBInfo) => (
        <Popconfirm title={t('pdbs.deleteConfirm')} onConfirm={() => handleDelete(record.name)}>
          <Button type="link" danger icon={<DeleteOutlined />}>{t('common.delete')}</Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <Card
      title={t('pdbs.management')}
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

export default PDBs;
