import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { Spin, message } from 'antd';
import { getPodDetails } from '@/api/pod';
import PodLogs from '@/components/k8s/pod/PodLogs';
import { useTranslation } from 'react-i18next';

const PodLogsPage: React.FC = () => {
  const { t } = useTranslation();
  const { clusterName, namespace, podName } = useParams<{
    clusterName: string;
    namespace: string;
    podName: string;
  }>();
  const [loading, setLoading] = useState(true);
  const [containerNames, setContainerNames] = useState<string[]>([]);

  useEffect(() => {
    const fetchPodDetails = async () => {
      if (!clusterName || !namespace || !podName) {
        message.error(t('common.error'));
        window.close();
        return;
      }

      try {
        const response = await getPodDetails(clusterName, namespace, podName);
        if (response.data.code === 0) {
          // 提取Pod的所有容器名称
          if (response.data.data.pod?.spec?.containers) {
            const containers = response.data.data.pod.spec.containers.map((container: any) => container.name);
            setContainerNames(containers);
          }
        } else {
          message.error(response.data.message || t('podDetail.fetchFailed'));
        }
      } catch (error) {
        console.error(t('podDetail.fetchFailed'), error);
        message.error(t('podDetail.fetchFailed'));
      } finally {
        setLoading(false);
      }
    };

    fetchPodDetails();
    
    // 设置页面标题
    document.title = `${t('podDetail.logs.title')} - ${podName}`;
    
    // 为页面设置样式以适应整个窗口
    document.body.style.margin = '0';
    document.body.style.padding = '0';
    document.body.style.overflow = 'hidden';
    document.body.style.height = '100vh';
    document.body.style.width = '100vw';
    
    return () => {
      // 恢复原始标题和样式
      document.title = 'Kube Tide';
      document.body.style.margin = '';
      document.body.style.padding = '';
      document.body.style.overflow = '';
      document.body.style.height = '';
      document.body.style.width = '';
    };
  }, [clusterName, namespace, podName, t]);

  if (loading) {
    return (
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        width: '100%',
        height: '100vh'
      }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div style={{
      display: 'flex',
      flexDirection: 'column',
      width: '100%',
      height: '100vh',
      overflow: 'hidden'
    }}>
      <PodLogs
        clusterName={clusterName!}
        namespace={namespace!}
        podName={podName!}
        containers={containerNames}
      />
    </div>
  );
};

export default PodLogsPage;