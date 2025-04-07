import React, { useState, useEffect } from 'react';
import { Modal, Form, Input, Button, Select, Space, Tag, Tooltip, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { NodeTaint, getNodeTaints, addNodeTaint, removeNodeTaint } from '@/api/node';
import { useTranslation } from 'react-i18next';

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
  const { t } = useTranslation();
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
      message.error(t('taintsManage.fetchFailed'));
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
      message.success(t('taintsManage.addSuccess'));
      form.resetFields();
      fetchTaints();
    } catch (err) {
      message.error(t('taintsManage.addFailed'));
    } finally {
      setLoading(false);
    }
  };

  // 删除污点
  const handleRemoveTaint = async (key: string, effect: string) => {
    setLoading(true);
    try {
      await removeNodeTaint(clusterName, nodeName, key, effect);
      message.success(t('taintsManage.deleteSuccess'));
      fetchTaints();
    } catch (err) {
      message.error(t('taintsManage.deleteFailed'));
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
      title={t('taintsManage.title', { nodeName })}
      open={open}
      onCancel={onClose}
      footer={null}
      width={600}
    >
      <div style={{ marginBottom: 16 }}>
        <h4>{t('taintsManage.currentTaints')}</h4>
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
            <div style={{ color: '#999' }}>{t('taintsManage.noTaints')}</div>
          )}
        </div>
      </div>

      <div style={{ marginTop: 24 }}>
        <h4>{t('taintsManage.addTaint')}</h4>
        <Form
          form={form}
          onFinish={handleAddTaint}
          layout="vertical"
        >
          <Form.Item
            name="key"
            label={t('taintsManage.keyLabel')}
            rules={[{ required: true, message: t('taintsManage.pleaseEnterKey') }]}
          >
            <Input placeholder={t('taintsManage.keyPlaceholder')} />
          </Form.Item>

          <Form.Item
            name="value"
            label={t('taintsManage.valueLabel')}
          >
            <Input placeholder={t('taintsManage.valuePlaceholder')} />
          </Form.Item>

          <Form.Item
            name="effect"
            label={t('taintsManage.effectLabel')}
            rules={[{ required: true, message: t('taintsManage.pleaseSelectEffect') }]}
          >
            <Select>
              <Select.Option value="NoSchedule">{t('taintsManage.effects.noSchedule')}</Select.Option>
              <Select.Option value="PreferNoSchedule">{t('taintsManage.effects.preferNoSchedule')}</Select.Option>
              <Select.Option value="NoExecute">{t('taintsManage.effects.noExecute')}</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item>
            <Button 
              type="primary" 
              htmlType="submit" 
              icon={<PlusOutlined />}
              loading={loading}
            >
              {t('taintsManage.addButton')}
            </Button>
          </Form.Item>
        </Form>
      </div>
    </Modal>
  );
};

export default TaintsManageModal;