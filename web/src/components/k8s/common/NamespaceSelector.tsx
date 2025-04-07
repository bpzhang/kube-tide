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

// 默认命名空间列表
const defaultNamespaces = ['default', 'kube-system', 'kube-public'];

// 全局缓存上次请求的命名空间数据
let lastFetchInfo: LastFetchInfo | null = null;

// 判断是否需要重新获取命名空间列表
// 如果与上次请求的集群不同，或者距离上次请求超过5分钟，则重新获取
const shouldFetchNamespaces = (clusterName: string): boolean => {
  if (!lastFetchInfo) return true;
  if (lastFetchInfo.cluster !== clusterName) return true;
  
  // 5分钟内使用缓存
  const now = Date.now();
  return (now - lastFetchInfo.timestamp) > 300000; // 5分钟 = 300000毫秒
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

  // 获取命名空间列表
  const fetchNamespaces = async (cluster: string) => {
    if (!cluster || !shouldFetchNamespaces(cluster)) {
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
        
        // 更新缓存信息
        setLastFetchInfo({
          cluster,
          timestamp: Date.now(),
          namespaces: namespacesList
        });
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

  // 当集群变化时获取命名空间列表
  useEffect(() => {
    // 如果有上次的缓存并且集群名相同，使用缓存数据
    if (lastFetchInfo && lastFetchInfo.cluster === clusterName && !shouldFetchNamespaces(clusterName)) {
      setNamespaces(lastFetchInfo.namespaces);
      return;
    }

    // 需要获取新数据
    if (clusterName) {
      fetchNamespaces(clusterName);
    }
  }, [clusterName, t]); // 只在clusterName变化时执行

  // 处理命名空间变化
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