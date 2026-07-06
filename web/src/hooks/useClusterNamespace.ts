import { useState, useEffect, useCallback } from 'react';
import { message } from 'antd';
import { getClusterList } from '@/api/cluster';

export function useClusterNamespace(t: (key: string) => string) {
  const [selectedCluster, setSelectedCluster] = useState('');
  const [clusters, setClusters] = useState<string[]>([]);
  const [namespace, setNamespace] = useState('default');
  const [clustersLoading, setClustersLoading] = useState(false);

  const fetchClusters = useCallback(async () => {
    try {
      setClustersLoading(true);
      const response = await getClusterList();
      if (response.data.code === 0) {
        const list = response.data.data.clusters;
        setClusters(list);
        if (list.length > 0) {
          setSelectedCluster((prev) => prev || list[0]);
        }
      }
    } catch {
      message.error(t('clusters.fetchFailed'));
    } finally {
      setClustersLoading(false);
    }
  }, [t]);

  useEffect(() => {
    fetchClusters();
  }, [fetchClusters]);

  return {
    selectedCluster,
    setSelectedCluster,
    clusters,
    namespace,
    setNamespace,
    clustersLoading,
    fetchClusters,
  };
}
