import React, { useState, useEffect } from 'react';
import { Modal, Form, Button, message } from 'antd';
import { CreateServiceRequest } from '@/api/service';
import ServiceForm from './ServiceForm';
import { ServiceFormData, processFormToCreateRequest } from './ServiceTypes';
import { useTranslation } from 'react-i18next';

interface CreateServiceModalProps {
  visible: boolean;
  onClose: () => void;
  onSubmit: (values: CreateServiceRequest) => Promise<void>;
  namespace: string;
}

const CreateServiceModal: React.FC<CreateServiceModalProps> = ({
  visible,
  onClose,
  onSubmit,
  namespace
}) => {
  const { t } = useTranslation();
  const [form] = Form.useForm();
  const [submitting, setSubmitting] = useState(false);

  // 初始化表单数据
  useEffect(() => {
    if (visible) {
      form.resetFields();
      form.setFieldsValue({
        namespace,
        type: 'ClusterIP',
        ports: [{
          protocol: 'TCP',
          port: 80,
          targetPort: 80
        }],
        labelsArray: [],
        selectorArray: []
      });
    }
  }, [visible, form, namespace]);

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setSubmitting(true);
      
      // 转换表单数据为API请求格式
      const serviceData = processFormToCreateRequest(values as ServiceFormData);
      
      await onSubmit(serviceData);
      message.success(t('services.createSuccess'));
      onClose();
    } catch (error) {
      console.error(t('services.createFailed'), error);
      message.error(t('services.createFailed') + ': ' + (error instanceof Error ? error.message : t('common.unknownError')));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal
      title={t('services.createService')}
      open={visible}
      onCancel={onClose}
      width={800}
      footer={[
        <Button key="cancel" onClick={onClose}>{t('common.cancel')}</Button>,
        <Button key="submit" type="primary" loading={submitting} onClick={handleSubmit}>{t('common.create')}</Button>
      ]}
    >
      <ServiceForm 
        form={form}
        mode="create"
        initialValues={{
          namespace,
          type: 'ClusterIP',
          ports: [{
            protocol: 'TCP',
            port: 80,
            targetPort: 80
          }],
          labelsArray: [],
          selectorArray: []
        }}
      />
    </Modal>
  );
};

export default CreateServiceModal;