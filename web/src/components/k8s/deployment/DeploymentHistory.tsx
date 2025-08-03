import React, { useState, useEffect } from 'react';
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

  // 获取版本历史
  const fetchHistory = async () => {
    if (!visible) return;
    
    setLoading(true);
    try {
      const response = await getDeploymentRolloutHistory(clusterName, namespace, deploymentName);
      setRevisions(response.data.revisions || []);
    } catch (error) {
      console.error('Failed to fetch deployment history:', error);
      message.error(t('deployment.history.fetchFailed'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchHistory();
  }, [visible, clusterName, namespace, deploymentName]);

  // 回滚到指定版本
  const handleRollback = async (revision: number) => {
    Modal.confirm({
      title: t('deployment.rollback.confirmTitle'),
      content: t('deployment.rollback.confirmContent', { revision }),
      okText: t('common.confirm'),
      cancelText: t('common.cancel'),
      onOk: async () => {
        setRollbackLoading(true);
        try {
          await rollbackDeployment(clusterName, namespace, deploymentName, revision);
          message.success(t('deployment.rollback.success', { revision }));
          onRollbackSuccess?.();
          onClose();
        } catch (error) {
          console.error('Failed to rollback deployment:', error);
          message.error(t('deployment.rollback.failed'));
        } finally {
          setRollbackLoading(false);
        }
      }
    });
  };

  // 查看版本详情
  const handleViewDetails = async (record: RevisionInfo) => {
    setSelectedRevision(record);
    setDetailsVisible(true);
  };

  // 格式化创建时间
  const formatTime = (time: string) => {
    return new Date(time).toLocaleString();
  };

  // 获取状态标签
  const getStatusTag = (record: RevisionInfo) => {
    const { readyReplicas, availableReplicas, replicas } = record;
    if (readyReplicas === replicas && availableReplicas === replicas) {
      return <Tag color="green">{t('common.ready')}</Tag>;
    } else if (readyReplicas > 0) {
      return <Tag color="orange">{t('common.updating')}</Tag>;
    } else {
      return <Tag color="red">{t('common.notReady')}</Tag>;
    }
  };

  const columns: ColumnsType<RevisionInfo> = [
    {
      title: t('deployment.history.revision'),
      dataIndex: 'revision',
      key: 'revision',
      width: 100,
      render: (revision: number) => (
        <Text strong>#{revision}</Text>
      ),
    },
    {
      title: t('deployment.history.changeReason'),
      dataIndex: 'changeReason',
      key: 'changeReason',
      ellipsis: true,
      render: (reason: string) => reason || t('common.noData'),
    },
    {
      title: t('deployment.history.replicaSet'),
      dataIndex: 'replicaSetName',
      key: 'replicaSetName',
      ellipsis: true,
    },
    {
      title: t('deployment.history.replicas'),
      key: 'replicas',
      width: 120,
      render: (_, record) => (
        <Space direction="vertical" size={0}>
          <Text>{record.readyReplicas}/{record.replicas || 0}</Text>
          {getStatusTag(record)}
        </Space>
      ),
    },
    {
      title: t('deployment.history.creationTime'),
      dataIndex: 'creationTime',
      key: 'creationTime',
      width: 180,
      render: (time: string) => formatTime(time),
    },
    {
      title: t('common.actions'),
      key: 'actions',
      width: 150,
      render: (_, record) => (
        <Space>
          <Tooltip title={t('deployment.history.viewDetails')}>
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => handleViewDetails(record)}
            />
          </Tooltip>
          <Tooltip title={t('deployment.rollback.title')}>
            <Button
              type="text"
              icon={<RollbackOutlined />}
              loading={rollbackLoading}
              onClick={() => handleRollback(record.revision)}
            />
          </Tooltip>
        </Space>
      ),
    },
  ];

  return (
    <>
      <Modal
        title={
          <Space>
            <HistoryOutlined />
            {t('deployment.history.title', { name: deploymentName })}
          </Space>
        }
        open={visible}
        onCancel={onClose}
        width={1000}
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
          pagination={{ pageSize: 10 }}
          size="small"
        />
      </Modal>

      {/* 版本详情抽屉 */}
      <Drawer
        title={
          <Space>
            <InfoCircleOutlined />
            {t('deployment.history.revisionDetails', { revision: selectedRevision?.revision })}
          </Space>
        }
        open={detailsVisible}
        onClose={() => setDetailsVisible(false)}
        width={600}
      >
        {selectedRevision && (
          <Space direction="vertical" size="middle" style={{ width: '100%' }}>
            <Card size="small" title={t('deployment.history.basicInfo')}>
              <Descriptions size="small" column={1}>
                <Descriptions.Item label={t('deployment.history.revision')}>
                  #{selectedRevision.revision}
                </Descriptions.Item>
                <Descriptions.Item label={t('deployment.history.replicaSet')}>
                  {selectedRevision.replicaSetName}
                </Descriptions.Item>
                <Descriptions.Item label={t('deployment.history.changeReason')}>
                  {selectedRevision.changeReason || t('common.noData')}
                </Descriptions.Item>
                <Descriptions.Item label={t('deployment.history.creationTime')}>
                  {formatTime(selectedRevision.creationTime)}
                </Descriptions.Item>
                <Descriptions.Item label={t('deployment.history.replicas')}>
                  {selectedRevision.readyReplicas}/{selectedRevision.replicas || 0}
                </Descriptions.Item>
              </Descriptions>
            </Card>

            {/* 容器信息 */}
            <Card size="small" title={t('deployment.containers')}>
              {selectedRevision.podTemplateSpec.spec?.containers?.map((container: any, index: number) => (
                <Card key={index} size="small" type="inner" title={container.name}>
                  <Descriptions size="small" column={1}>
                    <Descriptions.Item label={t('deployment.container.image')}>
                      {container.image}
                    </Descriptions.Item>
                    {container.command && (
                      <Descriptions.Item label={t('deployment.container.command')}>
                        <Paragraph code copyable={{ text: container.command.join(' ') }}>
                          {container.command.join(' ')}
                        </Paragraph>
                      </Descriptions.Item>
                    )}
                    {container.args && (
                      <Descriptions.Item label={t('deployment.container.args')}>
                        <Paragraph code copyable={{ text: container.args.join(' ') }}>
                          {container.args.join(' ')}
                        </Paragraph>
                      </Descriptions.Item>
                    )}
                  </Descriptions>
                </Card>
              ))}
            </Card>

            {/* 标签信息 */}
            {selectedRevision.labels && Object.keys(selectedRevision.labels).length > 0 && (
              <Card size="small" title={t('deployment.labels')}>
                <Space wrap>
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
