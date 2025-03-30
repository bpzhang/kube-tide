import React, { useState, useEffect, useCallback, useRef, useMemo } from 'react';
import { Table, Tag, Button, Popconfirm, message, Space, Input, Select, Card, Row, Col, Tooltip } from 'antd';
import { EyeOutlined, SearchOutlined, FilterOutlined, ReloadOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { deletePod, getPodsByNamespace } from '@/api/pod';

const { Option } = Select;

interface PodListProps {
  clusterName: string;
  namespace: string;
  pods: any[];
  onRefresh: () => void;
  isParamChange?: boolean; // 是否是参数变更触发的刷新
}

interface ProcessedPod {
  name: string;
  status: string;
  podIP: string;
  nodeName: string;
  createdAt: string;
  namespace: string;
  labels?: { [key: string]: string };
  // 保留原始 pod 数据以供详情查看
  rawPod: any;
}

const PodList: React.FC<PodListProps> = ({ clusterName, namespace, pods, onRefresh, isParamChange = false }) => {
  const navigate = useNavigate();
  const [filteredPods, setFilteredPods] = useState<ProcessedPod[]>([]);
  const [searchText, setSearchText] = useState('');
  const [statusFilter, setStatusFilter] = useState<string[]>([]);
  const [nodeFilter, setNodeFilter] = useState<string[]>([]);
  const [ipFilter, setIpFilter] = useState('');
  const [availableNodes, setAvailableNodes] = useState<string[]>([]);
  const [availableStatuses, setAvailableStatuses] = useState<string[]>([]);
  const [localPods, setLocalPods] = useState<any[]>([]);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const tableRef = useRef<any>(null);
  
  // 保存当前的排序和分页状态
  const [tableState, setTableState] = useState({
    pageSize: 10,
    current: 1,
    sortField: 'createdAt',
    sortOrder: 'descend',
  });

  // 使用接收到的 pods 来初始化或更新 localPods
  useEffect(() => {
    if (isParamChange) {
      // 如果是参数变更（如集群或命名空间改变），重置筛选条件
      resetFilters();
      setLocalPods(pods);
      // 重置排序和分页
      setTableState({
        pageSize: 10,
        current: 1,
        sortField: 'createdAt',
        sortOrder: 'descend',
      });
    } else if (!isRefreshing) {
      // 普通刷新，保留筛选条件和排序
      setLocalPods(prevPods => {
        // 使用引用比较优化性能，避免不必要的重新渲染
        if (pods === prevPods) return prevPods;
        
        // 只在数据真正改变时更新
        const podsChanged = JSON.stringify(pods) !== JSON.stringify(prevPods);
        return podsChanged ? pods : prevPods;
      });
    }
  }, [pods, isRefreshing, isParamChange]);

  // 处理原始 Pod 数据
  const processedPods: ProcessedPod[] = useMemo(() => 
    localPods.map(pod => ({
      name: pod.metadata?.name || '',
      namespace: pod.metadata?.namespace || namespace,
      status: pod.status?.phase || 'Unknown',
      podIP: pod.status?.podIP || '-',
      nodeName: pod.spec?.nodeName || '-',
      createdAt: pod.metadata?.creationTimestamp || '-',
      labels: pod.metadata?.labels || {},
      rawPod: pod,
    })), 
    [localPods, namespace]
  );

  // 当 localPods 数据变化时，重新计算可用的状态和节点列表
  useEffect(() => {
    // 节点优化：只在本地数据真正变化时才重新计算节点和状态列表
    const nodes = Array.from(new Set(processedPods.map(pod => pod.nodeName))).filter(node => node !== '-');
    const statuses = Array.from(new Set(processedPods.map(pod => pod.status)));
    
    // 只在列表真正变化时才更新状态，减少不必要的重渲染
    if (JSON.stringify(nodes) !== JSON.stringify(availableNodes)) {
      setAvailableNodes(nodes);
    }
    
    if (JSON.stringify(statuses) !== JSON.stringify(availableStatuses)) {
      setAvailableStatuses(statuses);
    }
    
    applyFilters(processedPods);
  }, [processedPods]);

  // 当筛选条件变化时，重新应用筛选
  useEffect(() => {
    applyFilters(processedPods);
  }, [searchText, statusFilter, nodeFilter, ipFilter, processedPods]);

  // 应用筛选条件
  const applyFilters = (pods: ProcessedPod[]) => {
    let result = [...pods];
    
    // 按名称搜索
    if (searchText) {
      result = result.filter(pod => 
        pod.name.toLowerCase().includes(searchText.toLowerCase()) ||
        (pod.labels && Object.entries(pod.labels).some(([k, v]) => 
          `${k}:${v}`.toLowerCase().includes(searchText.toLowerCase())
        ))
      );
    }
    
    // 按状态筛选
    if (statusFilter && statusFilter.length > 0) {
      result = result.filter(pod => statusFilter.includes(pod.status));
    }
    
    // 按节点筛选
    if (nodeFilter && nodeFilter.length > 0) {
      result = result.filter(pod => nodeFilter.includes(pod.nodeName));
    }
    
    // 按 IP 筛选
    if (ipFilter) {
      result = result.filter(pod => pod.podIP.includes(ipFilter));
    }
    
    setFilteredPods(result);
  };

  // 本地刷新Pod列表，不影响筛选条件
  const refreshPodList = useCallback(async () => {
    setIsRefreshing(true);
    try {
      const response = await getPodsByNamespace(clusterName, namespace);
      if (response.data.code === 0) {
        setLocalPods(response.data.data.pods || []);
        // 不重置筛选条件，保持当前筛选状态
      } else {
        message.error(response.data.message || '获取Pod列表失败');
      }
    } catch (err) {
      message.error('获取Pod列表失败');
    } finally {
      setIsRefreshing(false);
    }
  }, [clusterName, namespace]);

  const handleDelete = async (podName: string) => {
    try {
      await deletePod(clusterName, namespace, podName);
      message.success('Pod删除成功');
      // 在本地刷新数据，而不是调用父组件的onRefresh
      refreshPodList();
    } catch (err) {
      message.error('Pod删除失败');
    }
  };

  const handleViewDetails = (podName: string, podNamespace: string) => {
    navigate(`/workloads/pods/${clusterName}/${podNamespace}/${podName}`);
  };

  const getStatusColor = (status: string) => {
    const colors: { [key: string]: string } = {
      Running: 'green',
      Pending: 'gold',
      Failed: 'red',
      Unknown: 'grey',
      Succeeded: 'blue',
      Terminated: 'red',
    };
    return colors[status] || 'blue';
  };

  const resetFilters = () => {
    setSearchText('');
    setStatusFilter([]);
    setNodeFilter([]);
    setIpFilter('');
  };

  // 在顶层刷新按钮中使用本地刷新方法
  const handleRefresh = () => {
    refreshPodList();
  };

  // 处理表格变化事件，保存排序和分页状态
  const handleTableChange = (pagination: any, filters: any, sorter: any) => {
    setTableState({
      pageSize: pagination.pageSize,
      current: pagination.current,
      sortField: sorter.field,
      sortOrder: sorter.order,
    });
  };

  // 使用 useMemo 优化表格配置，避免不必要的重新渲染
  const columns = useMemo(() => [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: ProcessedPod) => (
        <a onClick={() => handleViewDetails(text, record.namespace)}>{text}</a>
      ),
      sorter: (a: ProcessedPod, b: ProcessedPod) => a.name.localeCompare(b.name),
    },
    {
      title: '命名空间',
      dataIndex: 'namespace',
      key: 'namespace',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>{status}</Tag>
      ),
      sorter: (a: ProcessedPod, b: ProcessedPod) => a.status.localeCompare(b.status),
    },
    {
      title: 'IP',
      dataIndex: 'podIP',
      key: 'podIP',
    },
    {
      title: '节点',
      dataIndex: 'nodeName',
      key: 'nodeName',
      sorter: (a: ProcessedPod, b: ProcessedPod) => a.nodeName.localeCompare(b.nodeName),
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (time: string) => new Date(time).toLocaleString(),
      sorter: (a: ProcessedPod, b: ProcessedPod) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime(),
      defaultSortOrder: 'descend',
    },
    {
      title: '标签',
      key: 'labels',
      render: (_: any, record: ProcessedPod) => (
        <div style={{ maxWidth: 200, maxHeight: 80, overflow: 'auto' }}>
          {record.labels && Object.entries(record.labels).map(([key, value]) => (
            <Tag key={key} color="blue">{`${key}: ${value}`}</Tag>
          ))}
        </div>
      ),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: ProcessedPod) => (
        <Space>
          <Button
            type="link"
            icon={<EyeOutlined />}
            onClick={() => handleViewDetails(record.name, record.namespace)}
          >
            详情
          </Button>
          <Popconfirm
            title="确定要删除这个Pod吗?"
            onConfirm={() => handleDelete(record.name)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" danger>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ], [clusterName, handleDelete, handleViewDetails]);

  return (
    <>
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={[16, 16]}>
          <Col span={6}>
            <Input 
              placeholder="搜索Pod名称或标签" 
              value={searchText}
              onChange={e => setSearchText(e.target.value)}
              prefix={<SearchOutlined />}
              allowClear
            />
          </Col>
          <Col span={5}>
            <Select
              mode="multiple"
              placeholder="筛选状态"
              value={statusFilter}
              onChange={setStatusFilter}
              style={{ width: '100%' }}
              allowClear
            >
              {availableStatuses.map(status => (
                <Option key={status} value={status}>
                  <Tag color={getStatusColor(status)}>{status}</Tag>
                </Option>
              ))}
            </Select>
          </Col>
          <Col span={5}>
            <Select
              mode="multiple"
              placeholder="筛选节点"
              value={nodeFilter}
              onChange={setNodeFilter}
              style={{ width: '100%' }}
              allowClear
            >
              {availableNodes.map(node => (
                <Option key={node} value={node}>{node}</Option>
              ))}
            </Select>
          </Col>
          <Col span={4}>
            <Input 
              placeholder="筛选IP" 
              value={ipFilter}
              onChange={e => setIpFilter(e.target.value)}
              allowClear
            />
          </Col>
          <Col span={4}>
            <Space>
              <Tooltip title="重置筛选条件">
                <Button icon={<FilterOutlined />} onClick={resetFilters}>
                  重置
                </Button>
              </Tooltip>
              <Tooltip title="刷新Pod列表">
                <Button type="primary" icon={<ReloadOutlined />} onClick={handleRefresh} loading={isRefreshing}>
                  刷新
                </Button>
              </Tooltip>
            </Space>
          </Col>
        </Row>
      </Card>
      <Table
        ref={tableRef}
        columns={columns}
        dataSource={filteredPods}
        rowKey="name"
        pagination={{ 
          pageSize: tableState.pageSize,
          current: tableState.current,
          showSizeChanger: true,
          showTotal: total => `共 ${total} 个Pod`,
          pageSizeOptions: ['10', '20', '50', '100']
        }}
        scroll={{ x: 'max-content' }}
        loading={isRefreshing}
        onChange={handleTableChange}
        // 保留当前排序状态
        sortDirections={['ascend', 'descend']}
      />
    </>
  );
};

export default PodList;