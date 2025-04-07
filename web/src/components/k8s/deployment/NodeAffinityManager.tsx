import React from 'react';
import { 
  Card, Form, Input, Button, Select, InputNumber, Space, 
  Row, Col, Divider, Alert, Typography 
} from 'antd';
import { PlusOutlined, MinusCircleOutlined, InfoCircleOutlined } from '@ant-design/icons';
import { NodeAffinityConfig, NodeSelectorTerm, PreferredNodeAffinity } from './DeploymentTypes';
import { useTranslation } from 'react-i18next';

const { Text, Paragraph } = Typography;
const { Option } = Select;

interface NodeAffinityManagerProps {
  value?: NodeAffinityConfig;
  onChange?: (value: NodeAffinityConfig) => void;
}

/**
 * 节点亲和性管理组件
 * 用于管理Deployment的节点亲和性规则
 */
const NodeAffinityManager: React.FC<NodeAffinityManagerProps> = ({ value, onChange }) => {
  const { t } = useTranslation();
  
  const defaultValue: NodeAffinityConfig = {
    requiredTerms: [],
    preferredTerms: []
  };

  const affinityConfig = value || defaultValue;

  const triggerChange = (changedValue: Partial<NodeAffinityConfig>) => {
    onChange?.({
      ...affinityConfig,
      ...changedValue
    });
  };

  // 渲染表达式匹配规则
  const renderExpressionItem = (field: any, prefix: string, remove: (name: number) => void) => {
    return (
      <Row key={field.key} gutter={[16, 0]} align="middle">
        <Col span={6}>
          <Form.Item
            name={[field.name, 'key']}
            rules={[{ required: true, message: t('deployments.form.nodeAffinity.pleaseEnterKey') }]}
            style={{ marginBottom: 8 }}
          >
            <Input placeholder={t('deployments.form.nodeAffinity.keyPlaceholder')} />
          </Form.Item>
        </Col>
        <Col span={5}>
          <Form.Item
            name={[field.name, 'operator']}
            rules={[{ required: true, message: t('deployments.form.nodeAffinity.pleaseSelectOperator') }]}
            style={{ marginBottom: 8 }}
            initialValue="In"
          >
            <Select>
              <Option value="In">In</Option>
              <Option value="NotIn">NotIn</Option>
              <Option value="Exists">Exists</Option>
              <Option value="DoesNotExist">DoesNotExist</Option>
              <Option value="Gt">Gt</Option>
              <Option value="Lt">Lt</Option>
            </Select>
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item 
            shouldUpdate={(prevValues, curValues) => {
              const prevOp = prevValues[prefix]?.[field.name]?.operator;
              const curOp = curValues[prefix]?.[field.name]?.operator;
              return prevOp !== curOp;
            }}
            style={{ marginBottom: 8 }}
          >
            {({ getFieldValue }) => {
              const operator = getFieldValue([prefix, field.name, 'operator']);
              if (operator === 'Exists' || operator === 'DoesNotExist') {
                return (
                  <Text type="secondary" style={{ lineHeight: '32px' }}>
                    {operator === 'Exists' 
                      ? t('deployments.form.nodeAffinity.keyExists') 
                      : t('deployments.form.nodeAffinity.keyDoesNotExist')}
                  </Text>
                );
              }
              return (
                <Form.Item
                  name={[field.name, 'values']}
                  rules={[{ required: true, message: t('deployments.form.nodeAffinity.pleaseEnterValue') }]}
                  style={{ marginBottom: 0 }}
                >
                  <Select
                    mode="tags"
                    placeholder={t('deployments.form.nodeAffinity.valuesPlaceholder')}
                    tokenSeparators={[',']}
                  />
                </Form.Item>
              );
            }}
          </Form.Item>
        </Col>
        <Col span={1}>
          <Button
            type="link"
            icon={<MinusCircleOutlined />}
            onClick={() => remove(field.name)}
          />
        </Col>
      </Row>
    );
  };

  // 渲染必须满足的节点选择规则
  const renderRequiredTerms = () => {
    return (
      <Card
        title={t('deployments.form.nodeAffinity.requiredRules')}
        bordered={false}
        extra={
          <Button 
            type="link" 
            icon={<InfoCircleOutlined />} 
            onClick={() => {}}
          >
            {t('common.help')}
          </Button>
        }
      >
        <Alert
          message={t('deployments.form.nodeAffinity.requiredRulesDescription')}
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
        />
        <Form.List name="nodeAffinity.requiredTerms">
          {(fields, { add, remove }) => (
            <>
              {fields.map(field => (
                <Card
                  key={field.key}
                  type="inner"
                  title={`${t('deployments.form.nodeAffinity.ruleGroup')} ${field.name + 1}`}
                  style={{ marginBottom: 16 }}
                  extra={
                    <Button
                      type="link"
                      danger
                      onClick={() => remove(field.name)}
                    >
                      {t('deployments.form.nodeAffinity.deleteRuleGroup')}
                    </Button>
                  }
                >
                  <Divider orientation="left">{t('deployments.form.nodeAffinity.matchLabels')}</Divider>
                  <Form.List name={[field.name, 'matchExpressions']}>
                    {(exprFields, { add: addExpr, remove: removeExpr }) => (
                      <>
                        {exprFields.map(exprField => renderExpressionItem(
                          exprField, 
                          `nodeAffinity.requiredTerms.${field.name}.matchExpressions`, 
                          removeExpr
                        ))}
                        <Form.Item>
                          <Button
                            type="dashed"
                            onClick={() => addExpr({ operator: 'In' })}
                            block
                            icon={<PlusOutlined />}
                          >
                            {t('deployments.form.nodeAffinity.addLabelRule')}
                          </Button>
                        </Form.Item>
                      </>
                    )}
                  </Form.List>

                  <Divider orientation="left">{t('deployments.form.nodeAffinity.matchFields')}</Divider>
                  <Paragraph type="secondary" style={{ marginBottom: 16 }}>
                    {t('deployments.form.nodeAffinity.fieldMatchDescription')}
                  </Paragraph>
                  <Form.List name={[field.name, 'matchFields']}>
                    {(fieldFields, { add: addField, remove: removeField }) => (
                      <>
                        {fieldFields.map(fieldField => renderExpressionItem(
                          fieldField, 
                          `nodeAffinity.requiredTerms.${field.name}.matchFields`, 
                          removeField
                        ))}
                        <Form.Item>
                          <Button
                            type="dashed"
                            onClick={() => addField({ operator: 'In' })}
                            block
                            icon={<PlusOutlined />}
                          >
                            {t('deployments.form.nodeAffinity.addFieldRule')}
                          </Button>
                        </Form.Item>
                      </>
                    )}
                  </Form.List>
                </Card>
              ))}
              <Form.Item>
                <Button
                  type="primary"
                  onClick={() => add({ matchExpressions: [{ operator: 'In' }] })}
                  icon={<PlusOutlined />}
                >
                  {t('deployments.form.nodeAffinity.addRequiredRuleGroup')}
                </Button>
              </Form.Item>
            </>
          )}
        </Form.List>
      </Card>
    );
  };

  // 渲染优先满足的节点选择规则
  const renderPreferredTerms = () => {
    return (
      <Card
        title={t('deployments.form.nodeAffinity.preferredRules')}
        bordered={false}
        extra={
          <Button 
            type="link" 
            icon={<InfoCircleOutlined />} 
            onClick={() => {}}
          >
            {t('common.help')}
          </Button>
        }
      >
        <Alert
          message={t('deployments.form.nodeAffinity.preferredRulesDescription')}
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
        />
        <Form.List name="nodeAffinity.preferredTerms">
          {(fields, { add, remove }) => (
            <>
              {fields.map(field => (
                <Card
                  key={field.key}
                  type="inner"
                  title={`${t('deployments.form.nodeAffinity.preferredRule')} ${field.name + 1}`}
                  style={{ marginBottom: 16 }}
                  extra={
                    <Button
                      type="link"
                      danger
                      onClick={() => remove(field.name)}
                    >
                      {t('deployments.form.nodeAffinity.deleteRule')}
                    </Button>
                  }
                >
                  <Form.Item
                    name={[field.name, 'weight']}
                    label={t('deployments.form.nodeAffinity.weight')}
                    rules={[{ required: true, message: t('deployments.form.nodeAffinity.pleaseEnterWeight') }]}
                    initialValue={10}
                  >
                    <InputNumber min={1} max={100} style={{ width: 120 }} />
                  </Form.Item>
                  
                  <Divider orientation="left">{t('deployments.form.nodeAffinity.matchLabels')}</Divider>
                  <Form.List name={[field.name, 'preference', 'matchExpressions']}>
                    {(exprFields, { add: addExpr, remove: removeExpr }) => (
                      <>
                        {exprFields.map(exprField => renderExpressionItem(
                          exprField, 
                          `nodeAffinity.preferredTerms.${field.name}.preference.matchExpressions`, 
                          removeExpr
                        ))}
                        <Form.Item>
                          <Button
                            type="dashed"
                            onClick={() => addExpr({ operator: 'In' })}
                            block
                            icon={<PlusOutlined />}
                          >
                            {t('deployments.form.nodeAffinity.addLabelRule')}
                          </Button>
                        </Form.Item>
                      </>
                    )}
                  </Form.List>

                  <Divider orientation="left">{t('deployments.form.nodeAffinity.matchFields')}</Divider>
                  <Paragraph type="secondary" style={{ marginBottom: 16 }}>
                    {t('deployments.form.nodeAffinity.fieldMatchDescription')}
                  </Paragraph>
                  <Form.List name={[field.name, 'preference', 'matchFields']}>
                    {(fieldFields, { add: addField, remove: removeField }) => (
                      <>
                        {fieldFields.map(fieldField => renderExpressionItem(
                          fieldField, 
                          `nodeAffinity.preferredTerms.${field.name}.preference.matchFields`, 
                          removeField
                        ))}
                        <Form.Item>
                          <Button
                            type="dashed"
                            onClick={() => addField({ operator: 'In' })}
                            block
                            icon={<PlusOutlined />}
                          >
                            {t('deployments.form.nodeAffinity.addFieldRule')}
                          </Button>
                        </Form.Item>
                      </>
                    )}
                  </Form.List>
                </Card>
              ))}
              <Form.Item>
                <Button
                  type="primary"
                  onClick={() => add({ weight: 10, preference: { matchExpressions: [{ operator: 'In' }] } })}
                  icon={<PlusOutlined />}
                >
                  {t('deployments.form.nodeAffinity.addPreferredRule')}
                </Button>
              </Form.Item>
            </>
          )}
        </Form.List>
      </Card>
    );
  };

  return (
    <div>
      <Alert
        message={t('deployments.form.nodeAffinity.description')}
        description={
          <div>
            <Paragraph>
              <strong>{t('deployments.form.nodeAffinity.requiredRulesTitle')}</strong>：
              {t('deployments.form.nodeAffinity.requiredRulesDetail')}
            </Paragraph>
            <Paragraph>
              <strong>{t('deployments.form.nodeAffinity.preferredRulesTitle')}</strong>：
              {t('deployments.form.nodeAffinity.preferredRulesDetail')}
            </Paragraph>
          </div>
        }
        type="info"
        showIcon
        style={{ marginBottom: 24 }}
      />
      {renderRequiredTerms()}
      {renderPreferredTerms()}
    </div>
  );
};

export default NodeAffinityManager;