import React, { useState, useEffect } from 'react';
import { Row, Col, Button, Modal, Form, Input, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { getClusterList, addCluster } from '../api/cluster';
import type { ClusterResponse } from '../api/cluster';
import ClusterCard from '../components/k8s/cluster/ClusterCard';

const Clusters: React.FC = () => {
  const { t } = useTranslation();
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
        message.error(response.data.message || t('clusters.fetchFailed'));
        setClusters([]);
      }
    } catch (err) {
      message.error(t('clusters.fetchFailed'));
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
        message.success(t('clusters.addSuccess'));
        setIsModalVisible(false);
        form.resetFields();
        fetchClusters();
      } else {
        message.error(response.data.message || t('clusters.addFailed'));
      }
    } catch (err) {
      message.error(t('clusters.addFailed'));
    }
  };

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ marginBottom: 16 }}>
        <Button 
          type="primary" 
          icon={<PlusOutlined />}
          onClick={() => setIsModalVisible(true)}
          loading={loading}
        >
          {t('clusters.addCluster')}
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
        title={t('clusters.addCluster')}
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
            label={t('clusters.clusterName')}
            rules={[{ required: true, message: t('clusters.pleaseInputName') }]}
          >
            <Input placeholder={t('clusters.clusterNamePlaceholder')} />
          </Form.Item>
          <Form.Item
            name="kubeconfigPath"
            label={t('clusters.kubeconfigPath')}
            rules={[{ required: true, message: t('clusters.pleaseInputKubeconfigPath') }]}
          >
            <Input placeholder={t('clusters.kubeconfigPathPlaceholder')} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default Clusters;