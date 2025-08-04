import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Button, Spin, message, Tabs, Space } from 'antd';
import { ArrowLeftOutlined, CodeOutlined, FileTextOutlined, SettingOutlined } from '@ant-design/icons';
import { getPodDetails } from '../../api/pod';
import PodDetail from '../../components/k8s/pod/PodDetail';
import PodMonitoring from '../../components/k8s/pod/PodMonitoring';
import PodRestartPolicyConfig from '../../components/k8s/pod/PodRestartPolicyConfig';
import { useTranslation } from 'react-i18next';

const PodDetailPage: React.FC = () => {
  const { t } = useTranslation();
  const { clusterName, namespace, podName } = useParams<{
    clusterName: string;
    namespace: string;
    podName: string;
  }>();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [pod, setPod] = useState<any>(null);

  useEffect(() => {
    const fetchPodDetails = async () => {
      if (!clusterName || !namespace || !podName) {
        message.error(t('common.error'));
        navigate('/workloads/pods');
        return;
      }

      try {
        const response = await getPodDetails(clusterName, namespace, podName);
        if (response.data.code === 0) {
          setPod(response.data.data.pod);
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
  }, [clusterName, namespace, podName, navigate, t]);

  // 在新标签页中打开日志页面
  const openLogsInNewTab = () => {
    const url = `/workloads/pods/${clusterName}/${namespace}/${podName}/logs`;
    window.open(url, '_blank');
  };

  // 在新标签页中打开终端页面
  const openTerminalInNewTab = () => {
    const url = `/workloads/pods/${clusterName}/${namespace}/${podName}/terminal`;
    window.open(url, '_blank');
  };

  if (loading) {
    return (
      <div style={{ padding: '24px', textAlign: 'center' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!pod) {
    return (
      <div style={{ padding: '24px', textAlign: 'center' }}>
        <h2>{t('podDetail.podNotFound')}</h2>
        <Button type="primary" onClick={() => navigate('/workloads/pods')}>
          {t('podDetail.backToPods')}
        </Button>
      </div>
    );
  }

  const tabItems = [
    {
      key: 'info',
      label: t('podDetail.tabs.info'),
      children: <PodDetail pod={pod} clusterName={clusterName!} />
    },
    {
      key: 'monitoring',
      label: t('podDetail.tabs.monitoring'),
      children: (
        <PodMonitoring
          clusterName={clusterName!}
          namespace={namespace!}
          podName={podName!}
        />
      )
    },
    {
      key: 'restartPolicy',
      label: (
        <span>
          <SettingOutlined />
          {t('pod.restartPolicy.configure')}
        </span>
      ),
      children: (
        <PodRestartPolicyConfig
          clusterName={clusterName!}
          namespace={namespace!}
          podName={podName!}
          disabled={pod.status.phase === 'Terminating' || pod.status.phase === 'Failed'}
        />
      )
    }
  ];

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '16px' }}>
        <Button
          icon={<ArrowLeftOutlined />}
          onClick={() => navigate('/workloads/pods')}
        >
          {t('podDetail.backToPods')}
        </Button>
        
        <Space>
          <Button 
            type="primary" 
            icon={<FileTextOutlined />} 
            onClick={openLogsInNewTab}
          >
            {t('podDetail.tabs.logs')}
          </Button>
          <Button 
            type="primary" 
            icon={<CodeOutlined />} 
            onClick={openTerminalInNewTab}
          >
            {t('podDetail.tabs.terminal')}
          </Button>
        </Space>
      </div>

      <Tabs defaultActiveKey="info" items={tabItems} />
    </div>
  );
};

export default PodDetailPage;