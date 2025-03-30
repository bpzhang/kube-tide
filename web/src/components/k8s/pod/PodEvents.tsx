import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, Button, Spin, Empty, message } from 'antd';
import { SyncOutlined } from '@ant-design/icons';
import { getPodEvents } from '@/api/pod';
import { formatDate } from '@/utils/format';

interface PodEventsProps {
  clusterName: string;
  namespace: string;
  podName: string;
}

/**
 * Pod事件组件，显示与Pod相关的Kubernetes事件
 */
const PodEvents: React.FC<PodEventsProps> = ({ clusterName, namespace, podName }) => {
  const [events, setEvents] = useState<any[]>([]);
  const [loading, setLoading] = useState<boolean>(true);

  // 获取事件的函数
  const fetchEvents = async () => {
    setLoading(true);
    try {
      const response = await getPodEvents(clusterName, namespace, podName);
      if (response.data.code === 0) {
        setEvents(response.data.data.events || []);
      } else {
        message.error(response.data.message || '获取Pod事件失败');
      }
    } catch (error) {
      console.error('获取Pod事件失败:', error);
      message.error('获取Pod事件失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  };

  // 在组件挂载时获取事件
  useEffect(() => {
    if (clusterName && namespace && podName) {
      fetchEvents();
    }
  }, [clusterName, namespace, podName]);

  // 获取事件类型对应的标签颜色
  const getEventTypeColor = (type: string) => {
    const typeColors: { [key: string]: string } = {
      Normal: 'green',
      Warning: 'orange',
    };
    return typeColors[type] || 'blue';
  };

  // 事件表格列定义
  const columns = [
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => <Tag color={getEventTypeColor(type)}>{type}</Tag>,
    },
    {
      title: '原因',
      dataIndex: 'reason',
      key: 'reason',
    },
    {
      title: '消息',
      dataIndex: 'message',
      key: 'message',
      width: '40%',
      ellipsis: true,
    },
    {
      title: '组件',
      dataIndex: 'source',
      key: 'source',
      render: (source: any) => source?.component || '-',
    },
    {
      title: '首次发生',
      dataIndex: 'firstTimestamp',
      key: 'firstTimestamp',
      render: (time: string) => formatDate(time),
    },
    {
      title: '最后发生',
      dataIndex: 'lastTimestamp',
      key: 'lastTimestamp',
      render: (time: string) => formatDate(time),
    },
    {
      title: '次数',
      dataIndex: 'count',
      key: 'count',
    },
  ];

  return (
    <Card 
      title="事件" 
      extra={
        <Button
          icon={<SyncOutlined />}
          onClick={fetchEvents}
          loading={loading}
        >
          刷新
        </Button>
      }
    >
      {loading ? (
        <div style={{ textAlign: 'center', padding: '20px' }}>
          <Spin />
        </div>
      ) : events.length > 0 ? (
        <Table
          columns={columns}
          dataSource={events}
          rowKey={(record) => `${record.metadata?.uid || ''}-${record.firstTimestamp || ''}`}
          pagination={false}
          size="middle"
        />
      ) : (
        <Empty description="没有找到相关事件" />
      )}
    </Card>
  );
};

export default PodEvents;