import React from 'react';
import { Form, Input, Select, Button, Divider, Row, Col } from 'antd';
import { PlusOutlined, MinusCircleOutlined } from '@ant-design/icons';
import { ServiceFormProps } from './ServiceTypes';
import PortNameSelect from '../common/PortNameSelect';

const { Option } = Select;

/**
 * 通用Service表单组件
 * 可用于创建和编辑Service
 */
const ServiceForm: React.FC<ServiceFormProps> = ({
  initialValues,
  onFormValuesChange,
  form,
  mode
}) => {
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
      {/* 在创建模式下才显示名称输入框 */}
      {mode === 'create' && (
        <Row gutter={16}>
          <Col span={12}>
            <Form.Item
              name="name"
              label="服务名称"
              rules={[
                { required: true, message: '请输入服务名称' },
                { 
                  pattern: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/, 
                  message: '名称必须由小写字母、数字和"-"组成，且以字母或数字开头和结尾' 
                }
              ]}
            >
              <Input placeholder="my-service" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item
              name="type"
              label="服务类型"
              rules={[{ required: true, message: '请选择服务类型' }]}
            >
              <Select>
                <Option value="ClusterIP">ClusterIP</Option>
                <Option value="NodePort">NodePort</Option>
                <Option value="LoadBalancer">LoadBalancer</Option>
              </Select>
            </Form.Item>
          </Col>
        </Row>
      )}

      {/* 在编辑模式下只显示服务类型 */}
      {mode === 'edit' && (
        <Form.Item
          name="type"
          label="服务类型"
          rules={[{ required: true, message: '请选择服务类型' }]}
        >
          <Select>
            <Option value="ClusterIP">ClusterIP</Option>
            <Option value="NodePort">NodePort</Option>
            <Option value="LoadBalancer">LoadBalancer</Option>
          </Select>
        </Form.Item>
      )}

      <Divider orientation="left">端口配置</Divider>
      <Form.List 
        name="ports"
        rules={[
          {
            validator: async (_, ports) => {
              if (!ports || ports.length < 1) {
                return Promise.reject(new Error('至少需要配置一个端口'));
              }
            },
          },
        ]}
      >
        {(fields, { add, remove }) => (
          <>
            {fields.map((field, index) => (
              <Row key={field.key} gutter={16} align="middle">
                <Col span={5}>
                  <Form.Item
                    key={field.key}
                    name={[field.name, 'name']}
                    label="名称"
                  >
                    <PortNameSelect placeholder="选择通信协议" />
                  </Form.Item>
                </Col>
                <Col span={5}>
                  <Form.Item
                    key={field.key}
                    name={[field.name, 'port']}
                    label="服务端口"
                    rules={[{ required: true, message: '请输入服务端口' }]}
                  >
                    <Input type="number" placeholder="80" />
                  </Form.Item>
                </Col>
                <Col span={5}>
                  <Form.Item
                    key={field.key}
                    name={[field.name, 'targetPort']}
                    label="目标端口"
                    rules={[{ required: true, message: '请输入目标端口' }]}
                  >
                    <Input type="number" placeholder="8080" />
                  </Form.Item>
                </Col>
                <Col span={4}>
                  <Form.Item
                    key={field.key}
                    name={[field.name, 'protocol']}
                    label="协议"
                  >
                    <Select defaultValue="TCP">
                      <Option value="TCP">TCP</Option>
                      <Option value="UDP">UDP</Option>
                      <Option value="SCTP">SCTP</Option>
                    </Select>
                  </Form.Item>
                </Col>
                {form.getFieldValue('type') === 'NodePort' && (
                  <Col span={4}>
                    <Form.Item
                      key={field.key}
                      name={[field.name, 'nodePort']}
                      label="节点端口"
                    >
                      <Input type="number" placeholder="30000-32767" />
                    </Form.Item>
                  </Col>
                )}
                <Col flex="none">
                  <Form.Item label=" ">
                    <Button
                      type="text"
                      danger
                      icon={<MinusCircleOutlined />}
                      onClick={() => remove(field.name)}
                      disabled={fields.length === 1}
                    />
                  </Form.Item>
                </Col>
              </Row>
            ))}
            <Form.Item>
              <Button type="dashed" onClick={() => add({ protocol: 'TCP' })} block icon={<PlusOutlined />}>
                添加端口
              </Button>
            </Form.Item>
          </>
        )}
      </Form.List>

      <Divider orientation="left">标签</Divider>
      <Form.List name="labelsArray">
        {(fields, { add, remove }) => (
          <>
            {fields.map((field) => (
              <Row key={field.key} gutter={8} align="middle">
                <Col span={11}>
                  <Form.Item
                    key={field.key}
                    name={[field.name, 'key']}
                    rules={[{ required: true, message: '请输入标签键' }]}
                  >
                    <Input placeholder="标签键" />
                  </Form.Item>
                </Col>
                <Col span={11}>
                  <Form.Item
                    key={field.key}
                    name={[field.name, 'value']}
                    rules={[{ required: true, message: '请输入标签值' }]}
                  >
                    <Input placeholder="标签值" />
                  </Form.Item>
                </Col>
                <Col span={2}>
                  <Button
                    type="text"
                    danger
                    icon={<MinusCircleOutlined />}
                    onClick={() => remove(field.name)}
                  />
                </Col>
              </Row>
            ))}
            <Form.Item>
              <Button type="dashed" onClick={() => add()} block icon={<PlusOutlined />}>
                添加标签
              </Button>
            </Form.Item>
          </>
        )}
      </Form.List>

      <Divider orientation="left">选择器</Divider>
      <Form.List name="selectorArray">
        {(fields, { add, remove }) => (
          <>
            {fields.map((field) => (
              <Row key={field.key} gutter={8} align="middle">
                <Col span={11}>
                  <Form.Item
                    key={field.key}
                    name={[field.name, 'key']}
                    rules={[{ required: true, message: '请输入选择器键' }]}
                  >
                    <Input placeholder="app" />
                  </Form.Item>
                </Col>
                <Col span={11}>
                  <Form.Item
                    key={field.key}
                    name={[field.name, 'value']}
                    rules={[{ required: true, message: '请输入选择器值' }]}
                  >
                    <Input placeholder="nginx" />
                  </Form.Item>
                </Col>
                <Col span={2}>
                  <Button
                    type="text"
                    danger
                    icon={<MinusCircleOutlined />}
                    onClick={() => remove(field.name)}
                  />
                </Col>
              </Row>
            ))}
            <Form.Item>
              <Button type="dashed" onClick={() => add()} block icon={<PlusOutlined />}>
                添加选择器
              </Button>
            </Form.Item>
          </>
        )}
      </Form.List>
    </Form>
  );
};

export default ServiceForm;