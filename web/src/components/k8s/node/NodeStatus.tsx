import { type FC } from 'react';
import { Card, Descriptions, Space, Row, Col, Button, Dropdown, Modal } from "antd";
import { CloudServerOutlined, MoreOutlined, ExclamationCircleOutlined, EyeOutlined } from "@ant-design/icons";
import { useNavigate } from "react-router-dom";
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

  const handleViewDetails = () => {
    navigate(`/nodes/${clusterName}/${node.name}`);
  };

  const handleDrain = () => {
    Modal.confirm({
      title: '节点排水确认',
      icon: <ExclamationCircleOutlined style={{ color: '#1677ff' }} />,
      content: (
        <div>
          <p>确定要对节点 <strong>{node.name}</strong> 进行排水操作吗？</p>
          <p>此操作将：</p>
          <ul>
            <li>将节点标记为不可调度</li>
            <li>安全地驱逐该节点上的所有Pod（DaemonSet除外）</li>
            <li>等待所有Pod迁移完成</li>
          </ul>
          <p>配置选项：</p>
          <ul>
            <li>忽略DaemonSet：是</li>
            <li>删除本地数据：否</li>
            <li>宽限期：300秒</li>
          </ul>
        </div>
      ),
      okText: "确认排水",
      cancelText: "取消",
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
        <div>加载中...</div>
      </Card>
    );
  }

  const items: MenuProps['items'] = [
    {
      key: 'view-details',
      label: "查看详情",
      onClick: handleViewDetails,
    },
    {
      key: 'drain',
      label: "排水(Drain)",
      onClick: handleDrain,
      disabled: !node.status || node.unschedulable,
      danger: true,
    },
    {
      key: 'cordon',
      label: "禁止调度(Cordon)",
      onClick: () => onCordon?.(node.name),
      disabled: node.unschedulable,
    },
    {
      key: 'uncordon',
      label: "允许调度(Uncordon)",
      onClick: () => onUncordon?.(node.name),
      disabled: !node.unschedulable,
    },
    {
      key: 'taint-management',
      label: "污点管理",
      onClick: () => onManageTaints?.(node.name),
    },
    {
      key: 'label-management',
      label: "标签管理",
      onClick: () => onManageLabels?.(node.name),
    },
    {
      type: 'divider',
    },
    {
      key: 'delete',
      label: "删除节点",
      danger: true,
      onClick: () => onDelete?.(node.name),
    },
  ];

  // Get node pool name from labels
  const nodePool = node.labels?.['k8s.io/pool-name'] || '未分配';

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
        <Descriptions.Item label="状态">
          <Space>
            <span style={{ color: node.status === 'Ready' ? '#52c41a' : '#ff4d4f' }}>
              {node.status}
            </span>
            {node.unschedulable && (
              <span style={{ color: '#faad14' }}>(不可调度)</span>
            )}
          </Space>
        </Descriptions.Item>
        <Descriptions.Item label="IP地址">{node.ip}</Descriptions.Item>
        <Descriptions.Item label="容器运行时">{node.containerRuntime}</Descriptions.Item>
        <Descriptions.Item label="系统版本">{node.os}</Descriptions.Item>
        <Descriptions.Item label="节点池">{nodePool}</Descriptions.Item>
      </Descriptions>

      <div style={{ marginTop: 16 }}>
        <Row gutter={[16, 16]}>
          <Col span={24}>
            <Card size="small" title="CPU资源信息" styles={{body: { display: 'flex', justifyContent: 'space-between' }}}>
              <span>总量: {formatCPU(metrics?.cpu_capacity)}</span>
              <span>已用: {formatCPU(metrics?.cpu_usage)}</span>
              <span>请求: {formatCPU(metrics?.cpu_requests)}</span>
              <span>限制: {formatCPU(metrics?.cpu_limits)}</span>
            </Card>
          </Col>
          <Col span={24}>
            <Card size="small" title="内存资源信息" styles={{body: { display: 'flex', justifyContent: 'space-between' }}}>
              <span>总量: {formatMemorySize(metrics?.memory_capacity)}</span>
              <span>已用: {formatMemorySize(metrics?.memory_usage)}</span>
              <span>请求: {formatMemorySize(metrics?.memory_requests)}</span>
              <span>限制: {formatMemorySize(metrics?.memory_limits)}</span>
            </Card>
          </Col>
          <Col span={24} style={{ textAlign: 'right' }}>
            <Button 
              type="primary" 
              icon={<EyeOutlined />} 
              onClick={handleViewDetails}
            >
              查看详情
            </Button>
          </Col>
        </Row>
      </div>
    </Card>
  );
};

export default NodeStatus;