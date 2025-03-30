import React from 'react';
import { Card, Descriptions, Tag, Space, Table } from 'antd';
import { formatDate } from '@/utils/format';
import PodEvents from './PodEvents';

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
      title: '名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '状态',
      key: 'state',
      render: (container: ContainerStatus) => {
        const state = getContainerStateText(container.state);
        return <Tag color={state.color}>{state.text}</Tag>;
      },
    },
    {
      title: '就绪',
      dataIndex: 'ready',
      key: 'ready',
      render: (ready: boolean) => (
        <Tag color={ready ? 'green' : 'red'}>{ready ? '是' : '否'}</Tag>
      ),
    },
    {
      title: '重启次数',
      dataIndex: 'restartCount',
      key: 'restartCount',
    },
    {
      title: '镜像',
      dataIndex: 'image',
      key: 'image',
    },
  ];

  return (
    <Space direction="vertical" size="middle" style={{ width: '100%' }}>
      <Card title="基本信息">
        <Descriptions column={2}>
          <Descriptions.Item label="名称">{pod.metadata.name}</Descriptions.Item>
          <Descriptions.Item label="命名空间">{pod.metadata.namespace}</Descriptions.Item>
          <Descriptions.Item label="创建时间">
            {formatDate(pod.metadata.creationTimestamp)}
          </Descriptions.Item>
          <Descriptions.Item label="状态">
            <Tag color={getStatusColor(pod.status.phase)}>{pod.status.phase}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="Pod IP">{pod.status.podIP || '-'}</Descriptions.Item>
          <Descriptions.Item label="主机 IP">{pod.status.hostIP || '-'}</Descriptions.Item>
          <Descriptions.Item label="节点">{pod.spec.nodeName || '-'}</Descriptions.Item>
        </Descriptions>
      </Card>

      <Card title="标签">
        <Space wrap>
          {Object.entries(pod.metadata.labels || {}).map(([key, value]) => (
            <Tag key={key}>{`${key}: ${value}`}</Tag>
          ))}
        </Space>
      </Card>

      <Card title="注解">
        <Space direction="vertical" style={{ width: '100%' }}>
          {Object.entries(pod.metadata.annotations || {}).map(([key, value]) => (
            <div key={key}>
              <strong>{key}:</strong> {value}
            </div>
          ))}
        </Space>
      </Card>

      <Card title="容器状态">
        <Table
          columns={containerColumns}
          dataSource={pod.status.containerStatuses}
          rowKey="name"
          pagination={false}
        />
      </Card>

      <Card title="存储卷挂载">
        {pod.spec.containers.map(container => (
          <div key={container.name} style={{ marginBottom: 16 }}>
            <h4>{container.name}</h4>
            <Table
              size="small"
              pagination={false}
              dataSource={container.volumeMounts || []}
              columns={[
                {
                  title: '名称',
                  dataIndex: 'name',
                  key: 'name',
                },
                {
                  title: '挂载路径',
                  dataIndex: 'mountPath',
                  key: 'mountPath',
                },
                {
                  title: '只读',
                  dataIndex: 'readOnly',
                  key: 'readOnly',
                  render: (readOnly?: boolean) => (
                    <Tag color={readOnly ? 'orange' : 'green'}>
                      {readOnly ? '是' : '否'}
                    </Tag>
                  ),
                },
              ]}
              rowKey="name"
            />
          </div>
        ))}
      </Card>

      <Card title="存储卷">
        <Table
          dataSource={pod.spec.volumes || []}
          columns={[
            {
              title: '名称',
              dataIndex: 'name',
              key: 'name',
            },
            {
              title: '类型',
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

      <PodEvents 
        clusterName={clusterName} 
        namespace={pod.metadata.namespace} 
        podName={pod.metadata.name} 
      />
    </Space>
  );
};

export default PodDetail;