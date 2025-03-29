import React, { useState } from 'react';
import { Modal, Form, Input, Button, Select, Radio, message } from 'antd';
import { addNode } from '@/api/node';
import type { AddNodeRequest } from '@/api/node';
import type { NodePool } from '@/api/nodepool';

interface AddNodeModalProps {
  open: boolean;
  onClose: () => void;
  clusterName: string;
  onSuccess?: () => void;
  nodePools: NodePool[];
}

const AddNodeModal: React.FC<AddNodeModalProps> = ({
  open,
  onClose,
  clusterName,
  onSuccess,
  nodePools
}) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [authType, setAuthType] = useState<'key' | 'password'>('key');

  const handleSubmit = async (values: any) => {
    setLoading(true);
    try {
      const nodeConfig: AddNodeRequest = {
        name: values.name,
        ip: values.ip,
        nodePool: values.nodePool,
        sshPort: values.sshPort || 22,
        sshUser: values.sshUser || 'root',
        authType: values.authType,
      };

      if (values.authType === 'password') {
        nodeConfig.sshPassword = values.sshPassword;
      } else {
        nodeConfig.sshKeyFile = values.sshKeyFile;
      }

      await addNode(clusterName, nodeConfig);
      message.success('节点添加成功');
      onSuccess?.();
      onClose();
      form.resetFields();
    } catch (err: any) {
      message.error(`添加节点失败: ${err.message}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      title="添加节点"
      open={open}
      onCancel={onClose}
      footer={null}
      width={600}
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={handleSubmit}
        initialValues={{
          authType: 'key',
          sshPort: 22,
          sshUser: 'root'
        }}
      >
        <Form.Item
          name="name"
          label="节点名称"
          rules={[{ required: true, message: '请输入节点名称' }]}
        >
          <Input placeholder="例如: worker-1" />
        </Form.Item>

        <Form.Item
          name="ip"
          label="节点IP地址"
          rules={[{ required: true, message: '请输入节点IP地址' }]}
        >
          <Input placeholder="例如: 192.168.1.100" />
        </Form.Item>

        <Form.Item
          name="nodePool"
          label="节点池"
          rules={[{ required: false }]}
        >
          <Select
            placeholder="选择节点池（可选）"
            allowClear
          >
            {nodePools.map(pool => (
              <Select.Option key={pool.name} value={pool.name}>
                {pool.name}
              </Select.Option>
            ))}
          </Select>
        </Form.Item>

        <Form.Item
          name="sshPort"
          label="SSH端口"
          rules={[{ required: true, message: '请输入SSH端口' }]}
        >
          <Input type="number" placeholder="22" />
        </Form.Item>

        <Form.Item
          name="sshUser"
          label="SSH用户名"
          rules={[{ required: true, message: '请输入SSH用户名' }]}
        >
          <Input placeholder="root" />
        </Form.Item>

        <Form.Item
          name="authType"
          label="认证方式"
          rules={[{ required: true, message: '请选择认证方式' }]}
        >
          <Radio.Group onChange={(e) => setAuthType(e.target.value)}>
            <Radio value="key">SSH密钥</Radio>
            <Radio value="password">密码</Radio>
          </Radio.Group>
        </Form.Item>

        {authType === 'key' ? (
          <Form.Item
            name="sshKeyFile"
            label="SSH密钥文件路径"
            rules={[{ required: true, message: '请输入SSH密钥文件路径' }]}
          >
            <Input placeholder="例如: /root/.ssh/id_rsa" />
          </Form.Item>
        ) : (
          <Form.Item
            name="sshPassword"
            label="SSH密码"
            rules={[{ required: true, message: '请输入SSH密码' }]}
          >
            <Input.Password placeholder="请输入SSH密码" />
          </Form.Item>
        )}

        <Form.Item>
          <Button type="primary" htmlType="submit" loading={loading}>
            添加节点
          </Button>
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default AddNodeModal;