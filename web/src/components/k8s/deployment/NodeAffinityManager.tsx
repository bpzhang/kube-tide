import React from 'react';
import { 
  Card, Form, Input, Button, Select, InputNumber, Space, 
  Row, Col, Divider, Alert, Typography 
} from 'antd';
import { PlusOutlined, MinusCircleOutlined, InfoCircleOutlined } from '@ant-design/icons';
import { NodeAffinityConfig, NodeSelectorTerm, PreferredNodeAffinity } from './DeploymentTypes';

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
            rules={[{ required: true, message: '请输入标签键' }]}
            style={{ marginBottom: 8 }}
          >
            <Input placeholder="例如: kubernetes.io/hostname" />
          </Form.Item>
        </Col>
        <Col span={5}>
          <Form.Item
            name={[field.name, 'operator']}
            rules={[{ required: true, message: '请选择操作符' }]}
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
                    {operator === 'Exists' ? '存在标签键' : '不存在标签键'}
                  </Text>
                );
              }
              return (
                <Form.Item
                  name={[field.name, 'values']}
                  rules={[{ required: true, message: '请输入标签值' }]}
                  style={{ marginBottom: 0 }}
                >
                  <Select
                    mode="tags"
                    placeholder="输入标签值并按回车，可输入多个值"
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
        title="必须满足的节点亲和性规则"
        bordered={false}
        extra={
          <Button 
            type="link" 
            icon={<InfoCircleOutlined />} 
            onClick={() => {}}
          >
            帮助
          </Button>
        }
      >
        <Alert
          message="Pod必须调度到满足所有规则组中至少一个规则组的节点上"
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
                  title={`规则组 ${field.name + 1}`}
                  style={{ marginBottom: 16 }}
                  extra={
                    <Button
                      type="link"
                      danger
                      onClick={() => remove(field.name)}
                    >
                      删除规则组
                    </Button>
                  }
                >
                  <Divider orientation="left">匹配标签</Divider>
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
                            添加标签匹配规则
                          </Button>
                        </Form.Item>
                      </>
                    )}
                  </Form.List>

                  <Divider orientation="left">匹配字段</Divider>
                  <Paragraph type="secondary" style={{ marginBottom: 16 }}>
                    字段匹配是高级功能，通常用于匹配节点的特殊字段，例如元数据。
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
                            添加字段匹配规则
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
                  添加必选规则组
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
        title="优先满足的节点亲和性规则"
        bordered={false}
        extra={
          <Button 
            type="link" 
            icon={<InfoCircleOutlined />} 
            onClick={() => {}}
          >
            帮助
          </Button>
        }
      >
        <Alert
          message="Pod会优先调度到满足这些规则的节点上，每条规则有一个权重值，满足规则的节点会获得相应的分数"
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
                  title={`优先规则 ${field.name + 1}`}
                  style={{ marginBottom: 16 }}
                  extra={
                    <Button
                      type="link"
                      danger
                      onClick={() => remove(field.name)}
                    >
                      删除规则
                    </Button>
                  }
                >
                  <Form.Item
                    name={[field.name, 'weight']}
                    label="权重"
                    rules={[{ required: true, message: '请输入权重值' }]}
                    initialValue={10}
                  >
                    <InputNumber min={1} max={100} style={{ width: 120 }} />
                  </Form.Item>
                  
                  <Divider orientation="left">匹配标签</Divider>
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
                            添加标签匹配规则
                          </Button>
                        </Form.Item>
                      </>
                    )}
                  </Form.List>

                  <Divider orientation="left">匹配字段</Divider>
                  <Paragraph type="secondary" style={{ marginBottom: 16 }}>
                    字段匹配是高级功能，通常用于匹配节点的特殊字段，例如元数据。
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
                            添加字段匹配规则
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
                  添加优先规则
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
        message="节点亲和性用于控制Pod可以调度到哪些节点上"
        description={
          <div>
            <Paragraph>
              <strong>必须满足的规则</strong>：Pod只能调度到满足所有规则组中至少一个规则组的节点上
            </Paragraph>
            <Paragraph>
              <strong>优先满足的规则</strong>：调度器会尝试调度Pod到满足这些规则的节点上，但如果无法满足，仍会调度到其他节点
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