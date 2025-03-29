import React, { useState } from 'react';
import { Card, Button, Space, message, Popconfirm } from 'antd';
import { useNavigate } from 'react-router-dom';
import { testClusterConnection, removeCluster } from '@/api/cluster';

interface ClusterCardProps {
  name: string;
  onRemove: () => void;
}

const ClusterCard: React.FC<ClusterCardProps> = ({ name, onRemove }) => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);

  const handleTest = async () => {
    try {
      setLoading(true);
      const response = await testClusterConnection(name);
      if (response.data.code === 0) {
        message.success('集群连接测试成功');
      } else {
        message.error(response.data.message || '集群连接测试失败');
      }
    } catch (err) {
      message.error('集群连接测试失败');
    } finally {
      setLoading(false);
    }
  };

  const handleRemove = async () => {
    try {
      setLoading(true);
      const response = await removeCluster(name);
      if (response.data.code === 0) {
        message.success('集群删除成功');
        onRemove();
      } else {
        message.error(response.data.message || '集群删除失败');
      }
    } catch (err) {
      message.error('集群删除失败');
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
          查看详情
        </Button>
        <Button 
          block 
          onClick={handleTest}
          loading={loading}
        >
          测试连接
        </Button>
        <Popconfirm
          title="确定要删除这个集群吗?"
          description="删除后将无法恢复，请谨慎操作。"
          onConfirm={handleRemove}
          okText="确定"
          cancelText="取消"
        >
          <Button 
            danger 
            block 
            loading={loading}
          >
            删除集群
          </Button>
        </Popconfirm>
      </Space>
    </Card>
  );
};

export default ClusterCard;