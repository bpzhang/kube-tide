import React, { useState, useEffect } from 'react';
import { Select, Space, Card, message } from 'antd';
import { getPodsByNamespace } from '../../api/pod';
import { getClusterList } from '../../api/cluster';
import PodList from '../../components/k8s/pod/PodList';

const { Option } = Select;

const Pods: React.FC = () => {
  const [selectedCluster, setSelectedCluster] = useState<string>('');
  const [clusters, setClusters] = useState<string[]>([]);
  const [namespace, setNamespace] = useState<string>('default');
  const [pods, setPods] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);

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
      message.error('获取集群列表失败');
    }
  };

  const fetchPods = async () => {
    if (!selectedCluster) return;
    
    setLoading(true);
    try {
      const response = await getPodsByNamespace(selectedCluster, namespace);
      if (response.data.code === 0) {
        setPods(response.data.data.pods || []);
      } else {
        message.error(response.data.message || '获取Pod列表失败');
        setPods([]);
      }
    } catch (err) {
      message.error('获取Pod列表失败');
      setPods([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchClusters();
  }, []);

  useEffect(() => {
    if (selectedCluster) {
      fetchPods();
      // 每30秒刷新一次
      const timer = setInterval(fetchPods, 30000);
      return () => clearInterval(timer);
    }
  }, [selectedCluster, namespace]);

  const handleClusterChange = (value: string) => {
    setSelectedCluster(value);
  };

  return (
    <Card
      title="Pod管理"
      extra={
        <Space>
          <span>集群:</span>
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
          <span>命名空间:</span>
          <Select 
            value={namespace} 
            onChange={setNamespace}
            style={{ width: 200 }}
          >
            <Option value="default">default</Option>
            <Option value="kube-system">kube-system</Option>
            {/* TODO: 通过API获取命名空间列表 */}
          </Select>
        </Space>
      }
      loading={loading}
    >
      <PodList
        clusterName={selectedCluster}
        namespace={namespace}
        pods={pods}
        onRefresh={fetchPods}
      />
    </Card>
  );
};

export default Pods;