import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { Spin, message, Select } from 'antd';
import { getPodDetails } from '@/api/pod';
import PodTerminal from '@/components/k8s/pod/PodTerminal';
import { useTranslation } from 'react-i18next';

const { Option } = Select;

const PodTerminalPage: React.FC = () => {
  const { t } = useTranslation();
  const { clusterName, namespace, podName } = useParams<{
    clusterName: string;
    namespace: string;
    podName: string;
  }>();
  const [loading, setLoading] = useState(true);
  const [selectedContainer, setSelectedContainer] = useState<string>('');
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
          // 提取 Pod 的所有容器名称
          if (response.data.data.pod?.spec?.containers) {
            const containers = response.data.data.pod.spec.containers.map((container: any) => container.name);
            setContainerNames(containers);
            // 默认选择第一个容器
            if (containers.length > 0) {
              setSelectedContainer(containers[0]);
            }
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
    document.title = `${t('podDetail.tabs.terminal')} - ${podName}`;
    
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
      overflow: 'hidden',
      padding: '8px',
      boxSizing: 'border-box'
    }}>
      {/* 容器选择器 */}
      {/* <div style={{
        marginBottom: '8px',
        backgroundColor: '#f0f0f0',
        padding: '8px',
        borderRadius: '4px',
        display: 'flex',
        alignItems: 'center',
        flexWrap: 'wrap'
      }}>
        <span style={{ marginRight: '8px' }}>{t('podDetail.container')}:</span>
        <Select
          value={selectedContainer}
          onChange={setSelectedContainer}
          style={{ minWidth: '150px', maxWidth: '100%', flex: '1 0 auto' }}
          size="small"
        >
          {containerNames.map((container) => (
            <Option key={container} value={container}>
              {container}
            </Option>
          ))}
        </Select>
      </div> */}
      
      {/* 终端组件 */}
      {selectedContainer && (
        <div style={{
          flex: '1 1 auto',
          minHeight: '0',
          display: 'flex',
          flexDirection: 'column'
        }}>
          <PodTerminal
            clusterName={clusterName!}
            namespace={namespace!}
            podName={podName!}
            containerName={selectedContainer}
          />
        </div>
      )}
    </div>
  );
};

export default PodTerminalPage;