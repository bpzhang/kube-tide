import React, { useState } from 'react';
import { Modal, Form, InputNumber, message, Slider, Radio } from 'antd';
import { useTranslation } from 'react-i18next';
import { scaleStatefulSet } from '@/api/statefulset';

interface ScaleStatefulSetModalProps {
  visible: boolean;
  onCancel: () => void;
  onSuccess: () => void;
  clusterName: string;
  namespace: string;
  statefulsetName: string;
  currentReplicas: number;
}

/**
 * StatefulSet扩缩容模态框组件
 */
const ScaleStatefulSetModal: React.FC<ScaleStatefulSetModalProps> = ({ 
  visible, 
  onCancel, 
  onSuccess, 
  clusterName, 
  namespace, 
  statefulsetName, 
  currentReplicas 
}) => {
  const { t } = useTranslation();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [replicaMethod, setReplicaMethod] = useState<'input' | 'slider'>('input');

  // 初始化表单
  const initialValues = {
    replicas: currentReplicas,
  };

  // 提交表单
  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setLoading(true);
      
      const response = await scaleStatefulSet(
        clusterName,
        namespace,
        statefulsetName,
        values.replicas
      );
      
      if (response.data.code === 0) {
        message.success(t('statefulsets.scaleSuccess'));
        onSuccess();
      } else {
        message.error(response.data.message || t('statefulsets.scaleFailed'));
      }
    } catch (err) {
      console.error('Scale StatefulSet error:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      title={t('statefulsets.scale')}
      open={visible}
      onCancel={onCancel}
      onOk={handleSubmit}
      confirmLoading={loading}
      destroyOnClose
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={initialValues}
        preserve={false}
      >
        <Form.Item>
          <Radio.Group 
            value={replicaMethod} 
            onChange={e => setReplicaMethod(e.target.value)}
            buttonStyle="solid"
          >
            <Radio.Button value="input">{t('common.manualInput')}</Radio.Button>
            <Radio.Button value="slider">{t('common.slider')}</Radio.Button>
          </Radio.Group>
        </Form.Item>
        
        {replicaMethod === 'input' ? (
          <Form.Item
            name="replicas"
            label={t('statefulsets.replicas')}
            rules={[
              { required: true, message: t('statefulsets.pleaseEnterReplicas') },
              { type: 'number', min: 0, message: t('statefulsets.replicasMustBePositive') }
            ]}
          >
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
        ) : (
          <Form.Item
            name="replicas"
            label={t('statefulsets.replicas')}
            rules={[{ required: true, message: t('statefulsets.pleaseEnterReplicas') }]}
          >
            <Slider min={0} max={20} marks={{ 0: '0', 5: '5', 10: '10', 15: '15', 20: '20' }} />
          </Form.Item>
        )}
        
        <div style={{ color: '#666', marginTop: 8 }}>
          {t('statefulsets.currentReplicas')}: {currentReplicas}
        </div>
      </Form>
    </Modal>
  );
};

export default ScaleStatefulSetModal;
