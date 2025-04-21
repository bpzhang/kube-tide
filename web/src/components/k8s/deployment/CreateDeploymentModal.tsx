import React from 'react';
import { CreateDeploymentRequest, UpdateDeploymentRequest } from '@/api/deployment';
import DeploymentModal from './DeploymentModal';
import { useTranslation } from 'react-i18next';

interface CreateDeploymentModalProps {
  visible: boolean;
  onClose: () => void;
  onSubmit: (values: CreateDeploymentRequest) => Promise<void>;
  clusterName: string;
  namespace: string;
}

/**
 * create deployment modal
 */
const CreateDeploymentModal: React.FC<CreateDeploymentModalProps> = ({
  visible,
  onClose,
  onSubmit,
  clusterName,
  namespace
}) => {
  const { t } = useTranslation();
  
  // create adapter function to match DeploymentModal's expected type
  const handleSubmit = async (values: CreateDeploymentRequest | UpdateDeploymentRequest) => {
    // in create mode, values must be of CreateDeploymentRequest type
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
      title={t('deployments.createDeployment')}
    />
  );
};

export default CreateDeploymentModal;