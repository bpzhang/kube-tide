import React, { useState, useEffect } from 'react';
import { Card, Button, Space, message, Popconfirm, Tag } from 'antd';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { testClusterConnection, removeCluster, getClusterDetails } from '../../../api/cluster';

interface ClusterCardProps {
  name: string;
  onRemove: () => void;
}

const ClusterCard: React.FC<ClusterCardProps> = ({ name, onRemove }) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [addType, setAddType] = useState<'path' | 'content' | 'unknown'>('unknown');

  // 获取集群添加方式
  useEffect(() => {
    const fetchClusterDetails = async () => {
      try {
        const response = await getClusterDetails(name);
        if (response.data.code === 0 && response.data.data.cluster.addType) {
          setAddType(response.data.data.cluster.addType);
        }
      } catch (error) {
        console.error("获取集群详情失败:", error);
      }
    };
    
    fetchClusterDetails();
  }, [name]);

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

  // 根据添加方式显示不同的标签颜色
  const getTagColor = () => {
    switch (addType) {
      case 'path':
        return 'blue';
      case 'content':
        return 'green';
      default:
        return 'default';
    }
  };

  // 获取添加方式的显示文本
  const getAddTypeText = () => {
    switch (addType) {
      case 'path':
        return t('clusters.addTypeFile') || '通过文件路径';
      case 'content':
        return t('clusters.addTypeContent') || '通过内容填写';
      default:
        return t('clusters.addTypeUnknown') || '未知方式';
    }
  };

  return (
    <Card 
      title={name} 
      style={{ width: 300, marginBottom: 16 }}
      extra={
        <Tag color={getTagColor()} style={{ marginRight: 0 }}>
          {getAddTypeText()}
        </Tag>
      }
    >
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