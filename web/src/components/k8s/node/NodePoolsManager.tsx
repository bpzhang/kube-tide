import React, { useState } from 'react';
import { Modal, Form, Input, Button, Space, Table, Typography, message, Popconfirm, Select } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
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
        message.success('节点池更新成功');
      } else {
        await createNodePool(clusterName, pool);
        message.success('节点池创建成功');
      }

      form.resetFields();
      setEditingPool(null);
      setShowForm(false);
      onSuccess();
    } catch (err) {
      message.error(`操作失败: ${(err as Error).message}`);
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
      message.success('节点池删除成功');
      onSuccess();
    } catch (err) {
      message.error(`删除失败: ${(err as Error).message}`);
    }
  };

  const handleCloseForm = () => {
    setShowForm(false);
    setEditingPool(null);
    form.resetFields();
  };

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '标签',
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
      title: '污点',
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
      title: '操作',
      key: 'action',
      render: (_: any, record: NodePool) => (
        <Space>
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定要删除此节点池吗？"
            onConfirm={() => handleDelete(record.name)}
            okText="确定"
            cancelText="取消"
          >
            <Button
              type="text"
              danger
              icon={<DeleteOutlined />}
            >
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Modal
      title="节点池管理"
      open={visible}
      onCancel={onClose}
      width={800}
      footer={null}
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
          创建节点池
        </Button>
      </div>

      <Table
        dataSource={nodePools}
        columns={columns}
        rowKey="name"
        pagination={false}
      />

      <Modal
        title={editingPool ? '编辑节点池' : '创建节点池'}
        open={showForm}
        onCancel={handleCloseForm}
        footer={null}
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
            label="节点池名称"
            rules={[{ required: true, message: '请输入节点池名称' }]}
          >
            <Input placeholder="请输入节点池名称" disabled={!!editingPool} />
          </Form.Item>

          <Form.List name="labels">
            {(fields, { add, remove }) => (
              <>
                {fields.map(({ key, name, ...restField }) => (
                  <Space key={key} style={{ display: 'flex', marginBottom: 8 }} align="baseline">
                    <Form.Item
                      {...restField}
                      name={[name, 'key']}
                      rules={[{ required: true, message: '请输入标签键' }]}
                    >
                      <Input placeholder="标签键" />
                    </Form.Item>
                    <Form.Item
                      {...restField}
                      name={[name, 'value']}
                      rules={[{ required: true, message: '请输入标签值' }]}
                    >
                      <Input placeholder="标签值" />
                    </Form.Item>
                    <Button type="text" onClick={() => remove(name)}>删除</Button>
                  </Space>
                ))}
                <Form.Item>
                  <Button type="dashed" onClick={() => add()} block icon={<PlusOutlined />}>
                    添加标签
                  </Button>
                </Form.Item>
              </>
            )}
          </Form.List>

          <Form.List name="taints">
            {(fields, { add, remove }) => (
              <>
                {fields.map(({ key, name, ...restField }) => (
                  <Space key={key} style={{ display: 'flex', marginBottom: 8 }} align="baseline">
                    <Form.Item
                      {...restField}
                      name={[name, 'key']}
                      rules={[{ required: true, message: '请输入污点键' }]}
                    >
                      <Input placeholder="污点键" />
                    </Form.Item>
                    <Form.Item
                      {...restField}
                      name={[name, 'value']}
                    >
                      <Input placeholder="污点值（可选）" />
                    </Form.Item>
                    <Form.Item
                      {...restField}
                      name={[name, 'effect']}
                      rules={[{ required: true, message: '请选择污点效果' }]}
                    >
                      <Select placeholder="选择效果">
                        <Select.Option value="NoSchedule">NoSchedule (不允许新Pod调度到节点)</Select.Option>
                        <Select.Option value="PreferNoSchedule">PreferNoSchedule (尽量避免新Pod调度到节点)</Select.Option>
                        <Select.Option value="NoExecute">NoExecute (不允许新Pod调度且驱逐现有Pod)</Select.Option>
                      </Select>
                    </Form.Item>
                    <Button type="text" onClick={() => remove(name)}>删除</Button>
                  </Space>
                ))}
                <Form.Item>
                  <Button type="dashed" onClick={() => add()} block icon={<PlusOutlined />}>
                    添加污点
                  </Button>
                </Form.Item>
              </>
            )}
          </Form.List>

          <Form.Item>
            <Space>
              <Button onClick={handleCloseForm}>
                取消
              </Button>
              <Button type="primary" htmlType="submit" loading={loading}>
                {editingPool ? '更新' : '创建'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </Modal>
  );
};

export default NodePoolsManager;