import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, Tabs, message, Button, Popconfirm } from 'antd';
import { DeleteOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useClusterNamespace } from '@/hooks/useClusterNamespace';
import ClusterNamespaceToolbar from '@/components/k8s/common/ClusterNamespaceToolbar';
import {
  listRoles,
  listClusterRoles,
  listRoleBindings,
  listClusterRoleBindings,
  deleteRoleBinding,
  deleteClusterRoleBinding,
  RoleInfo,
  RoleBindingInfo,
} from '@/api/rbac';

const RBAC: React.FC = () => {
  const { t } = useTranslation();
  const { selectedCluster, setSelectedCluster, clusters, namespace, setNamespace, clustersLoading } =
    useClusterNamespace(t);
  const [roles, setRoles] = useState<RoleInfo[]>([]);
  const [clusterRoles, setClusterRoles] = useState<RoleInfo[]>([]);
  const [roleBindings, setRoleBindings] = useState<RoleBindingInfo[]>([]);
  const [clusterRoleBindings, setClusterRoleBindings] = useState<RoleBindingInfo[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchAll = async () => {
    if (!selectedCluster) return;
    setLoading(true);
    try {
      const [rolesRes, crRes, rbRes, crbRes] = await Promise.all([
        listRoles(selectedCluster, namespace),
        listClusterRoles(selectedCluster),
        listRoleBindings(selectedCluster, namespace),
        listClusterRoleBindings(selectedCluster),
      ]);
      if (rolesRes.data.code === 0) setRoles(rolesRes.data.data.roles || []);
      if (crRes.data.code === 0) setClusterRoles(crRes.data.data.clusterroles || []);
      if (rbRes.data.code === 0) setRoleBindings(rbRes.data.data.rolebindings || []);
      if (crbRes.data.code === 0) setClusterRoleBindings(crbRes.data.data.clusterrolebindings || []);
    } catch {
      message.error(t('rbac.fetchFailed'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (selectedCluster) fetchAll();
  }, [selectedCluster, namespace]);

  const handleDeleteRoleBinding = async (name: string, isCluster: boolean) => {
    try {
      if (isCluster) {
        await deleteClusterRoleBinding(selectedCluster, name);
      } else {
        await deleteRoleBinding(selectedCluster, namespace, name);
      }
      message.success(t('rbac.deleteSuccess'));
      fetchAll();
    } catch {
      message.error(t('rbac.deleteFailed'));
    }
  };

  const roleColumns = [
    { title: t('common.name'), dataIndex: 'name', key: 'name' },
    { title: t('common.namespace'), dataIndex: 'namespace', key: 'namespace', render: (v: string) => v || '-' },
    { title: t('rbac.ruleCount'), dataIndex: 'ruleCount', key: 'ruleCount' },
  ];

  const bindingColumns = (isCluster: boolean) => [
    { title: t('common.name'), dataIndex: 'name', key: 'name' },
    { title: t('common.namespace'), dataIndex: 'namespace', key: 'namespace', render: (v: string) => v || '-' },
    {
      title: t('rbac.roleRef'),
      key: 'roleRef',
      render: (_: unknown, r: RoleBindingInfo) => `${r.roleRef?.kind}/${r.roleRef?.name}`,
    },
    {
      title: t('rbac.subjects'),
      dataIndex: 'subjectCount',
      key: 'subjectCount',
      render: (n: number, r: RoleBindingInfo) => (
        <>
          <Tag>{n}</Tag>
          {(r.subjects || []).slice(0, 2).map((s, i) => (
            <Tag key={i} color="blue">{`${s.kind}:${s.name}`}</Tag>
          ))}
        </>
      ),
    },
    {
      title: t('common.operations'),
      key: 'action',
      render: (_: unknown, record: RoleBindingInfo) => (
        <Popconfirm title={t('rbac.deleteConfirm')} onConfirm={() => handleDeleteRoleBinding(record.name, isCluster)}>
          <Button type="link" danger icon={<DeleteOutlined />}>{t('common.delete')}</Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <Card
      title={t('rbac.management')}
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
      <Tabs
        items={[
          {
            key: 'roles',
            label: t('rbac.roles'),
            children: <Table columns={roleColumns} dataSource={roles} rowKey="name" loading={loading} pagination={{ pageSize: 10 }} />,
          },
          {
            key: 'clusterRoles',
            label: t('rbac.clusterRoles'),
            children: <Table columns={roleColumns} dataSource={clusterRoles} rowKey="name" loading={loading} pagination={{ pageSize: 10 }} />,
          },
          {
            key: 'roleBindings',
            label: t('rbac.roleBindings'),
            children: <Table columns={bindingColumns(false)} dataSource={roleBindings} rowKey="name" loading={loading} pagination={{ pageSize: 10 }} />,
          },
          {
            key: 'clusterRoleBindings',
            label: t('rbac.clusterRoleBindings'),
            children: <Table columns={bindingColumns(true)} dataSource={clusterRoleBindings} rowKey="name" loading={loading} pagination={{ pageSize: 10 }} />,
          },
        ]}
      />
    </Card>
  );
};

export default RBAC;
