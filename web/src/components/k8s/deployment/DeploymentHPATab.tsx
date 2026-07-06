import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, Empty, Spin, message } from 'antd';
import { useTranslation } from 'react-i18next';
import { listHPAs, HPAInfo } from '@/api/hpa';

interface DeploymentHPATabProps {
  clusterName: string;
  namespace: string;
  deploymentName: string;
}

const DeploymentHPATab: React.FC<DeploymentHPATabProps> = ({
  clusterName,
  namespace,
  deploymentName,
}) => {
  const { t } = useTranslation();
  const [hpas, setHpas] = useState<HPAInfo[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchHPAs = async () => {
    setLoading(true);
    try {
      const response = await listHPAs(clusterName, namespace);
      if (response.data.code === 0) {
        const matched = (response.data.data.hpas || []).filter(
          (hpa) =>
            hpa.targetRef?.kind === 'Deployment' &&
            hpa.targetRef?.name === deploymentName,
        );
        setHpas(matched);
      } else {
        message.error(response.data.message || t('hpas.fetchFailed'));
      }
    } catch {
      message.error(t('hpas.fetchFailed'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (clusterName && namespace && deploymentName) {
      fetchHPAs();
    }
  }, [clusterName, namespace, deploymentName]);

  const columns = [
    { title: t('common.name'), dataIndex: 'name', key: 'name' },
    {
      title: t('hpas.replicas'),
      key: 'replicas',
      render: (_: unknown, r: HPAInfo) =>
        `${r.currentReplicas}/${r.desiredReplicas} (${r.minReplicas ?? 1}-${r.maxReplicas})`,
    },
    {
      title: t('hpas.metrics'),
      dataIndex: 'metrics',
      key: 'metrics',
      render: (metrics: HPAInfo['metrics']) =>
        (metrics || []).map((m, i) => (
          <Tag key={i}>
            {m.type}
            {m.utilization != null ? `: ${m.utilization}%` : m.average ? `: ${m.average}` : ''}
          </Tag>
        )),
    },
  ];

  if (loading) {
    return <Spin />;
  }

  return (
    <Card title={t('deployments.detail.hpa.title')}>
      {hpas.length > 0 ? (
        <Table columns={columns} dataSource={hpas} rowKey="name" pagination={false} />
      ) : (
        <Empty description={t('deployments.detail.hpa.noHpa')} />
      )}
    </Card>
  );
};

export default DeploymentHPATab;
