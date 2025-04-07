import React, { useState, useEffect } from 'react';
import { Modal, Form, Button, message } from 'antd';
import { updateService } from '@/api/service';
import ServiceForm from './ServiceForm';
import { ServiceFormData, processFormToUpdateRequest, processServiceToFormData } from './ServiceTypes';
import { useTranslation } from 'react-i18next';

interface EditServiceModalProps {
  visible: boolean;
  onClose: () => void;
  service: any;
  clusterName: string;
  onSuccess: () => void;
}

export const EditServiceModal: React.FC<EditServiceModalProps> = ({
  visible,
  onClose,
  service,
  clusterName,
  onSuccess,
}) => {
  const { t } = useTranslation();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);

  // 初始化表单数据
  useEffect(() => {
    if (visible && service) {
      // 使用工具函数将服务数据转换为表单格式
      const formData = processServiceToFormData(service);
      form.setFieldsValue(formData);
    }
  }, [visible, service, form]);

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setLoading(true);
      
      // 使用工具函数将表单数据转换为API请求格式
      const updateData = processFormToUpdateRequest(values as ServiceFormData);
      
      await updateService(clusterName, service.metadata.namespace, service.metadata.name, updateData);
      message.success(t('services.updateSuccess'));
      onSuccess();
      onClose();
    } catch (err) {
      console.error(t('services.updateFailed'), err);
      message.error(t('services.updateFailed'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      title={t('services.editService') + ': ' + service?.metadata?.name}
      open={visible}
      onCancel={onClose}
      width={800}
      footer={[
        <Button key="cancel" onClick={onClose}>{t('common.cancel')}</Button>,
        <Button key="submit" type="primary" loading={loading} onClick={handleSubmit}>{t('common.update')}</Button>,
      ]}
    >
      <ServiceForm 
        form={form}
        mode="edit"
        initialValues={service ? processServiceToFormData(service) : {
          type: 'ClusterIP',
          ports: [],
          labelsArray: [],
          selectorArray: []
        }}
      />
    </Modal>
  );
};