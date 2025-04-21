import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, Button, Spin, Empty, message, Tabs } from 'antd';
import { SyncOutlined } from '@ant-design/icons';
import { formatDate } from '@/utils/format';
import { getAllDeploymentEvents } from '@/api/deployment';
import { useTranslation } from 'react-i18next';

interface DeploymentEventsProps {
  clusterName: string;
  namespace: string;
  deploymentName: string;
}

/**
 * Deployment comprehensive event component, 
 * displaying all kubernetes events related to deployment and its associated replica sets and pods
 */
const DeploymentEvents: React.FC<DeploymentEventsProps> = ({ 
  clusterName, 
  namespace, 
  deploymentName
}) => {
  const { t } = useTranslation();
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

  // Function to fetch events
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
        message.error(response.data.message || 'Failed to fetch events');
      }
    } catch (error) {
      console.error('Failed to fetch events:', error);
      message.error('Failed to fetch events, please try again later');
    } finally {
      setLoading(false);
    }
  };

  // Function to fetch events
  useEffect(() => {
    if (clusterName && namespace && deploymentName) {
      fetchEvents();
    }
  }, [clusterName, namespace, deploymentName]);

  // Function to get the color of event type
  const getEventTypeColor = (type: string) => {
    const typeColors: { [key: string]: string } = {
      Normal: 'green',
      Warning: 'orange',
    };
    return typeColors[type] || 'blue';
  };

  // Event table column definitions
  const columns = [
    {
      title: t('events.columns.type'),
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => <Tag color={getEventTypeColor(type)}>{type}</Tag>,
      width: 100,
    },
    {
      title: t('events.columns.involved'),
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
      title: t('events.columns.reason'),
      dataIndex: 'reason',
      key: 'reason',
      width: 150,
    },
    {
      title: t('events.columns.message'),
      dataIndex: 'message',
      key: 'message',
      width: '40%',
      ellipsis: true,
    },
    {
      title: t('events.columns.source'),
      dataIndex: 'source',
      key: 'source',
      render: (source: any) => source?.component || '-',
      width: 120,
    },
    {
      title: t('events.columns.lastTimestamp'),
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
      title: t('events.columns.count'),
      dataIndex: 'count',
      key: 'count',
      width: 70,
    },
  ];

  // get all events from different sources
  const getAllEvents = () => {
    return [
      ...(events.deployment || []).map(evt => ({ ...evt, eventSource: 'deployment' })),
      ...(events.replicaSet || []).map(evt => ({ ...evt, eventSource: 'replicaSet' })),
      ...(events.pod || []).map(evt => ({ ...evt, eventSource: 'pod' }))
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

  // Get total event count
  const getTotalEventCount = () => {
    return (events.deployment?.length || 0) + (events.replicaSet?.length || 0) + (events.pod?.length || 0);
  };

  // Tab item configuration
  const tabItems = [
    {
      key: 'all',
      label: `${t('events.tabs.all')} (${getTotalEventCount()})`,
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
      label: `${t('events.tabs.deployment')} (${events.deployment?.length || 0})`,
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
      label: `${t('events.tabs.replicaSet')} (${events.replicaSet?.length || 0})`,
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
      label: `${t('events.tabs.pod')} (${events.pod?.length || 0})`,
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
      title={t('common.events')} 
      extra={
        <Button
          icon={<SyncOutlined />}
          onClick={fetchEvents}
          loading={loading}
        >
          {t('common.refresh')}
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

// events table component
interface EventsTableProps {
  events: any[];
  columns: any[];
  loading: boolean;
}

const EventsTable: React.FC<EventsTableProps> = ({ events, columns, loading }) => {
  const { t } = useTranslation();
  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '20px' }}>
        <Spin />
      </div>
    );
  }
  
  if (events && events.length === 0) {
    return <Empty description={t('common.noDataFound')} />;
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