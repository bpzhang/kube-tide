import React, { useState, useEffect } from 'react';
import { Row, Col, Card, Select, message, Space, Modal, Button, Pagination, Spin } from 'antd';
import { ExclamationCircleOutlined, PlusOutlined, SettingOutlined } from '@ant-design/icons';
import { 
  getNodeList, 
  getNodeMetrics, 
  drainNode, 
  cordonNode, 
  uncordonNode,
  removeNode,
} from '../api/node';
import { getNodePools } from '../api/nodepool';
import { getClusterList } from '../api/cluster';
import type { NodePool } from '../api/nodepool';
import NodeStatus from '../components/k8s/node/NodeStatus';
import TaintsManageModal from '../components/k8s/common/TaintsManageModal';
import LabelsManageModal from '../components/k8s/common/LabelsManageModal';
import AddNodeModal from '../components/k8s/node/AddNodeModal';
import NodePoolsManager from '../components/k8s/node/NodePoolsManager';

const { Option } = Select;

const Nodes: React.FC = () => {
  const [selectedCluster, setSelectedCluster] = useState<string>('');
  const [clusters, setClusters] = useState<string[]>([]);
  const [nodes, setNodes] = useState<any[]>([]);
  const [nodePools, setNodePools] = useState<NodePool[]>([]);
  const [metrics, setMetrics] = useState<{[key: string]: any}>({});
  const [loading, setLoading] = useState(false);
  const [actionLoading, setActionLoading] = useState(false);
  const [addModalVisible, setAddModalVisible] = useState(false);
  const [selectedNode, setSelectedNode] = useState<string>('');
  const [taintModalVisible, setTaintModalVisible] = useState(false);
  const [labelModalVisible, setLabelModalVisible] = useState(false);
  const [nodePoolManagerVisible, setNodePoolManagerVisible] = useState(false);
  
  // 分页相关状态
  const [currentPage, setCurrentPage] = useState<number>(1);
  const [pageSize, setPageSize] = useState<number>(8); // 每页显示数量，调整为适合卡片布局的数量
  const [total, setTotal] = useState<number>(0);
  const [totalPages, setTotalPages] = useState<number>(0);

  // 获取节点池列表
  const fetchNodePools = async () => {
    if (!selectedCluster) return;
    try {
      const response = await getNodePools(selectedCluster);
      if (response.data.code === 0) {
        setNodePools(response.data.data.pools || []);
      } else {
        console.error('获取节点池失败:', response.data.message);
        message.error(response.data.message || '获取节点池列表失败');
      }
    } catch (err) {
      console.error('获取节点池失败:', err);
      message.error('获取节点池列表失败');
    }
  };

  // 获取集群列表
  const fetchClusters = async () => {
    try {
      const response = await getClusterList();
      if (response.data.code === 0) {
        const clusterList = response.data.data.clusters;
        setClusters(clusterList);
        if (clusterList.length > 0 && !selectedCluster) {
          setSelectedCluster(clusterList[0]);
        }
      }
    } catch (err) {
      message.error('获取集群列表失败');
    }
  };

  const fetchNodes = async (page: number = currentPage, size: number = pageSize) => {
    if (!selectedCluster) return;
    
    setLoading(true);
    try {
      const response = await getNodeList(selectedCluster, page, size);
      if (response.data.code === 0) {
        const nodeList = response.data.data.nodes || [];
        setNodes(nodeList);
        
        // 更新分页信息
        if (response.data.data.pagination) {
          setTotal(response.data.data.pagination.total);
          setTotalPages(response.data.data.pagination.pages);
        }
        
        // 获取每个节点的指标
        const metricsData: {[key: string]: any} = {};
        for (const node of nodeList) {
          if (!node.metadata?.name) continue; // 跳过没有名称的节点
          
          try {
            const metricsResponse = await getNodeMetrics(selectedCluster, node.metadata.name);
            if (metricsResponse.data.code === 0) {
              metricsData[node.metadata.name] = metricsResponse.data.data.metrics;
            }
          } catch (err) {
            console.error(`获取节点 ${node.metadata.name} 的指标失败:`, err);
          }
        }
        setMetrics(metricsData);
      } else {
        message.error(response.data.message || '获取节点列表失败');
        setNodes([]);
      }
    } catch (err) {
      message.error('获取节点列表失败');
      setNodes([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchClusters();
  }, []);

  useEffect(() => {
    if (selectedCluster) {
      fetchNodes();
      fetchNodePools();
      // 每30秒刷新一次
      const timer = setInterval(() => {
        fetchNodes();
        fetchNodePools();
      }, 30000);
      return () => clearInterval(timer);
    }
  }, [selectedCluster]);

  const handleClusterChange = (value: string) => {
    setSelectedCluster(value);
  };

  // 处理污点管理
  const handleManageTaints = (nodeName: string) => {
    setSelectedNode(nodeName);
    setTaintModalVisible(true);
  };

  // 处理标签管理
  const handleManageLabels = (nodeName: string) => {
    setSelectedNode(nodeName);
    setLabelModalVisible(true);
  };

  // 节点排水操作
  const handleDrainNode = async (nodeName: string) => {
    try {
      setActionLoading(true);
      message.loading({ content: '正在执行节点排水操作...', key: 'drainNode', duration: 0 });
      
      await drainNode(selectedCluster, nodeName, {
        gracePeriodSeconds: 300,
        deleteLocalData: false,
        ignoreDaemonSets: true
      });
      
      message.success({ content: `节点 ${nodeName} 排水操作已完成`, key: 'drainNode' });
      await fetchNodes(); // 刷新节点列表
    } catch (err) {
      message.error({ 
        content: `节点排水失败: ${(err as Error).message}`, 
        key: 'drainNode'
      });
    } finally {
      setActionLoading(false);
    }
  };

  // 禁止节点调度
  const handleCordonNode = async (nodeName: string) => {
    setActionLoading(true);
    try {
      await cordonNode(selectedCluster, nodeName);
      message.success(`节点 ${nodeName} 已设置为不可调度`);
      fetchNodes(); // 刷新节点列表
    } catch (err) {
      message.error(`设置节点为不可调度失败: ${(err as Error).message}`);
    } finally {
      setActionLoading(false);
    }
  };

  // 允许节点调度
  const handleUncordonNode = async (nodeName: string) => {
    setActionLoading(true);
    try {
      await uncordonNode(selectedCluster, nodeName);
      message.success(`节点 ${nodeName} 已恢复调度功能`);
      fetchNodes(); // 刷新节点列表
    } catch (err) {
      message.error(`恢复节点调度功能失败: ${(err as Error).message}`);
    } finally {
      setActionLoading(false);
    }
  };

  // 处理删除节点
  const handleRemoveNode = (nodeName: string) => {
    Modal.confirm({
      title: '确认删除节点',
      icon: <ExclamationCircleOutlined />,
      content: (
        <div>
          <p>您确定要删除节点 <strong>{nodeName}</strong> 吗？</p>
          <p>此操作将从集群中移除该节点。如果节点上还有Pod运行，您可以选择强制删除。</p>
          <p>强制删除将：</p>
          <ul>
            <li>设置节点为不可调度</li>
            <li>驱逐节点上的所有Pod</li>
            <li>从集群中删除节点</li>
          </ul>
        </div>
      ),
      okText: '强制删除',
      cancelText: '取消',
      onOk: async () => {
        try {
          await removeNode(selectedCluster, nodeName, { force: true });
          message.success(`节点 ${nodeName} 删除成功`);
          fetchNodes(); // 刷新节点列表
        } catch (err) {
          message.error(`删除节点失败: ${(err as Error).message}`);
        }
      }
    });
  };

  // 处理页码变化
  const handlePageChange = (page: number, pageSize?: number) => {
    setCurrentPage(page);
    if (pageSize) {
      setPageSize(pageSize);
    }
    fetchNodes(page, pageSize);
  };

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ marginBottom: '16px' }}>
        <Space>
          <Select
            style={{ width: 200 }}
            value={selectedCluster}
            onChange={handleClusterChange}
            placeholder="选择集群"
          >
            {clusters.map(cluster => (
              <Option key={cluster} value={cluster}>{cluster}</Option>
            ))}
          </Select>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => setAddModalVisible(true)}
            disabled={!selectedCluster}
          >
            添加节点
          </Button>
          <Button
            icon={<SettingOutlined />}
            onClick={() => setNodePoolManagerVisible(true)}
            disabled={!selectedCluster}
          >
            节点池管理
          </Button>
        </Space>
      </div>

      {/* 节点列表 */}
      <Spin spinning={loading}>
        <Row gutter={[16, 16]}>
          {nodes.map(node => (
            <Col key={node.metadata?.name} xs={24} md={12}>
              <NodeStatus 
                node={{
                  name: node.metadata?.name || '',
                  status: node.status?.conditions?.find((c: any) => c.type === 'Ready')?.status === 'True' ? 'Ready' : 'NotReady',
                  ip: node.status?.addresses?.find((addr: any) => addr.type === 'InternalIP')?.address || '',
                  os: node.status?.nodeInfo?.osImage || '',
                  kernelVersion: node.status?.nodeInfo?.kernelVersion || '',
                  containerRuntime: node.status?.nodeInfo?.containerRuntimeVersion || '',
                  unschedulable: node.spec?.unschedulable || false, // 添加不可调度状态
                }}
                metrics={metrics[node.metadata?.name]}
                clusterName={selectedCluster}
                onDrain={handleDrainNode}
                onCordon={handleCordonNode}
                onUncordon={handleUncordonNode}
                onManageTaints={handleManageTaints}
                onManageLabels={handleManageLabels}
                onDelete={handleRemoveNode}
              />
            </Col>
          ))}
        </Row>
        
        {/* 分页控件 */}
        {total > 0 && (
          <div style={{ marginTop: '20px', textAlign: 'right' }}>
            <Pagination 
              current={currentPage}
              pageSize={pageSize}
              total={total}
              onChange={handlePageChange}
              showTotal={(total) => `共 ${total} 个节点`}
              showSizeChanger
              pageSizeOptions={['4', '8', '12', '16']}
            />
          </div>
        )}
      </Spin>

      {/* 添加节点弹窗 */}
      <AddNodeModal
        open={addModalVisible}
        onClose={() => setAddModalVisible(false)}
        clusterName={selectedCluster}
        onSuccess={fetchNodes}
        nodePools={nodePools}
      />

      {/* 节点池管理器 */}
      {selectedCluster && (
        <NodePoolsManager
          visible={nodePoolManagerVisible}
          onClose={() => setNodePoolManagerVisible(false)}
          clusterName={selectedCluster}
          nodePools={nodePools}
          onSuccess={fetchNodePools}
        />
      )}

      {/* 污点管理弹窗 */}
      <TaintsManageModal
        open={taintModalVisible}
        onClose={() => setTaintModalVisible(false)}
        clusterName={selectedCluster}
        nodeName={selectedNode}
        onSuccess={fetchNodes}
      />

      {/* 标签管理弹窗 */}
      <LabelsManageModal
        open={labelModalVisible}
        onClose={() => setLabelModalVisible(false)}
        clusterName={selectedCluster}
        nodeName={selectedNode}
        onSuccess={fetchNodes}
      />
    </div>
  );
};

export default Nodes;