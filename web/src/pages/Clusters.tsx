import React, { useState, useEffect } from 'react';
import { Row, Col, Button, Modal, Form, Input, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { getClusterList, addCluster } from '../api/cluster';
import type { ClusterResponse } from '../api/cluster';
import ClusterCard from '../components/k8s/cluster/ClusterCard';

const Clusters: React.FC = () => {
  const [clusters, setClusters] = useState<string[]>([]);
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();

  const fetchClusters = async () => {
    try {
      setLoading(true);
      const response = await getClusterList();
      if (response.data.code === 0) {
        setClusters(response.data.data.clusters);
      } else {
        message.error(response.data.message || '获取集群列表失败');
        setClusters([]);
      }
    } catch (err) {
      message.error('获取集群列表失败');
      setClusters([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchClusters();
  }, []);

  const handleAddCluster = async (values: any) => {
    try {
      const response = await addCluster(values);
      if (response.data.code === 0) {
        message.success('添加集群成功');
        setIsModalVisible(false);
        form.resetFields();
        fetchClusters();
      } else {
        message.error(response.data.message || '添加集群失败');
      }
    } catch (err) {
      message.error('添加集群失败');
    }
  };

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Button 
          type="primary" 
          icon={<PlusOutlined />}
          onClick={() => setIsModalVisible(true)}
          loading={loading}
        >
          添加集群
        </Button>
      </div>

      <Row gutter={[16, 16]}>
        {clusters.map(cluster => (
          <Col key={cluster} xs={24} sm={12} md={8} lg={6}>
            <ClusterCard 
              name={cluster} 
              onRemove={fetchClusters}
            />
          </Col>
        ))}
      </Row>

      <Modal
        title="添加集群"
        open={isModalVisible}
        onCancel={() => setIsModalVisible(false)}
        onOk={() => form.submit()}
        confirmLoading={loading}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleAddCluster}
        >
          <Form.Item
            name="name"
            label="集群名称"
            rules={[{ required: true, message: '请输入集群名称' }]}
          >
            <Input placeholder="请输入集群名称" />
          </Form.Item>
          <Form.Item
            name="kubeconfigPath"
            label="Kubeconfig路径"
            rules={[{ required: true, message: '请输入Kubeconfig文件路径' }]}
          >
            <Input placeholder="请输入Kubeconfig文件的完整路径" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default Clusters;