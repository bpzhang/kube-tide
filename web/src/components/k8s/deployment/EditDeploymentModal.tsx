import React from 'react';
import { UpdateDeploymentRequest } from '@/api/deployment';
import DeploymentModal from './DeploymentModal';
import { useTranslation } from 'react-i18next';

interface EditDeploymentModalProps {
  visible: boolean;
  onClose: () => void;
  onSubmit: (values: UpdateDeploymentRequest) => Promise<void>;
  deployment: {
    name: string;
    namespace: string;
    replicas: number;
    readyReplicas: number;
    strategy: string;
    labels: { [key: string]: string };
    annotations: { [key: string]: string };
    containers: any[];
    minReadySeconds?: number;
    revisionHistoryLimit?: number;
    paused?: boolean;
    selector?: { [key: string]: string };
  };
}

/**
 * 编辑Deployment的模态框
 */
const EditDeploymentModal: React.FC<EditDeploymentModalProps> = ({
  visible,
  onClose,
  onSubmit,
  deployment
}) => {
  const { t } = useTranslation();
  
  return (
    <DeploymentModal
      visible={visible}
      onClose={onClose}
      onSubmit={onSubmit}
      mode="edit"
      deployment={deployment}
      title={t('deployments.editDeployment') + ': ' + deployment?.name}
    />
  );
};

export default EditDeploymentModal;