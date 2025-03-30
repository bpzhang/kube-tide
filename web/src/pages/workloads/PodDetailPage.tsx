import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Button, Spin, message, Tabs, Select } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { getPodDetails } from '../../api/pod';
import PodDetail from '../../components/k8s/pod/PodDetail';
import PodTerminal from '../../components/k8s/pod/PodTerminal';
import PodLogs from '../../components/k8s/pod/PodLogs';

const { Option } = Select;

const PodDetailPage: React.FC = () => {
  const { clusterName, namespace, podName } = useParams<{
    clusterName: string;
    namespace: string;
    podName: string;
  }>();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [pod, setPod] = useState<any>(null);
  const [selectedContainer, setSelectedContainer] = useState<string>('');

  useEffect(() => {
    const fetchPodDetails = async () => {
      if (!clusterName || !namespace || !podName) {
        message.error('参数不完整');
        navigate('/workloads/pods');
        return;
      }

      try {
        const response = await getPodDetails(clusterName, namespace, podName);
        if (response.data.code === 0) {
          setPod(response.data.data.pod);
          // 默认选择第一个容器
          if (response.data.data.pod?.spec?.containers?.length > 0) {
            setSelectedContainer(response.data.data.pod.spec.containers[0].name);
          }
        } else {
          message.error(response.data.message || '获取Pod详情失败');
        }
      } catch (error) {
        console.error('获取Pod详情失败:', error);
        message.error('获取Pod详情失败');
      } finally {
        setLoading(false);
      }
    };

    fetchPodDetails();
  }, [clusterName, namespace, podName, navigate]);

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
        <h2>未找到Pod信息</h2>
        <Button type="primary" onClick={() => navigate('/workloads/pods')}>
          返回Pod列表
        </Button>
      </div>
    );
  }

  // 提取Pod的所有容器名称
  const containerNames = pod.spec.containers.map((container: any) => container.name);

  const tabItems = [
    {
      key: 'info',
      label: '基本信息',
      children: <PodDetail pod={pod} clusterName={clusterName!} />
    },
    {
      key: 'logs',
      label: '日志',
      children: (
        <PodLogs
          clusterName={clusterName!}
          namespace={namespace!}
          podName={podName!}
          containers={containerNames}
        />
      )
    },
    {
      key: 'terminal',
      label: '终端',
      children: (
        <Card>
          <div style={{ marginBottom: 16 }}>
            <span style={{ marginRight: 8 }}>容器:</span>
            <Select
              value={selectedContainer}
              onChange={setSelectedContainer}
              style={{ width: 200 }}
            >
              {pod.spec.containers.map((container: any) => (
                <Option key={container.name} value={container.name}>
                  {container.name}
                </Option>
              ))}
            </Select>
          </div>
          {selectedContainer && (
            <PodTerminal
              clusterName={clusterName!}
              namespace={namespace!}
              podName={podName!}
              containerName={selectedContainer}
            />
          )}
        </Card>
      )
    }
  ];

  return (
    <div style={{ padding: '24px' }}>
      <Button
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate('/workloads/pods')}
        style={{ marginBottom: '16px' }}
      >
        返回Pod列表
      </Button>

      <Tabs defaultActiveKey="info" items={tabItems} />
    </div>
  );
};

export default PodDetailPage;