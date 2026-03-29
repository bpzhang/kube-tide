import React, { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Button, Space, Spin, message } from 'antd';
import { ArrowLeftOutlined, ReloadOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { getDeploymentDetails } from '@/api/deployment';
import DeploymentDetail from '@/components/k8s/deployment/DeploymentDetail';

const DeploymentDetailPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { clusterName, namespace, deploymentName } = useParams<{
    clusterName: string;
    namespace: string;
    deploymentName: string;
  }>();
  const [loading, setLoading] = useState(true);
  const [deployment, setDeployment] = useState<any>(null);

  const fetchDeploymentDetail = async () => {
    if (!clusterName || !namespace || !deploymentName) {
      message.error(t('common.error'));
      navigate('/workloads/deployments');
      return;
    }

    setLoading(true);
    try {
      const response = await getDeploymentDetails(clusterName, namespace, deploymentName);
      if (response.data.code === 0) {
        setDeployment(response.data.data.deployment);
      } else {
        message.error(response.data.message || t('deployments.fetchDetailFailed'));
        setDeployment(null);
      }
    } catch (error) {
      console.error(t('deployments.fetchDetailFailed'), error);
      message.error(t('deployments.fetchDetailFailed'));
      setDeployment(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDeploymentDetail();
  }, [clusterName, namespace, deploymentName]);

  if (loading) {
    return (
      <div style={{ padding: '24px', textAlign: 'center' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!deployment) {
    return (
      <div style={{ padding: '24px', textAlign: 'center' }}>
        <h2>{t('deployments.noData')}</h2>
        <Button type="primary" onClick={() => navigate('/workloads/deployments')}>
          {t('common.back')}
        </Button>
      </div>
    );
  }

  return (
    <div style={{ padding: '24px' }}>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '16px',
        }}
      >
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate(-1)}>
          {t('common.back')}
        </Button>

        <Space>
          <Button icon={<ReloadOutlined />} onClick={fetchDeploymentDetail}>
            {t('common.refresh')}
          </Button>
        </Space>
      </div>

      <DeploymentDetail
        deployment={deployment}
        clusterName={clusterName}
        onUpdate={fetchDeploymentDetail}
      />
    </div>
  );
};

export default DeploymentDetailPage;