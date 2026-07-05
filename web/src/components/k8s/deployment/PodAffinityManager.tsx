import React from 'react';
import {
  Card, Form, Input, Button, Select, InputNumber,
  Row, Col, Divider, Alert, Typography
} from 'antd';
import { PlusOutlined, MinusCircleOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';

const { Text, Paragraph } = Typography;
const { Option } = Select;

interface PodAffinityManagerProps {
  fieldPrefix: 'podAffinity' | 'podAntiAffinity';
  mode: 'affinity' | 'antiAffinity';
}

const TOPOLOGY_PRESETS = [
  { value: 'kubernetes.io/hostname', labelKey: 'hostname' },
  { value: 'topology.kubernetes.io/zone', labelKey: 'zone' },
  { value: 'topology.kubernetes.io/region', labelKey: 'region' },
];

const PodAffinityManager: React.FC<PodAffinityManagerProps> = ({ fieldPrefix, mode }) => {
  const { t } = useTranslation();
  const ns = `deployments.form.${mode}`;

  const renderTermFields = (fieldName: number, showWeight?: boolean) => {
    const termBase = showWeight ? [fieldName, 'term'] : [fieldName];

    return (
      <>
        {showWeight && (
          <Form.Item
            name={[fieldName, 'weight']}
            label={t('deployments.form.nodeAffinity.weight')}
            initialValue={100}
            rules={[{ required: true }]}
          >
            <InputNumber min={1} max={100} style={{ width: 120 }} />
          </Form.Item>
        )}

        <Form.Item
          name={[...termBase, 'topologyKey']}
          label={t(`${ns}.topologyKey`)}
          rules={[{ required: true, message: t(`${ns}.pleaseEnterTopologyKey`) }]}
          initialValue="kubernetes.io/hostname"
          extra={t(`${ns}.topologyKeyHint`)}
        >
          <Select showSearch allowClear placeholder="kubernetes.io/hostname">
            {TOPOLOGY_PRESETS.map(p => (
              <Option key={p.value} value={p.value}>
                {t(`${ns}.topologyPresets.${p.labelKey}`)} ({p.value})
              </Option>
            ))}
          </Select>
        </Form.Item>

        <Form.Item
          name={[...termBase, 'namespaces']}
          label={t(`${ns}.namespaces`)}
          extra={t(`${ns}.namespacesHint`)}
        >
          <Select mode="tags" placeholder={t(`${ns}.namespacesPlaceholder`)} tokenSeparators={[',']} />
        </Form.Item>

        <Divider orientation="left">{t(`${ns}.matchLabels`)}</Divider>
        <Form.List name={[...termBase, 'labelSelector', 'matchLabels']}>
          {(labelFields, { add: addLabel, remove: removeLabel }) => (
            <>
              {labelFields.map(labelField => (
                <Row key={labelField.key} gutter={8} align="middle">
                  <Col span={10}>
                    <Form.Item name={[labelField.name, 'key']} rules={[{ required: true }]} style={{ marginBottom: 8 }}>
                      <Input placeholder={t('deployments.form.labels.keyPlaceholder')} />
                    </Form.Item>
                  </Col>
                  <Col span={10}>
                    <Form.Item name={[labelField.name, 'value']} rules={[{ required: true }]} style={{ marginBottom: 8 }}>
                      <Input placeholder={t('deployments.form.labels.valuePlaceholder')} />
                    </Form.Item>
                  </Col>
                  <Col span={4}>
                    <Button type="link" icon={<MinusCircleOutlined />} onClick={() => removeLabel(labelField.name)} />
                  </Col>
                </Row>
              ))}
              <Button type="dashed" onClick={() => addLabel()} block icon={<PlusOutlined />} style={{ marginBottom: 16 }}>
                {t(`${ns}.addMatchLabel`)}
              </Button>
            </>
          )}
        </Form.List>

        <Divider orientation="left">{t(`${ns}.matchExpressions`)}</Divider>
        <Form.List name={[...termBase, 'labelSelector', 'matchExpressions']}>
          {(exprFields, { add: addExpr, remove: removeExpr }) => (
            <>
              {exprFields.map(exprField => (
                <Row key={exprField.key} gutter={[16, 0]} align="middle">
                  <Col span={6}>
                    <Form.Item name={[exprField.name, 'key']} rules={[{ required: true }]} style={{ marginBottom: 8 }}>
                      <Input placeholder="app" />
                    </Form.Item>
                  </Col>
                  <Col span={5}>
                    <Form.Item name={[exprField.name, 'operator']} initialValue="In" style={{ marginBottom: 8 }}>
                      <Select>
                        <Option value="In">In</Option>
                        <Option value="NotIn">NotIn</Option>
                        <Option value="Exists">Exists</Option>
                        <Option value="DoesNotExist">DoesNotExist</Option>
                      </Select>
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item noStyle shouldUpdate>
                      {({ getFieldValue }) => {
                        const operator = getFieldValue([
                          fieldPrefix, 'requiredTerms', ...termBase, 'labelSelector', 'matchExpressions', exprField.name, 'operator',
                        ]) ?? getFieldValue([
                          fieldPrefix, 'preferredTerms', ...termBase, 'labelSelector', 'matchExpressions', exprField.name, 'operator',
                        ]);
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
                          <Form.Item name={[exprField.name, 'values']} style={{ marginBottom: 8 }}>
                            <Select mode="tags" placeholder={t('deployments.form.nodeAffinity.valuesPlaceholder')} tokenSeparators={[',']} />
                          </Form.Item>
                        );
                      }}
                    </Form.Item>
                  </Col>
                  <Col span={1}>
                    <Button type="link" icon={<MinusCircleOutlined />} onClick={() => removeExpr(exprField.name)} />
                  </Col>
                </Row>
              ))}
              <Button type="dashed" onClick={() => addExpr({ operator: 'In' })} block icon={<PlusOutlined />}>
                {t(`${ns}.addMatchExpression`)}
              </Button>
            </>
          )}
        </Form.List>
      </>
    );
  };

  return (
    <div>
      <Alert
        message={t(`${ns}.description`)}
        description={<Paragraph style={{ marginBottom: 0 }}>{t(`${ns}.detail`)}</Paragraph>}
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
      />

      <Card title={t(`${ns}.requiredRules`)} bordered={false} style={{ marginBottom: 16 }}>
        <Form.List name={[fieldPrefix, 'requiredTerms']}>
          {(fields, { add, remove }) => (
            <>
              {fields.map(field => (
                <Card
                  key={field.key}
                  type="inner"
                  title={`${t(`${ns}.requiredRule`)} ${field.name + 1}`}
                  style={{ marginBottom: 16 }}
                  extra={
                    <Button type="link" danger onClick={() => remove(field.name)}>
                      {t('common.delete')}
                    </Button>
                  }
                >
                  {renderTermFields(field.name)}
                </Card>
              ))}
              <Button
                type="primary"
                onClick={() => add({ topologyKey: 'kubernetes.io/hostname', labelSelector: { matchExpressions: [{ operator: 'In' }] } })}
                icon={<PlusOutlined />}
              >
                {t(`${ns}.addRequiredRule`)}
              </Button>
            </>
          )}
        </Form.List>
      </Card>

      <Card title={t(`${ns}.preferredRules`)} bordered={false}>
        <Form.List name={[fieldPrefix, 'preferredTerms']}>
          {(fields, { add, remove }) => (
            <>
              {fields.map(field => (
                <Card
                  key={field.key}
                  type="inner"
                  title={`${t(`${ns}.preferredRule`)} ${field.name + 1}`}
                  style={{ marginBottom: 16 }}
                  extra={
                    <Button type="link" danger onClick={() => remove(field.name)}>
                      {t('common.delete')}
                    </Button>
                  }
                >
                  {renderTermFields(field.name, true)}
                </Card>
              ))}
              <Button
                type="primary"
                onClick={() => add({
                  weight: 100,
                  term: { topologyKey: 'kubernetes.io/hostname', labelSelector: { matchExpressions: [{ operator: 'In' }] } },
                })}
                icon={<PlusOutlined />}
              >
                {t(`${ns}.addPreferredRule`)}
              </Button>
            </>
          )}
        </Form.List>
      </Card>
    </div>
  );
};

export default PodAffinityManager;
