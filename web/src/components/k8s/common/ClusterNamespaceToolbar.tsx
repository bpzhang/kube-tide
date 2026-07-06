import React from 'react';
import { Select, Space } from 'antd';
import NamespaceSelector from './NamespaceSelector';
import { useTranslation } from 'react-i18next';

const { Option } = Select;

interface ClusterNamespaceToolbarProps {
  selectedCluster: string;
  clusters: string[];
  namespace: string;
  onClusterChange: (value: string) => void;
  onNamespaceChange: (value: string) => void;
  loading?: boolean;
  extra?: React.ReactNode;
  showNamespace?: boolean;
}

const ClusterNamespaceToolbar: React.FC<ClusterNamespaceToolbarProps> = ({
  selectedCluster,
  clusters,
  namespace,
  onClusterChange,
  onNamespaceChange,
  loading = false,
  extra,
  showNamespace = true,
}) => {
  const { t } = useTranslation();

  return (
    <Space wrap>
      <span>{t('common.cluster')}</span>
      <Select
        value={selectedCluster || undefined}
        onChange={onClusterChange}
        style={{ width: 200 }}
        loading={loading}
      >
        {clusters.map((cluster) => (
          <Option key={cluster} value={cluster}>
            {cluster}
          </Option>
        ))}
      </Select>
      {showNamespace && (
        <>
          <span>{t('common.namespace')}</span>
          <NamespaceSelector
            clusterName={selectedCluster}
            value={namespace}
            onChange={onNamespaceChange}
          />
        </>
      )}
      {extra}
    </Space>
  );
};

export default ClusterNamespaceToolbar;
