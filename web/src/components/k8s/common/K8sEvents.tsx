import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, Button, Spin, Empty, message } from 'antd';
import { SyncOutlined } from '@ant-design/icons';
import { formatDate } from '@/utils/format';
import { useTranslation } from 'react-i18next';


interface K8sEventsProps {
  clusterName: string;
  namespace: string;
  resourceName: string;
  resourceKind: 'Pod' | 'Deployment' | 'Cluster';
  fetchEvents: (clusterName: string, namespace: string, resourceName: string) => Promise<any>;
}

/**
 * K8sEvents Component 
 */
const K8sEvents: React.FC<K8sEventsProps> = ({ 
  clusterName, 
  namespace, 
  resourceName,
  resourceKind,
  fetchEvents 
}) => {
  const { t } = useTranslation();
  const [events, setEvents] = useState<any[]>([]);
  const [loading, setLoading] = useState<boolean>(true);

  // load events from API
  const loadEvents = async () => {
    setLoading(true);
    try {
      const response = await fetchEvents(clusterName, namespace, resourceName);
      if (response.data.code === 0) {
        setEvents(response.data.data.events || []);
      } else {
        message.error(response.data.message || t('events.fetchFailed', { resourceKind }));
      }
    } catch (error) {
      console.error(t('events.fetchErrorLog', { resourceKind }), error);
      message.error(t('events.fetchErrorRetry', { resourceKind }));
    } finally {
      setLoading(false);
    }
  };

  // load events when component mounts
  useEffect(() => {
    if (clusterName && namespace && resourceName) {
      loadEvents();
    }
  }, [clusterName, namespace, resourceName]);

  // get event type color
  const getEventTypeColor = (type: string) => {
    const typeColors: { [key: string]: string } = {
      Normal: 'green',
      Warning: 'orange',
    };
    return typeColors[type] || 'blue';
  };

  // event table columns definition
  const columns = [
    {
      title: t('events.columns.type'),
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => <Tag color={getEventTypeColor(type)}>{type}</Tag>,
      width: 100,
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
      title: t('events.columns.firstTimestamp'),
      dataIndex: 'firstTimestamp',
      key: 'firstTimestamp',
      render: (time: string) => formatDate(time),
      width: 170,
    },
    {
      title: t('events.columns.lastTimestamp'),
      dataIndex: 'lastTimestamp',
      key: 'lastTimestamp',
      render: (time: string) => formatDate(time),
      width: 170,
    },
    {
      title: t('events.columns.count'),
      dataIndex: 'count',
      key: 'count',
      width: 80,
    },
  ];

  return (
    <Card 
      title={t('events.title', { resourceKind })}
      extra={
        <Button
          icon={<SyncOutlined />}
          onClick={loadEvents}
          loading={loading}
        >
          {t('common.refresh')}
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
        <Empty description={t('events.noEvents', { resourceKind })} />
      )}
    </Card>
  );
};

export default K8sEvents;