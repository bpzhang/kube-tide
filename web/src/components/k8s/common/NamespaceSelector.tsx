import React, { useState, useEffect, useMemo } from 'react';
import { Select, message } from 'antd';
import { getNamespaceList } from '@/api/namespace';

const { Option } = Select;

// 本地缓存，但使用React Hook方式实现而非全局变量
interface NamespaceSelectorProps {
  clusterName: string;  // 必须提供集群名称
  value?: string;      // 当前选中的命名空间
  onChange?: (namespace: string) => void; // 命名空间变更回调
  width?: number | string; // 选择器宽度
  disabled?: boolean;  // 是否禁用
  placeholder?: string; // 占位文本
  style?: React.CSSProperties; // 自定义样式
  defaultNamespaces?: string[]; // 默认的命名空间列表，当API调用失败时使用
}

/**
 * 命名空间选择器组件
 * 用于在各个页面中统一选择命名空间
 */
const NamespaceSelector: React.FC<NamespaceSelectorProps> = ({
  clusterName,
  value = 'default',
  onChange,
  width = 200,
  disabled = false,
  placeholder = '选择命名空间',
  style,
  defaultNamespaces = ['default', 'kube-system']
}) => {
  const [namespaces, setNamespaces] = useState<string[]>(defaultNamespaces);
  const [loading, setLoading] = useState<boolean>(false);
  
  // 使用useRef或useState来保存上次请求信息，避免使用模块级变量
  const [lastFetchInfo, setLastFetchInfo] = useState<{
    cluster: string; 
    timestamp: number;
    namespaces: string[]
  } | null>(null);

  // 判断是否需要获取新数据的函数
  const shouldFetchNamespaces = (currentCluster: string): boolean => {
    if (!currentCluster) return false;
    
    // 没有缓存或集群变化时需要重新获取
    if (!lastFetchInfo || lastFetchInfo.cluster !== currentCluster) {
      return true;
    }
    
    // 缓存时间超过30分钟需要重新获取（30 * 60 * 1000 = 1800000毫秒）
    const cacheAge = Date.now() - lastFetchInfo.timestamp;
    return cacheAge > 1800000;
  };

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
        message.error(response.data.message || '获取命名空间列表失败');
        setNamespaces(defaultNamespaces);
      }
    } catch (err) {
      console.warn('获取命名空间列表失败，使用默认值', err);
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
  }, [clusterName]); // 只在clusterName变化时执行

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
      {namespaces.map(ns => (
        <Option key={ns} value={ns}>{ns}</Option>
      ))}
    </Select>
  );
};

export default NamespaceSelector;