import { type FC } from 'react';
import { Card, Descriptions, Space, Row, Col, Button, Dropdown, Modal } from "antd";
import { CloudServerOutlined, MoreOutlined, ExclamationCircleOutlined, EyeOutlined } from "@ant-design/icons";
import { useNavigate } from "react-router-dom";
import { useTranslation } from 'react-i18next';
import type { MenuProps } from 'antd';

interface NodeInfo {
  name: string;
  status: string;
  ip: string;
  os: string;
  kernelVersion: string;
  containerRuntime: string;
  unschedulable: boolean;
  labels?: { [key: string]: string };
}

interface NodeMetrics {
  cpu_usage?: string;
  memory_usage?: string;
  cpu_capacity?: string;
  memory_capacity?: string;
  cpu_requests?: string;
  memory_requests?: string;
  cpu_limits?: string;
  memory_limits?: string;
}

interface NodeStatusProps {
  node: NodeInfo;
  metrics?: NodeMetrics;
  clusterName: string;
  onDrain: (nodeName: string) => void;
  onCordon: (nodeName: string) => void;
  onUncordon: (nodeName: string) => void;
  onManageTaints: (nodeName: string) => void;
  onManageLabels: (nodeName: string) => void;
  onDelete: (nodeName: string) => void;
}

const formatMemorySize = (size?: string): string => {
  if (!size || size === '0') return '0 GB';
  const value = parseInt(size.replace(/[^0-9]/g, ''));
  if (size.includes('Ki')) return `${(value / (1024 * 1024)).toFixed(2)} GB`;
  if (size.includes('Mi')) return `${(value / 1024).toFixed(2)} GB`;
  if (size.includes('Gi')) return `${value} GB`;
  return `${value} GB`;
};

const formatCPU = (cpu?: string): string => {
  if (!cpu || cpu === '0') return '0 Core';
  const value = parseInt(cpu.replace(/[^0-9]/g, ''));
  if (cpu.endsWith('m')) return `${(value / 1000).toFixed(2)} Core`;
  return `${value} Core`;
};

const NodeStatus: FC<NodeStatusProps> = ({ 
  node, 
  metrics = {}, 
  clusterName,
  onDrain, 
  onCordon, 
  onUncordon,
  onManageTaints,
  onManageLabels,
  onDelete
}) => {
  const navigate = useNavigate();
  const { t } = useTranslation();

  const handleViewDetails = () => {
    navigate(`/nodes/${clusterName}/${node.name}`);
  };

  const handleDrain = () => {
    Modal.confirm({
      title: t('nodeDetail.operations.drainConfirmTitle'),
      icon: <ExclamationCircleOutlined style={{ color: '#1677ff' }} />,
      content: (
        <div>
          <p>{t('nodeDetail.operations.drainConfirmMessage', { nodeName: node.name })}</p>
          <p>{t('nodeDetail.operations.drainExplanation')}</p>
          <ul>
            <li>{t('nodeDetail.operations.drainSetUnschedulable')}</li>
            <li>{t('nodeDetail.operations.drainEvictPods')}</li>
            <li>{t('nodeDetail.operations.drainWaitMigration')}</li>
          </ul>
          <p>{t('nodeDetail.operations.drainOptions')}</p>
          <ul>
            <li>{t('nodeDetail.operations.drainOptionDaemonSet')}</li>
            <li>{t('nodeDetail.operations.drainOptionLocalData')}</li>
            <li>{t('nodeDetail.operations.drainOptionGracePeriod')}</li>
          </ul>
        </div>
      ),
      okText: t('nodeDetail.operations.confirmDrain'),
      cancelText: t('common.cancel'),
      onOk: () => onDrain?.(node.name),
    });
  };

  if (!metrics) {
    return (
      <Card
        title={
          <Space>
            <CloudServerOutlined />
            {node.name}
          </Space>
        }
      >
        <div>{t('common.loading')}</div>
      </Card>
    );
  }

  const items: MenuProps['items'] = [
    {
      key: 'view-details',
      label: t('nodeDetail.operations.viewDetails'),
      onClick: handleViewDetails,
    },
    {
      key: 'drain',
      label: t('nodeDetail.operations.drain'),
      onClick: handleDrain,
      disabled: !node.status || node.unschedulable,
      danger: true,
    },
    {
      key: 'cordon',
      label: t('nodeDetail.operations.cordon'),
      onClick: () => onCordon?.(node.name),
      disabled: node.unschedulable,
    },
    {
      key: 'uncordon',
      label: t('nodeDetail.operations.uncordon'),
      onClick: () => onUncordon?.(node.name),
      disabled: !node.unschedulable,
    },
    {
      key: 'taint-management',
      label: t('nodeDetail.operations.manageTaints'),
      onClick: () => onManageTaints?.(node.name),
    },
    {
      key: 'label-management',
      label: t('nodeDetail.operations.manageLabels'),
      onClick: () => onManageLabels?.(node.name),
    },
    {
      type: 'divider',
    },
    {
      key: 'delete',
      label: t('nodeDetail.operations.delete'),
      danger: true,
      onClick: () => onDelete?.(node.name),
    },
  ];

  // Get node pool name from labels
  const nodePool = node.labels?.['k8s.io/pool-name'] || t('nodeDetail.nodePool.unassigned');

  return (
    <Card
      title={
        <Space>
          <CloudServerOutlined />
          {node.name}
        </Space>
      }
      extra={
        <Dropdown menu={{ items }} placement="bottomRight" arrow>
          <Button type="text" icon={<MoreOutlined />} />
        </Dropdown>
      }
    >
      <Descriptions column={2} size="small">
        <Descriptions.Item label={t('nodeDetail.basicInfo.status')}>
          <Space>
            <span style={{ color: node.status === 'Ready' ? '#52c41a' : '#ff4d4f' }}>
              {node.status}
            </span>
            {node.unschedulable && (
              <span style={{ color: '#faad14' }}>({t('nodeDetail.basicInfo.unschedulable')})</span>
            )}
          </Space>
        </Descriptions.Item>
        <Descriptions.Item label={t('nodeDetail.basicInfo.ipAddress')}>{node.ip}</Descriptions.Item>
        <Descriptions.Item label={t('nodeDetail.basicInfo.containerRuntime')}>{node.containerRuntime}</Descriptions.Item>
        <Descriptions.Item label={t('nodeDetail.basicInfo.os')}>{node.os}</Descriptions.Item>
        <Descriptions.Item label={t('nodeDetail.nodePool.title')}>{nodePool}</Descriptions.Item>
      </Descriptions>

      <div style={{ marginTop: 16 }}>
        <Row gutter={[16, 16]}>
          <Col span={24}>
            <Card size="small" title={t('nodeDetail.resourceUsage.cpu.title')} styles={{body: { display: 'flex', justifyContent: 'space-between' }}}>
              <span>{t('nodeDetail.resourceUsage.totalCapacity')}: {formatCPU(metrics?.cpu_capacity)}</span>
              <span>{t('nodeDetail.resourceUsage.used')}: {formatCPU(metrics?.cpu_usage)}</span>
              <span>{t('nodeDetail.resourceUsage.requested')}: {formatCPU(metrics?.cpu_requests)}</span>
              <span>{t('nodeDetail.resourceUsage.limited')}: {formatCPU(metrics?.cpu_limits)}</span>
            </Card>
          </Col>
          <Col span={24}>
            <Card size="small" title={t('nodeDetail.resourceUsage.memory.title')} styles={{body: { display: 'flex', justifyContent: 'space-between' }}}>
              <span>{t('nodeDetail.resourceUsage.totalCapacity')}: {formatMemorySize(metrics?.memory_capacity)}</span>
              <span>{t('nodeDetail.resourceUsage.used')}: {formatMemorySize(metrics?.memory_usage)}</span>
              <span>{t('nodeDetail.resourceUsage.requested')}: {formatMemorySize(metrics?.memory_requests)}</span>
              <span>{t('nodeDetail.resourceUsage.limited')}: {formatMemorySize(metrics?.memory_limits)}</span>
            </Card>
          </Col>
          <Col span={24} style={{ textAlign: 'right' }}>
            <Button 
              type="primary" 
              icon={<EyeOutlined />} 
              onClick={handleViewDetails}
            >
              {t('nodeDetail.operations.viewDetails')}
            </Button>
          </Col>
        </Row>
      </div>
    </Card>
  );
};

export default NodeStatus;