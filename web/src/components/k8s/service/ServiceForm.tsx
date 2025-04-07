import React from 'react';
import { Form, Input, Select, Button, Divider, Row, Col } from 'antd';
import { PlusOutlined, MinusCircleOutlined } from '@ant-design/icons';
import { ServiceFormProps } from './ServiceTypes';
import PortNameSelect from '../common/PortNameSelect';
import { useTranslation } from 'react-i18next';

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
  const { t } = useTranslation();
  
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
              label={t('services.form.name')}
              rules={[
                { required: true, message: t('services.form.pleaseEnterName') },
                { 
                  pattern: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/, 
                  message: t('services.form.namePattern') 
                }
              ]}
            >
              <Input placeholder="my-service" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item
              name="type"
              label={t('services.form.type')}
              rules={[{ required: true, message: t('services.form.pleaseSelectType') }]}
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
          label={t('services.form.type')}
          rules={[{ required: true, message: t('services.form.pleaseSelectType') }]}
        >
          <Select>
            <Option value="ClusterIP">ClusterIP</Option>
            <Option value="NodePort">NodePort</Option>
            <Option value="LoadBalancer">LoadBalancer</Option>
          </Select>
        </Form.Item>
      )}

      <Divider orientation="left">{t('services.form.portConfig')}</Divider>
      <Form.List 
        name="ports"
        rules={[
          {
            validator: async (_, ports) => {
              if (!ports || ports.length < 1) {
                return Promise.reject(new Error(t('services.form.atLeastOnePort')));
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
                    label={t('services.form.portName')}
                  >
                    <PortNameSelect placeholder={t('services.form.selectProtocol')} />
                  </Form.Item>
                </Col>
                <Col span={5}>
                  <Form.Item
                    key={field.key}
                    name={[field.name, 'port']}
                    label={t('services.form.servicePort')}
                    rules={[{ required: true, message: t('services.form.pleaseEnterServicePort') }]}
                  >
                    <Input type="number" placeholder="80" />
                  </Form.Item>
                </Col>
                <Col span={5}>
                  <Form.Item
                    key={field.key}
                    name={[field.name, 'targetPort']}
                    label={t('services.form.targetPort')}
                    rules={[{ required: true, message: t('services.form.pleaseEnterTargetPort') }]}
                  >
                    <Input type="number" placeholder="8080" />
                  </Form.Item>
                </Col>
                <Col span={4}>
                  <Form.Item
                    key={field.key}
                    name={[field.name, 'protocol']}
                    label={t('services.form.protocol')}
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
                      label={t('services.form.nodePort')}
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
                {t('services.form.addPort')}
              </Button>
            </Form.Item>
          </>
        )}
      </Form.List>

      <Divider orientation="left">{t('services.form.labels')}</Divider>
      <Form.List name="labelsArray">
        {(fields, { add, remove }) => (
          <>
            {fields.map((field) => (
              <Row key={field.key} gutter={8} align="middle">
                <Col span={11}>
                  <Form.Item
                    key={field.key}
                    name={[field.name, 'key']}
                    rules={[{ required: true, message: t('services.form.pleaseEnterLabelKey') }]}
                  >
                    <Input placeholder={t('services.form.labelKeyPlaceholder')} />
                  </Form.Item>
                </Col>
                <Col span={11}>
                  <Form.Item
                    key={field.key}
                    name={[field.name, 'value']}
                    rules={[{ required: true, message: t('services.form.pleaseEnterLabelValue') }]}
                  >
                    <Input placeholder={t('services.form.labelValuePlaceholder')} />
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
                {t('services.form.addLabel')}
              </Button>
            </Form.Item>
          </>
        )}
      </Form.List>

      <Divider orientation="left">{t('services.form.selectors')}</Divider>
      <Form.List name="selectorArray">
        {(fields, { add, remove }) => (
          <>
            {fields.map((field) => (
              <Row key={field.key} gutter={8} align="middle">
                <Col span={11}>
                  <Form.Item
                    key={field.key}
                    name={[field.name, 'key']}
                    rules={[{ required: true, message: t('services.form.pleaseEnterSelectorKey') }]}
                  >
                    <Input placeholder={t('services.form.selectorKeyPlaceholder')} />
                  </Form.Item>
                </Col>
                <Col span={11}>
                  <Form.Item
                    key={field.key}
                    name={[field.name, 'value']}
                    rules={[{ required: true, message: t('services.form.pleaseEnterSelectorValue') }]}
                  >
                    <Input placeholder={t('services.form.selectorValuePlaceholder')} />
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
                {t('services.form.addSelector')}
              </Button>
            </Form.Item>
          </>
        )}
      </Form.List>
    </Form>
  );
};

export default ServiceForm;