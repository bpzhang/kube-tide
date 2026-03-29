import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { Descriptions, Tag, Table, Card, Space, Typography, Button, message, Tabs, Empty, List } from 'antd';
import { EditOutlined } from '@ant-design/icons';
import { formatDate } from '@/utils/format';
import EditDeploymentModal from './EditDeploymentModal';
import PodList from '../pod/PodList';
import { updateDeployment, UpdateDeploymentRequest } from '@/api/deployment';
import { getIngressesByNamespace } from '@/api/ingress';
import { getPodsBySelector } from '@/api/pod';
import { getServicesByNamespace, getServiceEndpoints } from '@/api/service';
import DeploymentEvents from './DeploymentEvents';
import { useTranslation } from 'react-i18next';

const { Title } = Typography;

interface Container {
  name: string;
  image: string;
  ports: any[];
  env: any[];
  resources: {
    limits?: { cpu?: string; memory?: string };
    requests?: { cpu?: string; memory?: string };
  };
}

interface ServicePort {
  name?: string;
  port?: number;
  targetPort?: number | string;
  protocol?: string;
  nodePort?: number;
}

interface RelatedService {
  name: string;
  namespace: string;
  type: string;
  clusterIP: string;
  externalIPs: string[];
  selector: Record<string, string>;
  ports: ServicePort[];
}

interface AccessEntry {
  key: string;
  label: string;
  value: string;
}

interface RelatedRoute {
  name: string;
  namespace: string;
  ingressClassName?: string;
  host: string;
  path: string;
  pathType?: string;
  serviceName: string;
  servicePort?: string;
  tlsHosts: string[];
  tlsSecretName?: string;
}

interface ServiceEndpointSubset {
  addresses: Array<{
    ip: string;
    nodeName?: string;
    target?: string;
  }>;
  ports: Array<{
    name?: string;
    port: number;
    protocol?: string;
  }>;
}

interface DeploymentDetailProps {
  deployment: {
    name: string;
    namespace: string;
    replicas: number;
    readyReplicas: number;
    strategy: string;
    creationTime: string;
    labels: { [key: string]: string };
    selector: { [key: string]: string };
    annotations: { [key: string]: string };
    containers: Container[];
    conditions: Array<{
      type: string;
      status: string;
      lastUpdateTime: string;
      reason: string;
      message: string;
    }>;
    minReadySeconds?: number;
    revisionHistoryLimit?: number;
    paused?: boolean;
  };
  clusterName: string;
  onUpdate?: () => void;
}

const DeploymentDetail: React.FC<DeploymentDetailProps> = ({ 
  deployment, 
  clusterName, 
  onUpdate 
}) => {
  const { t } = useTranslation();
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [pods, setPods] = useState<any[]>([]);
  const [podsLoading, setPodsLoading] = useState(false);
  const [servicesLoading, setServicesLoading] = useState(false);
  const [relatedServices, setRelatedServices] = useState<RelatedService[]>([]);
  const [serviceEndpoints, setServiceEndpoints] = useState<Record<string, ServiceEndpointSubset[]>>({});
  const [routesLoading, setRoutesLoading] = useState(false);
  const [relatedRoutes, setRelatedRoutes] = useState<RelatedRoute[]>([]);
  const [serviceEndpointsSupported, setServiceEndpointsSupported] = useState(true);
  const [routesSupported, setRoutesSupported] = useState(true);

  // get Pods by selector
  const fetchPods = async () => {
    if (!deployment.selector) return;

    setPodsLoading(true);
    try {
      const response = await getPodsBySelector(
        clusterName,
        deployment.namespace,
        deployment.selector
      );
      if (response.data.code === 0) {
        // set Pods data
        setPods(response.data.data.pods);
        
        // extract health check probe data from Pods
        extractProbesFromPods(response.data.data.pods);
      } else {
        message.error(response.data.message || t('pods.fetchFailed'));
      }
    } catch (error) {
      console.error(t('pods.fetchFailed') + ':', error);
      message.error(t('pods.fetchFailed'));
    } finally {
      setPodsLoading(false);
    }
  };

  const isSelectorMatched = (selector: Record<string, string>, labels: Record<string, string>) => {
    const selectorEntries = Object.entries(selector || {});
    if (selectorEntries.length === 0) {
      return false;
    }

    return selectorEntries.every(([key, value]) => labels?.[key] === value);
  };

  const normalizeService = (service: any): RelatedService => {
    const metadata = service?.metadata || {};
    const spec = service?.spec || service || {};

    return {
      name: metadata.name || service?.name || '-',
      namespace: metadata.namespace || service?.namespace || deployment.namespace,
      type: spec.type || service?.type || '-',
      clusterIP: spec.clusterIP || service?.clusterIP || '-',
      externalIPs: spec.externalIPs || service?.externalIPs || [],
      selector: spec.selector || service?.selector || {},
      ports: spec.ports || service?.ports || [],
    };
  };

  const fetchRelatedServices = async () => {
    if (!clusterName || !deployment.namespace) return;

    setServicesLoading(true);
    try {
      const response = await getServicesByNamespace(clusterName, deployment.namespace);
      if (response.data.code !== 0) {
        message.error(response.data.message || t('services.fetchFailed'));
        setRelatedServices([]);
        return;
      }

      const serviceList = (response.data.data.services || []).map(normalizeService);
      const podLabelsList = pods
        .map((pod) => pod?.metadata?.labels || {})
        .filter((labels) => Object.keys(labels).length > 0);

      const matchedServices = serviceList.filter((service) => {
        const selector = service.selector || {};
        if (Object.keys(selector).length === 0) {
          return false;
        }

        if (podLabelsList.length > 0) {
          return podLabelsList.some((labels) => isSelectorMatched(selector, labels));
        }

        return isSelectorMatched(selector, deployment.selector || {});
      });

      setRelatedServices(matchedServices);

      if (!serviceEndpointsSupported) {
        setServiceEndpoints({});
        return;
      }

      const endpointMap: Record<string, ServiceEndpointSubset[]> = {};
      for (const service of matchedServices) {
        try {
          const endpointResponse = await getServiceEndpoints(clusterName, service.namespace, service.name);
          endpointMap[`${service.namespace}/${service.name}`] =
            endpointResponse.data.code === 0 ? endpointResponse.data.data.endpoints || [] : [];
        } catch (error) {
          if (axios.isAxiosError(error) && error.response?.status === 404) {
            setServiceEndpointsSupported(false);
            setServiceEndpoints({});
            return;
          }

          console.error(`${t('deployments.detail.fetchServiceEndpointsFailed')}:`, error);
          endpointMap[`${service.namespace}/${service.name}`] = [];
        }
      }

      setServiceEndpoints(endpointMap);
    } catch (error) {
      console.error(t('services.fetchFailed') + ':', error);
      message.error(t('services.fetchFailed'));
      setRelatedServices([]);
      setServiceEndpoints({});
    } finally {
      setServicesLoading(false);
    }
  };

  const fetchRelatedRoutes = async (services: RelatedService[]) => {
    if (!clusterName || !deployment.namespace) return;

    if (!routesSupported) {
      setRelatedRoutes([]);
      return;
    }

    setRoutesLoading(true);
    try {
      const response = await getIngressesByNamespace(clusterName, deployment.namespace);
      if (response.data.code !== 0) {
        message.error(response.data.message || t('deployments.detail.fetchRoutesFailed'));
        setRelatedRoutes([]);
        return;
      }

      const serviceNames = new Set(services.map((service) => service.name));
      const routes: RelatedRoute[] = [];

      (response.data.data.ingresses || []).forEach((ingress) => {
        const tlsHosts = (ingress.tls || []).flatMap((item) => item.hosts || []);
        const tlsSecretName = ingress.tls?.find((item) => item.secretName)?.secretName;

        (ingress.rules || []).forEach((rule) => {
          (rule.paths || []).forEach((path) => {
            const serviceName = path.backend?.serviceName || '';
            if (!serviceNames.has(serviceName)) {
              return;
            }

            routes.push({
              name: ingress.name,
              namespace: ingress.namespace,
              ingressClassName: ingress.ingressClassName,
              host: rule.host || '*',
              path: path.path || '/',
              pathType: path.pathType,
              serviceName,
              servicePort: path.backend?.servicePort,
              tlsHosts,
              tlsSecretName,
            });
          });
        });
      });

      setRelatedRoutes(routes);
    } catch (error) {
      if (axios.isAxiosError(error) && error.response?.status === 404) {
        setRoutesSupported(false);
        setRelatedRoutes([]);
        return;
      }

      console.error(t('deployments.detail.fetchRoutesFailed') + ':', error);
      message.error(t('deployments.detail.fetchRoutesFailed'));
      setRelatedRoutes([]);
    } finally {
      setRoutesLoading(false);
    }
  };

  // extract health check probe data from Pods
  const extractProbesFromPods = (pods: any[]) => {
    console.log(t('deployments.detail.extractingProbes'), pods.length);
    
    if (!pods || pods.length === 0) return;
    
    // create a mapping from container names to containers
    const containerMap: { [key: string]: any } = {};
    deployment.containers.forEach(container => {
      containerMap[container.name] = container;
    });
    
    // for each Pod, check its containers
    pods.forEach(pod => {
      console.log(t('deployments.detail.processingPod'), pod.metadata?.name);
      const containers = pod.spec?.containers || [];
      
      // for each container in the Pod
      containers.forEach((podContainer: any) => {
        const containerName = podContainer.name;
        // check if this container belongs to the current Deployment
        if (containerMap[containerName]) {
          console.log(t('deployments.detail.foundContainer', { name: containerName }));
          
          // update health check probes
          ['livenessProbe', 'readinessProbe', 'startupProbe'].forEach(probeType => {
            if (podContainer[probeType]) {
              console.log(t('deployments.detail.foundProbe', { container: containerName, type: probeType }));
              // add probe data to the deployment's container
              containerMap[containerName][probeType] = podContainer[probeType];
            }
          });
        }
      });
    });
    
    console.log(t('deployments.detail.processedContainers'), deployment.containers);
  };

  useEffect(() => {
    fetchPods();
    // refresh Pods every 30 seconds
    const timer = setInterval(fetchPods, 30000);
    return () => clearInterval(timer);
  }, [deployment.selector, deployment.namespace, clusterName]);

  useEffect(() => {
    fetchRelatedServices();
  }, [clusterName, deployment.namespace, deployment.selector, pods]);

  useEffect(() => {
    fetchRelatedRoutes(relatedServices);
  }, [clusterName, deployment.namespace, relatedServices]);

  // handle showing and hiding the edit modal
  const showEditModal = () => {
    setEditModalVisible(true);
  };

  const hideEditModal = () => {
    setEditModalVisible(false);
  };

  // handle update submission
  const handleUpdateSubmit = async (updateData: UpdateDeploymentRequest) => {
    try {
      await updateDeployment(
        clusterName,
        deployment.namespace,
        deployment.name,
        updateData
      );
      
      // If there is an update callback, call it
      if (onUpdate) {
        onUpdate();
      }
      
      return Promise.resolve();
    } catch (error) {
      console.error(t('deployments.editFailed') + ':', error);
      message.error(t('deployments.editFailed'));
      return Promise.reject(error);
    }
  };

  // container column definition
  const containerColumns = [
    {
      title: t('deployments.detail.containerColumns.name'),
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: t('deployments.detail.containerColumns.image'),
      dataIndex: 'image',
      key: 'image',
    },
    {
      title: t('deployments.detail.containerColumns.ports'),
      key: 'ports',
      render: (record: Container) => (
        <>
          {record.ports?.map((port, index) => (
            <Tag key={index}>
              {port.containerPort}/{port.protocol}
            </Tag>
          ))}
        </>
      ),
    },
  ];

  // condition column definition
  const conditionColumns = [
    {
      title: t('deployments.detail.conditionColumns.type'),
      dataIndex: 'type',
      key: 'type',
    },
    {
      title: t('deployments.detail.conditionColumns.status'),
      dataIndex: 'status',
      key: 'status',
      render: (text: string) => (
        <Tag color={text === 'True' ? 'success' : 'error'}>{text}</Tag>
      ),
    },
    {
      title: t('deployments.detail.conditionColumns.lastUpdateTime'),
      dataIndex: 'lastUpdateTime',
      key: 'lastUpdateTime',
      render: (text: string) => formatDate(text),
    },
    {
      title: t('deployments.detail.conditionColumns.reason'),
      dataIndex: 'reason',
      key: 'reason',
    },
    {
      title: t('deployments.detail.conditionColumns.message'),
      dataIndex: 'message',
      key: 'message',
    },
  ];

  const serviceColumns = [
    {
      title: t('services.name'),
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: t('services.type'),
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => <Tag color="blue">{type || '-'}</Tag>,
    },
    {
      title: t('services.clusterIp'),
      dataIndex: 'clusterIP',
      key: 'clusterIP',
      render: (value: string) => value || '-',
    },
    {
      title: t('services.externalIp'),
      dataIndex: 'externalIPs',
      key: 'externalIPs',
      render: (externalIPs: string[]) => (externalIPs?.length ? externalIPs.join(', ') : '-'),
    },
    {
      title: t('services.ports'),
      dataIndex: 'ports',
      key: 'ports',
      render: (ports: ServicePort[]) => (
        <Space direction="vertical" size={0}>
          {(ports || []).map((port, index) => (
            <span key={`${port.name || port.port}-${index}`}>
              {port.port ?? '-'} → {port.targetPort ?? '-'}
              {port.protocol ? ` (${port.protocol})` : ''}
              {port.nodePort ? ` / NodePort: ${port.nodePort}` : ''}
            </span>
          ))}
        </Space>
      ),
    },
    {
      title: t('services.selectors'),
      dataIndex: 'selector',
      key: 'selector',
      render: (selector: Record<string, string>) => (
        Object.keys(selector || {}).length > 0 ? (
          <Space wrap>
            {Object.entries(selector || {}).map(([key, value]) => (
              <Tag key={`${key}-${value}`}>{`${key}: ${value}`}</Tag>
            ))}
          </Space>
        ) : (
          <Tag color="red">{t('services.noSelectors')}</Tag>
        )
      ),
    },
    {
      title: t('services.endpoints'),
      key: 'endpoints',
      render: (_: unknown, record: RelatedService) => {
        const endpoints = serviceEndpoints[`${record.namespace}/${record.name}`] || [];
        const values = endpoints.flatMap((subset) =>
          subset.addresses.flatMap((address) =>
            (subset.ports.length ? subset.ports : [{ port: 0, protocol: '' }]).map((port) => ({
              label: `${address.ip}${port.port ? `:${port.port}` : ''}`,
              target: address.target,
            }))
          )
        );

        if (values.length === 0) {
          return <Tag>{t('deployments.detail.noEndpoints')}</Tag>;
        }

        return (
          <Space wrap>
            {values.map((item, index) => (
              <Tag key={`${item.label}-${index}`} color="geekblue">
                {item.target ? `${item.target} · ` : ''}{item.label}
              </Tag>
            ))}
          </Space>
        );
      },
    },
  ];

  const buildAccessEntries = (service: RelatedService): AccessEntry[] => {
    const entries: AccessEntry[] = [];
    const serviceKey = `${service.namespace}/${service.name}`;
    const ports = service.ports || [];
    const endpoints = serviceEndpoints[serviceKey] || [];

    if (service.clusterIP && service.clusterIP !== 'None') {
      ports.forEach((port, index) => {
        entries.push({
          key: `clusterip-${index}`,
          label: t('deployments.detail.access.clusterIP'),
          value: `${service.clusterIP}:${port.port ?? '-'}`,
        });
      });
    }

    (service.externalIPs || []).forEach((ip, ipIndex) => {
      if (ports.length > 0) {
        ports.forEach((port, portIndex) => {
          entries.push({
            key: `external-${ipIndex}-${portIndex}`,
            label: t('deployments.detail.access.externalIP'),
            value: `${ip}:${port.port ?? '-'}`,
          });
        });
      } else {
        entries.push({
          key: `external-${ipIndex}`,
          label: t('deployments.detail.access.externalIP'),
          value: ip,
        });
      }
    });

    ports.forEach((port, index) => {
      if (port.nodePort) {
        entries.push({
          key: `nodeport-${index}`,
          label: t('deployments.detail.access.nodePort'),
          value: `any-node-ip:${port.nodePort}`,
        });
      }
    });

    endpoints.forEach((subset, subsetIndex) => {
      subset.addresses.forEach((address, addressIndex) => {
        const subsetPorts = subset.ports.length > 0 ? subset.ports : [{ port: 0, protocol: '' }];
        subsetPorts.forEach((port, portIndex) => {
          entries.push({
            key: `endpoint-${subsetIndex}-${addressIndex}-${portIndex}`,
            label: t('deployments.detail.access.endpoint'),
            value: `${address.target ? `${address.target} · ` : ''}${address.ip}${port.port ? `:${port.port}` : ''}`,
          });
        });
      });
    });

    return entries;
  };

  const routeColumns = [
    {
      title: t('deployments.detail.routeColumns.route'),
      key: 'route',
      render: (record: RelatedRoute) => (
        <Space direction="vertical" size={0}>
          <span>{record.name}</span>
          {record.ingressClassName ? (
            <Typography.Text type="secondary">{record.ingressClassName}</Typography.Text>
          ) : null}
        </Space>
      ),
    },
    {
      title: t('deployments.detail.routeColumns.hostPath'),
      key: 'hostPath',
      render: (record: RelatedRoute) => (
        <Space direction="vertical" size={0}>
          <Typography.Text>{record.host}</Typography.Text>
          <Typography.Text code>{record.path}</Typography.Text>
        </Space>
      ),
    },
    {
      title: t('deployments.detail.routeColumns.backend'),
      key: 'backend',
      render: (record: RelatedRoute) => (
        <Typography.Text>{record.serviceName}{record.servicePort ? `:${record.servicePort}` : ''}</Typography.Text>
      ),
    },
    {
      title: t('deployments.detail.routeColumns.tls'),
      key: 'tls',
      render: (record: RelatedRoute) => (
        record.tlsHosts.length > 0 ? (
          <Space direction="vertical" size={0}>
            <Space wrap>
              {record.tlsHosts.map((host) => (
                <Tag key={host} color="green">{host}</Tag>
              ))}
            </Space>
            {record.tlsSecretName ? <Typography.Text type="secondary">{record.tlsSecretName}</Typography.Text> : null}
          </Space>
        ) : (
          <Tag>{t('deployments.detail.noTls')}</Tag>
        )
      ),
    },
    {
      title: t('deployments.detail.routeColumns.access'),
      key: 'access',
      render: (record: RelatedRoute) => {
        const protocol = record.tlsHosts.includes(record.host) ? 'https' : 'http';
        return <Typography.Text code>{`${protocol}://${record.host}${record.path}`}</Typography.Text>;
      },
    },
  ];

  const tabItems = [
    {
      key: 'overview',
      label: t('deployments.detail.tabs.overview'),
      children: (
        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
          <Card title={t('deployments.detail.basicInfo.title')}>
            <Descriptions column={2}>
              <Descriptions.Item label={t('deployments.namespace')}>{deployment.namespace}</Descriptions.Item>
              <Descriptions.Item label={t('deployments.createdAt')}>{formatDate(deployment.creationTime)}</Descriptions.Item>
              <Descriptions.Item label={t('deployments.detail.basicInfo.replicas')}>
                {deployment.readyReplicas || 0}/{deployment.replicas}
              </Descriptions.Item>
              <Descriptions.Item label={t('deployments.detail.basicInfo.strategy')}>{deployment.strategy}</Descriptions.Item>
            </Descriptions>
          </Card>

          <Card title={t('deployments.labels')}>
            {Object.keys(deployment.labels || {}).length > 0 ? (
              <Space wrap>
                {Object.entries(deployment.labels || {}).map(([key, value], index) => (
                  <Tag key={`label-${key}-${index}`}>{`${key}: ${value}`}</Tag>
                ))}
              </Space>
            ) : (
              <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={t('common.noDataFound')} />
            )}
          </Card>

          <Card title={t('deployments.detail.selector')}>
            {Object.keys(deployment.selector || {}).length > 0 ? (
              <Space wrap>
                {Object.entries(deployment.selector || {}).map(([key, value], index) => (
                  <Tag key={`selector-${key}-${index}`}>{`${key}: ${value}`}</Tag>
                ))}
              </Space>
            ) : (
              <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={t('common.noDataFound')} />
            )}
          </Card>
        </Space>
      ),
    },
    {
      key: 'containers',
      label: t('deployments.detail.tabs.containers'),
      children: (
        <Card title={t('deployments.detail.containers')}>
          <Table
            columns={containerColumns}
            dataSource={deployment.containers}
            rowKey="name"
            loading={podsLoading}
            pagination={false}
            scroll={{ x: 'max-content' }}
          />
        </Card>
      ),
    },
    {
      key: 'conditions',
      label: t('deployments.detail.tabs.conditions'),
      children: (
        <Card title={t('deployments.detail.conditions')}>
          <Table
            columns={conditionColumns}
            dataSource={deployment.conditions}
            rowKey="type"
            pagination={false}
            scroll={{ x: 'max-content' }}
          />
        </Card>
      ),
    },
    {
      key: 'pods',
      label: t('deployments.detail.tabs.pods'),
      children: (
        <Card title={t('deployments.detail.pods')} loading={podsLoading}>
          <PodList
            clusterName={clusterName}
            namespace={deployment.namespace}
            pods={pods}
            onRefresh={fetchPods}
          />
        </Card>
      ),
    },
    {
      key: 'access',
      label: t('deployments.detail.tabs.access'),
      children: (
        <Tabs
          defaultActiveKey="service"
          items={[
            {
              key: 'service',
              label: t('deployments.detail.access.tabs.service'),
              children: relatedServices.length > 0 ? (
                <Space direction="vertical" size="middle" style={{ width: '100%' }}>
                  {relatedServices.map((service) => {
                    const accessEntries = buildAccessEntries(service);
                    return (
                      <Card
                        key={`${service.namespace}/${service.name}`}
                        title={`${service.name} (${service.type})`}
                        size="small"
                      >
                        {accessEntries.length > 0 ? (
                          <List
                            dataSource={accessEntries}
                            renderItem={(item) => (
                              <List.Item key={item.key}>
                                <Space direction="vertical" size={2} style={{ width: '100%' }}>
                                  <Typography.Text type="secondary">{item.label}</Typography.Text>
                                  <Typography.Text code>{item.value}</Typography.Text>
                                </Space>
                              </List.Item>
                            )}
                          />
                        ) : (
                          <Empty
                            image={Empty.PRESENTED_IMAGE_SIMPLE}
                            description={t('deployments.detail.noAccessEntries')}
                          />
                        )}
                      </Card>
                    );
                  })}

                  <Card title={t('deployments.detail.relatedServices')} loading={servicesLoading}>
                    <Table
                      columns={serviceColumns}
                      dataSource={relatedServices}
                      rowKey={(record) => `${record.namespace}/${record.name}`}
                      pagination={false}
                      scroll={{ x: 'max-content' }}
                    />
                  </Card>
                </Space>
              ) : (
                <Card title={t('deployments.detail.access.title')}>
                  <Empty
                    image={Empty.PRESENTED_IMAGE_SIMPLE}
                    description={t('deployments.detail.noRelatedServices')}
                  />
                </Card>
              ),
            },
            {
              key: 'route',
              label: t('deployments.detail.access.tabs.route'),
              children: (
                <Card title={t('deployments.detail.access.routeTitle')} loading={routesLoading}>
                  {relatedRoutes.length > 0 ? (
                    <Table
                      columns={routeColumns}
                      dataSource={relatedRoutes}
                      rowKey={(record) => `${record.namespace}/${record.name}/${record.host}/${record.path}/${record.serviceName}`}
                      pagination={false}
                      scroll={{ x: 'max-content' }}
                    />
                  ) : (
                    <Empty
                      image={Empty.PRESENTED_IMAGE_SIMPLE}
                      description={t('deployments.detail.noRoutes')}
                    />
                  )}
                </Card>
              ),
            },
          ]}
        />
      ),
    },
    {
      key: 'events',
      label: t('deployments.detail.tabs.events'),
      children: (
        <DeploymentEvents
          clusterName={clusterName}
          namespace={deployment.namespace}
          deploymentName={deployment.name}
        />
      ),
    },
  ];

  return (
    <Space direction="vertical" size="middle" style={{ width: '100%' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Title level={4}>{deployment.name}</Title>
        <Button 
          type="primary" 
          icon={<EditOutlined />} 
          onClick={showEditModal}
        >
          {t('common.edit')}
        </Button>
      </div>

      <Tabs defaultActiveKey="overview" items={tabItems} destroyOnHidden />

      <EditDeploymentModal
        visible={editModalVisible}
        onClose={hideEditModal}
        onSubmit={handleUpdateSubmit}
        deployment={deployment}
      />
    </Space>
  );
};

export default DeploymentDetail;