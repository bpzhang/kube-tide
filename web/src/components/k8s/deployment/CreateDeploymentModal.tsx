import React from 'react';
import { CreateDeploymentRequest, UpdateDeploymentRequest } from '@/api/deployment';
import DeploymentModal from './DeploymentModal';

interface CreateDeploymentModalProps {
  visible: boolean;
  onClose: () => void;
  onSubmit: (values: CreateDeploymentRequest) => Promise<void>;
  clusterName: string;
  namespace: string;
}

/**
 * 创建Deployment的模态框
 */
const CreateDeploymentModal: React.FC<CreateDeploymentModalProps> = ({
  visible,
  onClose,
  onSubmit,
  clusterName,
  namespace
}) => {
  // 创建适配器函数，使其符合 DeploymentModal 期望的类型
  const handleSubmit = async (values: CreateDeploymentRequest | UpdateDeploymentRequest) => {
    // 在创建模式下，values 一定是 CreateDeploymentRequest 类型
    await onSubmit(values as CreateDeploymentRequest);
  };

  return (
    <DeploymentModal
      visible={visible}
      onClose={onClose}
      onSubmit={handleSubmit}
      mode="create"
      clusterName={clusterName}
      namespace={namespace}
      title="创建 Deployment"
    />
  );
};

export default CreateDeploymentModal;