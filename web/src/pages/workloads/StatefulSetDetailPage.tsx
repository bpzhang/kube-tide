import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Button, Spin, message, Tabs, Space, Popconfirm, Tooltip } from 'antd';
import { 
  ArrowLeftOutlined, 
  ReloadOutlined, 
  ScissorOutlined, 
  DeleteOutlined,
  SyncOutlined
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { getStatefulSetDetails, deleteStatefulSet, restartStatefulSet, getStatefulSetPods, getAllStatefulSetEvents } from '@/api/statefulset';
import ScaleStatefulSetModal from '@/components/k8s/statefulset/ScaleStatefulSetModal';
import K8sEvents from '@/components/k8s/common/K8sEvents';
import './StatefulSetDetailPage.css';

const { TabPane } = Tabs;

interface StatefulSetParams {
  clusterName: string;
  namespace: string;
  statefulsetName: string;
}

interface Container {
  name: string;
  image: string;
  resources?: {
    requests?: Record<string, string | number>;
    limits?: Record<string, string | number>;
  };
  env?: Array<{
    name: string;
    value?: string;
    valueFrom?: any;
  }>;
  ports?: Array<{
    name?: string;
    containerPort: number;
    protocol: string;
  }>;
}

interface VolumeClaimTemplate {
  name: string;
  storageClassName: string;
  accessModes: string[];
  storage: string;
}

interface StatefulSet {
  name: string;
  namespace: string;
  replicas: number;
  readyReplicas: number;
  serviceName: string;
  updateStrategy: string;
  podManagementPolicy: string;
  creationTime: string;
  labels?: Record<string, string>;
  annotations?: Record<string, string>;
  containers?: Container[];
  volumeClaimTemplates?: VolumeClaimTemplate[];
}

interface Pod {
  metadata: {
    name: string;
    namespace: string;
    creationTimestamp: string;
  };
  status: {
    phase: string;
  };
}

interface Event {
  type: string;
  reason: string;
  message: string;
  count: number;
  firstTimestamp: string;
  lastTimestamp: string;
  involvedObject: {
    kind: string;
    name: string;
  };
}

interface Events {
  statefulset: Event[];
  pod: Event[];
}

/**
 * StatefulSet详情页面
 */
const StatefulSetDetailPage: React.FC = () => {
  const { t } = useTranslation();
  const { clusterName, namespace, statefulsetName } = useParams<StatefulSetParams>();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [statefulset, setStatefulSet] = useState<StatefulSet | null>(null);
  const [pods, setPods] = useState<Pod[]>([]);
  const [events, setEvents] = useState<Events>({ statefulset: [], pod: [] });
  const [scaleModalVisible, setScaleModalVisible] = useState(false);

  // 获取StatefulSet详情
  const fetchStatefulSetDetails = async () => {
    if (!clusterName || !namespace || !statefulsetName) {
      message.error(t('common.error'));
      navigate('/workloads/statefulsets');
      return;
    }

    setLoading(true);
    try {
      const response = await getStatefulSetDetails(clusterName, namespace, statefulsetName);
      if (response.data.code === 0) {
        setStatefulSet(response.data.data.statefulset);
      } else {
        message.error(response.data.message || t('statefulsets.fetchDetailsFailed'));
      }
    } catch (error) {
      console.error(t('statefulsets.fetchDetailsFailed'), error);
      message.error(t('statefulsets.fetchDetailsFailed'));
    } finally {
      setLoading(false);
    }
  };

  // 获取StatefulSet关联的Pod
  const fetchStatefulSetPods = async () => {
    if (!clusterName || !namespace || !statefulsetName) return;

    try {
      const response = await getStatefulSetPods(clusterName, namespace, statefulsetName);
      if (response.data.code === 0) {
        setPods(response.data.data.pods || []);
      } else {
        message.error(response.data.message || t('statefulsets.fetchPodsFailed'));
      }
    } catch (error) {
      console.error(t('statefulsets.fetchPodsFailed'), error);
      message.error(t('statefulsets.fetchPodsFailed'));
    }
  };

  // 获取StatefulSet相关事件
  const fetchStatefulSetEvents = async () => {
    if (!clusterName || !namespace || !statefulsetName) return;

    try {
      const response = await getAllStatefulSetEvents(clusterName, namespace, statefulsetName);
      if (response.data.code === 0) {
        setEvents(response.data.data.events || { statefulset: [], pod: [] });
      } else {
        message.error(response.data.message || t('statefulsets.fetchEventsFailed'));
      }
    } catch (error) {
      console.error(t('statefulsets.fetchEventsFailed'), error);
      message.error(t('statefulsets.fetchEventsFailed'));
    }
  };

  // 初始化加载
  useEffect(() => {
    fetchStatefulSetDetails();
    fetchStatefulSetPods();
    fetchStatefulSetEvents();

    // 定时刷新
    const timer = setInterval(() => {
      fetchStatefulSetDetails();
      fetchStatefulSetPods();
      fetchStatefulSetEvents();
    }, 30000);

    return () => clearInterval(timer);
  }, [clusterName, namespace, statefulsetName]);

  // 返回列表
  const handleBack = () => {
    navigate('/workloads/statefulsets');
  };

  // 刷新数据
  const handleRefresh = () => {
    fetchStatefulSetDetails();
    fetchStatefulSetPods();
    fetchStatefulSetEvents();
    message.success(t('common.refreshSuccess'));
  };

  // 处理扩缩容
  const handleScale = () => {
    setScaleModalVisible(true);
  };

  // 处理重启
  const handleRestart = async () => {
    if (!clusterName || !namespace || !statefulsetName) return;

    try {
      const response = await restartStatefulSet(clusterName, namespace, statefulsetName);
      if (response.data.code === 0) {
        message.success(t('statefulsets.restartSuccess'));
        fetchStatefulSetDetails();
      } else {
        message.error(response.data.message || t('statefulsets.restartFailed'));
      }
    } catch (error) {
      console.error(t('statefulsets.restartFailed'), error);
      message.error(t('statefulsets.restartFailed'));
    }
  };

  // 处理删除
  const handleDelete = async () => {
    if (!clusterName || !namespace || !statefulsetName) return;

    try {
      const response = await deleteStatefulSet(clusterName, namespace, statefulsetName);
      if (response.data.code === 0) {
        message.success(t('statefulsets.deleteSuccess'));
        navigate('/workloads/statefulsets');
      } else {
        message.error(response.data.message || t('statefulsets.deleteFailed'));
      }
    } catch (error) {
      console.error(t('statefulsets.deleteFailed'), error);
      message.error(t('statefulsets.deleteFailed'));
    }
  };

  // 扩缩容成功回调
  const handleScaleSuccess = () => {
    setScaleModalVisible(false);
    fetchStatefulSetDetails();
    message.success(t('statefulsets.scaleSuccess'));
  };

  // 渲染基础信息
  const renderBasicInfo = () => {
    if (!statefulset) return null;

    return (
      <div className="basic-info">
        <Card title={t('common.basicInfo')} bordered={false}>
          <div className="info-item">
            <span className="info-label">{t('common.name')}:</span>
            <span className="info-value">{statefulset.name}</span>
          </div>
          <div className="info-item">
            <span className="info-label">{t('common.namespace')}:</span>
            <span className="info-value">{statefulset.namespace}</span>
          </div>
          <div className="info-item">
            <span className="info-label">{t('statefulsets.replicas')}:</span>
            <span className="info-value">{statefulset.replicas}</span>
          </div>
          <div className="info-item">
            <span className="info-label">{t('statefulsets.readyReplicas')}:</span>
            <span className="info-value">{statefulset.readyReplicas}</span>
          </div>
          <div className="info-item">
            <span className="info-label">{t('statefulsets.serviceName')}:</span>
            <span className="info-value">{statefulset.serviceName}</span>
          </div>
          <div className="info-item">
            <span className="info-label">{t('statefulsets.updateStrategy')}:</span>
            <span className="info-value">{statefulset.updateStrategy}</span>
          </div>
          <div className="info-item">
            <span className="info-label">{t('statefulsets.podManagementPolicy')}:</span>
            <span className="info-value">{statefulset.podManagementPolicy}</span>
          </div>
          <div className="info-item">
            <span className="info-label">{t('common.createTime')}:</span>
            <span className="info-value">
              {new Date(statefulset.creationTime).toLocaleString()}
            </span>
          </div>
        </Card>
      </div>
    );
  };

  // 渲染标签和注解
  const renderLabelsAndAnnotations = () => {
    if (!statefulset) return null;

    return (
      <div className="labels-annotations">
        <Card title={t('common.labelsAndAnnotations')} bordered={false}>
          <h3>{t('common.labels')}</h3>
          <div className="labels-container">
            {statefulset.labels && Object.keys(statefulset.labels).length > 0 ? (
              Object.entries(statefulset.labels).map(([key, value]) => (
                <div key={key} className="label-item">
                  <span className="label-key">{key}</span>
                  <span className="label-value">{String(value)}</span>
                </div>
              ))
            ) : (
              <div className="empty-message">{t('common.noLabels')}</div>
            )}
          </div>

          <h3>{t('common.annotations')}</h3>
          <div className="annotations-container">
            {statefulset.annotations && Object.keys(statefulset.annotations).length > 0 ? (
              Object.entries(statefulset.annotations).map(([key, value]) => (
                <div key={key} className="annotation-item">
                  <span className="annotation-key">{key}</span>
                  <span className="annotation-value">{String(value)}</span>
                </div>
              ))
            ) : (
              <div className="empty-message">{t('common.noAnnotations')}</div>
            )}
          </div>
        </Card>
      </div>
    );
  };

  // 渲染容器信息
  const renderContainers = () => {
    if (!statefulset?.containers?.length) return null;

    return (
      <div className="containers">
        <Card title={t('statefulsets.containers')} bordered={false}>
          {statefulset.containers.map((container) => (
            <div key={container.name} className="container-item">
              <h3>{container.name}</h3>
              <div className="container-info">
                <div className="info-item">
                  <span className="info-label">{t('statefulsets.image')}:</span>
                  <span className="info-value">{container.image}</span>
                </div>

                {/* 资源请求与限制 */}
                {container.resources && (
                  <div className="resources">
                    <h4>{t('statefulsets.resources')}</h4>
                    <div className="resources-info">
                      {container.resources.requests && (
                        <div className="requests">
                          <h5>{t('statefulsets.requests')}</h5>
                          {Object.entries(container.resources.requests).map(([key, value]) => (
                            <div key={key} className="resource-item">
                              <span className="resource-key">{key}:</span>
                              <span className="resource-value">{String(value)}</span>
                            </div>
                          ))}
                        </div>
                      )}

                      {container.resources.limits && (
                        <div className="limits">
                          <h5>{t('statefulsets.limits')}</h5>
                          {Object.entries(container.resources.limits).map(([key, value]) => (
                            <div key={key} className="resource-item">
                              <span className="resource-key">{key}:</span>
                              <span className="resource-value">{String(value)}</span>
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                  </div>
                )}

                {/* 环境变量 */}
                {container.env && container.env.length > 0 && (
                  <div className="environment">
                    <h4>{t('statefulsets.environment')}</h4>
                    <div className="env-list">
                      {container.env.map((env, envIndex) => (
                        <div key={envIndex} className="env-item">
                          <span className="env-key">{env.name}:</span>
                          <span className="env-value">
                            {env.value !== undefined ? env.value : (env.valueFrom ? '(From Source)' : '')}
                          </span>
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {/* 端口 */}
                {container.ports && container.ports.length > 0 && (
                  <div className="ports">
                    <h4>{t('statefulsets.ports')}</h4>
                    <div className="ports-list">
                      {container.ports.map((port, portIndex) => (
                        <div key={portIndex} className="port-item">
                          <span className="port-name">
                            {port.name ? `${port.name}: ` : ''}
                          </span>
                          <span className="port-value">
                            {port.containerPort} {port.protocol}
                          </span>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </div>
          ))}
        </Card>
      </div>
    );
  };

  // 渲染PVC模板
  const renderVolumeClaimTemplates = () => {
    if (!statefulset?.volumeClaimTemplates?.length) return null;

    return (
      <div className="volume-claim-templates">
        <Card title={t('statefulsets.volumeClaimTemplates')} bordered={false}>
          {statefulset.volumeClaimTemplates.map((pvc) => (
            <div key={pvc.name} className="pvc-item">
              <h3>{pvc.name}</h3>
              <div className="pvc-info">
                <div className="info-item">
                  <span className="info-label">{t('statefulsets.storageClassName')}:</span>
                  <span className="info-value">{pvc.storageClassName || '(default)'}</span>
                </div>
                <div className="info-item">
                  <span className="info-label">{t('statefulsets.accessModes')}:</span>
                  <span className="info-value">{pvc.accessModes.join(', ')}</span>
                </div>
                <div className="info-item">
                  <span className="info-label">{t('statefulsets.storage')}:</span>
                  <span className="info-value">{pvc.storage}</span>
                </div>
              </div>
            </div>
          ))}
        </Card>
      </div>
    );
  };

  // 渲染Pod列表
  const renderPods = () => {
    return (
      <div className="pods-list">
        <Card title={t('statefulsets.pods')} bordered={false}>
          {pods.length > 0 ? (
            <div className="pod-items">
              {pods.map((pod) => (
                <div key={pod.metadata.name} className="pod-item">
                  <div className="pod-name">
                    <a onClick={() => navigate(`/workloads/pods/detail/${clusterName}/${pod.metadata.namespace}/${pod.metadata.name}`)}>
                      {pod.metadata.name}
                    </a>
                  </div>
                  <div className="pod-status">
                    <span className={`status-badge ${pod.status.phase.toLowerCase()}`}>
                      {pod.status.phase}
                    </span>
                  </div>
                  <div className="pod-age">
                    <span>{new Date(pod.metadata.creationTimestamp).toLocaleString()}</span>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="empty-message">{t('common.noPods')}</div>
          )}
        </Card>
      </div>
    );
  };

  // 渲染事件
  const renderEvents = () => {
    const allEvents = [...events.statefulset, ...events.pod];
    
    return (
      <div className="events">
        <Card title={t('common.events')} bordered={false}>
          <K8sEvents events={allEvents} />
        </Card>
      </div>
    );
  };

  // 渲染YAML
  const renderYaml = () => {
    if (!statefulset) return null;

    // 实际应用中应当获取完整的YAML
    const yaml = JSON.stringify(statefulset, null, 2);
    
    return (
      <div className="yaml">
        <Card title={t('common.yaml')} bordered={false}>
          <pre className="yaml-content">{yaml}</pre>
        </Card>
      </div>
    );
  };

  if (loading && !statefulset) {
    return (
      <div className="loading-container">
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div className="statefulset-detail-page">
      <Card
        title={
          <div className="page-title">
            <Button icon={<ArrowLeftOutlined />} onClick={handleBack} style={{ marginRight: 16 }}>
              {t('common.back')}
            </Button>
            <span>{t('statefulsets.detail')}: {statefulsetName}</span>
          </div>
        }
        extra={
          <Space>
            <Tooltip title={t('common.refresh')}>
              <Button icon={<SyncOutlined />} onClick={handleRefresh} />
            </Tooltip>
            <Tooltip title={t('common.scale')}>
              <Button 
                icon={<ScissorOutlined />} 
                onClick={handleScale}
                disabled={!statefulset}
              />
            </Tooltip>
            <Tooltip title={t('common.restart')}>
              <Popconfirm
                title={t('statefulsets.confirmRestart')}
                onConfirm={handleRestart}
                okText={t('common.confirm')}
                cancelText={t('common.cancel')}
              >
                <Button icon={<ReloadOutlined />} disabled={!statefulset} />
              </Popconfirm>
            </Tooltip>
            <Tooltip title={t('common.delete')}>
              <Popconfirm
                title={t('statefulsets.confirmDelete')}
                onConfirm={handleDelete}
                okText={t('common.confirm')}
                cancelText={t('common.cancel')}
              >
                <Button icon={<DeleteOutlined />} danger disabled={!statefulset} />
              </Popconfirm>
            </Tooltip>
          </Space>
        }
      >
        <Tabs defaultActiveKey="basic">
          <TabPane tab={t('common.basic')} key="basic">
            {renderBasicInfo()}
            {renderLabelsAndAnnotations()}
            {renderContainers()}
            {renderVolumeClaimTemplates()}
          </TabPane>
          <TabPane tab={t('common.pods')} key="pods">
            {renderPods()}
          </TabPane>
          <TabPane tab={t('common.events')} key="events">
            {renderEvents()}
          </TabPane>
          <TabPane tab="YAML" key="yaml">
            {renderYaml()}
          </TabPane>
        </Tabs>
      </Card>

      {/* 扩缩容模态框 */}
      {statefulset && (
        <ScaleStatefulSetModal
          visible={scaleModalVisible}
          onCancel={() => setScaleModalVisible(false)}
          onSuccess={handleScaleSuccess}
          clusterName={clusterName || ''}
          namespace={namespace || ''}
          statefulsetName={statefulsetName || ''}
          currentReplicas={statefulset.replicas}
        />
      )}
    </div>
  );
};

export default StatefulSetDetailPage;
