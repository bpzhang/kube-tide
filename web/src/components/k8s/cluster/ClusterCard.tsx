import React, { useState } from 'react';
import { Card, Button, Space, message, Popconfirm } from 'antd';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { testClusterConnection, removeCluster } from '@/api/cluster';

interface ClusterCardProps {
  name: string;
  onRemove: () => void;
}

const ClusterCard: React.FC<ClusterCardProps> = ({ name, onRemove }) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);

  const handleTest = async () => {
    try {
      setLoading(true);
      const response = await testClusterConnection(name);
      if (response.data.code === 0) {
        message.success(t('clusterDetail.actions.testSuccess'));
      } else {
        message.error(response.data.message || t('clusterDetail.actions.testFailed'));
      }
    } catch (err) {
      message.error(t('clusterDetail.actions.testFailed'));
    } finally {
      setLoading(false);
    }
  };

  const handleRemove = async () => {
    try {
      setLoading(true);
      const response = await removeCluster(name);
      if (response.data.code === 0) {
        message.success(t('clusters.deleteSuccess', { name }));
        onRemove();
      } else {
        message.error(response.data.message || t('clusters.deleteFailed'));
      }
    } catch (err) {
      message.error(t('clusters.deleteFailed'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card title={name} style={{ width: 300, marginBottom: 16 }}>
      <Space direction="vertical" style={{ width: '100%' }}>
        <Button 
          type="primary" 
          block 
          onClick={() => navigate(`/clusters/${name}`)}
        >
          {t('common.details')}
        </Button>
        <Button 
          block 
          onClick={handleTest}
          loading={loading}
        >
          {t('clusterDetail.actions.test')}
        </Button>
        <Popconfirm
          title={t('clusters.deleteConfirm')}
          description={t('clusters.deleteConfirmMessage', { name })}
          onConfirm={handleRemove}
          okText={t('common.confirm')}
          cancelText={t('common.cancel')}
        >
          <Button 
            danger 
            block 
            loading={loading}
          >
            {t('common.delete')}
          </Button>
        </Popconfirm>
      </Space>
    </Card>
  );
};

export default ClusterCard;