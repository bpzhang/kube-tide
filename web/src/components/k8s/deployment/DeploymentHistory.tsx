import React, { useState, useEffect, useMemo } from 'react';
import { Table, Button, Modal, message, Card, Tag, Space, Tooltip, Typography, Descriptions, Drawer } from 'antd';
import { HistoryOutlined, RollbackOutlined, EyeOutlined, InfoCircleOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { getDeploymentRolloutHistory, rollbackDeployment, RevisionInfo } from '../../../api/deployment';
import type { ColumnsType } from 'antd/es/table';

const { Text, Paragraph } = Typography;

interface DeploymentHistoryProps {
  visible: boolean;
  onClose: () => void;
  clusterName: string;
  namespace: string;
  deploymentName: string;
  onRollbackSuccess?: () => void;
}

const getContainerImages = (record: RevisionInfo): string[] => {
  return (record.podTemplateSpec?.spec?.containers || [])
    .map((container: { image?: string }) => container.image)
    .filter(Boolean);
};

const DeploymentHistory: React.FC<DeploymentHistoryProps> = ({
  visible,
  onClose,
  clusterName,
  namespace,
  deploymentName,
  onRollbackSuccess
}) => {
  const { t } = useTranslation();
  const [revisions, setRevisions] = useState<RevisionInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [rollbackLoading, setRollbackLoading] = useState(false);
  const [detailsVisible, setDetailsVisible] = useState(false);
  const [selectedRevision, setSelectedRevision] = useState<RevisionInfo | null>(null);

  const currentRevision = revisions[0]?.revision;

  const fetchHistory = async () => {
    if (!visible) return;

    setLoading(true);
    try {
      const response = await getDeploymentRolloutHistory(clusterName, namespace, deploymentName);
      if (response.data.code === 0) {
        setRevisions(response.data.data?.revisions || []);
      } else {
        message.error(response.data.message || t('deployments.history.fetchFailed'));
        setRevisions([]);
      }
    } catch (error) {
      console.error('Failed to fetch deployment history:', error);
      message.error(t('deployments.history.fetchFailed'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchHistory();
  }, [visible, clusterName, namespace, deploymentName]);

  const handleRollback = async (revision: number) => {
    Modal.confirm({
      title: t('deployments.rollback.confirmTitle'),
      content: t('deployments.rollback.confirmContent', { revision }),
      okText: t('common.confirm'),
      cancelText: t('common.cancel'),
      onOk: async () => {
        setRollbackLoading(true);
        try {
          await rollbackDeployment(clusterName, namespace, deploymentName, revision);
          message.success(t('deployments.rollback.success', { revision }));
          onRollbackSuccess?.();
          onClose();
        } catch (error) {
          console.error('Failed to rollback deployment:', error);
          message.error(t('deployments.rollback.failed'));
        } finally {
          setRollbackLoading(false);
        }
      }
    });
  };

  const handleViewDetails = (record: RevisionInfo) => {
    setSelectedRevision(record);
    setDetailsVisible(true);
  };

  const formatTime = (time: string) => {
    return new Date(time).toLocaleString(undefined, {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getStatusTag = (record: RevisionInfo) => {
    const { readyReplicas, availableReplicas, replicas } = record;
    const target = replicas || 0;

    if (target === 0) {
      return <Tag color="default">{t('deployments.history.noReplicas')}</Tag>;
    }
    if (readyReplicas === target && availableReplicas === target) {
      return <Tag color="success">{t('common.ready')}</Tag>;
    }
    if (readyReplicas > 0) {
      return <Tag color="warning">{t('common.updating')}</Tag>;
    }
    return <Tag color="error">{t('common.notReady')}</Tag>;
  };

  const hasChangeReason = useMemo(
    () => revisions.some((item) => Boolean(item.changeReason)),
    [revisions]
  );

  const columns: ColumnsType<RevisionInfo> = useMemo(() => {
    const baseColumns: ColumnsType<RevisionInfo> = [
      {
        title: t('deployments.history.revision'),
        dataIndex: 'revision',
        key: 'revision',
        width: 96,
        fixed: 'left',
        render: (revision: number) => (
          <Space size={4}>
            <Text strong>#{revision}</Text>
            {revision === currentRevision && (
              <Tag color="blue" style={{ margin: 0 }}>
                {t('deployments.history.current')}
              </Tag>
            )}
          </Space>
        ),
      },
      {
        title: t('deployments.image'),
        key: 'image',
        ellipsis: { showTitle: false },
        render: (_, record) => {
          const images = getContainerImages(record);
          if (images.length === 0) {
            return <Text type="secondary">—</Text>;
          }
          return (
            <Tooltip title={images.join('\n')}>
              <Text ellipsis style={{ maxWidth: 220 }}>
                {images.join(', ')}
              </Text>
            </Tooltip>
          );
        },
      },
    ];

    if (hasChangeReason) {
      baseColumns.push({
        title: t('deployments.history.changeReason'),
        dataIndex: 'changeReason',
        key: 'changeReason',
        width: 160,
        ellipsis: { showTitle: false },
        render: (reason: string) => (
          <Tooltip title={reason}>
            <Text ellipsis style={{ maxWidth: 140 }}>{reason}</Text>
          </Tooltip>
        ),
      });
    }

    baseColumns.push(
      {
        title: t('deployments.history.replicaSet'),
        dataIndex: 'replicaSetName',
        key: 'replicaSetName',
        width: 180,
        ellipsis: { showTitle: false },
        render: (name: string) => (
          <Tooltip title={name}>
            <Text code ellipsis style={{ maxWidth: 160, fontSize: 12 }}>
              {name}
            </Text>
          </Tooltip>
        ),
      },
      {
        title: t('deployments.history.replicas'),
        key: 'replicas',
        width: 112,
        render: (_, record) => (
          <Space size={6}>
            <Text>{record.readyReplicas}/{record.replicas || 0}</Text>
            {getStatusTag(record)}
          </Space>
        ),
      },
      {
        title: t('deployments.history.creationTime'),
        dataIndex: 'creationTime',
        key: 'creationTime',
        width: 148,
        render: (time: string) => (
          <Text style={{ fontSize: 12 }}>{formatTime(time)}</Text>
        ),
      },
      {
        title: t('common.operations'),
        key: 'actions',
        width: 88,
        fixed: 'right',
        align: 'center',
        render: (_, record) => (
          <Space size={0}>
            <Tooltip title={t('deployments.history.viewDetails')}>
              <Button
                type="text"
                size="small"
                icon={<EyeOutlined />}
                onClick={() => handleViewDetails(record)}
              />
            </Tooltip>
            <Tooltip title={t('deployments.rollback.title')}>
              <Button
                type="text"
                size="small"
                icon={<RollbackOutlined />}
                loading={rollbackLoading}
                disabled={record.revision === currentRevision}
                onClick={() => handleRollback(record.revision)}
              />
            </Tooltip>
          </Space>
        ),
      }
    );

    return baseColumns;
  }, [currentRevision, hasChangeReason, rollbackLoading, t]);

  return (
    <>
      <Modal
        title={
          <Space>
            <HistoryOutlined />
            {t('deployments.history.title', { name: deploymentName })}
          </Space>
        }
        open={visible}
        onCancel={onClose}
        width={960}
        styles={{ body: { paddingTop: 12 } }}
        footer={[
          <Button key="close" onClick={onClose}>
            {t('common.close')}
          </Button>,
        ]}
      >
        <Table
          columns={columns}
          dataSource={revisions}
          loading={loading}
          rowKey="revision"
          size="small"
          tableLayout="fixed"
          scroll={{ x: 860 }}
          pagination={{
            pageSize: 8,
            showSizeChanger: false,
            showTotal: (total) => t('deployments.history.total', { total }),
          }}
        />
      </Modal>

      <Drawer
        title={
          <Space>
            <InfoCircleOutlined />
            {t('deployments.history.revisionDetails', { revision: selectedRevision?.revision })}
          </Space>
        }
        open={detailsVisible}
        onClose={() => setDetailsVisible(false)}
        width={640}
      >
        {selectedRevision && (
          <Space direction="vertical" size="middle" style={{ width: '100%' }}>
            <Card size="small" title={t('deployments.history.basicInfo')}>
              <Descriptions size="small" column={1} bordered>
                <Descriptions.Item label={t('deployments.history.revision')}>
                  #{selectedRevision.revision}
                </Descriptions.Item>
                <Descriptions.Item label={t('deployments.history.replicaSet')}>
                  <Text code copyable>{selectedRevision.replicaSetName}</Text>
                </Descriptions.Item>
                <Descriptions.Item label={t('deployments.history.changeReason')}>
                  {selectedRevision.changeReason || '—'}
                </Descriptions.Item>
                <Descriptions.Item label={t('deployments.history.creationTime')}>
                  {formatTime(selectedRevision.creationTime)}
                </Descriptions.Item>
                <Descriptions.Item label={t('deployments.history.replicas')}>
                  <Space size={8}>
                    <Text>{selectedRevision.readyReplicas}/{selectedRevision.replicas || 0}</Text>
                    {getStatusTag(selectedRevision)}
                  </Space>
                </Descriptions.Item>
              </Descriptions>
            </Card>

            <Card size="small" title={t('deployments.containers')}>
              {selectedRevision.podTemplateSpec.spec?.containers?.map((container: any, index: number) => (
                <Card key={index} size="small" type="inner" title={container.name}>
                  <Descriptions size="small" column={1}>
                    <Descriptions.Item label={t('deployments.container.image')}>
                      <Text copyable>{container.image}</Text>
                    </Descriptions.Item>
                    {container.command && (
                      <Descriptions.Item label={t('deployments.container.command')}>
                        <Paragraph code copyable={{ text: container.command.join(' ') }} style={{ marginBottom: 0 }}>
                          {container.command.join(' ')}
                        </Paragraph>
                      </Descriptions.Item>
                    )}
                    {container.args && (
                      <Descriptions.Item label={t('deployments.container.args')}>
                        <Paragraph code copyable={{ text: container.args.join(' ') }} style={{ marginBottom: 0 }}>
                          {container.args.join(' ')}
                        </Paragraph>
                      </Descriptions.Item>
                    )}
                  </Descriptions>
                </Card>
              ))}
            </Card>

            {selectedRevision.labels && Object.keys(selectedRevision.labels).length > 0 && (
              <Card size="small" title={t('deployments.labels')}>
                <Space wrap size={[4, 4]}>
                  {Object.entries(selectedRevision.labels).map(([key, value]) => (
                    <Tag key={key}>{key}={value}</Tag>
                  ))}
                </Space>
              </Card>
            )}
          </Space>
        )}
      </Drawer>
    </>
  );
};

export default DeploymentHistory;
