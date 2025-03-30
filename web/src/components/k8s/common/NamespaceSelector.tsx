import React, { useState, useEffect } from 'react';
import { Select, message } from 'antd';
import { getNamespaceList } from '@/api/namespace';

const { Option } = Select;

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

  // 当集群变化时获取命名空间列表
  useEffect(() => {
    if (!clusterName) return;
    
    const fetchNamespaces = async () => {
      setLoading(true);
      try {
        const response = await getNamespaceList(clusterName);
        if (response.data.code === 0) {
          // 如果返回的命名空间列表为空，则使用默认值
          const namespacesList = response.data.data.namespaces.length > 0 
            ? response.data.data.namespaces 
            : defaultNamespaces;
          setNamespaces(namespacesList);
        } else {
          message.error(response.data.message || '获取命名空间列表失败');
          setNamespaces(defaultNamespaces);
        }
      } catch (err) {
        console.warn('获取命名空间列表失败，使用默认值');
        setNamespaces(defaultNamespaces);
      } finally {
        setLoading(false);
      }
    };

    fetchNamespaces();
  }, [clusterName, defaultNamespaces]);

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