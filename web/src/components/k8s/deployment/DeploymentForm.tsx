import React, { useCallback } from 'react';
import { 
  Form, Input, InputNumber, Button, Tabs,
  Card, Row, Col, Select, Switch, Divider, 
  Space
} from 'antd';
import { PlusOutlined, MinusCircleOutlined } from '@ant-design/icons';
import { DeploymentFormProps } from './DeploymentTypes';
import PortNameSelect from '../common/PortNameSelect';
import NodeAffinityManager from './NodeAffinityManager';
import { useTranslation } from 'react-i18next';

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
  const { t } = useTranslation();
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
                label={t('deployments.form.name')}
                rules={[
                  { required: true, message: t('deployments.form.pleaseEnterName') },
                  { pattern: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/, message: t('deployments.form.namePattern') }
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
              label={t('deployments.form.replicas')}
              rules={[{ required: true, message: t('deployments.form.pleaseEnterReplicas') }]}
            >
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item
              name="strategy"
              label={t('deployments.form.strategy')}
              rules={[{ required: true, message: t('deployments.form.pleaseSelectStrategy') }]}
            >
              <Select>
                <Option value="RollingUpdate">{t('deployments.form.strategies.rollingUpdate')}</Option>
                <Option value="Recreate">{t('deployments.form.strategies.recreate')}</Option>
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
                  <Form.Item name="maxSurgeValue" label={t('deployments.form.maxSurge')}>
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
                  <Form.Item name="maxUnavailableValue" label={t('deployments.form.maxUnavailable')}>
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
            <Form.Item name="minReadySeconds" label={t('deployments.form.minReadySeconds')}>
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item name="revisionHistoryLimit" label={t('deployments.form.revisionHistoryLimit')}>
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item name="paused" label={t('deployments.form.pauseDeploy')} valuePropName="checked">
              <Switch />
            </Form.Item>
          </Col>
        </Row>
      </>
    );
  }, [mode, t]);
  
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
                        label={t('deployments.form.container.name')}
                        rules={[
                          { required: true, message: t('deployments.form.container.pleaseEnterName') },
                          { pattern: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/, message: t('deployments.form.namePattern') }
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
                    label={t('deployments.form.container.image')}
                    rules={[{ required: true, message: t('deployments.form.container.pleaseEnterImage') }]}
                  >
                    <Input placeholder={t('deployments.form.container.imagePlaceholder')} />
                  </Form.Item>
                  
                  <Divider orientation="left">{t('deployments.form.container.resources')}</Divider>
                  <Row gutter={16}>
                    <Col span={12}>
                      <Form.Item
                        key={`${field.key}-resources-requests-cpu`}
                        name={[field.name, 'resources', 'requests', 'cpu']}
                        label={t('deployments.form.container.cpuRequest')}
                      >
                        <Space.Compact style={{ width: '100%' }}>
                          <Form.Item
                            key={`${field.key}-resources-requests-cpuValue`}
                            name={[field.name, 'resources', 'requests', 'cpuValue']}
                            noStyle
                            rules={[{ required: true, message: t('deployments.form.container.pleaseEnterCpuRequest') }]}
                          >
                            <InputNumber 
                              style={{ width: '60%' }} 
                              min={0} 
                              placeholder={t('deployments.form.container.enterValue')}
                            />
                          </Form.Item>
                          <Form.Item
                            key={`${field.key}-resources-requests-cpuUnit`}
                            name={[field.name, 'resources', 'requests', 'cpuUnit']}
                            noStyle
                            initialValue="m"
                          >
                            <Select style={{ width: '40%' }}>
                              <Option value="m">{t('deployments.form.container.cpuUnits.millicore')}</Option>
                              <Option value="">{t('deployments.form.container.cpuUnits.core')}</Option>
                            </Select>
                          </Form.Item>
                        </Space.Compact>
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        key={`${field.key}-resources-limits-cpu`}
                        name={[field.name, 'resources', 'limits', 'cpu']}
                        label={t('deployments.form.container.cpuLimit')}
                      >
                        <Space.Compact style={{ width: '100%' }}>
                          <Form.Item
                            key={`${field.key}-resources-limits-cpuValue`}
                            name={[field.name, 'resources', 'limits', 'cpuValue']}
                            noStyle
                            rules={[{ required: true, message: t('deployments.form.container.pleaseEnterCpuLimit') }]}
                          >
                            <InputNumber 
                              style={{ width: '60%' }} 
                              min={0} 
                              placeholder={t('deployments.form.container.enterValue')}
                            />
                          </Form.Item>
                          <Form.Item
                            key={`${field.key}-resources-limits-cpuUnit`}
                            name={[field.name, 'resources', 'limits', 'cpuUnit']}
                            noStyle
                            initialValue="m"
                          >
                            <Select style={{ width: '40%' }}>
                              <Option value="m">{t('deployments.form.container.cpuUnits.millicore')}</Option>
                              <Option value="">{t('deployments.form.container.cpuUnits.core')}</Option>
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
                        label={t('deployments.form.container.memoryRequest')}
                      >
                        <Space.Compact style={{ width: '100%' }}>
                          <Form.Item
                            key={`${field.key}-resources-requests-memoryValue`}
                            name={[field.name, 'resources', 'requests', 'memoryValue']}
                            noStyle
                            rules={[{ required: true, message: t('deployments.form.container.pleaseEnterMemoryRequest') }]}
                          >
                            <InputNumber 
                              style={{ width: '60%' }} 
                              min={0} 
                              placeholder={t('deployments.form.container.enterValue')}
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
                        label={t('deployments.form.container.memoryLimit')}
                      >
                        <Space.Compact style={{ width: '100%' }}>
                          <Form.Item
                            key={`${field.key}-resources-limits-memoryValue`}
                            name={[field.name, 'resources', 'limits', 'memoryValue']}
                            noStyle
                            rules={[{ required: true, message: t('deployments.form.container.pleaseEnterMemoryLimit') }]}
                          >
                            <InputNumber 
                              style={{ width: '60%' }} 
                              min={0} 
                              placeholder={t('deployments.form.container.enterValue')}
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

                  <Divider orientation="left">{t('deployments.form.container.ports')}</Divider>
                  <Form.List name={[field.name, 'ports']}>
                    {(portFields, { addPort, removePort }) => (
                      <>
                        {portFields.map(portField => (
                          <Row key={`port-${field.key}-${portField.key}`} gutter={16} align="middle">
                            <Col span={6}>
                              <Form.Item
                                {...portField}
                                name={[portField.name, 'name']}
                                label={t('deployments.form.container.portName')}
                              >
                                <PortNameSelect placeholder={t('deployments.form.container.selectProtocol')} />
                              </Form.Item>
                            </Col>
                            <Col span={6}>
                              <Form.Item
                                {...portField}
                                name={[portField.name, 'containerPort']}
                                label={t('deployments.form.container.containerPort')}
                                rules={[{ required: true, message: t('deployments.form.container.pleaseEnterPort') }]}
                              >
                                <InputNumber min={1} max={65535} style={{ width: '100%' }} />
                              </Form.Item>
                            </Col>
                            <Col span={6}>
                              <Form.Item
                                {...portField}
                                name={[portField.name, 'protocol']}
                                label={t('deployments.form.container.protocol')}
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
                            {t('deployments.form.container.addPort')}
                          </Button>
                        </Form.Item>
                      </>
                    )}
                  </Form.List>
                  
                  <Divider orientation="left">{t('deployments.form.container.environmentVariables')}</Divider>
                  <Form.List name={[field.name, 'env']}>
                    {(envFields, { addEnv, removeEnv }) => (
                      <>
                        {envFields.map(envField => (
                          <Row key={`env-${field.key}-${envField.key}`} gutter={8} align="middle">
                            <Col span={10}>
                              <Form.Item
                                {...envField}
                                name={[envField.name, 'name']}
                                rules={[{ required: true, message: t('deployments.form.container.pleaseEnterVarName') }]}
                                style={{ marginBottom: 8 }}
                              >
                                <Input placeholder={t('deployments.form.container.variableName')} />
                              </Form.Item>
                            </Col>
                            <Col span={10}>
                              <Form.Item
                                {...envField}
                                name={[envField.name, 'value']}
                                rules={[{ required: true, message: t('deployments.form.container.pleaseEnterVarValue') }]}
                                style={{ marginBottom: 8 }}
                              >
                                <Input placeholder={t('deployments.form.container.variableValue')} />
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
                            {t('deployments.form.container.addEnvVar')}
                          </Button>
                        </Form.Item>
                      </>
                    )}
                  </Form.List>
                  
                  <Divider orientation="left">{t('deployments.form.container.healthChecks')}</Divider>
                  
                  {/* 存活探针配置 */}
                  <Card
                    size="small"
                    title={t('deployments.form.container.probes.liveness')}
                    style={{ marginBottom: 16 }}
                    styles={{ body: { padding: '12px' } }}
                  >
                    <Form.Item
                      name={[field.name, 'livenessProbe', 'type']}
                      label={t('deployments.form.container.probes.type')}
                    >
                      <Select placeholder={t('deployments.form.container.probes.selectType')}>
                        <Option value="httpGet">HTTP GET</Option>
                        <Option value="tcpSocket">TCP Socket</Option>
                        <Option value="exec">{t('deployments.form.container.probes.command')}</Option>
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
                                  label={t('deployments.form.container.probes.path')}
                                  rules={[{ required: true, message: t('deployments.form.container.probes.pleaseEnterPath') }]}
                                >
                                  <Input placeholder="/health" />
                                </Form.Item>
                                <Form.Item
                                  name={[field.name, 'livenessProbe', 'httpGet', 'port']}
                                  label={t('deployments.form.container.probes.port')}
                                  rules={[{ required: true, message: t('deployments.form.container.probes.pleaseEnterPort') }]}
                                >
                                  <InputNumber min={1} max={65535} />
                                </Form.Item>
                                <Form.Item
                                  name={[field.name, 'livenessProbe', 'httpGet', 'scheme']}
                                  label={t('deployments.form.container.probes.scheme')}
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
                                label={t('deployments.form.container.probes.port')}
                                rules={[{ required: true, message: t('deployments.form.container.probes.pleaseEnterPort') }]}
                              >
                                <InputNumber min={1} max={65535} />
                              </Form.Item>
                            );
                          case 'exec':
                            return (
                              <Form.Item
                                name={[field.name, 'livenessProbe', 'exec', 'command']}
                                label={t('deployments.form.container.probes.command')}
                                rules={[{ required: true, message: t('deployments.form.container.probes.pleaseEnterCommand') }]}
                              >
                                <Input.TextArea placeholder={t('deployments.form.container.probes.commandPlaceholder')} />
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
                          label={t('deployments.form.container.probes.initialDelay')}
                        >
                          <InputNumber min={0} />
                        </Form.Item>
                      </Col>
                      <Col span={8}>
                        <Form.Item
                          name={[field.name, 'livenessProbe', 'periodSeconds']}
                          label={t('deployments.form.container.probes.period')}
                        >
                          <InputNumber min={1} />
                        </Form.Item>
                      </Col>
                      <Col span={8}>
                        <Form.Item
                          name={[field.name, 'livenessProbe', 'timeoutSeconds']}
                          label={t('deployments.form.container.probes.timeout')}
                        >
                          <InputNumber min={1} />
                        </Form.Item>
                      </Col>
                    </Row>
                    <Row gutter={16}>
                      <Col span={12}>
                        <Form.Item
                          name={[field.name, 'livenessProbe', 'successThreshold']}
                          label={t('deployments.form.container.probes.successThreshold')}
                        >
                          <InputNumber min={1} />
                        </Form.Item>
                      </Col>
                      <Col span={12}>
                        <Form.Item
                          name={[field.name, 'livenessProbe', 'failureThreshold']}
                          label={t('deployments.form.container.probes.failureThreshold')}
                        >
                          <InputNumber min={1} />
                        </Form.Item>
                      </Col>
                    </Row>
                  </Card>
                  
                  {/* 就绪探针配置 */}
                  <Card
                    size="small"
                    title={t('deployments.form.container.probes.readiness')}
                    style={{ marginBottom: 16 }}
                    styles={{ body: { padding: '12px' } }}
                  >
                    <Form.Item
                      name={[field.name, 'readinessProbe', 'type']}
                      label={t('deployments.form.container.probes.type')}
                    >
                      <Select placeholder={t('deployments.form.container.probes.selectType')}>
                        <Option value="httpGet">HTTP GET</Option>
                        <Option value="tcpSocket">TCP Socket</Option>
                        <Option value="exec">{t('deployments.form.container.probes.command')}</Option>
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
                                  label={t('deployments.form.container.probes.path')}
                                  rules={[{ required: true, message: t('deployments.form.container.probes.pleaseEnterPath') }]}
                                >
                                  <Input placeholder="/ready" />
                                </Form.Item>
                                <Form.Item
                                  name={[field.name, 'readinessProbe', 'httpGet', 'port']}
                                  label={t('deployments.form.container.probes.port')}
                                  rules={[{ required: true, message: t('deployments.form.container.probes.pleaseEnterPort') }]}
                                >
                                  <InputNumber min={1} max={65535} />
                                </Form.Item>
                                <Form.Item
                                  name={[field.name, 'readinessProbe', 'httpGet', 'scheme']}
                                  label={t('deployments.form.container.probes.scheme')}
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
                                label={t('deployments.form.container.probes.port')}
                                rules={[{ required: true, message: t('deployments.form.container.probes.pleaseEnterPort') }]}
                              >
                                <InputNumber min={1} max={65535} />
                              </Form.Item>
                            );
                          case 'exec':
                            return (
                              <Form.Item
                                name={[field.name, 'readinessProbe', 'exec', 'command']}
                                label={t('deployments.form.container.probes.command')}
                                rules={[{ required: true, message: t('deployments.form.container.probes.pleaseEnterCommand') }]}
                              >
                                <Input.TextArea placeholder={t('deployments.form.container.probes.commandPlaceholder')} />
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
                          label={t('deployments.form.container.probes.initialDelay')}
                        >
                          <InputNumber min={0} />
                        </Form.Item>
                      </Col>
                      <Col span={8}>
                        <Form.Item
                          name={[field.name, 'readinessProbe', 'periodSeconds']}
                          label={t('deployments.form.container.probes.period')}
                        >
                          <InputNumber min={1} />
                        </Form.Item>
                      </Col>
                      <Col span={8}>
                        <Form.Item
                          name={[field.name, 'readinessProbe', 'timeoutSeconds']}
                          label={t('deployments.form.container.probes.timeout')}
                        >
                          <InputNumber min={1} />
                        </Form.Item>
                      </Col>
                    </Row>
                    <Row gutter={16}>
                      <Col span={12}>
                        <Form.Item
                          name={[field.name, 'readinessProbe', 'successThreshold']}
                          label={t('deployments.form.container.probes.successThreshold')}
                        >
                          <InputNumber min={1} />
                        </Form.Item>
                      </Col>
                      <Col span={12}>
                        <Form.Item
                          name={[field.name, 'readinessProbe', 'failureThreshold']}
                          label={t('deployments.form.container.probes.failureThreshold')}
                        >
                          <InputNumber min={1} />
                        </Form.Item>
                      </Col>
                    </Row>
                  </Card>
                  
                  {/* 启动探针配置 */}
                  <Card
                    size="small"
                    title={t('deployments.form.container.probes.startup')}
                    style={{ marginBottom: 16 }}
                    styles={{ body: { padding: '12px' } }}
                  >
                    <Form.Item
                      name={[field.name, 'startupProbe', 'type']}
                      label={t('deployments.form.container.probes.type')}
                    >
                      <Select placeholder={t('deployments.form.container.probes.selectType')}>
                        <Option value="httpGet">HTTP GET</Option>
                        <Option value="tcpSocket">TCP Socket</Option>
                        <Option value="exec">{t('deployments.form.container.probes.command')}</Option>
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
                                  label={t('deployments.form.container.probes.path')}
                                  rules={[{ required: true, message: t('deployments.form.container.probes.pleaseEnterPath') }]}
                                >
                                  <Input placeholder="/startup" />
                                </Form.Item>
                                <Form.Item
                                  name={[field.name, 'startupProbe', 'httpGet', 'port']}
                                  label={t('deployments.form.container.probes.port')}
                                  rules={[{ required: true, message: t('deployments.form.container.probes.pleaseEnterPort') }]}
                                >
                                  <InputNumber min={1} max={65535} />
                                </Form.Item>
                                <Form.Item
                                  name={[field.name, 'startupProbe', 'httpGet', 'scheme']}
                                  label={t('deployments.form.container.probes.scheme')}
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
                                label={t('deployments.form.container.probes.port')}
                                rules={[{ required: true, message: t('deployments.form.container.probes.pleaseEnterPort') }]}
                              >
                                <InputNumber min={1} max={65535} />
                              </Form.Item>
                            );
                          case 'exec':
                            return (
                              <Form.Item
                                name={[field.name, 'startupProbe', 'exec', 'command']}
                                label={t('deployments.form.container.probes.command')}
                                rules={[{ required: true, message: t('deployments.form.container.probes.pleaseEnterCommand') }]}
                              >
                                <Input.TextArea placeholder={t('deployments.form.container.probes.commandPlaceholder')} />
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
                          label={t('deployments.form.container.probes.initialDelay')}
                        >
                          <InputNumber min={0} />
                        </Form.Item>
                      </Col>
                      <Col span={8}>
                        <Form.Item
                          name={[field.name, 'startupProbe', 'periodSeconds']}
                          label={t('deployments.form.container.probes.period')}
                        >
                          <InputNumber min={1} />
                        </Form.Item>
                      </Col>
                      <Col span={8}>
                        <Form.Item
                          name={[field.name, 'startupProbe', 'timeoutSeconds']}
                          label={t('deployments.form.container.probes.timeout')}
                        >
                          <InputNumber min={1} />
                        </Form.Item>
                      </Col>
                    </Row>
                    <Row gutter={16}>
                      <Col span={12}>
                        <Form.Item
                          name={[field.name, 'startupProbe', 'successThreshold']}
                          label={t('deployments.form.container.probes.successThreshold')}
                        >
                          <InputNumber min={1} />
                        </Form.Item>
                      </Col>
                      <Col span={12}>
                        <Form.Item
                          name={[field.name, 'startupProbe', 'failureThreshold']}
                          label={t('deployments.form.container.probes.failureThreshold')}
                        >
                          <InputNumber min={1} />
                        </Form.Item>
                      </Col>
                    </Row>
                  </Card>

                  {/* 卷挂载配置 */}
                  {form.getFieldValue('volumes')?.length > 0 && (
                    <>
                      <Divider orientation="left">{t('deployments.form.container.volumeMounts')}</Divider>
                      <Form.List name={[field.name, 'volumeMounts']}>
                        {(mountFields, { add: addMount, remove: removeMount }) => (
                          <>
                            {mountFields.map(mountField => (
                              <Row key={`mount-${field.key}-${mountField.key}`} gutter={16} align="middle">
                                <Col span={6}>
                                  <Form.Item
                                    key={`mount-${field.key}-${mountField.key}-name`}
                                    name={[mountField.name, 'name']}
                                    label={t('deployments.form.container.volumeMount.name')}
                                    rules={[{ required: true, message: t('deployments.form.container.volumeMount.pleaseSelectVolume') }]}
                                  >
                                    <Select placeholder={t('deployments.form.container.volumeMount.selectVolume')}>
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
                                    label={t('deployments.form.container.volumeMount.mountPath')}
                                    rules={[{ required: true, message: t('deployments.form.container.volumeMount.pleaseEnterMountPath') }]}
                                  >
                                    <Input placeholder="/data" />
                                  </Form.Item>
                                </Col>
                                <Col span={6}>
                                  <Form.Item
                                    key={`mount-${field.key}-${mountField.key}-subPath`}
                                    name={[mountField.name, 'subPath']}
                                    label={t('deployments.form.container.volumeMount.subPath')}
                                  >
                                    <Input placeholder={t('deployments.form.container.volumeMount.optional')} />
                                  </Form.Item>
                                </Col>
                                <Col span={3}>
                                  <Form.Item
                                    key={`mount-${field.key}-${mountField.key}-readOnly`}
                                    name={[mountField.name, 'readOnly']}
                                    label={t('deployments.form.container.volumeMount.readOnly')}
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
                                {t('deployments.form.container.volumeMount.add')}
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
                  {t('deployments.form.container.add')}
                </Button>
              </Form.Item>
            )}
          </div>
        )}
      </Form.List>
    );
  }, [form, mode, t]);
  
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
                    rules={[{ required: true, message: t('deployments.form.volume.pleaseEnterName') }]}
                    style={{ marginBottom: 0 }}
                  >
                    <Input placeholder={t('deployments.form.volume.name')} />
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
                  label={t('deployments.form.volume.type')}
                  rules={[{ required: true, message: t('deployments.form.volume.pleaseSelectType') }]}
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
                              label={t('deployments.form.volume.configMap.name')}
                              rules={[{ required: true, message: t('deployments.form.volume.configMap.pleaseEnterName') }]}
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
                                          label={t('deployments.form.volume.key')}
                                          rules={[{ required: true, message: t('deployments.form.volume.pleaseEnterKey') }]}
                                        >
                                          <Input placeholder="config-key" />
                                        </Form.Item>
                                      </Col>
                                      <Col span={8}>
                                        <Form.Item
                                          key={itemField.key}
                                          name={[itemField.name, 'path']}
                                          label={t('deployments.form.volume.path')}
                                          rules={[{ required: true, message: t('deployments.form.volume.pleaseEnterPath') }]}
                                        >
                                          <Input placeholder="config.yaml" />
                                        </Form.Item>
                                      </Col>
                                      <Col span={7}>
                                        <Form.Item
                                          key={itemField.key}
                                          name={[itemField.name, 'mode']}
                                          label={t('deployments.form.volume.mode')}
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
                                      {t('deployments.form.volume.configMap.addMapping')}
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
                              label={t('deployments.form.volume.secret.name')}
                              rules={[{ required: true, message: t('deployments.form.volume.secret.pleaseEnterName') }]}
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
                                          label={t('deployments.form.volume.key')}
                                          rules={[{ required: true, message: t('deployments.form.volume.pleaseEnterKey') }]}
                                        >
                                          <Input placeholder="secret-key" />
                                        </Form.Item>
                                      </Col>
                                      <Col span={8}>
                                        <Form.Item
                                          key={itemField.key}
                                          name={[itemField.name, 'path']}
                                          label={t('deployments.form.volume.path')}
                                          rules={[{ required: true, message: t('deployments.form.volume.pleaseEnterPath') }]}
                                        >
                                          <Input placeholder="secret.txt" />
                                        </Form.Item>
                                      </Col>
                                      <Col span={7}>
                                        <Form.Item
                                          key={itemField.key}
                                          name={[itemField.name, 'mode']}
                                          label={t('deployments.form.volume.mode')}
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
                                      {t('deployments.form.volume.secret.addMapping')}
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
                              label={t('deployments.form.volume.pvc.name')}
                              rules={[{ required: true, message: t('deployments.form.volume.pvc.pleaseEnterName') }]}
                            >
                              <Input placeholder="my-pvc" />
                            </Form.Item>
                            <Form.Item
                              key={field.key}
                              name={[field.name, 'persistentVolumeClaim', 'readOnly']}
                              label={t('deployments.form.volume.pvc.readOnly')}
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
                              label={t('deployments.form.volume.emptyDir.medium')}
                            >
                              <Select placeholder={t('deployments.form.volume.emptyDir.default')}>
                                <Option value="">{t('deployments.form.volume.emptyDir.default')}</Option>
                                <Option value="Memory">{t('deployments.form.volume.emptyDir.memory')}</Option>
                              </Select>
                            </Form.Item>
                            <Form.Item
                              key={field.key}
                              name={[field.name, 'emptyDir', 'sizeLimit']}
                              label={t('deployments.form.volume.emptyDir.sizeLimit')}
                            >
                              <Input placeholder={t('deployments.form.volume.emptyDir.sizeLimitPlaceholder')} />
                            </Form.Item>
                          </>
                        );
                      case 'hostPath':
                        return (
                          <>
                            <Form.Item
                              key={field.key}
                              name={[field.name, 'hostPath', 'path']}
                              label={t('deployments.form.volume.hostPath.path')}
                              rules={[{ required: true, message: t('deployments.form.volume.hostPath.pleaseEnterPath') }]}
                            >
                              <Input placeholder="/data" />
                            </Form.Item>
                            <Form.Item
                              key={field.key}
                              name={[field.name, 'hostPath', 'type']}
                              label={t('deployments.form.volume.hostPath.type')}
                            >
                              <Select placeholder={t('deployments.form.volume.hostPath.default')}>
                                <Option value="">{t('deployments.form.volume.hostPath.default')}</Option>
                                <Option value="Directory">{t('deployments.form.volume.hostPath.directory')}</Option>
                                <Option value="DirectoryOrCreate">{t('deployments.form.volume.hostPath.directoryOrCreate')}</Option>
                                <Option value="File">{t('deployments.form.volume.hostPath.file')}</Option>
                                <Option value="FileOrCreate">{t('deployments.form.volume.hostPath.fileOrCreate')}</Option>
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
                {t('deployments.form.volume.add')}
              </Button>
            </Form.Item>
          </>
        )}
      </Form.List>
    );
  }, [t]);
  
  // 渲染标签和注解Tab
  const renderLabelsAndAnnotationsTab = useCallback(() => {
    return (
      <>
        <Divider orientation="left">{t('deployments.labels')}</Divider>
        <Form.List name="labels">
          {(fields, { add, remove }) => (
            <>
              {fields.map(field => (
                <Row key={field.key} gutter={8} align="middle">
                  <Col span={10}>
                    <Form.Item
                      key={field.key}
                      name={[field.name, 'key']}
                      rules={[{ required: true, message: t('deployments.form.labels.pleaseEnterKey') }]}
                      style={{ marginBottom: 8 }}
                    >
                      <Input placeholder={t('deployments.form.labels.keyPlaceholder')} />
                    </Form.Item>
                  </Col>
                  <Col span={10}>
                    <Form.Item
                      key={field.key}
                      name={[field.name, 'value']}
                      rules={[{ required: true, message: t('deployments.form.labels.pleaseEnterValue') }]}
                      style={{ marginBottom: 8 }}
                    >
                      <Input placeholder={t('deployments.form.labels.valuePlaceholder')} />
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
                  {t('deployments.form.labels.add')}
                </Button>
              </Form.Item>
            </>
          )}
        </Form.List>
        
        <Divider orientation="left">{t('deployments.form.nodeSelector.title')}</Divider>
        <Form.List name="nodeSelector">
          {(fields, { add, remove }) => (
            <>
              {fields.map(field => (
                <Row key={field.key} gutter={8} align="middle">
                  <Col span={10}>
                    <Form.Item
                      key={field.key}
                      name={[field.name, 'key']}
                      rules={[{ required: true, message: t('deployments.form.nodeSelector.pleaseEnterKey') }]}
                      style={{ marginBottom: 8 }}
                    >
                      <Input placeholder={t('deployments.form.nodeSelector.keyPlaceholder')} />
                    </Form.Item>
                  </Col>
                  <Col span={10}>
                    <Form.Item
                      key={field.key}
                      name={[field.name, 'value']}
                      rules={[{ required: true, message: t('deployments.form.nodeSelector.pleaseEnterValue') }]}
                      style={{ marginBottom: 8 }}
                    >
                      <Input placeholder={t('deployments.form.nodeSelector.valuePlaceholder')} />
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
                  {t('deployments.form.nodeSelector.add')}
                </Button>
              </Form.Item>
            </>
          )}
        </Form.List>
        
        <Divider orientation="left">{t('deployments.annotations')}</Divider>
        <Form.List name="annotations">
          {(fields, { add, remove }) => (
            <>
              {fields.map(field => (
                <Row key={field.key} gutter={8} align="middle">
                  <Col span={10}>
                    <Form.Item
                      key={field.key}
                      name={[field.name, 'key']}
                      rules={[{ required: true, message: t('deployments.form.annotations.pleaseEnterKey') }]}
                      style={{ marginBottom: 8 }}
                    >
                      <Input placeholder={t('deployments.form.annotations.keyPlaceholder')} />
                    </Form.Item>
                  </Col>
                  <Col span={10}>
                    <Form.Item
                      key={field.key}
                      name={[field.name, 'value']}
                      rules={[{ required: true, message: t('deployments.form.annotations.pleaseEnterValue') }]}
                      style={{ marginBottom: 8 }}
                    >
                      <Input placeholder={t('deployments.form.annotations.valuePlaceholder')} />
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
                  {t('deployments.form.annotations.add')}
                </Button>
              </Form.Item>
            </>
          )}
        </Form.List>
      </>
    );
  }, [t]);
  
  // 渲染高级选项Tab
  const renderAdvancedTab = useCallback(() => {
    return (
      <>
        <Form.Item
          name="serviceAccountName"
          label={t('deployments.form.advanced.serviceAccountName')}
        >
          <Input placeholder={t('deployments.form.advanced.serviceAccountPlaceholder')} />
        </Form.Item>
        <Form.Item
          name="hostNetwork"
          label={t('deployments.form.advanced.hostNetwork')}
          valuePropName="checked"
        >
          <Switch />
        </Form.Item>
        <Form.Item
          name="dnsPolicy"
          label={t('deployments.form.advanced.dnsPolicy')}
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
  }, [t]);
  
  // 渲染节点亲和性Tab
  const renderNodeAffinityTab = useCallback(() => {
    return (
      <Form.Item name="nodeAffinity">
        <NodeAffinityManager />
      </Form.Item>
    );
  }, []);
  
  // 定义Tab项
  const tabItems = [
    {
      key: 'basic',
      label: t('deployments.form.tabs.basic'),
      children: renderBasicTab()
    },
    {
      key: 'containers',
      label: t('deployments.form.tabs.containers'),
      children: renderContainersTab()
    },
    {
      key: 'volumes',
      label: t('deployments.form.tabs.volumes'),
      children: renderVolumesTab()
    },
    {
      key: 'labelsAndAnnotations',
      label: t('deployments.form.tabs.labelsAndAnnotations'),
      children: renderLabelsAndAnnotationsTab()
    },
    {
      key: 'nodeAffinity',
      label: t('deployments.form.tabs.nodeAffinity'),
      children: renderNodeAffinityTab()
    },
    {
      key: 'advanced',
      label: t('deployments.form.tabs.advanced'),
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