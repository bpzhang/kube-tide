import { CreateServiceRequest, UpdateServiceRequest, ServicePort } from '@/api/service';

// 表单数据中端口的类型定义
export interface ServicePortFormData {
  name?: string;
  port: number;
  targetPort: number;
  protocol: 'TCP' | 'UDP' | 'SCTP';
  nodePort?: number;
}

// 键值对类型定义，用于标签、选择器等
export interface KeyValuePair {
  key: string;
  value: string;
}

// 服务表单数据类型定义
export interface ServiceFormData {
  name?: string;
  namespace?: string;
  type: 'ClusterIP' | 'NodePort' | 'LoadBalancer';
  ports: ServicePortFormData[];
  labelsArray: KeyValuePair[];
  selectorArray: KeyValuePair[];
}

// 服务表单组件属性类型定义
export interface ServiceFormProps {
  initialValues: ServiceFormData;
  onFormValuesChange?: (changedValues: any, allValues: ServiceFormData) => void;
  form: any; // Form实例
  mode: 'create' | 'edit';
}

// 将表单数据转换为创建服务的API请求数据
export const processFormToCreateRequest = (values: ServiceFormData): CreateServiceRequest => {
  // 构建创建Service的请求
  const serviceData: CreateServiceRequest = {
    name: values.name || '',
    namespace: values.namespace || 'default',
    type: values.type,
    ports: values.ports.map(port => ({
      name: port.name,
      port: port.port,
      targetPort: port.targetPort,
      protocol: port.protocol,
      nodePort: port.nodePort
    })),
    selector: {},
    labels: {},
    annotations: {}
  };

  // 处理选择器
  (values.selectorArray || []).forEach((item: KeyValuePair) => {
    if (item?.key) {
      serviceData.selector![item.key] = item.value || '';
    }
  });

  // 处理标签
  (values.labelsArray || []).forEach((item: KeyValuePair) => {
    if (item?.key) {
      serviceData.labels![item.key] = item.value || '';
    }
  });

  return serviceData;
};

// 将表单数据转换为更新服务的API请求数据
export const processFormToUpdateRequest = (values: ServiceFormData): UpdateServiceRequest => {
  const updateData: UpdateServiceRequest = {
    type: values.type,
    ports: values.ports.map(port => ({
      name: port.name,
      port: port.port,
      targetPort: port.targetPort,
      protocol: port.protocol,
      nodePort: port.nodePort
    })),
    selector: {},
    labels: {}
  };

  // 处理选择器
  (values.selectorArray || []).forEach((item: KeyValuePair) => {
    if (item?.key) {
      updateData.selector![item.key] = item.value || '';
    }
  });

  // 处理标签
  (values.labelsArray || []).forEach((item: KeyValuePair) => {
    if (item?.key) {
      updateData.labels![item.key] = item.value || '';
    }
  });

  return updateData;
};

// 将后端服务数据转换为表单数据格式
export const processServiceToFormData = (service: any): ServiceFormData => {
  // 转换标签和选择器为数组格式，方便表单处理
  const labelsArray = Object.entries(service?.metadata?.labels || {}).map(([key, value]) => ({ key, value: String(value) }));
  const selectorArray = Object.entries(service?.spec?.selector || {}).map(([key, value]) => ({ key, value: String(value) }));
  
  // 格式化端口信息
  const formattedPorts = (service?.spec?.ports || []).map((port: any) => ({
    name: port.name || '',
    port: port.port,
    targetPort: typeof port.targetPort === 'object' ? port.targetPort.intVal : port.targetPort,
    protocol: port.protocol || 'TCP',
    nodePort: port.nodePort
  }));

  return {
    name: service?.metadata?.name,
    namespace: service?.metadata?.namespace,
    type: service?.spec?.type || 'ClusterIP',
    ports: formattedPorts,
    labelsArray,
    selectorArray
  };
};