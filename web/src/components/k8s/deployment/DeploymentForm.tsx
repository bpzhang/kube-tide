import React, { useCallback } from 'react';
import { 
  Form, Input, InputNumber, Button, Tabs,
  Card, Row, Col, Select, Switch, Divider, 
  Space
} from 'antd';
import { PlusOutlined, MinusCircleOutlined } from '@ant-design/icons';
import { DeploymentFormProps } from './DeploymentTypes';
import PortNameSelect from '../common/PortNameSelect';

const { Option } = Select;

/**
 * 通用Deployment表单组件
 * 可用于创建和编辑Deployment
 */
const DeploymentForm: React.FC<DeploymentFormProps> = ({
  initialValues,
  onFormValuesChange,
  form,
  mode
}) => {
  const [activeTab, setActiveTab] = React.useState('basic');
  
  // 渲染基本信息Tab
  const renderBasicTab = useCallback(() => {
    return (
      <>
        {/* 在创建模式下才显示名称输入框 */}
        {mode === 'create' && (
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="name"
                label="Deployment名称"
                rules={[
                  { required: true, message: '请输入Deployment名称' },
                  { pattern: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/, message: '名称必须由小写字母、数字、和"-"组成，且不能以"-"开头或结尾' }
                ]}
              >
                <Input placeholder="my-deployment" />
              </Form.Item>
            </Col>
          </Row>
        )}
        
        <Row gutter={16}>
          <Col span={12}>
            <Form.Item
              name="replicas"
              label="副本数"
              rules={[{ required: true, message: '请输入副本数' }]}
            >
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item
              name="strategy"
              label="部署策略"
              rules={[{ required: true, message: '请选择部署策略' }]}
            >
              <Select>
                <Option value="RollingUpdate">滚动更新(RollingUpdate)</Option>
                <Option value="Recreate">重建(Recreate)</Option>
              </Select>
            </Form.Item>
          </Col>
        </Row>
        
        <Form.Item
          noStyle
          shouldUpdate={(prevValues, currentValues) => 
            prevValues.strategy !== currentValues.strategy
          }
        >
          {({ getFieldValue }) => {
            const strategy = getFieldValue('strategy');
            return strategy === 'RollingUpdate' ? (
              <Row gutter={16}>
                <Col span={12}>
                  <Form.Item name="maxSurgeValue" label="最大超出数">
                    <span style={{ display: 'flex', alignItems: 'center' }}>
                      <InputNumber
                        style={{ width: '100%' }}
                        min={0}
                        max={100}
                        placeholder="25"
                      />
                      <span style={{ marginLeft: 8 }}>%</span>
                    </span>
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item name="maxUnavailableValue" label="最大不可用数">
                    <span style={{ display: 'flex', alignItems: 'center' }}>
                      <InputNumber
                        style={{ width: '100%' }}
                        min={0}
                        max={100}
                        placeholder="25"
                      />
                      <span style={{ marginLeft: 8 }}>%</span>
                    </span>
                  </Form.Item>
                </Col>
              </Row>
            ) : null;
          }}
        </Form.Item>
        
        <Row gutter={16}>
          <Col span={8}>
            <Form.Item name="minReadySeconds" label="最小就绪时间(秒)">
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item name="revisionHistoryLimit" label="历史版本保留数">
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item name="paused" label="暂停部署" valuePropName="checked">
              <Switch />
            </Form.Item>
          </Col>
        </Row>
      </>
    );
  }, [mode]);
  
  // 渲染容器配置Tab
  const renderContainersTab = useCallback(() => {
    return (
      <Form.List name="containers">
        {(fields, { add, remove }) => (
          <div>
            {fields.map(field => {
              const containerName = form.getFieldValue(['containers', field.name, 'name']);
              return (
                <Card 
                  key={field.key} 
                  title={
                    mode === 'create' ? (
                      <Form.Item
                        key={field.key}
                        name={[field.name, 'name']}
                        label="容器名称"
                        rules={[
                          { required: true, message: '请输入容器名称' },
                          { pattern: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/, message: '名称必须由小写字母、数字、和"-"组成，且不能以"-"开头或结尾' }
                        ]}
                        style={{ marginBottom: 0 }}
                      >
                        <Input placeholder="container-1" />
                      </Form.Item>
                    ) : containerName
                  }
                  style={{ marginBottom: 16 }}
                  extra={fields.length > 1 && mode === 'create' ? (
                    <Button
                      type="link"
                      onClick={() => remove(field.name)}
                      icon={<MinusCircleOutlined />}
                    />
                  ) : null}
                >
                  {/* 在编辑模式下隐藏容器名称字段，但保留表单值 */}
                  {mode === 'edit' && (
                    <Form.Item
                      key={`${field.key}-name-hidden`}
                      name={[field.name, 'name']}
                      hidden
                    >
                      <Input />
                    </Form.Item>
                  )}
                  
                  <Form.Item
                    key={`${field.key}-image`}
                    name={[field.name, 'image']}
                    label="镜像"
                    rules={[{ required: true, message: '请输入容器镜像' }]}
                  >
                    <Input placeholder="例如: nginx:1.19" />
                  </Form.Item>
                  
                  <Divider orientation="left">资源限制</Divider>
                  <Row gutter={16}>
                    <Col span={12}>
                      <Form.Item
                        key={`${field.key}-resources-requests-cpu`}
                        name={[field.name, 'resources', 'requests', 'cpu']}
                        label="CPU请求"
                      >
                        <Space.Compact style={{ width: '100%' }}>
                          <Form.Item
                            key={`${field.key}-resources-requests-cpuValue`}
                            name={[field.name, 'resources', 'requests', 'cpuValue']}
                            noStyle
                            rules={[{ required: true, message: '请输入CPU请求值' }]}
                          >
                            <InputNumber 
                              style={{ width: '60%' }} 
                              min={0} 
                              placeholder="请输入数值"
                            />
                          </Form.Item>
                          <Form.Item
                            key={`${field.key}-resources-requests-cpuUnit`}
                            name={[field.name, 'resources', 'requests', 'cpuUnit']}
                            noStyle
                            initialValue="m"
                          >
                            <Select style={{ width: '40%' }}>
                              <Option value="m">毫核(m)</Option>
                              <Option value="">核心</Option>
                            </Select>
                          </Form.Item>
                        </Space.Compact>
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        key={`${field.key}-resources-limits-cpu`}
                        name={[field.name, 'resources', 'limits', 'cpu']}
                        label="CPU限制"
                      >
                        <Space.Compact style={{ width: '100%' }}>
                          <Form.Item
                            key={`${field.key}-resources-limits-cpuValue`}
                            name={[field.name, 'resources', 'limits', 'cpuValue']}
                            noStyle
                            rules={[{ required: true, message: '请输入CPU限制值' }]}
                          >
                            <InputNumber 
                              style={{ width: '60%' }} 
                              min={0} 
                              placeholder="请输入数值"
                            />
                          </Form.Item>
                          <Form.Item
                            key={`${field.key}-resources-limits-cpuUnit`}
                            name={[field.name, 'resources', 'limits', 'cpuUnit']}
                            noStyle
                            initialValue="m"
                          >
                            <Select style={{ width: '40%' }}>
                              <Option value="m">毫核(m)</Option>
                              <Option value="">核心</Option>
                            </Select>
                          </Form.Item>
                        </Space.Compact>
                      </Form.Item>
                    </Col>
                  </Row>
                  <Row gutter={16}>
                    <Col span={12}>
                      <Form.Item
                        key={`${field.key}-resources-requests-memory`}
                        name={[field.name, 'resources', 'requests', 'memory']}
                        label="内存请求"
                      >
                        <Space.Compact style={{ width: '100%' }}>
                          <Form.Item
                            key={`${field.key}-resources-requests-memoryValue`}
                            name={[field.name, 'resources', 'requests', 'memoryValue']}
                            noStyle
                            rules={[{ required: true, message: '请输入内存请求值' }]}
                          >
                            <InputNumber 
                              style={{ width: '60%' }} 
                              min={0} 
                              placeholder="请输入数值"
                            />
                          </Form.Item>
                          <Form.Item
                            key={`${field.key}-resources-requests-memoryUnit`}
                            name={[field.name, 'resources', 'requests', 'memoryUnit']}
                            noStyle
                            initialValue="Mi"
                          >
                            <Select style={{ width: '40%' }}>
                              <Option value="Mi">MiB</Option>
                              <Option value="Gi">GiB</Option>
                              <Option value="M">MB</Option>
                              <Option value="G">GB</Option>
                            </Select>
                          </Form.Item>
                        </Space.Compact>
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        key={`${field.key}-resources-limits-memory`}
                        name={[field.name, 'resources', 'limits', 'memory']}
                        label="内存限制"
                      >
                        <Space.Compact style={{ width: '100%' }}>
                          <Form.Item
                            key={`${field.key}-resources-limits-memoryValue`}
                            name={[field.name, 'resources', 'limits', 'memoryValue']}
                            noStyle
                            rules={[{ required: true, message: '请输入内存限制值' }]}
                          >
                            <InputNumber 
                              style={{ width: '60%' }} 
                              min={0} 
                              placeholder="请输入数值"
                            />
                          </Form.Item>
                          <Form.Item
                            key={`${field.key}-resources-limits-memoryUnit`}
                            name={[field.name, 'resources', 'limits', 'memoryUnit']}
                            noStyle
                            initialValue="Mi"
                          >
                            <Select style={{ width: '40%' }}>
                              <Option value="Mi">MiB</Option>
                              <Option value="Gi">GiB</Option>
                              <Option value="M">MB</Option>
                              <Option value="G">GB</Option>
                            </Select>
                          </Form.Item>
                        </Space.Compact>
                      </Form.Item>
                    </Col>
                  </Row>

                  <Divider orientation="left">端口映射</Divider>
                  <Form.List name={[field.name, 'ports']}>
                    {(portFields, { add: addPort, remove: removePort }) => (
                      <>
                        {portFields.map(portField => (
                          <Row key={`port-${field.key}-${portField.key}`} gutter={16} align="middle">
                            <Col span={6}>
                              <Form.Item
                                {...portField}
                                name={[portField.name, 'name']}
                                label="名称"
                              >
                                <PortNameSelect placeholder="选择通信协议" />
                              </Form.Item>
                            </Col>
                            <Col span={6}>
                              <Form.Item
                                {...portField}
                                name={[portField.name, 'containerPort']}
                                label="容器端口"
                                rules={[{ required: true, message: '请输入容器端口' }]}
                              >
                                <InputNumber min={1} max={65535} style={{ width: '100%' }} />
                              </Form.Item>
                            </Col>
                            <Col span={6}>
                              <Form.Item
                                {...portField}
                                name={[portField.name, 'protocol']}
                                label="协议"
                                initialValue="TCP"
                              >
                                <Select>
                                  <Option value="TCP">TCP</Option>
                                  <Option value="UDP">UDP</Option>
                                  <Option value="SCTP">SCTP</Option>
                                </Select>
                              </Form.Item>
                            </Col>
                            <Col span={4}>
                              <Form.Item label=" " style={{ marginBottom: 0 }}>
                                <Button
                                  type="link"
                                  onClick={() => removePort(portField.name)}
                                  icon={<MinusCircleOutlined />}
                                />
                              </Form.Item>
                            </Col>
                          </Row>
                        ))}
                        <Form.Item>
                          <Button
                            type="dashed"
                            onClick={() => addPort({ containerPort: 80, protocol: 'TCP' })}
                            block
                            icon={<PlusOutlined />}
                          >
                            添加端口
                          </Button>
                        </Form.Item>
                      </>
                    )}
                  </Form.List>
                  
                  <Divider orientation="left">环境变量</Divider>
                  <Form.List name={[field.name, 'env']}>
                    {(envFields, { add: addEnv, remove: removeEnv }) => (
                      <>
                        {envFields.map(envField => (
                          <Row key={`env-${field.key}-${envField.key}`} gutter={8} align="middle">
                            <Col span={10}>
                              <Form.Item
                                {...envField}
                                name={[envField.name, 'name']}
                                rules={[{ required: true, message: '请输入变量名' }]}
                                style={{ marginBottom: 8 }}
                              >
                                <Input placeholder="变量名" />
                              </Form.Item>
                            </Col>
                            <Col span={10}>
                              <Form.Item
                                {...envField}
                                name={[envField.name, 'value']}
                                rules={[{ required: true, message: '请输入变量值' }]}
                                style={{ marginBottom: 8 }}
                              >
                                <Input placeholder="变量值" />
                              </Form.Item>
                            </Col>
                            <Col span={4}>
                              <Button
                                type="link"
                                icon={<MinusCircleOutlined />}
                                onClick={() => removeEnv(envField.name)}
                              />
                            </Col>
                          </Row>
                        ))}
                        <Form.Item>
                          <Button
                            type="dashed"
                            onClick={() => addEnv()}
                            block
                            icon={<PlusOutlined />}
                          >
                            添加环境变量
                          </Button>
                        </Form.Item>
                      </>
                    )}
                  </Form.List>
                  
                  <Divider orientation="left">健康检查</Divider>
                  
                  {/* 存活探针配置 */}
                  <Card
                    size="small"
                    title="存活探针 (Liveness Probe)"
                    style={{ marginBottom: 16 }}
                    styles={{ body: { padding: '12px' } }}
                  >
                    <Form.Item
                      name={[field.name, 'livenessProbe', 'type']}
                      label="检查类型"
                    >
                      <Select placeholder="选择检查类型">
                        <Option value="httpGet">HTTP GET</Option>
                        <Option value="tcpSocket">TCP Socket</Option>
                        <Option value="exec">命令行</Option>
                      </Select>
                    </Form.Item>
                    <Form.Item
                      noStyle
                      shouldUpdate={(prevValues, curValues) => {
                        const prevType = prevValues?.containers?.[field.name]?.livenessProbe?.type;
                        const curType = curValues?.containers?.[field.name]?.livenessProbe?.type;
                        return prevType !== curType;
                      }}
                    >
                      {({ getFieldValue }) => {
                        const probeType = getFieldValue(['containers', field.name, 'livenessProbe', 'type']);
                        
                        switch (probeType) {
                          case 'httpGet':
                            return (
                              <>
                                <Form.Item
                                  name={[field.name, 'livenessProbe', 'httpGet', 'path']}
                                  label="路径"
                                  rules={[{ required: true, message: '请输入检查路径' }]}
                                >
                                  <Input placeholder="/health" />
                                </Form.Item>
                                <Form.Item
                                  name={[field.name, 'livenessProbe', 'httpGet', 'port']}
                                  label="端口"
                                  rules={[{ required: true, message: '请输入端口号' }]}
                                >
                                  <InputNumber min={1} max={65535} />
                                </Form.Item>
                                <Form.Item
                                  name={[field.name, 'livenessProbe', 'httpGet', 'scheme']}
                                  label="协议"
                                  initialValue="HTTP"
                                >
                                  <Select>
                                    <Option value="HTTP">HTTP</Option>
                                    <Option value="HTTPS">HTTPS</Option>
                                  </Select>
                                </Form.Item>
                              </>
                            );
                          case 'tcpSocket':
                            return (
                              <Form.Item
                                name={[field.name, 'livenessProbe', 'tcpSocket', 'port']}
                                label="端口"
                                rules={[{ required: true, message: '请输入端口号' }]}
                              >
                                <InputNumber min={1} max={65535} />
                              </Form.Item>
                            );
                          case 'exec':
                            return (
                              <Form.Item
                                name={[field.name, 'livenessProbe', 'exec', 'command']}
                                label="命令"
                                rules={[{ required: true, message: '请输入检查命令' }]}
                              >
                                <Input.TextArea placeholder="输入要执行的命令，每行一条" />
                              </Form.Item>
                            );
                          default:
                            return null;
                        }
                      }}
                    </Form.Item>
                    <Row gutter={16}>
                      <Col span={8}>
                        <Form.Item
                          name={[field.name, 'livenessProbe', 'initialDelaySeconds']}
                          label="初始延迟(秒)"
                        >
                          <InputNumber min={0} />
                        </Form.Item>
                      </Col>
                      <Col span={8}>
                        <Form.Item
                          name={[field.name, 'livenessProbe', 'periodSeconds']}
                          label="检查周期(秒)"
                        >
                          <InputNumber min={1} />
                        </Form.Item>
                      </Col>
                      <Col span={8}>
                        <Form.Item
                          name={[field.name, 'livenessProbe', 'timeoutSeconds']}
                          label="超时时间(秒)"
                        >
                          <InputNumber min={1} />
                        </Form.Item>
                      </Col>
                    </Row>
                    <Row gutter={16}>
                      <Col span={12}>
                        <Form.Item
                          name={[field.name, 'livenessProbe', 'successThreshold']}
                          label="成功阈值"
                        >
                          <InputNumber min={1} />
                        </Form.Item>
                      </Col>
                      <Col span={12}>
                        <Form.Item
                          name={[field.name, 'livenessProbe', 'failureThreshold']}
                          label="失败阈值"
                        >
                          <InputNumber min={1} />
                        </Form.Item>
                      </Col>
                    </Row>
                  </Card>
                  
                  {/* 就绪探针配置 */}
                  <Card
                    size="small"
                    title="就绪探针 (Readiness Probe)"
                    style={{ marginBottom: 16 }}
                    styles={{ body: { padding: '12px' } }}
                  >
                    <Form.Item
                      name={[field.name, 'readinessProbe', 'type']}
                      label="检查类型"
                    >
                      <Select placeholder="选择检查类型">
                        <Option value="httpGet">HTTP GET</Option>
                        <Option value="tcpSocket">TCP Socket</Option>
                        <Option value="exec">命令行</Option>
                      </Select>
                    </Form.Item>
                    <Form.Item
                      noStyle
                      shouldUpdate={(prevValues, curValues) => {
                        const prevType = prevValues?.containers?.[field.name]?.readinessProbe?.type;
                        const curType = curValues?.containers?.[field.name]?.readinessProbe?.type;
                        return prevType !== curType;
                      }}
                    >
                      {({ getFieldValue }) => {
                        const probeType = getFieldValue(['containers', field.name, 'readinessProbe', 'type']);
                        
                        switch (probeType) {
                          case 'httpGet':
                            return (
                              <>
                                <Form.Item
                                  name={[field.name, 'readinessProbe', 'httpGet', 'path']}
                                  label="路径"
                                  rules={[{ required: true, message: '请输入检查路径' }]}
                                >
                                  <Input placeholder="/ready" />
                                </Form.Item>
                                <Form.Item
                                  name={[field.name, 'readinessProbe', 'httpGet', 'port']}
                                  label="端口"
                                  rules={[{ required: true, message: '请输入端口号' }]}
                                >
                                  <InputNumber min={1} max={65535} />
                                </Form.Item>
                                <Form.Item
                                  name={[field.name, 'readinessProbe', 'httpGet', 'scheme']}
                                  label="协议"
                                  initialValue="HTTP"
                                >
                                  <Select>
                                    <Option value="HTTP">HTTP</Option>
                                    <Option value="HTTPS">HTTPS</Option>
                                  </Select>
                                </Form.Item>
                              </>
                            );
                          case 'tcpSocket':
                            return (
                              <Form.Item
                                name={[field.name, 'readinessProbe', 'tcpSocket', 'port']}
                                label="端口"
                                rules={[{ required: true, message: '请输入端口号' }]}
                              >
                                <InputNumber min={1} max={65535} />
                              </Form.Item>
                            );
                          case 'exec':
                            return (
                              <Form.Item
                                name={[field.name, 'readinessProbe', 'exec', 'command']}
                                label="命令"
                                rules={[{ required: true, message: '请输入检查命令' }]}
                              >
                                <Input.TextArea placeholder="输入要执行的命令，每行一条" />
                              </Form.Item>
                            );
                          default:
                            return null;
                        }
                      }}
                    </Form.Item>
                    <Row gutter={16}>
                      <Col span={8}>
                        <Form.Item
                          name={[field.name, 'readinessProbe', 'initialDelaySeconds']}
                          label="初始延迟(秒)"
                        >
                          <InputNumber min={0} />
                        </Form.Item>
                      </Col>
                      <Col span={8}>
                        <Form.Item
                          name={[field.name, 'readinessProbe', 'periodSeconds']}
                          label="检查周期(秒)"
                        >
                          <InputNumber min={1} />
                        </Form.Item>
                      </Col>
                      <Col span={8}>
                        <Form.Item
                          name={[field.name, 'readinessProbe', 'timeoutSeconds']}
                          label="超时时间(秒)"
                        >
                          <InputNumber min={1} />
                        </Form.Item>
                      </Col>
                    </Row>
                    <Row gutter={16}>
                      <Col span={12}>
                        <Form.Item
                          name={[field.name, 'readinessProbe', 'successThreshold']}
                          label="成功阈值"
                        >
                          <InputNumber min={1} />
                        </Form.Item>
                      </Col>
                      <Col span={12}>
                        <Form.Item
                          name={[field.name, 'readinessProbe', 'failureThreshold']}
                          label="失败阈值"
                        >
                          <InputNumber min={1} />
                        </Form.Item>
                      </Col>
                    </Row>
                  </Card>
                  
                  {/* 启动探针配置 */}
                  <Card
                    size="small"
                    title="启动探针 (Startup Probe)"
                    style={{ marginBottom: 16 }}
                    styles={{ body: { padding: '12px' } }}
                  >
                    <Form.Item
                      name={[field.name, 'startupProbe', 'type']}
                      label="检查类型"
                    >
                      <Select placeholder="选择检查类型">
                        <Option value="httpGet">HTTP GET</Option>
                        <Option value="tcpSocket">TCP Socket</Option>
                        <Option value="exec">命令行</Option>
                      </Select>
                    </Form.Item>
                    <Form.Item
                      noStyle
                      shouldUpdate={(prevValues, curValues) => {
                        const prevType = prevValues?.containers?.[field.name]?.startupProbe?.type;
                        const curType = curValues?.containers?.[field.name]?.startupProbe?.type;
                        return prevType !== curType;
                      }}
                    >
                      {({ getFieldValue }) => {
                        const probeType = getFieldValue(['containers', field.name, 'startupProbe', 'type']);
                        
                        switch (probeType) {
                          case 'httpGet':
                            return (
                              <>
                                <Form.Item
                                  name={[field.name, 'startupProbe', 'httpGet', 'path']}
                                  label="路径"
                                  rules={[{ required: true, message: '请输入检查路径' }]}
                                >
                                  <Input placeholder="/startup" />
                                </Form.Item>
                                <Form.Item
                                  name={[field.name, 'startupProbe', 'httpGet', 'port']}
                                  label="端口"
                                  rules={[{ required: true, message: '请输入端口号' }]}
                                >
                                  <InputNumber min={1} max={65535} />
                                </Form.Item>
                                <Form.Item
                                  name={[field.name, 'startupProbe', 'httpGet', 'scheme']}
                                  label="协议"
                                  initialValue="HTTP"
                                >
                                  <Select>
                                    <Option value="HTTP">HTTP</Option>
                                    <Option value="HTTPS">HTTPS</Option>
                                  </Select>
                                </Form.Item>
                              </>
                            );
                          case 'tcpSocket':
                            return (
                              <Form.Item
                                name={[field.name, 'startupProbe', 'tcpSocket', 'port']}
                                label="端口"
                                rules={[{ required: true, message: '请输入端口号' }]}
                              >
                                <InputNumber min={1} max={65535} />
                              </Form.Item>
                            );
                          case 'exec':
                            return (
                              <Form.Item
                                name={[field.name, 'startupProbe', 'exec', 'command']}
                                label="命令"
                                rules={[{ required: true, message: '请输入检查命令' }]}
                              >
                                <Input.TextArea placeholder="输入要执行的命令，每行一条" />
                              </Form.Item>
                            );
                          default:
                            return null;
                        }
                      }}
                    </Form.Item>
                    <Row gutter={16}>
                      <Col span={8}>
                        <Form.Item
                          name={[field.name, 'startupProbe', 'initialDelaySeconds']}
                          label="初始延迟(秒)"
                        >
                          <InputNumber min={0} />
                        </Form.Item>
                      </Col>
                      <Col span={8}>
                        <Form.Item
                          name={[field.name, 'startupProbe', 'periodSeconds']}
                          label="检查周期(秒)"
                        >
                          <InputNumber min={1} />
                        </Form.Item>
                      </Col>
                      <Col span={8}>
                        <Form.Item
                          name={[field.name, 'startupProbe', 'timeoutSeconds']}
                          label="超时时间(秒)"
                        >
                          <InputNumber min={1} />
                        </Form.Item>
                      </Col>
                    </Row>
                    <Row gutter={16}>
                      <Col span={12}>
                        <Form.Item
                          name={[field.name, 'startupProbe', 'successThreshold']}
                          label="成功阈值"
                        >
                          <InputNumber min={1} />
                        </Form.Item>
                      </Col>
                      <Col span={12}>
                        <Form.Item
                          name={[field.name, 'startupProbe', 'failureThreshold']}
                          label="失败阈值"
                        >
                          <InputNumber min={1} />
                        </Form.Item>
                      </Col>
                    </Row>
                  </Card>

                  {/* 卷挂载配置 */}
                  {form.getFieldValue('volumes')?.length > 0 && (
                    <>
                      <Divider orientation="left">卷挂载</Divider>
                      <Form.List name={[field.name, 'volumeMounts']}>
                        {(mountFields, { add: addMount, remove: removeMount }) => (
                          <>
                            {mountFields.map(mountField => (
                              <Row key={`mount-${field.key}-${mountField.key}`} gutter={16} align="middle">
                                <Col span={6}>
                                  <Form.Item
                                    key={`mount-${field.key}-${mountField.key}-name`}
                                    name={[mountField.name, 'name']}
                                    label="卷名称"
                                    rules={[{ required: true, message: '请选择要挂载的卷' }]}
                                  >
                                    <Select placeholder="选择卷">
                                      {form.getFieldValue('volumes')?.map((volume: any) => (
                                        <Option key={volume.name} value={volume.name}>
                                          {volume.name}
                                        </Option>
                                      )) || []}
                                    </Select>
                                  </Form.Item>
                                </Col>
                                <Col span={8}>
                                  <Form.Item
                                    key={`mount-${field.key}-${mountField.key}-mountPath`}
                                    name={[mountField.name, 'mountPath']}
                                    label="挂载路径"
                                    rules={[{ required: true, message: '请输入挂载路径' }]}
                                  >
                                    <Input placeholder="/data" />
                                  </Form.Item>
                                </Col>
                                <Col span={6}>
                                  <Form.Item
                                    key={`mount-${field.key}-${mountField.key}-subPath`}
                                    name={[mountField.name, 'subPath']}
                                    label="子路径"
                                  >
                                    <Input placeholder="可选" />
                                  </Form.Item>
                                </Col>
                                <Col span={3}>
                                  <Form.Item
                                    key={`mount-${field.key}-${mountField.key}-readOnly`}
                                    name={[mountField.name, 'readOnly']}
                                    label="只读"
                                    valuePropName="checked"
                                  >
                                    <Switch />
                                  </Form.Item>
                                </Col>
                                <Col span={1}>
                                  <Button
                                    type="link"
                                    icon={<MinusCircleOutlined />}
                                    onClick={() => removeMount(mountField.name)}
                                  />
                                </Col>
                              </Row>
                            ))}
                            <Form.Item>
                              <Button
                                type="dashed"
                                onClick={() => addMount()}
                                block
                                icon={<PlusOutlined />}
                              >
                                添加卷挂载
                              </Button>
                            </Form.Item>
                          </>
                        )}
                      </Form.List>
                    </>
                  )}
                </Card>
              );
            })}
            
            {/* 只在创建模式下显示添加容器按钮 */}
            {mode === 'create' && (
              <Form.Item>
                <Button
                  type="dashed"
                  onClick={() => add({
                    name: `container-${fields.length + 1}`,
                    resources: {
                      limits: { 
                        cpuValue: 500, 
                        cpuUnit: 'm', 
                        memoryValue: 512, 
                        memoryUnit: 'Mi' 
                      },
                      requests: { 
                        cpuValue: 100, 
                        cpuUnit: 'm', 
                        memoryValue: 128, 
                        memoryUnit: 'Mi' 
                      }
                    },
                    env: []
                  })}
                  block
                  icon={<PlusOutlined />}
                >
                  添加容器
                </Button>
              </Form.Item>
            )}
          </div>
        )}
      </Form.List>
    );
  }, [form, mode]);
  
  // 渲染存储卷Tab
  const renderVolumesTab = useCallback(() => {
    return (
      <Form.List name="volumes">
        {(fields, { add, remove }) => (
          <>
            {fields.map((field) => (
              <Card
                key={field.key}
                style={{ marginBottom: 16 }}
                title={
                  <Form.Item
                    key={field.key}
                    name={[field.name, 'name']}
                    rules={[{ required: true, message: '请输入卷名称' }]}
                    style={{ marginBottom: 0 }}
                  >
                    <Input placeholder="卷名称" />
                  </Form.Item>
                }
                extra={
                  <Button
                    type="link"
                    onClick={() => remove(field.name)}
                    icon={<MinusCircleOutlined />}
                  />
                }
              >
                <Form.Item
                  key={field.key}
                  name={[field.name, 'type']}
                  label="卷类型"
                  rules={[{ required: true, message: '请选择卷类型' }]}
                >
                  <Select>
                    <Option value="configMap">ConfigMap</Option>
                    <Option value="secret">Secret</Option>
                    <Option value="persistentVolumeClaim">PersistentVolumeClaim</Option>
                    <Option value="emptyDir">EmptyDir</Option>
                    <Option value="hostPath">HostPath</Option>
                  </Select>
                </Form.Item>
                <Form.Item
                  noStyle
                  shouldUpdate={(prevValues, curValues) => {
                    const prevType = prevValues?.volumes?.[field.name]?.type;
                    const curType = curValues?.volumes?.[field.name]?.type;
                    return prevType !== curType;
                  }}
                >
                  {({ getFieldValue }) => {
                    const type = getFieldValue(['volumes', field.name, 'type']);
                    switch (type) {
                      case 'configMap':
                        return (
                          <>
                            <Form.Item
                              key={field.key}
                              name={[field.name, 'configMap', 'name']}
                              label="ConfigMap名称"
                              rules={[{ required: true, message: '请输入ConfigMap名称' }]}
                            >
                              <Input placeholder="my-config" />
                            </Form.Item>
                            <Form.List name={[field.name, 'configMap', 'items']}>
                              {(itemFields, { add: addItem, remove: removeItem }) => (
                                <>
                                  {itemFields.map((itemField) => (
                                    <Row key={itemField.key} gutter={16} align="middle">
                                      <Col span={8}>
                                        <Form.Item
                                          key={itemField.key}
                                          name={[itemField.name, 'key']}
                                          label="键"
                                          rules={[{ required: true, message: '请输入键名' }]}
                                        >
                                          <Input placeholder="config-key" />
                                        </Form.Item>
                                      </Col>
                                      <Col span={8}>
                                        <Form.Item
                                          key={itemField.key}
                                          name={[itemField.name, 'path']}
                                          label="路径"
                                          rules={[{ required: true, message: '请输入文件路径' }]}
                                        >
                                          <Input placeholder="config.yaml" />
                                        </Form.Item>
                                      </Col>
                                      <Col span={7}>
                                        <Form.Item
                                          key={itemField.key}
                                          name={[itemField.name, 'mode']}
                                          label="权限"
                                        >
                                          <Input placeholder="0644" />
                                        </Form.Item>
                                      </Col>
                                      <Col span={1}>
                                        <Button
                                          type="link"
                                          onClick={() => removeItem(itemField.name)}
                                          icon={<MinusCircleOutlined />}
                                        />
                                      </Col>
                                    </Row>
                                  ))}
                                  <Form.Item>
                                    <Button
                                      type="dashed"
                                      onClick={() => addItem()}
                                      block
                                      icon={<PlusOutlined />}
                                    >
                                      添加键值映射
                                    </Button>
                                  </Form.Item>
                                </>
                              )}
                            </Form.List>
                          </>
                        );
                      case 'secret':
                        return (
                          <>
                            <Form.Item
                              key={field.key}
                              name={[field.name, 'secret', 'secretName']}
                              label="Secret名称"
                              rules={[{ required: true, message: '请输入Secret名称' }]}
                            >
                              <Input placeholder="my-secret" />
                            </Form.Item>
                            <Form.List name={[field.name, 'secret', 'items']}>
                              {(itemFields, { add: addItem, remove: removeItem }) => (
                                <>
                                  {itemFields.map((itemField) => (
                                    <Row key={itemField.key} gutter={16} align="middle">
                                      <Col span={8}>
                                        <Form.Item
                                          key={itemField.key}
                                          name={[itemField.name, 'key']}
                                          label="键"
                                          rules={[{ required: true, message: '请输入键名' }]}
                                        >
                                          <Input placeholder="secret-key" />
                                        </Form.Item>
                                      </Col>
                                      <Col span={8}>
                                        <Form.Item
                                          key={itemField.key}
                                          name={[itemField.name, 'path']}
                                          label="路径"
                                          rules={[{ required: true, message: '请输入文件路径' }]}
                                        >
                                          <Input placeholder="secret.txt" />
                                        </Form.Item>
                                      </Col>
                                      <Col span={7}>
                                        <Form.Item
                                          key={itemField.key}
                                          name={[itemField.name, 'mode']}
                                          label="权限"
                                        >
                                          <Input placeholder="0600" />
                                        </Form.Item>
                                      </Col>
                                      <Col span={1}>
                                        <Button
                                          type="link"
                                          onClick={() => removeItem(itemField.name)}
                                          icon={<MinusCircleOutlined />}
                                        />
                                      </Col>
                                    </Row>
                                  ))}
                                  <Form.Item>
                                    <Button
                                      type="dashed"
                                      onClick={() => addItem()}
                                      block
                                      icon={<PlusOutlined />}
                                    >
                                      添加键值映射
                                    </Button>
                                  </Form.Item>
                                </>
                              )}
                            </Form.List>
                          </>
                        );
                      case 'persistentVolumeClaim':
                        return (
                          <>
                            <Form.Item
                              key={field.key}
                              name={[field.name, 'persistentVolumeClaim', 'claimName']}
                              label="PVC名称"
                              rules={[{ required: true, message: '请输入PVC名称' }]}
                            >
                              <Input placeholder="my-pvc" />
                            </Form.Item>
                            <Form.Item
                              key={field.key}
                              name={[field.name, 'persistentVolumeClaim', 'readOnly']}
                              label="只读"
                              valuePropName="checked"
                            >
                              <Switch />
                            </Form.Item>
                          </>
                        );
                      case 'emptyDir':
                        return (
                          <>
                            <Form.Item
                              key={field.key}
                              name={[field.name, 'emptyDir', 'medium']}
                              label="存储介质"
                            >
                              <Select placeholder="默认">
                                <Option value="">默认</Option>
                                <Option value="Memory">内存</Option>
                              </Select>
                            </Form.Item>
                            <Form.Item
                              key={field.key}
                              name={[field.name, 'emptyDir', 'sizeLimit']}
                              label="大小限制"
                            >
                              <Input placeholder="例如：1Gi" />
                            </Form.Item>
                          </>
                        );
                      case 'hostPath':
                        return (
                          <>
                            <Form.Item
                              key={field.key}
                              name={[field.name, 'hostPath', 'path']}
                              label="主机路径"
                              rules={[{ required: true, message: '请输入主机路径' }]}
                            >
                              <Input placeholder="/data" />
                            </Form.Item>
                            <Form.Item
                              key={field.key}
                              name={[field.name, 'hostPath', 'type']}
                              label="类型"
                            >
                              <Select placeholder="默认">
                                <Option value="">默认</Option>
                                <Option value="Directory">目录</Option>
                                <Option value="DirectoryOrCreate">目录或创建</Option>
                                <Option value="File">文件</Option>
                                <Option value="FileOrCreate">文件或创建</Option>
                              </Select>
                            </Form.Item>
                          </>
                        );
                      default:
                        return null;
                    }
                  }}
                </Form.Item>
              </Card>
            ))}
            <Form.Item>
              <Button
                type="dashed"
                onClick={() => add()}
                block
                icon={<PlusOutlined />}
              >
                添加存储卷
              </Button>
            </Form.Item>
          </>
        )}
      </Form.List>
    );
  }, []);
  
  // 渲染标签和注解Tab
  const renderLabelsAndAnnotationsTab = useCallback(() => {
    return (
      <>
        <Divider orientation="left">标签</Divider>
        <Form.List name="labels">
          {(fields, { add, remove }) => (
            <>
              {fields.map(field => (
                <Row key={field.key} gutter={8} align="middle">
                  <Col span={10}>
                    <Form.Item
                      key={field.key}
                      name={[field.name, 'key']}
                      rules={[{ required: true, message: '请输入标签键' }]}
                      style={{ marginBottom: 8 }}
                    >
                      <Input placeholder="标签键" />
                    </Form.Item>
                  </Col>
                  <Col span={10}>
                    <Form.Item
                      key={field.key}
                      name={[field.name, 'value']}
                      rules={[{ required: true, message: '请输入标签值' }]}
                      style={{ marginBottom: 8 }}
                    >
                      <Input placeholder="标签值" />
                    </Form.Item>
                  </Col>
                  <Col span={4}>
                    <Button
                      type="link"
                      icon={<MinusCircleOutlined />}
                      onClick={() => remove(field.name)}
                    />
                  </Col>
                </Row>
              ))}
              <Form.Item>
                <Button
                  type="dashed"
                  onClick={() => add()}
                  block
                  icon={<PlusOutlined />}
                >
                  添加标签
                </Button>
              </Form.Item>
            </>
          )}
        </Form.List>
        
        <Divider orientation="left">节点选择器</Divider>
        <Form.List name="nodeSelector">
          {(fields, { add, remove }) => (
            <>
              {fields.map(field => (
                <Row key={field.key} gutter={8} align="middle">
                  <Col span={10}>
                    <Form.Item
                      key={field.key}
                      name={[field.name, 'key']}
                      rules={[{ required: true, message: '请输入选择器键' }]}
                      style={{ marginBottom: 8 }}
                    >
                      <Input placeholder="例如: disk-type" />
                    </Form.Item>
                  </Col>
                  <Col span={10}>
                    <Form.Item
                      key={field.key}
                      name={[field.name, 'value']}
                      rules={[{ required: true, message: '请输入选择器值' }]}
                      style={{ marginBottom: 8 }}
                    >
                      <Input placeholder="例如: ssd" />
                    </Form.Item>
                  </Col>
                  <Col span={4}>
                    <Button
                      type="link"
                      icon={<MinusCircleOutlined />}
                      onClick={() => remove(field.name)}
                    />
                  </Col>
                </Row>
              ))}
              <Form.Item>
                <Button
                  type="dashed"
                  onClick={() => add()}
                  block
                  icon={<PlusOutlined />}
                >
                  添加节点选择器
                </Button>
              </Form.Item>
            </>
          )}
        </Form.List>
        
        <Divider orientation="left">注解</Divider>
        <Form.List name="annotations">
          {(fields, { add, remove }) => (
            <>
              {fields.map(field => (
                <Row key={field.key} gutter={8} align="middle">
                  <Col span={10}>
                    <Form.Item
                      key={field.key}
                      name={[field.name, 'key']}
                      rules={[{ required: true, message: '请输入注解键' }]}
                      style={{ marginBottom: 8 }}
                    >
                      <Input placeholder="注解键" />
                    </Form.Item>
                  </Col>
                  <Col span={10}>
                    <Form.Item
                      key={field.key}
                      name={[field.name, 'value']}
                      rules={[{ required: true, message: '请输入注解值' }]}
                      style={{ marginBottom: 8 }}
                    >
                      <Input placeholder="注解值" />
                    </Form.Item>
                  </Col>
                  <Col span={4}>
                    <Button
                      type="link"
                      icon={<MinusCircleOutlined />}
                      onClick={() => remove(field.name)}
                    />
                  </Col>
                </Row>
              ))}
              <Form.Item>
                <Button
                  type="dashed"
                  onClick={() => add()}
                  block
                  icon={<PlusOutlined />}
                >
                  添加注解
                </Button>
              </Form.Item>
            </>
          )}
        </Form.List>
      </>
    );
  }, []);
  
  // 渲染高级选项Tab
  const renderAdvancedTab = useCallback(() => {
    return (
      <>
        <Form.Item
          name="serviceAccountName"
          label="服务账号名称"
        >
          <Input placeholder="默认使用default服务账号" />
        </Form.Item>
        <Form.Item
          name="hostNetwork"
          label="使用主机网络"
          valuePropName="checked"
        >
          <Switch />
        </Form.Item>
        <Form.Item
          name="dnsPolicy"
          label="DNS策略"
        >
          <Select placeholder="ClusterFirst">
            <Option value="ClusterFirst">ClusterFirst</Option>
            <Option value="ClusterFirstWithHostNet">ClusterFirstWithHostNet</Option>
            <Option value="Default">Default</Option>
            <Option value="None">None</Option>
          </Select>
        </Form.Item>
      </>
    );
  }, []);
  
  // 定义Tab项
  const tabItems = [
    {
      key: 'basic',
      label: '基本信息',
      children: renderBasicTab()
    },
    {
      key: 'containers',
      label: '容器',
      children: renderContainersTab()
    },
    {
      key: 'volumes',
      label: '存储卷',
      children: renderVolumesTab()
    },
    {
      key: 'labelsAndAnnotations',
      label: '标签和注解',
      children: renderLabelsAndAnnotationsTab()
    },
    {
      key: 'advanced',
      label: '高级选项',
      children: renderAdvancedTab()
    }
  ];
  
  // 处理表单值变化
  const handleValuesChange = (changedValues: any, allValues: any) => {
    if (onFormValuesChange) {
      onFormValuesChange(changedValues, allValues);
    }
  };
  
  return (
    <Form 
      form={form} 
      layout="vertical" 
      initialValues={initialValues}
      onValuesChange={handleValuesChange}
    >
      <Tabs activeKey={activeTab} onChange={setActiveTab} items={tabItems} />
    </Form>
  );
};

export default DeploymentForm;