import React, { useState, useEffect } from 'react';
import { Modal, Form, Button, message } from 'antd';
import { updateService } from '@/api/service';
import ServiceForm from './ServiceForm';
import { ServiceFormData, processFormToUpdateRequest, processServiceToFormData } from './ServiceTypes';

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
      message.success('服务更新成功');
      onSuccess();
      onClose();
    } catch (err) {
      console.error('更新服务失败:', err);
      message.error('更新服务失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      title={`编辑服务: ${service?.metadata?.name}`}
      open={visible}
      onCancel={onClose}
      width={800}
      footer={[
        <Button key="cancel" onClick={onClose}>取消</Button>,
        <Button key="submit" type="primary" loading={loading} onClick={handleSubmit}>更新</Button>,
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