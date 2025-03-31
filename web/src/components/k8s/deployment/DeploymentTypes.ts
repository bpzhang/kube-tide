import { CreateDeploymentRequest, UpdateDeploymentRequest } from '@/api/deployment';

// 容器类型定义
export interface ContainerFormData {
  name: string;
  image: string;
  resources?: {
    limits?: {
      cpu?: string;
      memory?: string;
      cpuValue?: number;
      cpuUnit?: string;
      memoryValue?: number;
      memoryUnit?: string;
    };
    requests?: {
      cpu?: string;
      memory?: string;
      cpuValue?: number;
      cpuUnit?: string;
      memoryValue?: number;
      memoryUnit?: string;
    };
  };
  env?: Array<{ name: string; value: string }>;
  ports?: Array<{
    name?: string;
    containerPort: number;
    protocol?: string;
  }>;
  volumeMounts?: Array<{
    name: string;
    mountPath: string;
    subPath?: string;
    readOnly?: boolean;
  }>;
  livenessProbe?: {
    type?: string;
    httpGet?: {
      path: string;
      port: number;
      scheme: string;
    };
    tcpSocket?: {
      port: number;
    };
    exec?: {
      command: string | string[];
    };
    initialDelaySeconds?: number;
    timeoutSeconds?: number;
    periodSeconds?: number;
    successThreshold?: number;
    failureThreshold?: number;
  };
  readinessProbe?: {
    type?: string;
    httpGet?: {
      path: string;
      port: number;
      scheme: string;
    };
    tcpSocket?: {
      port: number;
    };
    exec?: {
      command: string | string[];
    };
    initialDelaySeconds?: number;
    timeoutSeconds?: number;
    periodSeconds?: number;
    successThreshold?: number;
    failureThreshold?: number;
  };
  startupProbe?: {
    type?: string;
    httpGet?: {
      path: string;
      port: number;
      scheme: string;
    };
    tcpSocket?: {
      port: number;
    };
    exec?: {
      command: string | string[];
    };
    initialDelaySeconds?: number;
    timeoutSeconds?: number;
    periodSeconds?: number;
    successThreshold?: number;
    failureThreshold?: number;
  };
  [key: string]: any;
}

// 存储卷类型定义
export interface VolumeFormData {
  name: string;
  type: 'configMap' | 'secret' | 'persistentVolumeClaim' | 'emptyDir' | 'hostPath';
  configMap?: {
    name: string;
    items?: Array<{
      key: string;
      path: string;
      mode?: string;
    }>;
  };
  secret?: {
    secretName: string;
    items?: Array<{
      key: string;
      path: string;
      mode?: string;
    }>;
  };
  persistentVolumeClaim?: {
    claimName: string;
    readOnly?: boolean;
  };
  emptyDir?: {
    medium?: string;
    sizeLimit?: string;
  };
  hostPath?: {
    path: string;
    type?: string;
  };
}

// 节点亲和性匹配表达式
export interface NodeAffinityExpression {
  key: string;
  operator: 'In' | 'NotIn' | 'Exists' | 'DoesNotExist' | 'Gt' | 'Lt';
  values?: string[];
}

// 节点亲和性选择器
export interface NodeSelectorTerm {
  matchExpressions?: NodeAffinityExpression[];
  matchFields?: NodeAffinityExpression[];
}

// 首选节点亲和性规则
export interface PreferredNodeAffinity {
  weight: number;
  preference: NodeSelectorTerm;
}

// 节点亲和性配置
export interface NodeAffinityConfig {
  // 必须满足的节点亲和性规则
  requiredTerms: NodeSelectorTerm[];
  // 优先满足的节点亲和性规则
  preferredTerms: PreferredNodeAffinity[];
}

// 表单数据类型定义
export interface DeploymentFormData {
  name?: string;
  replicas: number;
  strategy: 'RollingUpdate' | 'Recreate';
  maxSurgeValue?: number;
  maxUnavailableValue?: number;
  minReadySeconds?: number;
  revisionHistoryLimit?: number;
  paused?: boolean;
  containers: ContainerFormData[];
  volumes?: VolumeFormData[];
  labels?: Array<{ key: string; value: string }>;
  annotations?: Array<{ key: string; value: string }>;
  nodeSelector?: Array<{ key: string; value: string }>;
  nodeAffinity?: NodeAffinityConfig;
  serviceAccountName?: string;
  hostNetwork?: boolean;
  dnsPolicy?: string;
}

// 表单属性类型定义
export interface DeploymentFormProps {
  initialValues: DeploymentFormData;
  onFormValuesChange?: (changedValues: any, allValues: DeploymentFormData) => void;
  form: any; // Form实例
  mode: 'create' | 'edit';
}

// 处理表单数据为API请求格式
export const processFormToCreateRequest = (values: DeploymentFormData): CreateDeploymentRequest => {
  // 构建创建Deployment的请求
  const deploymentData: CreateDeploymentRequest = {
    name: values.name || '',
    replicas: values.replicas,
    containers: []
  };
  
  // 处理基本信息
  if (values.minReadySeconds !== undefined) {
    deploymentData.minReadySeconds = values.minReadySeconds;
  }
  
  if (values.revisionHistoryLimit !== undefined) {
    deploymentData.revisionHistoryLimit = values.revisionHistoryLimit;
  }
  
  if (values.paused !== undefined) {
    deploymentData.paused = values.paused;
  }
  
  // 处理部署策略
  if (values.strategy) {
    deploymentData.strategy = {
      type: values.strategy,
      rollingUpdate: values.strategy === 'RollingUpdate' ? {
        maxSurge: values.maxSurgeValue ? `${values.maxSurgeValue}%` : '25%',
        maxUnavailable: values.maxUnavailableValue ? `${values.maxUnavailableValue}%` : '25%'
      } : undefined
    };
  }
  
  // 处理标签
  if (values.labels && values.labels.length > 0) {
    deploymentData.labels = {};
    values.labels.forEach((item: { key: string; value: string }) => {
      if (item.key && item.value) {
        deploymentData.labels![item.key] = item.value;
      }
    });
  }
  
  // 处理注解
  if (values.annotations && values.annotations.length > 0) {
    deploymentData.annotations = {};
    values.annotations.forEach((item: { key: string; value: string }) => {
      if (item.key && item.value) {
        deploymentData.annotations![item.key] = item.value;
      }
    });
  }
  
  // 处理容器信息
  if (values.containers && values.containers.length > 0) {
    deploymentData.containers = values.containers.map((container: any) => {
      const containerSpec: any = {
        name: container.name,
        image: container.image,
      };
      
      // 处理资源限制
      if (container.resources) {
        containerSpec.resources = {
          limits: {},
          requests: {}
        };
        
        // 处理CPU资源
        if (container.resources.requests?.cpuValue) {
          const cpuValue = container.resources.requests.cpuValue;
          const cpuUnit = container.resources.requests.cpuUnit || 'm';
          containerSpec.resources.requests.cpu = cpuUnit === 'm' ? `${cpuValue}m` : `${cpuValue}`;
        }
        
        if (container.resources.limits?.cpuValue) {
          const cpuValue = container.resources.limits.cpuValue;
          const cpuUnit = container.resources.limits.cpuUnit || 'm';
          containerSpec.resources.limits.cpu = cpuUnit === 'm' ? `${cpuValue}m` : `${cpuValue}`;
        }
        
        // 处理内存资源
        if (container.resources.requests?.memoryValue) {
          const memoryValue = container.resources.requests.memoryValue;
          const memoryUnit = container.resources.requests.memoryUnit || 'Mi';
          containerSpec.resources.requests.memory = `${memoryValue}${memoryUnit}`;
        }
        
        if (container.resources.limits?.memoryValue) {
          const memoryValue = container.resources.limits.memoryValue;
          const memoryUnit = container.resources.limits.memoryUnit || 'Mi';
          containerSpec.resources.limits.memory = `${memoryValue}${memoryUnit}`;
        }
        
        // 如果limits或requests为空对象，则移除
        if (Object.keys(containerSpec.resources.limits).length === 0) {
          delete containerSpec.resources.limits;
        }
        
        if (Object.keys(containerSpec.resources.requests).length === 0) {
          delete containerSpec.resources.requests;
        }
        
        // 如果整个resources为空对象，则移除
        if (Object.keys(containerSpec.resources).length === 0) {
          delete containerSpec.resources;
        }
      }
      
      // 处理环境变量
      if (container.env && container.env.length > 0) {
        containerSpec.env = container.env.map((env: any) => ({
          name: env.name,
          value: env.value
        }));
      }
      
      // 处理端口映射
      if (container.ports && container.ports.length > 0) {
        containerSpec.ports = container.ports.map((port: any) => ({
          name: port.name,
          containerPort: port.containerPort,
          protocol: port.protocol || 'TCP'
        }));
      }
      
      // 处理健康检查探针
      ['livenessProbe', 'readinessProbe', 'startupProbe'].forEach(probeType => {
        if (container[probeType]?.type) {
          containerSpec[probeType] = {
            initialDelaySeconds: container[probeType].initialDelaySeconds,
            timeoutSeconds: container[probeType].timeoutSeconds,
            periodSeconds: container[probeType].periodSeconds,
            successThreshold: container[probeType].successThreshold,
            failureThreshold: container[probeType].failureThreshold
          };
          
          switch (container[probeType].type) {
            case 'httpGet':
              containerSpec[probeType].httpGet = {
                path: container[probeType].httpGet.path,
                port: container[probeType].httpGet.port,
                scheme: container[probeType].httpGet.scheme || 'HTTP'
              };
              break;
            case 'tcpSocket':
              containerSpec[probeType].tcpSocket = {
                port: container[probeType].tcpSocket.port
              };
              break;
            case 'exec':
              containerSpec[probeType].exec = {
                command: typeof container[probeType].exec.command === 'string' 
                  ? container[probeType].exec.command.split('\n').filter(Boolean)
                  : container[probeType].exec.command
              };
              break;
          }
        }
      });
      
      // 处理卷挂载
      if (container.volumeMounts && container.volumeMounts.length > 0) {
        containerSpec.volumeMounts = container.volumeMounts.map((mount: any) => ({
          name: mount.name,
          mountPath: mount.mountPath,
          subPath: mount.subPath,
          readOnly: mount.readOnly
        }));
      }
      
      return containerSpec;
    });
  }
  
  // 处理卷配置
  if (values.volumes && values.volumes.length > 0) {
    deploymentData.volumes = values.volumes.map((volume: any) => {
      const volumeConfig: any = {
        name: volume.name
      };
      
      switch (volume.type) {
        case 'configMap':
          volumeConfig.configMap = {
            name: volume.configMap.name,
            items: volume.configMap.items?.map((item: any) => ({
              key: item.key,
              path: item.path,
              mode: item.mode ? parseInt(item.mode, 8) : undefined
            }))
          };
          break;
        case 'secret':
          volumeConfig.secret = {
            secretName: volume.secret.secretName,
            items: volume.secret.items?.map((item: any) => ({
              key: item.key,
              path: item.path,
              mode: item.mode ? parseInt(item.mode, 8) : undefined
            }))
          };
          break;
        case 'persistentVolumeClaim':
          volumeConfig.persistentVolumeClaim = {
            claimName: volume.persistentVolumeClaim.claimName,
            readOnly: volume.persistentVolumeClaim.readOnly
          };
          break;
        case 'emptyDir':
          volumeConfig.emptyDir = {
            medium: volume.emptyDir.medium,
            sizeLimit: volume.emptyDir.sizeLimit
          };
          break;
        case 'hostPath':
          volumeConfig.hostPath = {
            path: volume.hostPath.path,
            type: volume.hostPath.type
          };
          break;
      }
      
      return volumeConfig;
    });
  }
  
  // 处理节点选择器
  if (values.nodeSelector && values.nodeSelector.length > 0) {
    deploymentData.nodeSelector = {};
    values.nodeSelector.forEach((item: { key: string; value: string }) => {
      if (item.key && item.value) {
        deploymentData.nodeSelector![item.key] = item.value;
      }
    });
  }

  // 处理节点亲和性
  if (values.nodeAffinity) {
    deploymentData.affinity = {
      nodeAffinity: {}
    };

    // 处理必须满足的节点选择规则
    if (values.nodeAffinity.requiredTerms && values.nodeAffinity.requiredTerms.length > 0) {
      deploymentData.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution = {
        nodeSelectorTerms: values.nodeAffinity.requiredTerms.map(term => {
          const result: any = {};
          
          if (term.matchExpressions && term.matchExpressions.length > 0) {
            result.matchExpressions = term.matchExpressions.filter(expr => expr.key && expr.operator)
              .map(expr => ({
                key: expr.key,
                operator: expr.operator,
                values: (expr.operator === 'Exists' || expr.operator === 'DoesNotExist') 
                  ? undefined 
                  : (expr.values || [])
              }));
          }
          
          if (term.matchFields && term.matchFields.length > 0) {
            result.matchFields = term.matchFields.filter(field => field.key && field.operator)
              .map(field => ({
                key: field.key,
                operator: field.operator,
                values: (field.operator === 'Exists' || field.operator === 'DoesNotExist') 
                  ? undefined 
                  : (field.values || [])
              }));
          }
          
          return result;
        }).filter(term => 
          (term.matchExpressions && term.matchExpressions.length > 0) || 
          (term.matchFields && term.matchFields.length > 0)
        )
      };
    }

    // 处理优先满足的节点选择规则
    if (values.nodeAffinity.preferredTerms && values.nodeAffinity.preferredTerms.length > 0) {
      deploymentData.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution = 
        values.nodeAffinity.preferredTerms
          .filter(preferred => preferred.weight > 0 && preferred.preference)
          .map(preferred => {
            const term: any = {
              weight: preferred.weight,
              preference: {}
            };

            if (preferred.preference.matchExpressions && preferred.preference.matchExpressions.length > 0) {
              term.preference.matchExpressions = preferred.preference.matchExpressions
                .filter(expr => expr.key && expr.operator)
                .map(expr => ({
                  key: expr.key,
                  operator: expr.operator,
                  values: (expr.operator === 'Exists' || expr.operator === 'DoesNotExist') 
                    ? undefined 
                    : (expr.values || [])
                }));
            }

            if (preferred.preference.matchFields && preferred.preference.matchFields.length > 0) {
              term.preference.matchFields = preferred.preference.matchFields
                .filter(field => field.key && field.operator)
                .map(field => ({
                  key: field.key,
                  operator: field.operator,
                  values: (field.operator === 'Exists' || field.operator === 'DoesNotExist') 
                    ? undefined 
                    : (field.values || [])
                }));
            }

            return term;
          })
          .filter(preferred => 
            (preferred.preference.matchExpressions && preferred.preference.matchExpressions.length > 0) || 
            (preferred.preference.matchFields && preferred.preference.matchFields.length > 0)
          );
    }

    // 如果没有亲和性规则，则移除affinity字段
    if (!deploymentData.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution &&
        !deploymentData.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution) {
      delete deploymentData.affinity;
    }
  }
  
  return deploymentData;
};

// 处理表单数据为更新API请求格式
export const processFormToUpdateRequest = (values: DeploymentFormData): UpdateDeploymentRequest => {
  const updateData: UpdateDeploymentRequest = {
    replicas: values.replicas,
    strategy: {
      type: values.strategy,
      rollingUpdate: values.strategy === 'RollingUpdate' ? {
        maxSurge: values.maxSurgeValue ? `${values.maxSurgeValue}%` : '25%',
        maxUnavailable: values.maxUnavailableValue ? `${values.maxUnavailableValue}%` : '25%'
      } : undefined
    },
    minReadySeconds: values.minReadySeconds,
    revisionHistoryLimit: values.revisionHistoryLimit,
    paused: values.paused,
  };

  // 处理容器资源
  if (values.containers) {
    const resources: { [key: string]: any } = {};
    const livenessProbe: { [key: string]: any } = {};
    const readinessProbe: { [key: string]: any } = {};
    const startupProbe: { [key: string]: any } = {};
    
    values.containers.forEach((container: any) => {
      if (container.resources) {
        resources[container.name] = {
          limits: {},
          requests: {}
        };
        
        // 处理CPU资源
        if (container.resources.requests?.cpuValue) {
          const cpuValue = container.resources.requests.cpuValue;
          const cpuUnit = container.resources.requests.cpuUnit || 'm';
          resources[container.name].requests.cpu = cpuUnit === 'm' ? `${cpuValue}m` : `${cpuValue}`;
        }
        
        if (container.resources.limits?.cpuValue) {
          const cpuValue = container.resources.limits.cpuValue;
          const cpuUnit = container.resources.limits.cpuUnit || 'm';
          resources[container.name].limits.cpu = cpuUnit === 'm' ? `${cpuValue}m` : `${cpuValue}`;
        }
        
        // 处理内存资源
        if (container.resources.requests?.memoryValue) {
          const memoryValue = container.resources.requests.memoryValue;
          const memoryUnit = container.resources.requests.memoryUnit || 'Mi';
          resources[container.name].requests.memory = `${memoryValue}${memoryUnit}`;
        }
        
        if (container.resources.limits?.memoryValue) {
          const memoryValue = container.resources.limits.memoryValue;
          const memoryUnit = container.resources.limits.memoryUnit || 'Mi';
          resources[container.name].limits.memory = `${memoryValue}${memoryUnit}`;
        }
        
        // 如果limits或requests为空，则删除它们
        if (Object.keys(resources[container.name].limits).length === 0) {
          delete resources[container.name].limits;
        }
        if (Object.keys(resources[container.name].requests).length === 0) {
          delete resources[container.name].requests;
        }
        // 如果整个resources为空，则删除它
        if (Object.keys(resources[container.name]).length === 0) {
          delete resources[container.name];
        }
      }
      
      // 处理健康检查探针
      ['livenessProbe', 'readinessProbe', 'startupProbe'].forEach(probeType => {
        if (container[probeType]?.type) {
          const probe: any = {
            initialDelaySeconds: container[probeType].initialDelaySeconds,
            timeoutSeconds: container[probeType].timeoutSeconds,
            periodSeconds: container[probeType].periodSeconds,
            successThreshold: container[probeType].successThreshold,
            failureThreshold: container[probeType].failureThreshold
          };
          
          switch (container[probeType].type) {
            case 'httpGet':
              probe.httpGet = {
                path: container[probeType].httpGet.path,
                port: container[probeType].httpGet.port,
                scheme: container[probeType].httpGet.scheme || 'HTTP'
              };
              break;
            case 'tcpSocket':
              probe.tcpSocket = {
                port: container[probeType].tcpSocket.port
              };
              break;
            case 'exec':
              probe.exec = {
                command: typeof container[probeType].exec.command === 'string'
                  ? container[probeType].exec.command.split('\n').filter(Boolean)
                  : container[probeType].exec.command
              };
              break;
          }
          
          // 根据探针类型存储到对应的对象中
          if (probeType === 'livenessProbe') {
            livenessProbe[container.name] = probe;
          } else if (probeType === 'readinessProbe') {
            readinessProbe[container.name] = probe;
          } else if (probeType === 'startupProbe') {
            startupProbe[container.name] = probe;
          }
        }
      });
    });
    
    // 如果有资源配置，添加到updateData中
    if (Object.keys(resources).length > 0) {
      updateData.resources = resources;
    }
    
    // 如果有健康检查配置，添加到updateData中
    if (Object.keys(livenessProbe).length > 0) {
      updateData.livenessProbe = livenessProbe;
    }
    if (Object.keys(readinessProbe).length > 0) {
      updateData.readinessProbe = readinessProbe;
    }
    if (Object.keys(startupProbe).length > 0) {
      updateData.startupProbe = startupProbe;
    }
  }
  
  // 处理labels
  if (values.labels && values.labels.length > 0) {
    updateData.labels = {};
    values.labels.forEach((item: { key: string; value: string }) => {
      if (item.key && item.value) {
        updateData.labels![item.key] = item.value;
      }
    });
  }
  
  // 处理annotations
  if (values.annotations && values.annotations.length > 0) {
    updateData.annotations = {};
    values.annotations.forEach((item: { key: string; value: string }) => {
      if (item.key && item.value) {
        updateData.annotations![item.key] = item.value;
      }
    });
  }
  
  // 处理节点选择器
  if (values.nodeSelector && values.nodeSelector.length > 0) {
    updateData.nodeSelector = {};
    values.nodeSelector.forEach((item: { key: string; value: string }) => {
      if (item.key && item.value) {
        updateData.nodeSelector![item.key] = item.value;
      }
    });
  }

  // 处理节点亲和性
  if (values.nodeAffinity) {
    updateData.affinity = {
      nodeAffinity: {}
    };

    // 处理必须满足的节点选择规则
    if (values.nodeAffinity.requiredTerms && values.nodeAffinity.requiredTerms.length > 0) {
      updateData.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution = {
        nodeSelectorTerms: values.nodeAffinity.requiredTerms.map(term => {
          const result: any = {};
          
          if (term.matchExpressions && term.matchExpressions.length > 0) {
            result.matchExpressions = term.matchExpressions.filter(expr => expr.key && expr.operator)
              .map(expr => ({
                key: expr.key,
                operator: expr.operator,
                values: (expr.operator === 'Exists' || expr.operator === 'DoesNotExist') 
                  ? undefined 
                  : (expr.values || [])
              }));
          }
          
          if (term.matchFields && term.matchFields.length > 0) {
            result.matchFields = term.matchFields.filter(field => field.key && field.operator)
              .map(field => ({
                key: field.key,
                operator: field.operator,
                values: (field.operator === 'Exists' || field.operator === 'DoesNotExist') 
                  ? undefined 
                  : (field.values || [])
              }));
          }
          
          return result;
        }).filter(term => 
          (term.matchExpressions && term.matchExpressions.length > 0) || 
          (term.matchFields && term.matchFields.length > 0)
        )
      };
    }

    // 处理优先满足的节点选择规则
    if (values.nodeAffinity.preferredTerms && values.nodeAffinity.preferredTerms.length > 0) {
      updateData.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution = 
        values.nodeAffinity.preferredTerms
          .filter(preferred => preferred.weight > 0 && preferred.preference)
          .map(preferred => {
            const term: any = {
              weight: preferred.weight,
              preference: {}
            };

            if (preferred.preference.matchExpressions && preferred.preference.matchExpressions.length > 0) {
              term.preference.matchExpressions = preferred.preference.matchExpressions
                .filter(expr => expr.key && expr.operator)
                .map(expr => ({
                  key: expr.key,
                  operator: expr.operator,
                  values: (expr.operator === 'Exists' || expr.operator === 'DoesNotExist') 
                    ? undefined 
                    : (expr.values || [])
                }));
            }

            if (preferred.preference.matchFields && preferred.preference.matchFields.length > 0) {
              term.preference.matchFields = preferred.preference.matchFields
                .filter(field => field.key && field.operator)
                .map(field => ({
                  key: field.key,
                  operator: field.operator,
                  values: (field.operator === 'Exists' || field.operator === 'DoesNotExist') 
                    ? undefined 
                    : (field.values || [])
                }));
            }

            return term;
          })
          .filter(preferred => 
            (preferred.preference.matchExpressions && preferred.preference.matchExpressions.length > 0) || 
            (preferred.preference.matchFields && preferred.preference.matchFields.length > 0)
          );
    }

    // 如果没有亲和性规则，则移除affinity字段
    if (!updateData.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution &&
        !updateData.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution) {
      delete updateData.affinity;
    }
  }
  
  return updateData;
};

// 处理后端数据为表单数据
export const processDeploymentToFormData = (deployment: any): DeploymentFormData => {
  // 转换labels和annotations为字符串数组格式，方便表单处理
  const labelsArray = Object.entries(deployment.labels || {}).map(([key, value]) => ({ key, value }));
  const annotationsArray = Object.entries(deployment.annotations || {}).map(([key, value]) => ({ key, value }));
  const nodeSelectorArray = Object.entries(deployment.nodeSelector || {}).map(([key, value]) => ({ key, value }));
  
  // 处理节点亲和性
  let nodeAffinity: NodeAffinityConfig | undefined = undefined;
  
  if (deployment.affinity?.nodeAffinity) {
    const affinity = deployment.affinity.nodeAffinity;
    
    nodeAffinity = {
      requiredTerms: [],
      preferredTerms: []
    };

    // 处理必须满足的节点选择规则
    if (affinity.requiredDuringSchedulingIgnoredDuringExecution?.nodeSelectorTerms) {
      nodeAffinity.requiredTerms = affinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms.map((term: any) => {
        const result: NodeSelectorTerm = {
          matchExpressions: [],
          matchFields: []
        };
        
        // 处理表达式匹配
        if (term.matchExpressions) {
          result.matchExpressions = term.matchExpressions.map((expr: any) => ({
            key: expr.key,
            operator: expr.operator,
            values: expr.values || []
          }));
        }
        
        // 处理字段匹配
        if (term.matchFields) {
          result.matchFields = term.matchFields.map((field: any) => ({
            key: field.key,
            operator: field.operator,
            values: field.values || []
          }));
        }
        
        return result;
      });
    }
    
    // 处理优先满足的节点选择规则
    if (affinity.preferredDuringSchedulingIgnoredDuringExecution) {
      nodeAffinity.preferredTerms = affinity.preferredDuringSchedulingIgnoredDuringExecution.map((term: any) => {
        const preference: NodeSelectorTerm = {
          matchExpressions: [],
          matchFields: []
        };
        
        // 处理表达式匹配
        if (term.preference.matchExpressions) {
          preference.matchExpressions = term.preference.matchExpressions.map((expr: any) => ({
            key: expr.key,
            operator: expr.operator,
            values: expr.values || []
          }));
        }
        
        // 处理字段匹配
        if (term.preference.matchFields) {
          preference.matchFields = term.preference.matchFields.map((field: any) => ({
            key: field.key,
            operator: field.operator,
            values: field.values || []
          }));
        }
        
        return {
          weight: term.weight,
          preference
        };
      });
    }
  }
  
  // 处理容器资源限制的数据格式
  const processedContainers = deployment.containers.map((container: any) => {
    const processedContainer = { ...container };
    
    if (container.resources) {
      processedContainer.resources = {
        requests: {},
        limits: {}
      };
      
      // 处理CPU资源
      if (container.resources.requests?.cpu) {
        const cpuMatch = container.resources.requests.cpu.match(/^(\d+)(m?)$/);
        if (cpuMatch) {
          processedContainer.resources.requests = {
            ...processedContainer.resources.requests,
            cpuValue: parseInt(cpuMatch[1]),
            cpuUnit: cpuMatch[2] || ""
          };
        }
      }
      
      if (container.resources.limits?.cpu) {
        const cpuMatch = container.resources.limits.cpu.match(/^(\d+)(m?)$/);
        if (cpuMatch) {
          processedContainer.resources.limits = {
            ...processedContainer.resources.limits,
            cpuValue: parseInt(cpuMatch[1]),
            cpuUnit: cpuMatch[2] || ""
          };
        }
      }
      
      // 处理内存资源
      if (container.resources.requests?.memory) {
        const memoryMatch = container.resources.requests.memory.match(/^(\d+)(Mi|Gi|M|G)$/);
        if (memoryMatch) {
          processedContainer.resources.requests = {
            ...processedContainer.resources.requests,
            memoryValue: parseInt(memoryMatch[1]),
            memoryUnit: memoryMatch[2]
          };
        }
      }
      
      if (container.resources.limits?.memory) {
        const memoryMatch = container.resources.limits.memory.match(/^(\d+)(Mi|Gi|M|G)$/);
        if (memoryMatch) {
          processedContainer.resources.limits = {
            ...processedContainer.resources.limits,
            memoryValue: parseInt(memoryMatch[1]),
            memoryUnit: memoryMatch[2]
          };
        }
      }
    }
    
    // 处理健康检查探针
    ['livenessProbe', 'readinessProbe', 'startupProbe'].forEach(probeType => {
      if (container[probeType]) {
        // 确定探针类型
        let type: string | undefined = undefined;
        if (container[probeType].httpGet) type = 'httpGet';
        else if (container[probeType].tcpSocket) type = 'tcpSocket';
        else if (container[probeType].exec) type = 'exec';
        
        processedContainer[probeType] = {
          type,
          initialDelaySeconds: container[probeType].initialDelaySeconds || 0,
          timeoutSeconds: container[probeType].timeoutSeconds || 1,
          periodSeconds: container[probeType].periodSeconds || 10,
          successThreshold: container[probeType].successThreshold || 1,
          failureThreshold: container[probeType].failureThreshold || 3,
          httpGet: container[probeType].httpGet || { path: '/', port: 80, scheme: 'HTTP' },
          tcpSocket: container[probeType].tcpSocket || { port: 80 },
          exec: container[probeType].exec || { command: [] }
        };
        
        // 如果是exec类型，处理command为字符串
        if (type === 'exec' && Array.isArray(processedContainer[probeType].exec.command)) {
          processedContainer[probeType].exec.command = processedContainer[probeType].exec.command.join('\n');
        }
      } else {
        processedContainer[probeType] = {
          type: undefined,
          initialDelaySeconds: 0,
          timeoutSeconds: 1,
          periodSeconds: 10,
          successThreshold: 1,
          failureThreshold: 3,
          httpGet: { path: '/', port: 80, scheme: 'HTTP' },
          tcpSocket: { port: 80 },
          exec: { command: '' }
        };
      }
    });
    
    return processedContainer;
  });
  
  return {
    name: deployment.name,
    replicas: deployment.replicas,
    strategy: deployment.strategy || 'RollingUpdate',
    minReadySeconds: deployment.minReadySeconds,
    revisionHistoryLimit: deployment.revisionHistoryLimit,
    paused: deployment.paused,
    containers: processedContainers,
    labels: labelsArray,
    annotations: annotationsArray,
    nodeSelector: nodeSelectorArray,
    nodeAffinity: nodeAffinity,
    serviceAccountName: deployment.serviceAccountName,
    hostNetwork: deployment.hostNetwork,
    dnsPolicy: deployment.dnsPolicy
  };
};