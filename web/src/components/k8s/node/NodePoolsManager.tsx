import React, { useState } from 'react';
import { Modal, Form, Input, Button, Space, Table, Typography, message, Popconfirm, Select } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import type { NodePool } from '@/api/nodepool';
import { createNodePool, updateNodePool, deleteNodePool } from '@/api/nodepool';

const { Text } = Typography;

interface NodePoolsManagerProps {
  visible: boolean;
  onClose: () => void;
  clusterName: string;
  nodePools: NodePool[];
  onSuccess: () => void;
}

interface NodePoolFormData {
  name: string;
  labels: { key: string; value: string }[];
  taints: { key: string; value?: string; effect: string }[];
}

const NodePoolsManager: React.FC<NodePoolsManagerProps> = ({
  visible,
  onClose,
  clusterName,
  nodePools,
  onSuccess,
}) => {
  const { t } = useTranslation();
  const [form] = Form.useForm<NodePoolFormData>();
  const [editingPool, setEditingPool] = useState<NodePool | null>(null);
  const [loading, setLoading] = useState(false);
  const [showForm, setShowForm] = useState(false);

  const handleSubmit = async (values: NodePoolFormData) => {
    setLoading(true);
    try {
      const pool: NodePool = {
        name: values.name,
        labels: values.labels?.reduce((acc, curr) => ({
          ...acc,
          [curr.key]: curr.value
        }), {}) || {},
        taints: values.taints?.map(t => ({
          key: t.key,
          value: t.value,
          effect: t.effect,
        })) || [],
      };

      if (editingPool) {
        await updateNodePool(clusterName, editingPool.name, pool);
        message.success(t('nodes.nodePool.updateSuccess'));
      } else {
        await createNodePool(clusterName, pool);
        message.success(t('nodes.nodePool.createSuccess'));
      }

      form.resetFields();
      setEditingPool(null);
      setShowForm(false);
      onSuccess();
    } catch (err) {
      message.error(t('nodes.nodePool.operationFailed', { message: (err as Error).message }));
    } finally {
      setLoading(false);
    }
  };

  const handleEdit = (pool: NodePool) => {
    setEditingPool(pool);
    setShowForm(true);
    // 将节点池数据转换为表单格式
    const formData: NodePoolFormData = {
      name: pool.name,
      labels: Object.entries(pool.labels || {}).map(([key, value]) => ({
        key,
        value,
      })),
      taints: pool.taints || [],
    };
    form.setFieldsValue(formData);
  };

  const handleDelete = async (poolName: string) => {
    try {
      await deleteNodePool(clusterName, poolName);
      message.success(t('nodes.nodePool.deleteSuccess'));
      onSuccess();
    } catch (err) {
      message.error(t('nodes.nodePool.operationFailed', { message: (err as Error).message }));
    }
  };

  const handleCloseForm = () => {
    setShowForm(false);
    setEditingPool(null);
    form.resetFields();
  };

  const columns = [
    {
      title: t('nodes.nodePool.poolName'),
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: t('nodes.nodePool.labels'),
      dataIndex: 'labels',
      key: 'labels',
      render: (labels: Record<string, string>) => (
        <div>
          {Object.entries(labels || {}).map(([key, value]) => (
            <Text key={key} style={{ display: 'block' }}>
              {key}: {value}
            </Text>
          ))}
        </div>
      ),
    },
    {
      title: t('nodes.nodePool.taints'),
      dataIndex: 'taints',
      key: 'taints',
      render: (taints: NodePool['taints']) => (
        <div>
          {taints?.map((taint, index) => (
            <Text key={index} style={{ display: 'block' }}>
              {taint.key}={taint.value || ''}:{taint.effect}
            </Text>
          ))}
        </div>
      ),
    },
    {
      title: t('common.operations'),
      key: 'action',
      render: (_: any, record: NodePool) => (
        <Space>
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            {t('common.edit')}
          </Button>
          <Popconfirm
            title={t('nodes.nodePool.confirmDelete')}
            onConfirm={() => handleDelete(record.name)}
            okText={t('common.confirm')}
            cancelText={t('common.cancel')}
          >
            <Button
              type="text"
              danger
              icon={<DeleteOutlined />}
            >
              {t('common.delete')}
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Modal
      title={t('nodes.nodePool.manage')}
      open={visible}
      onCancel={onClose}
      width={1000} // 增加主模态框宽度从800px到1000px
      footer={null}
      bodyStyle={{ maxHeight: '80vh', overflow: 'auto' }} // 设置最大高度并添加滚动条
    >
      <div style={{ marginBottom: 16 }}>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => {
            setEditingPool(null);
            form.resetFields();
            setShowForm(true);
          }}
        >
          {t('nodes.nodePool.createPool')}
        </Button>
      </div>

      <Table
        dataSource={nodePools}
        columns={columns}
        rowKey="name"
        pagination={false}
        scroll={{ x: 'max-content' }} // 添加水平滚动支持
      />

      <Modal
        title={editingPool ? t('nodes.nodePool.editPool') : t('nodes.nodePool.addPool')}
        open={showForm}
        onCancel={handleCloseForm}
        footer={null}
        width={800} // 设置子模态框宽度为800px
        bodyStyle={{ maxHeight: '70vh', overflow: 'auto' }} // 添加滚动支持
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            labels: [],
            taints: [],
          }}
        >
          <Form.Item
            name="name"
            label={t('nodes.nodePool.poolName')}
            rules={[{ required: true, message: t('nodes.nodePool.pleaseEnterPoolName') }]}
          >
            <Input placeholder={t('nodes.nodePool.pleaseEnterPoolName')} disabled={!!editingPool} />
          </Form.Item>

          <Typography.Title level={5} style={{ marginTop: 16 }}>{t('nodes.nodePool.labels')}</Typography.Title>
          <Form.List name="labels">
            {(fields, { add, remove }) => (
              <>
                {fields.map(({ key, name, ...restField }) => (
                  <Space key={key} style={{ display: 'flex', marginBottom: 8, width: '100%' }} align="baseline" wrap>
                    <Form.Item
                      {...restField}
                      name={[name, 'key']}
                      rules={[{ required: true, message: t('nodes.nodePool.pleaseEnterLabelKey') }]}
                      style={{ minWidth: '200px', marginRight: '8px' }}
                    >
                      <Input placeholder={t('nodes.nodePool.labelKey')} />
                    </Form.Item>
                    <Form.Item
                      {...restField}
                      name={[name, 'value']}
                      rules={[{ required: true, message: t('nodes.nodePool.pleaseEnterLabelValue') }]}
                      style={{ minWidth: '200px', marginRight: '8px' }}
                    >
                      <Input placeholder={t('nodes.nodePool.labelValue')} />
                    </Form.Item>
                    <Button type="text" danger onClick={() => remove(name)}>{t('common.delete')}</Button>
                  </Space>
                ))}
                <Form.Item>
                  <Button type="dashed" onClick={() => add()} block icon={<PlusOutlined />}>
                    {t('nodes.nodePool.addLabel')}
                  </Button>
                </Form.Item>
              </>
            )}
          </Form.List>

          <Typography.Title level={5} style={{ marginTop: 16 }}>{t('nodes.nodePool.taints')}</Typography.Title>
          <Form.List name="taints">
            {(fields, { add, remove }) => (
              <>
                {fields.map(({ key, name, ...restField }) => (
                  <div key={key} style={{ display: 'flex', marginBottom: 8, flexWrap: 'wrap', alignItems: 'flex-start' }}>
                    <Form.Item
                      {...restField}
                      name={[name, 'key']}
                      rules={[{ required: true, message: t('nodes.nodePool.pleaseEnterTaintKey') }]}
                      style={{ minWidth: '150px', marginRight: '8px' }}
                    >
                      <Input placeholder={t('nodes.nodePool.taintKey')} />
                    </Form.Item>
                    <Form.Item
                      {...restField}
                      name={[name, 'value']}
                      style={{ minWidth: '150px', marginRight: '8px' }}
                    >
                      <Input placeholder={t('nodes.nodePool.taintValue')} />
                    </Form.Item>
                    <Form.Item
                      {...restField}
                      name={[name, 'effect']}
                      rules={[{ required: true, message: t('nodes.nodePool.pleaseSelectEffect') }]}
                      style={{ minWidth: '200px', marginRight: '8px' }}
                    >
                      <Select placeholder={t('nodes.nodePool.pleaseSelectEffect')}>
                        <Select.Option value="NoSchedule">{t('nodes.nodePool.taintEffects.NoSchedule')}</Select.Option>
                        <Select.Option value="PreferNoSchedule">{t('nodes.nodePool.taintEffects.PreferNoSchedule')}</Select.Option>
                        <Select.Option value="NoExecute">{t('nodes.nodePool.taintEffects.NoExecute')}</Select.Option>
                      </Select>
                    </Form.Item>
                    <Button type="text" danger onClick={() => remove(name)} style={{ marginTop: '5px' }}>{t('common.delete')}</Button>
                  </div>
                ))}
                <Form.Item>
                  <Button type="dashed" onClick={() => add()} block icon={<PlusOutlined />}>
                    {t('nodes.nodePool.addTaint')}
                  </Button>
                </Form.Item>
              </>
            )}
          </Form.List>

          <Form.Item style={{ marginTop: 24 }}>
            <Space>
              <Button onClick={handleCloseForm}>
                {t('common.cancel')}
              </Button>
              <Button type="primary" htmlType="submit" loading={loading}>
                {editingPool ? t('common.update') : t('common.create')}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </Modal>
  );
};

export default NodePoolsManager;