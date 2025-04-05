import React, { useState } from 'react';
import { Modal, Form, Input, Button, Select, Radio, message } from 'antd';
import { useTranslation } from 'react-i18next';
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
  const { t } = useTranslation();
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
      message.success(t('nodes.addNodeModal.addSuccess'));
      onSuccess?.();
      onClose();
      form.resetFields();
    } catch (err: any) {
      message.error(t('nodes.addNodeModal.addFailed', { message: err.message }));
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      title={t('nodes.addNodeModal.title')}
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
          label={t('nodes.addNodeModal.nodeName')}
          rules={[{ required: true, message: t('nodes.addNodeModal.pleaseEnterNodeName') }]}
        >
          <Input placeholder={t('nodes.addNodeModal.nodeNamePlaceholder')} />
        </Form.Item>

        <Form.Item
          name="ip"
          label={t('nodes.addNodeModal.nodeIP')}
          rules={[{ required: true, message: t('nodes.addNodeModal.pleaseEnterNodeIP') }]}
        >
          <Input placeholder={t('nodes.addNodeModal.nodeIPPlaceholder')} />
        </Form.Item>

        <Form.Item
          name="nodePool"
          label={t('nodes.addNodeModal.nodePool')}
          rules={[{ required: false }]}
        >
          <Select
            placeholder={t('nodes.addNodeModal.selectNodePool')}
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
          label={t('nodes.addNodeModal.sshPort')}
          rules={[{ required: true, message: t('nodes.addNodeModal.pleaseEnterSSHPort') }]}
        >
          <Input type="number" placeholder="22" />
        </Form.Item>

        <Form.Item
          name="sshUser"
          label={t('nodes.addNodeModal.sshUser')}
          rules={[{ required: true, message: t('nodes.addNodeModal.pleaseEnterSSHUser') }]}
        >
          <Input placeholder="root" />
        </Form.Item>

        <Form.Item
          name="authType"
          label={t('nodes.addNodeModal.authType')}
          rules={[{ required: true, message: t('nodes.addNodeModal.pleaseEnterAuthType') }]}
        >
          <Radio.Group onChange={(e) => setAuthType(e.target.value)}>
            <Radio value="key">{t('nodes.addNodeModal.sshKey')}</Radio>
            <Radio value="password">{t('nodes.addNodeModal.password')}</Radio>
          </Radio.Group>
        </Form.Item>

        {authType === 'key' ? (
          <Form.Item
            name="sshKeyFile"
            label={t('nodes.addNodeModal.sshKeyFile')}
            rules={[{ required: true, message: t('nodes.addNodeModal.pleaseEnterSSHKeyFile') }]}
          >
            <Input placeholder={t('nodes.addNodeModal.sshKeyFilePlaceholder')} />
          </Form.Item>
        ) : (
          <Form.Item
            name="sshPassword"
            label={t('nodes.addNodeModal.sshPassword')}
            rules={[{ required: true, message: t('nodes.addNodeModal.pleaseEnterSSHPassword') }]}
          >
            <Input.Password placeholder={t('nodes.addNodeModal.pleaseEnterSSHPassword')} />
          </Form.Item>
        )}

        <Form.Item>
          <Button type="primary" htmlType="submit" loading={loading}>
            {t('nodes.addNodeModal.addNode')}
          </Button>
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default AddNodeModal;