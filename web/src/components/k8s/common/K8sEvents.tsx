import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, Button, Spin, Empty, message } from 'antd';
import { SyncOutlined } from '@ant-design/icons';
import { formatDate } from '@/utils/format';

interface K8sEventsProps {
  clusterName: string;
  namespace: string;
  resourceName: string;
  resourceKind: 'Pod' | 'Deployment';
  fetchEvents: (clusterName: string, namespace: string, resourceName: string) => Promise<any>;
}

/**
 * K8s事件组件，显示与Kubernetes资源相关的事件
 */
const K8sEvents: React.FC<K8sEventsProps> = ({ 
  clusterName, 
  namespace, 
  resourceName,
  resourceKind,
  fetchEvents 
}) => {
  const [events, setEvents] = useState<any[]>([]);
  const [loading, setLoading] = useState<boolean>(true);

  // 获取事件的函数
  const loadEvents = async () => {
    setLoading(true);
    try {
      const response = await fetchEvents(clusterName, namespace, resourceName);
      if (response.data.code === 0) {
        setEvents(response.data.data.events || []);
      } else {
        message.error(response.data.message || `获取${resourceKind}事件失败`);
      }
    } catch (error) {
      console.error(`获取${resourceKind}事件失败:`, error);
      message.error(`获取${resourceKind}事件失败，请稍后重试`);
    } finally {
      setLoading(false);
    }
  };

  // 在组件挂载时获取事件
  useEffect(() => {
    if (clusterName && namespace && resourceName) {
      loadEvents();
    }
  }, [clusterName, namespace, resourceName]);

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
      width: 100,
    },
    {
      title: '原因',
      dataIndex: 'reason',
      key: 'reason',
      width: 150,
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
      width: 120,
    },
    {
      title: '首次发生',
      dataIndex: 'firstTimestamp',
      key: 'firstTimestamp',
      render: (time: string) => formatDate(time),
      width: 170,
    },
    {
      title: '最后发生',
      dataIndex: 'lastTimestamp',
      key: 'lastTimestamp',
      render: (time: string) => formatDate(time),
      width: 170,
    },
    {
      title: '次数',
      dataIndex: 'count',
      key: 'count',
      width: 80,
    },
  ];

  return (
    <Card 
      title="事件" 
      extra={
        <Button
          icon={<SyncOutlined />}
          onClick={loadEvents}
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
          scroll={{ x: 'max-content' }}
        />
      ) : (
        <Empty description={`没有找到相关${resourceKind}事件`} />
      )}
    </Card>
  );
};

export default K8sEvents;