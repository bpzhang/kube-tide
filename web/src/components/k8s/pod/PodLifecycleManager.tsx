import React, { useState, useEffect } from 'react';
import {
  Card,
  Button,
  Space,
  Modal,
  Form,
  InputNumber,
  Switch,
  message,
  Spin,
  Alert,
  Timeline,
  Tag,
  Row,
  Col,
} from 'antd';
import {
  PlayCircleOutlined,
  PauseCircleOutlined,
  StopOutlined,
  ReloadOutlined,
  HistoryOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import {
  managePodLifecycle,
  getPodLifecycleHistory,
  getPodLifecycleStatus,
  PodLifecycleRequest,
  PodLifecycleStatus,
  PodLifecycleEvent,
  ContainerLifecycleStatus,
} from '@/api/pod';

interface PodLifecycleManagerProps {
  clusterName: string;
  namespace: string;
  podName: string;
  onStatusChange?: (status: PodLifecycleStatus) => void;
}

const PodLifecycleManager: React.FC<PodLifecycleManagerProps> = ({
  clusterName,
  namespace,
  podName,
  onStatusChange,
}) => {
  const { t } = useTranslation();
  const [form] = Form.useForm();
  
  const [loading, setLoading] = useState(false);
  const [status, setStatus] = useState<PodLifecycleStatus | null>(null);
  const [history, setHistory] = useState<PodLifecycleEvent[]>([]);
  const [modalVisible, setModalVisible] = useState(false);
  const [historyModalVisible, setHistoryModalVisible] = useState(false);
  const [selectedAction, setSelectedAction] = useState<string>('');
  const [operationLoading, setOperationLoading] = useState(false);

  // 获取Pod生命周期状态
  const fetchStatus = async () => {
    try {
      setLoading(true);
      const response = await getPodLifecycleStatus(clusterName, namespace, podName);
      if (response.data.code === 0) {
        setStatus(response.data.data.status);
        onStatusChange?.(response.data.data.status);
      }
    } catch (error) {
      console.error('获取Pod生命周期状态失败:', error);
      message.error(t('podLifecycle.fetchStatusFailed'));
    } finally {
      setLoading(false);
    }
  };

  // 获取Pod生命周期历史
  const fetchHistory = async () => {
    try {
      const response = await getPodLifecycleHistory(clusterName, namespace, podName);
      if (response.data.code === 0) {
        setHistory(response.data.data.history);
      }
    } catch (error) {
      console.error('获取Pod生命周期历史失败:', error);
      message.error(t('podLifecycle.fetchHistoryFailed'));
    }
  };

  useEffect(() => {
    fetchStatus();
  }, [clusterName, namespace, podName]);

  // 执行生命周期操作
  const handleLifecycleAction = async (values: any) => {
    try {
      setOperationLoading(true);
      
      const request: PodLifecycleRequest = {
        action: selectedAction as any,
        gracePeriod: values.gracePeriod,
        force: values.force || false,
        waitTimeout: (values.waitTimeout || 300) * 1000, // 转换为毫秒
      };

      const response = await managePodLifecycle(clusterName, namespace, podName, request);
      
      if (response.data) {
        const result = response.data;
        if (result.data.success) {
          message.success(result.data.message);
          setStatus(result.data.podStatus);
          onStatusChange?.(result.data.podStatus);
        } else {
          message.error(result.data.message);
        }
      } else {
        message.error(t('podLifecycle.operationFailed'));
      }
    } catch (error) {
      console.error('Pod生命周期操作失败:', error);
      message.error(t('podLifecycle.operationFailed'));
    } finally {
      setOperationLoading(false);
      setModalVisible(false);
      form.resetFields();
    }
  };

  // 打开操作模态框
  const openActionModal = (action: string) => {
    setSelectedAction(action);
    setModalVisible(true);
    
    // 设置默认值
    form.setFieldsValue({
      gracePeriod: 30,
      force: false,
      waitTimeout: 300,
    });
  };

  // 打开历史模态框
  const openHistoryModal = () => {
    setHistoryModalVisible(true);
    fetchHistory();
  };

  // 获取状态颜色
  const getStatusColor = (phase: string) => {
    const colors: { [key: string]: string } = {
      Running: 'green',
      Pending: 'gold',
      Failed: 'red',
      Succeeded: 'blue',
      Unknown: 'grey',
    };
    return colors[phase] || 'default';
  };

  // 获取容器状态描述
  const getContainerStateDescription = (container: ContainerLifecycleStatus) => {
    if (container.state.running) {
      return {
        status: 'Running',
        color: 'green',
        detail: `Started at ${new Date(container.state.running.startedAt).toLocaleString()}`,
      };
    }
    if (container.state.waiting) {
      return {
        status: 'Waiting',
        color: 'gold',
        detail: `${container.state.waiting.reason}: ${container.state.waiting.message}`,
      };
    }
    if (container.state.terminated) {
      return {
        status: 'Terminated',
        color: 'red',
        detail: `Exit code: ${container.state.terminated.exitCode}, Reason: ${container.state.terminated.reason}`,
      };
    }
    return {
      status: 'Unknown',
      color: 'grey',
      detail: 'Unknown state',
    };
  };

  // 渲染操作按钮
  const renderActionButtons = () => {
    if (!status) return null;

    const isRunning = status.phase === 'Running';

    return (
      <Space>
        {!isRunning && (
          <Button
            type="primary"
            icon={<PlayCircleOutlined />}
            onClick={() => openActionModal('start')}
            disabled={operationLoading}
          >
            {t('podLifecycle.start')}
          </Button>
        )}
        
        {isRunning && (
          <>
            <Button
              icon={<PauseCircleOutlined />}
              onClick={() => openActionModal('pause')}
              disabled={operationLoading}
            >
              {t('podLifecycle.pause')}
            </Button>
            
            <Button
              icon={<StopOutlined />}
              onClick={() => openActionModal('stop')}
              disabled={operationLoading}
            >
              {t('podLifecycle.stop')}
            </Button>
          </>
        )}
        
        <Button
          icon={<ReloadOutlined />}
          onClick={() => openActionModal('restart')}
          disabled={operationLoading}
        >
          {t('podLifecycle.restart')}
        </Button>
        
        {status.phase === 'Pending' && (
          <Button
            type="link"
            icon={<PlayCircleOutlined />}
            onClick={() => openActionModal('resume')}
            disabled={operationLoading}
          >
            {t('podLifecycle.resume')}
          </Button>
        )}
      </Space>
    );
  };

  if (loading) {
    return (
      <Card title={t('podLifecycle.title')}>
        <div style={{ textAlign: 'center', padding: '20px' }}>
          <Spin size="large" />
        </div>
      </Card>
    );
  }

  return (
    <>
      <Card
        title={t('podLifecycle.title')}
        extra={
          <Space>
            <Button
              icon={<HistoryOutlined />}
              onClick={openHistoryModal}
            >
              {t('podLifecycle.viewHistory')}
            </Button>
            <Button
              icon={<ReloadOutlined />}
              onClick={fetchStatus}
              loading={loading}
            >
              {t('common.refresh')}
            </Button>
          </Space>
        }
      >
        {status && (
          <>
            {/* Pod状态概览 */}
            <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
              <Col xs={24} sm={8}>
                <Card size="small" title={t('podLifecycle.currentStatus')}>
                  <Space direction="vertical" style={{ width: '100%' }}>
                    <div>
                      <Tag color={getStatusColor(status.phase)} style={{ fontSize: '14px' }}>
                        {status.phase}
                      </Tag>
                    </div>
                    <div>
                      <span>{t('podLifecycle.ready')}: </span>
                      <Tag color={status.ready ? 'green' : 'red'}>
                        {status.ready ? t('common.yes') : t('common.no')}
                      </Tag>
                    </div>
                    <div>
                      <span>{t('podLifecycle.totalRestarts')}: </span>
                      <span style={{ fontWeight: 'bold' }}>{status.restartCount}</span>
                    </div>
                  </Space>
                </Card>
              </Col>
              
              <Col xs={24} sm={16}>
                <Card size="small" title={t('podLifecycle.containerStatus')}>
                  <Space direction="vertical" style={{ width: '100%' }}>
                    {status.containerStatuses.map((container, index) => {
                      const stateInfo = getContainerStateDescription(container);
                      return (
                        <div key={index} style={{ marginBottom: '8px' }}>
                          <Row justify="space-between" align="middle">
                            <Col>
                              <Space>
                                <strong>{container.name}</strong>
                                <Tag color={stateInfo.color}>{stateInfo.status}</Tag>
                                <span style={{ fontSize: '12px', color: '#666' }}>
                                  {t('podLifecycle.restarts')}: {container.restartCount}
                                </span>
                              </Space>
                            </Col>
                            <Col>
                              <Tag color={container.ready ? 'green' : 'red'}>
                                {container.ready ? t('podLifecycle.ready') : t('podLifecycle.notReady')}
                              </Tag>
                            </Col>
                          </Row>
                          {stateInfo.detail && (
                            <div style={{ fontSize: '12px', color: '#666', marginTop: '4px' }}>
                              {stateInfo.detail}
                            </div>
                          )}
                        </div>
                      );
                    })}
                  </Space>
                </Card>
              </Col>
            </Row>

            {/* 操作按钮 */}
            <div style={{ marginTop: '16px', textAlign: 'center' }}>
              {renderActionButtons()}
            </div>

            {/* 提示信息 */}
            {status.phase === 'Failed' && (
              <Alert
                style={{ marginTop: '16px' }}
                message={t('podLifecycle.failedPodTip')}
                type="warning"
                showIcon
              />
            )}
          </>
        )}
      </Card>

      {/* 操作确认模态框 */}
      <Modal
        title={t(`podLifecycle.${selectedAction}`)}
        open={modalVisible}
        onCancel={() => {
          setModalVisible(false);
          form.resetFields();
        }}
        footer={null}
        width={500}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleLifecycleAction}
        >
          <Alert
            message={t(`podLifecycle.${selectedAction}Confirm`)}
            type="info"
            showIcon
            style={{ marginBottom: '16px' }}
          />

          {(selectedAction === 'stop' || selectedAction === 'restart') && (
            <Form.Item
              name="gracePeriod"
              label={t('podLifecycle.gracePeriod')}
              help={t('podLifecycle.gracePeriodHelp')}
            >
              <InputNumber
                min={0}
                max={3600}
                addonAfter={t('common.seconds')}
                style={{ width: '100%' }}
              />
            </Form.Item>
          )}

          {(selectedAction === 'stop' || selectedAction === 'restart') && (
            <Form.Item
              name="force"
              valuePropName="checked"
            >
              <Switch checkedChildren={t('podLifecycle.forceDelete')} unCheckedChildren={t('podLifecycle.gracefulDelete')} />
            </Form.Item>
          )}

          <Form.Item
            name="waitTimeout"
            label={t('podLifecycle.waitTimeout')}
            help={t('podLifecycle.waitTimeoutHelp')}
          >
            <InputNumber
              min={30}
              max={1800}
              addonAfter={t('common.seconds')}
              style={{ width: '100%' }}
            />
          </Form.Item>

          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => setModalVisible(false)}>
                {t('common.cancel')}
              </Button>
              <Button
                type="primary"
                htmlType="submit"
                loading={operationLoading}
              >
                {t('common.confirm')}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 历史记录模态框 */}
      <Modal
        title={t('podLifecycle.lifecycleHistory')}
        open={historyModalVisible}
        onCancel={() => setHistoryModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setHistoryModalVisible(false)}>
            {t('common.close')}
          </Button>
        ]}
        width={800}
      >
        {history.length > 0 ? (
          <Timeline>
            {history.map((event, index) => (
              <Timeline.Item
                key={index}
                color={event.type === 'Warning' ? 'red' : 'green'}
              >
                <div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <Tag color={event.type === 'Warning' ? 'red' : 'green'}>
                      {event.reason}
                    </Tag>
                    <span style={{ fontSize: '12px', color: '#666' }}>
                      {new Date(event.timestamp).toLocaleString()}
                    </span>
                  </div>
                  <div style={{ marginTop: '8px' }}>
                    {event.message}
                  </div>
                  <div style={{ fontSize: '12px', color: '#999', marginTop: '4px' }}>
                    {t('podLifecycle.source')}: {event.source}
                  </div>
                </div>
              </Timeline.Item>
            ))}
          </Timeline>
        ) : (
          <div style={{ textAlign: 'center', padding: '40px' }}>
            <InfoCircleOutlined style={{ fontSize: '48px', color: '#ccc' }} />
            <div style={{ marginTop: '16px', color: '#666' }}>
              {t('podLifecycle.noHistoryData')}
            </div>
          </div>
        )}
      </Modal>
    </>
  );
};

export default PodLifecycleManager;
