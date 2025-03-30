import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, Button, Spin, Empty, message, Tabs } from 'antd';
import { SyncOutlined } from '@ant-design/icons';
import { formatDate } from '@/utils/format';
import { getAllDeploymentEvents } from '@/api/deployment';

interface DeploymentEventsProps {
  clusterName: string;
  namespace: string;
  deploymentName: string;
}

/**
 * Deployment综合事件组件，显示与Deployment及其关联的ReplicaSet和Pod相关的所有Kubernetes事件
 */
const DeploymentEvents: React.FC<DeploymentEventsProps> = ({ 
  clusterName, 
  namespace, 
  deploymentName
}) => {
  const [events, setEvents] = useState<{
    deployment: any[],
    replicaSet: any[],
    pod: any[]
  }>({
    deployment: [],
    replicaSet: [],
    pod: []
  });
  const [loading, setLoading] = useState<boolean>(true);
  const [activeKey, setActiveKey] = useState<string>('all');

  // 获取事件的函数
  const fetchEvents = async () => {
    setLoading(true);
    try {
      const response = await getAllDeploymentEvents(clusterName, namespace, deploymentName);
      if (response.data.code === 0) {
        setEvents(response.data.data.events || {
          deployment: [],
          replicaSet: [],
          pod: []
        });
      } else {
        message.error(response.data.message || '获取事件失败');
      }
    } catch (error) {
      console.error('获取事件失败:', error);
      message.error('获取事件失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  };

  // 在组件挂载时获取事件
  useEffect(() => {
    if (clusterName && namespace && deploymentName) {
      fetchEvents();
    }
  }, [clusterName, namespace, deploymentName]);

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
      title: '对象',
      key: 'involved',
      render: (text: any, record: any) => {
        const kind = record.involvedObject?.kind || '';
        const name = record.involvedObject?.name || '';
        let color = 'blue';
        if (kind === 'Deployment') color = 'purple';
        else if (kind === 'ReplicaSet') color = 'blue';
        else if (kind === 'Pod') color = 'cyan';
        
        return <Tag color={color}>{`${kind}/${name}`}</Tag>;
      },
      width: 180,
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
      title: '最后发生',
      dataIndex: 'lastTimestamp',
      key: 'lastTimestamp',
      render: (time: string) => formatDate(time),
      width: 170,
      sorter: (a: any, b: any) => {
        if (!a.lastTimestamp || !b.lastTimestamp) return 0;
        return new Date(b.lastTimestamp).getTime() - new Date(a.lastTimestamp).getTime();
      },
      defaultSortOrder: 'descend',
    },
    {
      title: '次数',
      dataIndex: 'count',
      key: 'count',
      width: 70,
    },
  ];

  // 获取所有事件合并后的数组
  const getAllEvents = () => {
    return [
      ...events.deployment.map(evt => ({ ...evt, eventSource: 'deployment' })),
      ...events.replicaSet.map(evt => ({ ...evt, eventSource: 'replicaSet' })),
      ...events.pod.map(evt => ({ ...evt, eventSource: 'pod' }))
    ].sort((a, b) => {
      if (!a.lastTimestamp || !b.lastTimestamp) return 0;
      return new Date(b.lastTimestamp).getTime() - new Date(a.lastTimestamp).getTime();
    });
  };

  // 根据当前选择的标签页返回对应的事件数据
  const getEventsForActiveTab = () => {
    switch (activeKey) {
      case 'deployment':
        return events.deployment;
      case 'replicaSet':
        return events.replicaSet;
      case 'pod':
        return events.pod;
      case 'all':
      default:
        return getAllEvents();
    }
  };

  // 获取总事件数
  const getTotalEventCount = () => {
    return events.deployment.length + events.replicaSet.length + events.pod.length;
  };

  // Tab项配置
  const tabItems = [
    {
      key: 'all',
      label: `全部 (${getTotalEventCount()})`,
      children: (
        <EventsTable 
          events={getAllEvents()} 
          columns={columns} 
          loading={loading} 
        />
      ),
    },
    {
      key: 'deployment',
      label: `Deployment (${events.deployment.length})`,
      children: (
        <EventsTable 
          events={events.deployment} 
          columns={columns} 
          loading={loading} 
        />
      ),
    },
    {
      key: 'replicaSet',
      label: `ReplicaSet (${events.replicaSet.length})`,
      children: (
        <EventsTable 
          events={events.replicaSet} 
          columns={columns} 
          loading={loading} 
        />
      ),
    },
    {
      key: 'pod',
      label: `Pod (${events.pod.length})`,
      children: (
        <EventsTable 
          events={events.pod} 
          columns={columns} 
          loading={loading} 
        />
      ),
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
      <Tabs 
        activeKey={activeKey} 
        onChange={setActiveKey} 
        items={tabItems}
      />
    </Card>
  );
};

// 事件表格子组件
interface EventsTableProps {
  events: any[];
  columns: any[];
  loading: boolean;
}

const EventsTable: React.FC<EventsTableProps> = ({ events, columns, loading }) => {
  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '20px' }}>
        <Spin />
      </div>
    );
  }
  
  if (events.length === 0) {
    return <Empty description="没有找到相关事件" />;
  }
  
  return (
    <Table
      columns={columns}
      dataSource={events}
      rowKey={(record) => `${record.metadata?.uid || ''}-${record.firstTimestamp || ''}-${record.eventSource || ''}`}
      pagination={{ pageSize: 10 }}
      size="middle"
      scroll={{ x: 'max-content' }}
    />
  );
};

export default DeploymentEvents;