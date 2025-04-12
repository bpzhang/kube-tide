import React, { useState } from 'react';
import { Modal, Form, Input, InputNumber, message, Select, Switch, Button, Space, Row, Col } from 'antd';
import { PlusOutlined, MinusCircleOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { createStatefulSet } from '@/api/statefulset';
import NamespaceSelector from '@/components/k8s/common/NamespaceSelector';

const { Option } = Select;

interface CreateStatefulSetModalProps {
  visible: boolean;
  onCancel: () => void;
  onSuccess: () => void;
  clusterName: string;
  namespace: string;
}

/**
 * 创建StatefulSet模态框组件
 */
const CreateStatefulSetModal: React.FC<CreateStatefulSetModalProps> = ({ 
  visible, 
  onCancel, 
  onSuccess, 
  clusterName,
  namespace: initialNamespace
}) => {
  const { t } = useTranslation();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [namespace, setNamespace] = useState(initialNamespace);

  // 初始化表单
  const initialValues = {
    namespace,
    replicas: 1,
    podManagementPolicy: 'OrderedReady',
    updateStrategy: 'RollingUpdate',
    containers: [
      {
        name: 'container-1',
        image: '',
        resources: {
          requests: {
            cpu: '100m',
            memory: '128Mi'
          },
          limits: {
            cpu: '500m',
            memory: '512Mi'
          }
        }
      }
    ],
    volumeClaimTemplates: [
      {
        name: 'data',
        storageClassName: '',
        accessModes: ['ReadWriteOnce'],
        storage: '1Gi'
      }
    ]
  };

  // 重置表单
  const resetForm = () => {
    form.resetFields();
  };

  // 提交表单
  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setLoading(true);
      
      const response = await createStatefulSet(clusterName, values.namespace, {
        ...values,
        labels: values.labels ? convertLabelsToObject(values.labels) : undefined,
        annotations: values.annotations ? convertLabelsToObject(values.annotations) : undefined
      });
      
      if (response.data.code === 0) {
        message.success(t('statefulsets.createSuccess'));
        resetForm();
        onSuccess();
      } else {
        message.error(response.data.message || t('statefulsets.createFailed'));
      }
    } catch (err) {
      console.error('Create StatefulSet error:', err);
    } finally {
      setLoading(false);
    }
  };

  // 处理命名空间变化
  const handleNamespaceChange = (value: string) => {
    setNamespace(value);
    form.setFieldsValue({ namespace: value });
  };

  // 将标签数组转换为对象
  const convertLabelsToObject = (labels: { key: string; value: string }[]) => {
    const result: Record<string, string> = {};
    labels.forEach(item => {
      if (item.key && item.value) {
        result[item.key] = item.value;
      }
    });
    return result;
  };

  return (
    <Modal
      title={t('statefulsets.create')}
      open={visible}
      onCancel={onCancel}
      onOk={handleSubmit}
      confirmLoading={loading}
      width={800}
      destroyOnClose
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={initialValues}
        preserve={false}
      >
        <Row gutter={16}>
          <Col span={12}>
            <Form.Item
              name="name"
              label={t('common.name')}
              rules={[{ required: true, message: t('statefulsets.pleaseEnterName') }]}
            >
              <Input placeholder={t('statefulsets.namePlaceholder')} />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item
              name="namespace"
              label={t('common.namespace')}
              rules={[{ required: true, message: t('statefulsets.pleaseSelectNamespace') }]}
            >
              <NamespaceSelector 
                clusterName={clusterName} 
                value={namespace} 
                onChange={handleNamespaceChange} 
              />
            </Form.Item>
          </Col>
        </Row>

        <Row gutter={16}>
          <Col span={12}>
            <Form.Item
              name="serviceName"
              label={t('statefulsets.serviceName')}
              rules={[{ required: true, message: t('statefulsets.pleaseEnterServiceName') }]}
            >
              <Input placeholder={t('statefulsets.serviceNamePlaceholder')} />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item
              name="replicas"
              label={t('statefulsets.replicas')}
              rules={[
                { required: true, message: t('statefulsets.pleaseEnterReplicas') },
                { type: 'number', min: 0, message: t('statefulsets.replicasMustBePositive') }
              ]}
            >
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
          </Col>
        </Row>

        <Row gutter={16}>
          <Col span={12}>
            <Form.Item
              name="podManagementPolicy"
              label={t('statefulsets.podManagementPolicy')}
              rules={[{ required: true, message: t('statefulsets.pleaseSelectPolicy') }]}
            >
              <Select>
                <Option value="OrderedReady">OrderedReady</Option>
                <Option value="Parallel">Parallel</Option>
              </Select>
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item
              name="updateStrategy"
              label={t('statefulsets.updateStrategy')}
              rules={[{ required: true, message: t('statefulsets.pleaseSelectStrategy') }]}
            >
              <Select>
                <Option value="RollingUpdate">RollingUpdate</Option>
                <Option value="OnDelete">OnDelete</Option>
              </Select>
            </Form.Item>
          </Col>
        </Row>

        <h3>{t('statefulsets.containers')}</h3>
        <Form.List name="containers">
          {(fields, { add, remove }) => (
            <>
              {fields.map(field => (
                <div key={field.key} style={{ marginBottom: 24, padding: 16, border: '1px solid #f0f0f0', borderRadius: 4 }}>
                  <Row gutter={16}>
                    <Col span={12}>
                      <Form.Item
                        {...field}
                        name={[field.name, 'name']}
                        label={t('common.name')}
                        rules={[{ required: true, message: t('statefulsets.pleaseEnterContainerName') }]}
                      >
                        <Input placeholder={t('statefulsets.containerNamePlaceholder')} />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        {...field}
                        name={[field.name, 'image']}
                        label={t('statefulsets.image')}
                        rules={[{ required: true, message: t('statefulsets.pleaseEnterImage') }]}
                      >
                        <Input placeholder={t('statefulsets.imagePlaceholder')} />
                      </Form.Item>
                    </Col>
                  </Row>

                  <h4>{t('statefulsets.resources')}</h4>
                  <Row gutter={16}>
                    <Col span={12}>
                      <Form.Item
                        {...field}
                        name={[field.name, 'resources', 'requests', 'cpu']}
                        label={t('statefulsets.requestsCpu')}
                      >
                        <Input placeholder="100m" />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        {...field}
                        name={[field.name, 'resources', 'requests', 'memory']}
                        label={t('statefulsets.requestsMemory')}
                      >
                        <Input placeholder="128Mi" />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        {...field}
                        name={[field.name, 'resources', 'limits', 'cpu']}
                        label={t('statefulsets.limitsCpu')}
                      >
                        <Input placeholder="500m" />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        {...field}
                        name={[field.name, 'resources', 'limits', 'memory']}
                        label={t('statefulsets.limitsMemory')}
                      >
                        <Input placeholder="512Mi" />
                      </Form.Item>
                    </Col>
                  </Row>

                  {fields.length > 1 && (
                    <Button 
                      type="dashed" 
                      danger 
                      onClick={() => remove(field.name)} 
                      block 
                      icon={<MinusCircleOutlined />}
                    >
                      {t('statefulsets.removeContainer')}
                    </Button>
                  )}
                </div>
              ))}
              <Form.Item>
                <Button 
                  type="dashed" 
                  onClick={() => add()} 
                  block 
                  icon={<PlusOutlined />}
                >
                  {t('statefulsets.addContainer')}
                </Button>
              </Form.Item>
            </>
          )}
        </Form.List>

        <h3>{t('statefulsets.volumeClaimTemplates')}</h3>
        <Form.List name="volumeClaimTemplates">
          {(fields, { add, remove }) => (
            <>
              {fields.map(field => (
                <div key={field.key} style={{ marginBottom: 24, padding: 16, border: '1px solid #f0f0f0', borderRadius: 4 }}>
                  <Row gutter={16}>
                    <Col span={12}>
                      <Form.Item
                        {...field}
                        name={[field.name, 'name']}
                        label={t('common.name')}
                        rules={[{ required: true, message: t('statefulsets.pleaseEnterPvcName') }]}
                      >
                        <Input placeholder={t('statefulsets.pvcNamePlaceholder')} />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        {...field}
                        name={[field.name, 'storageClassName']}
                        label={t('statefulsets.storageClassName')}
                      >
                        <Input placeholder={t('statefulsets.storageClassNamePlaceholder')} />
                      </Form.Item>
                    </Col>
                  </Row>

                  <Row gutter={16}>
                    <Col span={12}>
                      <Form.Item
                        {...field}
                        name={[field.name, 'accessModes']}
                        label={t('statefulsets.accessModes')}
                        rules={[{ required: true, message: t('statefulsets.pleaseSelectAccessMode') }]}
                      >
                        <Select mode="multiple">
                          <Option value="ReadWriteOnce">ReadWriteOnce</Option>
                          <Option value="ReadOnlyMany">ReadOnlyMany</Option>
                          <Option value="ReadWriteMany">ReadWriteMany</Option>
                        </Select>
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        {...field}
                        name={[field.name, 'storage']}
                        label={t('statefulsets.storage')}
                        rules={[{ required: true, message: t('statefulsets.pleaseEnterStorage') }]}
                      >
                        <Input placeholder="1Gi" />
                      </Form.Item>
                    </Col>
                  </Row>

                  {fields.length > 1 && (
                    <Button 
                      type="dashed" 
                      danger 
                      onClick={() => remove(field.name)} 
                      block 
                      icon={<MinusCircleOutlined />}
                    >
                      {t('statefulsets.removePvc')}
                    </Button>
                  )}
                </div>
              ))}
              <Form.Item>
                <Button 
                  type="dashed" 
                  onClick={() => add()} 
                  block 
                  icon={<PlusOutlined />}
                >
                  {t('statefulsets.addPvc')}
                </Button>
              </Form.Item>
            </>
          )}
        </Form.List>

        <h3>{t('common.labels')}</h3>
        <Form.List name="labels">
          {(fields, { add, remove }) => (
            <>
              {fields.map(field => (
                <Row key={field.key} gutter={16} align="middle">
                  <Col span={10}>
                    <Form.Item
                      {...field}
                      name={[field.name, 'key']}
                      rules={[{ required: true, message: t('common.pleaseEnterKey') }]}
                    >
                      <Input placeholder={t('common.key')} />
                    </Form.Item>
                  </Col>
                  <Col span={10}>
                    <Form.Item
                      {...field}
                      name={[field.name, 'value']}
                      rules={[{ required: true, message: t('common.pleaseEnterValue') }]}
                    >
                      <Input placeholder={t('common.value')} />
                    </Form.Item>
                  </Col>
                  <Col span={4}>
                    <MinusCircleOutlined onClick={() => remove(field.name)} />
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
                  {t('common.addLabel')}
                </Button>
              </Form.Item>
            </>
          )}
        </Form.List>
      </Form>
    </Modal>
  );
};

export default CreateStatefulSetModal;
