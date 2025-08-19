import React, { useState, useEffect } from 'react';
import { Card, Select, Button, message, Space, Typography, Alert, Modal, Checkbox } from 'antd';
import { ExclamationCircleOutlined, SettingOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { getPodRestartPolicy, updatePodRestartPolicy } from '@/api/pod';

const { Text } = Typography;
const { Option } = Select;
const { confirm } = Modal;

interface PodRestartPolicyConfigProps {
  clusterName: string;
  namespace: string;
  podName: string;
  disabled?: boolean;
}

const PodRestartPolicyConfig: React.FC<PodRestartPolicyConfigProps> = ({
  clusterName,
  namespace,
  podName,
  disabled = false,
}) => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [updating, setUpdating] = useState(false);
  const [currentPolicy, setCurrentPolicy] = useState<string>('');
  const [selectedPolicy, setSelectedPolicy] = useState<string>('');
  const [deleteOriginal, setDeleteOriginal] = useState(false);

  // 重启策略选项
  const restartPolicyOptions = [
    {
      value: 'Always',
      label: t('pod.restartPolicy.always'),
      description: 'Container will always be restarted',
    },
    {
      value: 'OnFailure',
      label: t('pod.restartPolicy.onFailure'),
      description: 'Container will be restarted only when it fails',
    },
    {
      value: 'Never',
      label: t('pod.restartPolicy.never'),
      description: 'Container will never be restarted',
    },
  ];

  // 获取当前重启策略
  const fetchRestartPolicy = async () => {
    setLoading(true);
    try {
      const response = await getPodRestartPolicy(clusterName, namespace, podName);
      if (response.data.code === 200) {
        const policy = response.data.data.restartPolicy;
        setCurrentPolicy(policy);
        setSelectedPolicy(policy);
      } else {
        message.error(t('pod.restartPolicy.fetchFailed'));
      }
    } catch (error: any) {
      console.error('Failed to fetch restart policy:', error);
      // 显示更详细的错误信息
      let errorMsg = t('pod.restartPolicy.fetchFailed');
      if (error?.response?.data?.message) {
        errorMsg += ': ' + error.response.data.message;
      } else if (error?.message) {
        errorMsg += ': ' + error.message;
      }
      message.error(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  // 更新重启策略
  const handleUpdatePolicy = async () => {
    if (selectedPolicy === currentPolicy) {
      message.info(t('pod.restartPolicy.noChanges') || 'No changes detected in restart policy.');
      return;
    }

    confirm({
      title: t('pod.restartPolicy.confirm'),
      icon: <ExclamationCircleOutlined />,
      content: (
        <div>
          <Alert
            message="Important Note"
            description={deleteOriginal 
              ? "This will delete the current Pod and create a new one with the specified restart policy."
              : "This will create a new Pod with the specified restart policy. The original Pod will remain unchanged."
            }
            type="warning"
            showIcon
            style={{ marginBottom: 16 }}
          />
          <p>
            <Text strong>Current:</Text> {getOptionLabel(currentPolicy)}
          </p>
          <p>
            <Text strong>New:</Text> {getOptionLabel(selectedPolicy)}
          </p>
        </div>
      ),
      onOk: async () => {
        setUpdating(true);
        try {
          const response = await updatePodRestartPolicy(
            clusterName,
            namespace,
            podName,
            selectedPolicy,
            deleteOriginal
          );
          if (response.data.code === 200) {
            if (deleteOriginal) {
              message.success(t('pod.restartPolicy.recreateSuccess'));
            } else {
              message.success(t('pod.restartPolicy.createSuccess'));
            }
            // 刷新当前策略
            fetchRestartPolicy();
          } else {
            message.error(t('pod.restartPolicy.updateFailed'));
          }
        } catch (error: any) {
          console.error('Failed to update restart policy:', error);
          // 显示更详细的错误信息
          let errorMsg = t('pod.restartPolicy.updateFailed');
          if (error?.response?.data?.message) {
            errorMsg += ': ' + error.response.data.message;
          } else if (error?.message) {
            errorMsg += ': ' + error.message;
          }
          message.error(errorMsg);
          // 重置选择
          setSelectedPolicy(currentPolicy);
        } finally {
          setUpdating(false);
        }
      },
    });
  };

  // 获取选项标签
  const getOptionLabel = (value: string) => {
    const option = restartPolicyOptions.find(opt => opt.value === value);
    return option ? option.label : value;
  };

  useEffect(() => {
    fetchRestartPolicy();
  }, [clusterName, namespace, podName]);

  return (
    <Card
      title={
        <Space>
          <SettingOutlined />
          {t('pod.restartPolicy.title')}
        </Space>
      }
      size="small"
      loading={loading}
    >
      <Space direction="vertical" style={{ width: '100%' }}>
        <Alert
          message={t('pod.restartPolicy.alertMessage')}
          description={t('pod.restartPolicy.alertDescription')}
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
        />
        
        <div>
          <Text strong>{t('pod.restartPolicy.current')}: </Text>
          <Text code>{getOptionLabel(currentPolicy)}</Text>
        </div>

        <div>
          <label style={{ display: 'block', marginBottom: 8 }}>
            <Text strong>{t('pod.restartPolicy.update')}:</Text>
          </label>
          <Select
            value={selectedPolicy}
            onChange={setSelectedPolicy}
            style={{ width: '100%', marginBottom: 16 }}
            disabled={disabled || updating}
            placeholder={t('pod.restartPolicy.title')}
          >
            {restartPolicyOptions.map(option => (
              <Option key={option.value} value={option.value}>
                <div>
                  <div style={{ fontWeight: 'bold' }}>{option.label}</div>
                  <div style={{ fontSize: '12px', color: '#666' }}>
                    {option.description}
                  </div>
                </div>
              </Option>
            ))}
          </Select>
        </div>

        <div style={{ marginBottom: 16 }}>
          <Checkbox
            checked={deleteOriginal}
            onChange={(e) => setDeleteOriginal(e.target.checked)}
            disabled={disabled || updating}
          >
            {t('pod.restartPolicy.deleteOriginal')}
          </Checkbox>
        </div>

        <Button
          type="primary"
          onClick={handleUpdatePolicy}
          loading={updating}
          disabled={disabled || selectedPolicy === currentPolicy}
          style={{ width: '100%' }}
        >
          {deleteOriginal ? 'Recreate Pod' : 'Create New Pod'}
        </Button>
      </Space>
    </Card>
  );
};

export default PodRestartPolicyConfig;
