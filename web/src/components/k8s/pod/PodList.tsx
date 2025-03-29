import React from 'react';
import { Table, Tag, Button, Popconfirm, message, Space } from 'antd';
import { EyeOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { deletePod } from '@/api/pod';

interface PodListProps {
  clusterName: string;
  namespace: string;
  pods: any[];
  onRefresh: () => void;
}

interface ProcessedPod {
  name: string;
  status: string;
  podIP: string;
  nodeName: string;
  createdAt: string;
}

const PodList: React.FC<PodListProps> = ({ clusterName, namespace, pods, onRefresh }) => {
  const navigate = useNavigate();

  const handleDelete = async (podName: string) => {
    try {
      await deletePod(clusterName, namespace, podName);
      message.success('Pod删除成功');
      onRefresh();
    } catch (err) {
      message.error('Pod删除失败');
    }
  };

  const handleViewDetails = (podName: string) => {
    navigate(`/workloads/pods/${clusterName}/${namespace}/${podName}`);
  };

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

  const processedPods: ProcessedPod[] = pods.map(pod => ({
    name: pod.metadata?.name || '',
    status: pod.status?.phase || 'Unknown',
    podIP: pod.status?.podIP || '-',
    nodeName: pod.spec?.nodeName || '-',
    createdAt: pod.metadata?.creationTimestamp || '-',
  }));

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string) => (
        <a onClick={() => handleViewDetails(text)}>{text}</a>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>{status}</Tag>
      ),
    },
    {
      title: 'IP',
      dataIndex: 'podIP',
      key: 'podIP',
    },
    {
      title: '节点',
      dataIndex: 'nodeName',
      key: 'nodeName',
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (time: string) => new Date(time).toLocaleString(),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: ProcessedPod) => (
        <Space>
          <Button
            type="link"
            icon={<EyeOutlined />}
            onClick={() => handleViewDetails(record.name)}
          >
            详情
          </Button>
          <Popconfirm
            title="确定要删除这个Pod吗?"
            onConfirm={() => handleDelete(record.name)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" danger>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Table
      columns={columns}
      dataSource={processedPods}
      rowKey="name"
      pagination={{ pageSize: 10 }}
    />
  );
};

export default PodList;