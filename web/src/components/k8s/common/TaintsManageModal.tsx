import React, { useState, useEffect } from 'react';
import { Modal, Form, Input, Select, Button, Tag, Tooltip, Space, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { NodeTaint, getNodeTaints, addNodeTaint, removeNodeTaint } from '@/api/node';

interface TaintsManageModalProps {
  open: boolean;
  onClose: () => void;
  nodeName: string;
  clusterName: string;
  onSuccess?: () => void;
}

const TaintsManageModal: React.FC<TaintsManageModalProps> = ({
  open,
  onClose,
  nodeName,
  clusterName,
}) => {
  const [form] = Form.useForm();
  const [taints, setTaints] = useState<NodeTaint[]>([]);
  const [loading, setLoading] = useState(false);

  // 获取节点污点
  const fetchTaints = async () => {
    try {
      const response = await getNodeTaints(clusterName, nodeName);
      if (response.data.code === 0) {
        setTaints(response.data.data.taints || []);
      }
    } catch (err) {
      message.error('获取节点污点失败');
    }
  };

  useEffect(() => {
    if (open) {
      fetchTaints();
    }
  }, [open, clusterName, nodeName]);

  // 添加污点
  const handleAddTaint = async (values: any) => {
    setLoading(true);
    try {
      await addNodeTaint(clusterName, nodeName, {
        key: values.key,
        value: values.value,
        effect: values.effect,
      });
      message.success('添加污点成功');
      form.resetFields();
      fetchTaints();
    } catch (err) {
      message.error('添加污点失败');
    } finally {
      setLoading(false);
    }
  };

  // 删除污点
  const handleRemoveTaint = async (key: string, effect: string) => {
    setLoading(true);
    try {
      await removeNodeTaint(clusterName, nodeName, key, effect);
      message.success('删除污点成功');
      fetchTaints();
    } catch (err) {
      message.error('删除污点失败');
    } finally {
      setLoading(false);
    }
  };

  // 获取污点标签颜色
  const getTaintColor = (effect: string) => {
    switch (effect) {
      case 'NoSchedule':
        return 'red';
      case 'PreferNoSchedule':
        return 'orange';
      case 'NoExecute':
        return 'volcano';
      default:
        return 'blue';
    }
  };

  return (
    <Modal
      title={`节点污点管理 - ${nodeName}`}
      open={open}
      onCancel={onClose}
      footer={null}
      width={600}
    >
      <div style={{ marginBottom: 16 }}>
        <h4>当前污点：</h4>
        <div style={{ marginBottom: 16 }}>
          {taints.length > 0 ? (
            <Space size={[0, 8]} wrap>
              {taints.map((taint, index) => (
                <Tooltip 
                  key={index}
                  title={`${taint.key}${taint.value ? '=' + taint.value : ''}`}
                >
                  <Tag
                    color={getTaintColor(taint.effect)}
                    closable
                    onClose={() => handleRemoveTaint(taint.key, taint.effect)}
                  >
                    {taint.key}: {taint.effect}
                  </Tag>
                </Tooltip>
              ))}
            </Space>
          ) : (
            <div style={{ color: '#999' }}>暂无污点</div>
          )}
        </div>
      </div>

      <div style={{ marginTop: 24 }}>
        <h4>添加污点：</h4>
        <Form
          form={form}
          onFinish={handleAddTaint}
          layout="vertical"
        >
          <Form.Item
            name="key"
            label="键(Key)"
            rules={[{ required: true, message: '请输入污点键' }]}
          >
            <Input placeholder="例如: node-role.kubernetes.io/master" />
          </Form.Item>

          <Form.Item
            name="value"
            label="值(Value)"
          >
            <Input placeholder="可选，例如: true" />
          </Form.Item>

          <Form.Item
            name="effect"
            label="效果(Effect)"
            rules={[{ required: true, message: '请选择污点效果' }]}
          >
            <Select>
              <Select.Option value="NoSchedule">NoSchedule (不允许新Pod调度到节点)</Select.Option>
              <Select.Option value="PreferNoSchedule">PreferNoSchedule (尽量避免新Pod调度到节点)</Select.Option>
              <Select.Option value="NoExecute">NoExecute (不允许新Pod调度且驱逐现有Pod)</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item>
            <Button 
              type="primary" 
              htmlType="submit" 
              icon={<PlusOutlined />}
              loading={loading}
            >
              添加污点
            </Button>
          </Form.Item>
        </Form>
      </div>
    </Modal>
  );
};

export default TaintsManageModal;