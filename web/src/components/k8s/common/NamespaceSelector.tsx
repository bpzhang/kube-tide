import React, { useState, useEffect } from 'react';
import { Select, message } from 'antd';
import { getNamespaceList } from '@/api/namespace';
import { useTranslation } from 'react-i18next';

const { Option } = Select;

interface LastFetchInfo {
  cluster: string;
  timestamp: number;
  namespaces: string[];
}

interface NamespaceSelectorProps {
  clusterName: string;
  value?: string;
  onChange?: (value: string) => void;
  width?: number | string;
  placeholder?: string;
  disabled?: boolean;
  style?: React.CSSProperties;
}

// default namespaces
const defaultNamespaces = ['default', 'kube-system', 'kube-public'];

// global cache for last fetched namespace data
const namespaceCache: Record<string, LastFetchInfo> = {};

// Determine whether you need to re-get the namespace list
// If it is different from the last requested cluster, or is more than 5 minutes away from the last request, re-get it
const shouldFetchNamespaces = (clusterName: string): boolean => {
  const cachedInfo = namespaceCache[clusterName];
  if (!cachedInfo) return true;
  
  // use cache within 5 minutes
  const now = Date.now();
  return (now - cachedInfo.timestamp) > 300000; // 5 minutes = 300000 milliseconds
};

const NamespaceSelector: React.FC<NamespaceSelectorProps> = ({
  clusterName,
  value = 'default',
  onChange,
  width = 200,
  placeholder,
  disabled = false,
  style = {},
}) => {
  const { t } = useTranslation();
  const [namespaces, setNamespaces] = useState<string[]>(defaultNamespaces);
  const [loading, setLoading] = useState<boolean>(false);

  // get namespace list
  const fetchNamespaces = async (cluster: string) => {
    if (!cluster) return;

    // check cache 
    if (!shouldFetchNamespaces(cluster) && namespaceCache[cluster]) {
      setNamespaces(namespaceCache[cluster].namespaces);
      return;
    }

    setLoading(true);
    try {
      const response = await getNamespaceList(cluster);
      if (response.data.code === 0) {
        const namespacesList = response.data.data.namespaces.length > 0 
          ? response.data.data.namespaces 
          : defaultNamespaces;
        
        setNamespaces(namespacesList);
        
        // update cache
        // If the namespace list is empty, use defaultNamespaces
        namespaceCache[cluster] = {
          cluster,
          timestamp: Date.now(),
          namespaces: namespacesList
        };
      } else {
        message.error(response.data.message || t('namespaceSelector.fetchFailed'));
        setNamespaces(defaultNamespaces);
      }
    } catch (err) {
      console.warn(t('namespaceSelector.fetchError'), err);
      setNamespaces(defaultNamespaces);
    } finally {
      setLoading(false);
    }
  };

  // fetch namespaces when component mounts or clusterName changes
  useEffect(() => {
    if (clusterName) {
      fetchNamespaces(clusterName);
    }
  }, [clusterName]); // fetch namespaces when component mounts or clusterName changes

  // handle namespace change
  const handleNamespaceChange = (value: string) => {
    if (onChange) {
      onChange(value);
    }
  };

  return (
    <Select
      value={value}
      onChange={handleNamespaceChange}
      style={{ width, ...style }}
      loading={loading}
      disabled={disabled || !clusterName}
      placeholder={placeholder}
      showSearch
      filterOption={(input, option) => 
        (option?.children as unknown as string)
          .toLowerCase()
          .includes(input.toLowerCase())
      }
    >
      {namespaces.map(namespace => (
        <Option key={namespace} value={namespace}>{namespace}</Option>
      ))}
    </Select>
  );
};

export default NamespaceSelector;