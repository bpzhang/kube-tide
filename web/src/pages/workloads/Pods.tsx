import React, { useState, useEffect, useCallback } from 'react';
import { Select, Space, Card, message, Spin } from 'antd';
import { useTranslation } from 'react-i18next';
import { getPodsByNamespace } from '../../api/pod';
import { getClusterList } from '../../api/cluster';
import PodList from '../../components/k8s/pod/PodList';
import NamespaceSelector from '../../components/k8s/common/NamespaceSelector';

const { Option } = Select;

const Pods: React.FC = () => {
  const { t } = useTranslation();
  const [selectedCluster, setSelectedCluster] = useState<string>('');
  const [clusters, setClusters] = useState<string[]>([]);
  const [namespace, setNamespace] = useState<string>('default');
  const [pods, setPods] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [isChangeParams, setIsChangeParams] = useState(false); // 标记是参数变更还是刷新

  // 获取集群列表
  const fetchClusters = async () => {
    try {
      const response = await getClusterList();
      if (response.data.code === 0) {
        setClusters(response.data.data.clusters);
        if (response.data.data.clusters.length > 0 && !selectedCluster) {
          setSelectedCluster(response.data.data.clusters[0]);
        }
      }
    } catch (err) {
      message.error(t('clusters.fetchFailed'));
    }
  };

  // 获取 Pod 列表，使用 useCallback 确保函数引用稳定
  const fetchPods = useCallback(async (isParamChange = false) => {
    if (!selectedCluster) return;
    
    setLoading(true);
    setIsChangeParams(isParamChange); // 设置是否为参数变更触发的刷新
    
    try {
      const response = await getPodsByNamespace(selectedCluster, namespace);
      if (response.data.code === 0) {
        setPods(response.data.data.pods || []);
      } else {
        message.error(response.data.message || t('pods.fetchFailed'));
        setPods([]);
      }
    } catch (err) {
      message.error(t('pods.fetchFailed'));
      setPods([]);
    } finally {
      setLoading(false);
    }
  }, [selectedCluster, namespace, t]);

  // 初始化加载
  useEffect(() => {
    fetchClusters();
  }, []);

  // 当集群或命名空间变化时重新获取Pod列表
  useEffect(() => {
    if (selectedCluster) {
      fetchPods(true); // 传递 true 表示这是参数变更
      
      // 每30秒刷新一次
      const timer = setInterval(() => fetchPods(false), 30000);
      return () => clearInterval(timer);
    }
  }, [selectedCluster, namespace, fetchPods]);

  // 处理集群变化
  const handleClusterChange = (value: string) => {
    setSelectedCluster(value);
  };

  // 处理命名空间变化
  const handleNamespaceChange = (value: string) => {
    setNamespace(value);
  };

  // 刷新Pod列表
  const handleRefresh = () => {
    fetchPods(false);
  };

  return (
    <Card
      title={t('pods.management')}
      extra={
        <Space>
          <span>{t('pods.cluster')}</span>
          <Select
            value={selectedCluster}
            onChange={handleClusterChange}
            style={{ width: 200 }}
            loading={loading}
          >
            {clusters.map(cluster => (
              <Option key={cluster} value={cluster}>{cluster}</Option>
            ))}
          </Select>
          <span>{t('pods.namespace')}:</span>
          <NamespaceSelector
            clusterName={selectedCluster}
            value={namespace}
            onChange={handleNamespaceChange}
          />
        </Space>
      }
    >
      {/* 只有在初始加载时显示整体加载状态 */}
      {loading && pods.length === 0 ? (
        <div style={{ padding: '40px 0', textAlign: 'center' }}>
          <Spin size="large">
            <div style={{ padding: '50px', textAlign: 'center' }}>
              <p>{t('pods.loading')}</p>
            </div>
          </Spin>
        </div>
      ) : (
        <PodList
          clusterName={selectedCluster}
          namespace={namespace}
          pods={pods}
          onRefresh={handleRefresh}
          isParamChange={isChangeParams}
        />
      )}
    </Card>
  );
};

export default Pods;