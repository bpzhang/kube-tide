import React, { useState, useEffect } from 'react';
import { Modal, Form, Button, message } from 'antd';
import { DeploymentFormData, processFormToCreateRequest, processFormToUpdateRequest, processDeploymentToFormData } from './DeploymentTypes';
import DeploymentForm from './DeploymentForm';
import { CreateDeploymentRequest, UpdateDeploymentRequest } from '@/api/deployment';

export interface DeploymentModalProps {
  visible: boolean;
  onClose: () => void;
  mode: 'create' | 'edit';
  onSubmit: (values: CreateDeploymentRequest | UpdateDeploymentRequest) => Promise<void>;
  clusterName?: string;
  namespace?: string;
  deployment?: any;
  title?: string;
}

/**
 * 通用Deployment模态框组件
 * 可用于创建和编辑Deployment
 */
const DeploymentModal: React.FC<DeploymentModalProps> = ({
  visible,
  onClose,
  mode,
  onSubmit,
  clusterName,
  namespace,
  deployment,
  title
}) => {
  const [form] = Form.useForm();
  const [submitting, setSubmitting] = useState(false);
  const [formValues, setFormValues] = useState<DeploymentFormData | null>(null);
  
  // 表单初始值 - 创建模式
  const initialCreateValues: DeploymentFormData = {
    replicas: 1,
    strategy: 'RollingUpdate',
    containers: [
      {
        name: 'container-1',
        image: '',
        resources: {
          limits: {
            cpuValue: 500,
            cpuUnit: 'm',
            memoryValue: 512,
            memoryUnit: 'Mi',
          },
          requests: {
            cpuValue: 100,
            cpuUnit: 'm',
            memoryValue: 128,
            memoryUnit: 'Mi',
          }
        },
        env: []
      }
    ]
  };

  // 重置表单和初始化表单数据
  useEffect(() => {
    if (visible) {
      form.resetFields();
      
      if (mode === 'create') {
        // 创建模式：使用默认初始值
        form.setFieldsValue(initialCreateValues);
        setFormValues(initialCreateValues);
      } else if (mode === 'edit' && deployment) {
        // 编辑模式：将API返回的数据转换为表单数据格式
        const processedFormData = processDeploymentToFormData(deployment);
        setFormValues(processedFormData);
        form.setFieldsValue(processedFormData);
      }
    }
  }, [visible, deployment, form, mode]);

  // 处理提交
  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setSubmitting(true);
      
      // 根据模式处理表单数据
      let submitData;
      if (mode === 'create') {
        // 创建模式：转换为创建请求格式
        submitData = processFormToCreateRequest(values);
        if (clusterName) {
          submitData.clusterName = clusterName;
        }
        if (namespace) {
          submitData.namespace = namespace;
        }
      } else {
        // 编辑模式：转换为更新请求格式
        submitData = processFormToUpdateRequest(values);
      }
      
      // 调用API提交
      await onSubmit(submitData);
      message.success(`Deployment${mode === 'create' ? '创建' : '更新'}成功`);
      onClose();
    } catch (error) {
      console.error(`Error ${mode === 'create' ? 'creating' : 'updating'} deployment:`, error);
      message.error('表单验证失败或操作失败');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal
      title={title || `${mode === 'create' ? '创建' : '编辑'} Deployment${deployment?.name ? `: ${deployment.name}` : ''}`}
      open={visible}
      onCancel={onClose}
      width={800}
      footer={[
        <Button key="cancel" onClick={onClose}>
          取消
        </Button>,
        <Button key="submit" type="primary" loading={submitting} onClick={handleSubmit}>
          {mode === 'create' ? '创建' : '更新'}
        </Button>
      ]}
    >
      {(mode === 'create' || (mode === 'edit' && formValues)) && (
        <DeploymentForm 
          form={form}
          initialValues={formValues || initialCreateValues}
          mode={mode}
        />
      )}
    </Modal>
  );
};

export default DeploymentModal;