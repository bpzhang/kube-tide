import React, { useState, useEffect } from 'react';
import {
  Card,
  Descriptions,
  Tag,
  Button,
  Space,
  Table,
  Modal,
  Form,
  Input,
  InputNumber,
  Steps,
  message,
  Spin,
} from 'antd';
import { PauseCircleOutlined, PlayCircleOutlined, RocketOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import {
  getRolloutStatus,
  pauseRollout,
  resumeRollout,
  createCanaryDeployment,
  RolloutStatus,
} from '@/api/deployment';

interface DeploymentRolloutTabProps {
  clusterName: string;
  namespace: string;
  deploymentName: string;
}

const DeploymentRolloutTab: React.FC<DeploymentRolloutTabProps> = ({
  clusterName,
  namespace,
  deploymentName,
}) => {
  const { t } = useTranslation();
  const [status, setStatus] = useState<RolloutStatus | null>(null);
  const [loading, setLoading] = useState(false);
  const [actionLoading, setActionLoading] = useState(false);
  const [canaryVisible, setCanaryVisible] = useState(false);
  const [canaryStep, setCanaryStep] = useState(0);
  const [form] = Form.useForm();

  const fetchStatus = async () => {
    setLoading(true);
    try {
      const response = await getRolloutStatus(clusterName, namespace, deploymentName);
      if (response.data.code === 0) {
        setStatus(response.data.data.rollout);
      } else {
        message.error(response.data.message || t('deployments.detail.rollout.fetchFailed'));
      }
    } catch {
      message.error(t('deployments.detail.rollout.fetchFailed'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (clusterName && namespace && deploymentName) {
      fetchStatus();
      const timer = setInterval(fetchStatus, 15000);
      return () => clearInterval(timer);
    }
  }, [clusterName, namespace, deploymentName]);

  const handlePause = async () => {
    setActionLoading(true);
    try {
      await pauseRollout(clusterName, namespace, deploymentName);
      message.success(t('deployments.detail.rollout.pauseSuccess'));
      fetchStatus();
    } catch {
      message.error(t('deployments.detail.rollout.pauseFailed'));
    } finally {
      setActionLoading(false);
    }
  };

  const handleResume = async () => {
    setActionLoading(true);
    try {
      await resumeRollout(clusterName, namespace, deploymentName);
      message.success(t('deployments.detail.rollout.resumeSuccess'));
      fetchStatus();
    } catch {
      message.error(t('deployments.detail.rollout.resumeFailed'));
    } finally {
      setActionLoading(false);
    }
  };

  const handleCanaryCreate = async (values: {
    name: string;
    replicas: number;
    canaryLabelKey?: string;
    canaryLabelValue?: string;
  }) => {
    setActionLoading(true);
    try {
      await createCanaryDeployment(clusterName, namespace, deploymentName, {
        name: values.name,
        replicas: values.replicas,
        canaryLabelKey: values.canaryLabelKey,
        canaryLabelValue: values.canaryLabelValue,
      });
      message.success(t('deployments.detail.rollout.canarySuccess'));
      setCanaryVisible(false);
      setCanaryStep(0);
      form.resetFields();
      fetchStatus();
    } catch {
      message.error(t('deployments.detail.rollout.canaryFailed'));
    } finally {
      setActionLoading(false);
    }
  };

  const conditionColumns = [
    { title: t('deployments.detail.conditionColumns.type'), dataIndex: 'type', key: 'type' },
    {
      title: t('deployments.detail.conditionColumns.status'),
      dataIndex: 'status',
      key: 'status',
      render: (s: string) => <Tag color={s === 'True' ? 'success' : 'error'}>{s}</Tag>,
    },
    { title: t('deployments.detail.conditionColumns.reason'), dataIndex: 'reason', key: 'reason' },
    { title: t('deployments.detail.conditionColumns.message'), dataIndex: 'message', key: 'message' },
  ];

  if (loading && !status) {
    return <Spin />;
  }

  return (
    <Space direction="vertical" size="middle" style={{ width: '100%' }}>
      <Card
        title={t('deployments.detail.rollout.title')}
        extra={
          <Space>
            <Button
              icon={<PauseCircleOutlined />}
              onClick={handlePause}
              loading={actionLoading}
              disabled={status?.paused}
            >
              {t('deployments.detail.rollout.pause')}
            </Button>
            <Button
              icon={<PlayCircleOutlined />}
              onClick={handleResume}
              loading={actionLoading}
              disabled={!status?.paused}
            >
              {t('deployments.detail.rollout.resume')}
            </Button>
            <Button
              type="primary"
              icon={<RocketOutlined />}
              onClick={() => {
                form.setFieldsValue({
                  name: `${deploymentName}-canary`,
                  replicas: 1,
                  canaryLabelKey: 'track',
                  canaryLabelValue: 'canary',
                });
                setCanaryVisible(true);
              }}
            >
              {t('deployments.detail.rollout.canary')}
            </Button>
          </Space>
        }
      >
        <Descriptions column={2}>
          <Descriptions.Item label={t('deployments.detail.rollout.paused')}>
            <Tag color={status?.paused ? 'warning' : 'success'}>
              {status?.paused ? t('common.yes') : t('common.no')}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label={t('deployments.detail.rollout.replicas')}>
            {status?.readyReplicas ?? 0}/{status?.replicas ?? 0}
          </Descriptions.Item>
          <Descriptions.Item label={t('deployments.detail.rollout.updated')}>
            {status?.updatedReplicas ?? 0}
          </Descriptions.Item>
          <Descriptions.Item label={t('deployments.detail.rollout.available')}>
            {status?.availableReplicas ?? 0}
          </Descriptions.Item>
          <Descriptions.Item label={t('deployments.detail.rollout.unavailable')}>
            {status?.unavailableReplicas ?? 0}
          </Descriptions.Item>
          <Descriptions.Item label={t('deployments.detail.rollout.generation')}>
            {status?.observedGeneration ?? '-'}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {(status?.conditions?.length ?? 0) > 0 && (
        <Card title={t('deployments.detail.rollout.conditions')}>
          <Table
            columns={conditionColumns}
            dataSource={status?.conditions || []}
            rowKey="type"
            pagination={false}
            scroll={{ x: 'max-content' }}
          />
        </Card>
      )}

      <Modal
        title={t('deployments.detail.rollout.canaryWizard')}
        open={canaryVisible}
        onCancel={() => {
          setCanaryVisible(false);
          setCanaryStep(0);
        }}
        footer={
          <Space>
            {canaryStep > 0 && (
              <Button onClick={() => setCanaryStep(canaryStep - 1)}>{t('common.back')}</Button>
            )}
            {canaryStep < 2 ? (
              <Button type="primary" onClick={() => setCanaryStep(canaryStep + 1)}>
                {t('common.confirm')}
              </Button>
            ) : (
              <Button type="primary" loading={actionLoading} onClick={() => form.submit()}>
                {t('deployments.detail.rollout.createCanary')}
              </Button>
            )}
          </Space>
        }
        width={560}
      >
        <Steps
          current={canaryStep}
          style={{ marginBottom: 24 }}
          items={[
            { title: t('deployments.detail.rollout.stepName') },
            { title: t('deployments.detail.rollout.stepLabels') },
            { title: t('deployments.detail.rollout.stepReview') },
          ]}
        />
        <Form form={form} layout="vertical" onFinish={handleCanaryCreate}>
          <div style={{ display: canaryStep === 0 ? 'block' : 'none' }}>
            <Form.Item name="name" label={t('common.name')} rules={[{ required: true }]}>
              <Input />
            </Form.Item>
            <Form.Item name="replicas" label={t('deployments.detail.rollout.canaryReplicas')} rules={[{ required: true }]}>
              <InputNumber min={1} style={{ width: '100%' }} />
            </Form.Item>
          </div>
          <div style={{ display: canaryStep === 1 ? 'block' : 'none' }}>
            <Form.Item name="canaryLabelKey" label={t('deployments.detail.rollout.labelKey')}>
              <Input />
            </Form.Item>
            <Form.Item name="canaryLabelValue" label={t('deployments.detail.rollout.labelValue')}>
              <Input />
            </Form.Item>
          </div>
          <div style={{ display: canaryStep === 2 ? 'block' : 'none' }}>
            <Form.Item shouldUpdate>
              {() => (
                <Descriptions column={1} size="small">
                  <Descriptions.Item label={t('common.name')}>{form.getFieldValue('name')}</Descriptions.Item>
                  <Descriptions.Item label={t('deployments.detail.rollout.canaryReplicas')}>
                    {form.getFieldValue('replicas')}
                  </Descriptions.Item>
                  <Descriptions.Item label={t('deployments.detail.rollout.label')}>
                    {`${form.getFieldValue('canaryLabelKey') || 'track'}=${form.getFieldValue('canaryLabelValue') || 'canary'}`}
                  </Descriptions.Item>
                </Descriptions>
              )}
            </Form.Item>
          </div>
        </Form>
      </Modal>
    </Space>
  );
};

export default DeploymentRolloutTab;
