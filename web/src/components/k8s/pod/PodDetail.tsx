import React from 'react';
import { Card, Descriptions, Tag, Space, Table } from 'antd';
import { formatDate } from '@/utils/format';
import K8sEvents from '../common/K8sEvents';
import { getPodEvents } from '@/api/pod';
import { useTranslation } from 'react-i18next';

interface ContainerStatus {
  name: string;
  ready: boolean;
  restartCount: number;
  state: any;
  image: string;
}

interface PodDetailProps {
  pod: {
    metadata: {
      name: string;
      namespace: string;
      creationTimestamp: string;
      labels?: { [key: string]: string };
      annotations?: { [key: string]: string };
    };
    spec: {
      nodeName: string;
      containers: Array<{
        name: string;
        image: string;
        ports?: Array<{
          containerPort: number;
          protocol: string;
        }>;
        volumeMounts?: Array<{
          name: string;
          mountPath: string;
          readOnly?: boolean;
        }>;
      }>;
      volumes?: Array<{
        name: string;
        [key: string]: any;
      }>;
    };
    status: {
      phase: string;
      podIP: string;
      hostIP: string;
      conditions?: Array<{
        type: string;
        status: string;
        lastTransitionTime: string;
        reason?: string;
        message?: string;
      }>;
      containerStatuses?: ContainerStatus[];
    };
  };
  clusterName: string; // 添加集群名称参数
}

const PodDetail: React.FC<PodDetailProps> = ({ pod, clusterName }) => {
  const { t } = useTranslation();
  
  const getStatusColor = (status: string) => {
    const colors: { [key: string]: string } = {
      Running: 'green',
      Pending: 'gold',
      Failed: 'red',
      Unknown: 'grey',
      Succeeded: 'blue',
      Terminated: 'red',
    };
    return colors[status] || 'blue';
  };

  const getContainerStateText = (state: any): { text: string; color: string } => {
    if (state.running) {
      return { text: 'Running', color: 'green' };
    } else if (state.waiting) {
      return { text: `Waiting (${state.waiting.reason})`, color: 'gold' };
    } else if (state.terminated) {
      return { text: `Terminated (${state.terminated.reason})`, color: 'red' };
    }
    return { text: 'Unknown', color: 'grey' };
  };

  const containerColumns = [
    {
      title: t('podDetail.containers.name'),
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: t('podDetail.containers.status'),
      key: 'state',
      render: (container: ContainerStatus) => {
        const state = getContainerStateText(container.state);
        return <Tag color={state.color}>{state.text}</Tag>;
      },
    },
    {
      title: t('podDetail.containers.ready'),
      dataIndex: 'ready',
      key: 'ready',
      render: (ready: boolean) => (
        <Tag color={ready ? 'green' : 'red'}>{ready ? t('common.yes') : t('common.no')}</Tag>
      ),
    },
    {
      title: t('podDetail.containers.restarts'),
      dataIndex: 'restartCount',
      key: 'restartCount',
    },
    {
      title: t('podDetail.containers.image'),
      dataIndex: 'image',
      key: 'image',
    },
  ];

  return (
    <Space direction="vertical" size="middle" style={{ width: '100%' }}>
      <Card title={t('podDetail.basicInfo.title')}>
        <Descriptions column={2}>
          <Descriptions.Item label={t('podDetail.basicInfo.name')}>{pod.metadata.name}</Descriptions.Item>
          <Descriptions.Item label={t('podDetail.basicInfo.namespace')}>{pod.metadata.namespace}</Descriptions.Item>
          <Descriptions.Item label={t('podDetail.basicInfo.creationTime')}>
            {formatDate(pod.metadata.creationTimestamp)}
          </Descriptions.Item>
          <Descriptions.Item label={t('podDetail.basicInfo.status')}>
            <Tag color={getStatusColor(pod.status.phase)}>{pod.status.phase}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label={t('podDetail.basicInfo.ip')}>{pod.status.podIP || '-'}</Descriptions.Item>
          <Descriptions.Item label={t('podDetail.basicInfo.hostIP')}>{pod.status.hostIP || '-'}</Descriptions.Item>
          <Descriptions.Item label={t('podDetail.basicInfo.node')}>{pod.spec.nodeName || '-'}</Descriptions.Item>
        </Descriptions>
      </Card>

      <Card title={t('pods.labels')}>
        <Space wrap>
          {Object.entries(pod.metadata.labels || {}).map(([key, value]) => (
            <Tag key={key}>{`${key}: ${value}`}</Tag>
          ))}
        </Space>
      </Card>

      <Card title={t('pods.annotations')}>
        <Space direction="vertical" style={{ width: '100%' }}>
          {Object.entries(pod.metadata.annotations || {}).map(([key, value]) => (
            <div key={key}>
              <strong>{key}:</strong> {value}
            </div>
          ))}
        </Space>
      </Card>

      <Card title={t('podDetail.containers.title')}>
        <Table
          columns={containerColumns}
          dataSource={pod.status.containerStatuses}
          rowKey="name"
          pagination={false}
        />
      </Card>

      <Card title={t('podDetail.volumeMounts.title')}>
        {pod.spec.containers.map(container => (
          <div key={container.name} style={{ marginBottom: 16 }}>
            <h4>{container.name}</h4>
            <Table
              size="small"
              pagination={false}
              dataSource={container.volumeMounts || []}
              columns={[
                {
                  title: t('podDetail.volumeMounts.name'),
                  dataIndex: 'name',
                  key: 'name',
                },
                {
                  title: t('podDetail.volumeMounts.mountPath'),
                  dataIndex: 'mountPath',
                  key: 'mountPath',
                },
                {
                  title: t('podDetail.volumeMounts.readOnly'),
                  dataIndex: 'readOnly',
                  key: 'readOnly',
                  render: (readOnly?: boolean) => (
                    <Tag color={readOnly ? 'orange' : 'green'}>
                      {readOnly ? t('common.yes') : t('common.no')}
                    </Tag>
                  ),
                },
              ]}
              rowKey="name"
            />
          </div>
        ))}
      </Card>

      <Card title={t('podDetail.volumes.title')}>
        <Table
          dataSource={pod.spec.volumes || []}
          columns={[
            {
              title: t('podDetail.volumes.name'),
              dataIndex: 'name',
              key: 'name',
            },
            {
              title: t('podDetail.volumes.type'),
              key: 'type',
              render: (volume: any) => {
                const type = Object.keys(volume).find(key => 
                  key !== 'name' && key !== 'key'
                );
                return <Tag>{type}</Tag>;
              },
            },
          ]}
          rowKey="name"
          pagination={false}
        />
      </Card>

      <K8sEvents 
        clusterName={clusterName} 
        namespace={pod.metadata.namespace} 
        resourceName={pod.metadata.name} 
        resourceKind="Pod"
        fetchEvents={getPodEvents} 
      />
    </Space>
  );
};

export default PodDetail;