import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, message, Button, Popconfirm } from 'antd';
import { DeleteOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useClusterNamespace } from '@/hooks/useClusterNamespace';
import ClusterNamespaceToolbar from '@/components/k8s/common/ClusterNamespaceToolbar';
import { listNetworkPolicies, deleteNetworkPolicy, NetworkPolicyInfo } from '@/api/networkpolicy';

const NetworkPolicies: React.FC = () => {
  const { t } = useTranslation();
  const { selectedCluster, setSelectedCluster, clusters, namespace, setNamespace, clustersLoading } =
    useClusterNamespace(t);
  const [items, setItems] = useState<NetworkPolicyInfo[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchItems = async () => {
    if (!selectedCluster) return;
    setLoading(true);
    try {
      const response = await listNetworkPolicies(selectedCluster, namespace);
      if (response.data.code === 0) {
        setItems(response.data.data.networkpolicies || []);
      } else {
        message.error(response.data.message || t('networkPolicies.fetchFailed'));
        setItems([]);
      }
    } catch {
      message.error(t('networkPolicies.fetchFailed'));
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
      await deleteNetworkPolicy(selectedCluster, namespace, name);
      message.success(t('networkPolicies.deleteSuccess'));
      fetchItems();
    } catch {
      message.error(t('networkPolicies.deleteFailed'));
    }
  };

  const columns = [
    { title: t('common.name'), dataIndex: 'name', key: 'name' },
    { title: t('common.namespace'), dataIndex: 'namespace', key: 'namespace' },
    {
      title: t('networkPolicies.policyTypes'),
      dataIndex: 'policyTypes',
      key: 'policyTypes',
      render: (types: string[]) => (types || []).map((tp) => <Tag key={tp}>{tp}</Tag>),
    },
    {
      title: t('networkPolicies.podSelector'),
      dataIndex: 'podSelector',
      key: 'podSelector',
      render: (sel: Record<string, string>) =>
        sel && Object.keys(sel).length > 0
          ? Object.entries(sel).map(([k, v]) => <Tag key={k}>{`${k}=${v}`}</Tag>)
          : '-',
    },
    {
      title: t('networkPolicies.rules'),
      key: 'rules',
      render: (_: unknown, r: NetworkPolicyInfo) =>
        `${t('networkPolicies.ingress')}: ${r.ingressRuleCount}, ${t('networkPolicies.egress')}: ${r.egressRuleCount}`,
    },
    {
      title: t('common.operations'),
      key: 'action',
      render: (_: unknown, record: NetworkPolicyInfo) => (
        <Popconfirm title={t('networkPolicies.deleteConfirm')} onConfirm={() => handleDelete(record.name)}>
          <Button type="link" danger icon={<DeleteOutlined />}>{t('common.delete')}</Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <Card
      title={t('networkPolicies.management')}
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

export default NetworkPolicies;
