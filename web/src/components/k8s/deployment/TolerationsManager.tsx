import React from 'react';
import { Card, Form, Input, Button, Select, InputNumber, Alert, Typography } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';

const { Paragraph } = Typography;

/**
 * Pod 容忍度配置，配合节点污点实现专用节点池调度
 */
const TolerationsManager: React.FC = () => {
  const { t } = useTranslation();

  return (
    <div>
      <Alert
        message={t('deployments.form.tolerations.description')}
        description={
          <Paragraph style={{ marginBottom: 0 }}>
            {t('deployments.form.tolerations.detail')}
          </Paragraph>
        }
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
      />
      <Card title={t('deployments.form.tolerations.title')} bordered={false}>
        <Form.List name="tolerations">
          {(fields, { add, remove }) => (
            <>
              {fields.map(({ key, name, ...restField }) => (
                <div key={key} style={{ display: 'flex', flexWrap: 'wrap', gap: 8, marginBottom: 12, alignItems: 'flex-start' }}>
                  <Form.Item
                    {...restField}
                    name={[name, 'key']}
                    style={{ minWidth: 140, marginBottom: 0 }}
                  >
                    <Input placeholder={t('deployments.form.tolerations.keyPlaceholder')} />
                  </Form.Item>
                  <Form.Item
                    {...restField}
                    name={[name, 'operator']}
                    initialValue="Equal"
                    style={{ minWidth: 100, marginBottom: 0 }}
                  >
                    <Select>
                      <Select.Option value="Equal">Equal</Select.Option>
                      <Select.Option value="Exists">Exists</Select.Option>
                    </Select>
                  </Form.Item>
                  <Form.Item
                    {...restField}
                    name={[name, 'value']}
                    style={{ minWidth: 120, marginBottom: 0 }}
                  >
                    <Input placeholder={t('deployments.form.tolerations.valuePlaceholder')} />
                  </Form.Item>
                  <Form.Item
                    {...restField}
                    name={[name, 'effect']}
                    style={{ minWidth: 160, marginBottom: 0 }}
                  >
                    <Select allowClear placeholder={t('deployments.form.tolerations.effectPlaceholder')}>
                      <Select.Option value="NoSchedule">{t('nodes.nodePool.taintEffects.NoSchedule')}</Select.Option>
                      <Select.Option value="PreferNoSchedule">{t('nodes.nodePool.taintEffects.PreferNoSchedule')}</Select.Option>
                      <Select.Option value="NoExecute">{t('nodes.nodePool.taintEffects.NoExecute')}</Select.Option>
                    </Select>
                  </Form.Item>
                  <Form.Item
                    {...restField}
                    name={[name, 'tolerationSeconds']}
                    style={{ minWidth: 120, marginBottom: 0 }}
                  >
                    <InputNumber placeholder={t('deployments.form.tolerations.secondsPlaceholder')} min={0} style={{ width: '100%' }} />
                  </Form.Item>
                  <Button type="text" danger onClick={() => remove(name)}>
                    {t('common.delete')}
                  </Button>
                </div>
              ))}
              <Button type="dashed" onClick={() => add({ operator: 'Equal' })} block icon={<PlusOutlined />}>
                {t('deployments.form.tolerations.add')}
              </Button>
            </>
          )}
        </Form.List>
      </Card>
    </div>
  );
};

export default TolerationsManager;
