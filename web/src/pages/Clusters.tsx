import React, { useState, useEffect } from 'react';
import { Row, Col, Button, Modal, Form, Input, message, Radio, Tabs } from 'antd';
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
  const [addType, setAddType] = useState<'path' | 'content'>('path'); // 默认使用路径方式
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
      // 添加集群类型字段
      const clusterData = {
        ...values,
        addType
      };
      
      const response = await addCluster(clusterData);
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
  
  // 处理添加方式切换
  const handleAddTypeChange = (type: 'path' | 'content') => {
    setAddType(type);
    form.resetFields(['kubeconfigPath', 'kubeconfigContent']);
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
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleAddCluster}
          initialValues={{ addTypeField: 'path' }} // 默认选择通过文件路径
        >
          <Form.Item
            name="name"
            label={t('clusters.clusterName')}
            rules={[{ required: true, message: t('clusters.pleaseInputName') }]}
          >
            <Input placeholder={t('clusters.clusterNamePlaceholder')} />
          </Form.Item>

          <div style={{ marginBottom: 24 }}>
            <div style={{ marginBottom: 8, fontWeight: 'bold' }}>添加方式</div>
            <Radio.Group 
              value={addType} 
              onChange={(e) => handleAddTypeChange(e.target.value)}
              buttonStyle="solid"
              size="large"
              style={{ width: '100%', marginBottom: 8 }}
            >
              <Radio.Button value="path" style={{ width: '50%', textAlign: 'center', height: '40px', lineHeight: '40px' }}>
                通过文件路径
              </Radio.Button>
              <Radio.Button value="content" style={{ width: '50%', textAlign: 'center', height: '40px', lineHeight: '40px' }}>
                通过内容填写
              </Radio.Button>
            </Radio.Group>
          </div>
          
          {addType === 'path' ? (
            <Form.Item
              name="kubeconfigPath"
              label="Kubeconfig路径"
              rules={[{ required: addType === 'path', message: "请输入kubeconfig文件路径" }]}
            >
              <Input placeholder="请输入kubeconfig文件的完整路径" />
            </Form.Item>
          ) : (
            <Form.Item
              name="kubeconfigContent"
              label="Kubeconfig内容"
              rules={[{ required: addType === 'content', message: "请输入kubeconfig内容" }]}
            >
              <Input.TextArea 
                placeholder="在此粘贴kubeconfig的YAML内容"
                rows={10}
                style={{ fontFamily: 'monospace' }}
              />
            </Form.Item>
          )}
        </Form>
      </Modal>
    </div>
  );
};

export default Clusters;