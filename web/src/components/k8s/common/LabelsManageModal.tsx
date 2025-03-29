import React, { useEffect } from 'react';
import { Modal, Form, Input, Button, Space } from 'antd';
import { PlusOutlined, MinusCircleOutlined } from '@ant-design/icons';


interface LabelsManageModalProps {
  open: boolean;
  onClose: () => void;
  onSave: (labels: Record<string, string>, selector: Record<string, string>) => void;
  initialLabels?: Record<string, string>;
  initialSelector?: Record<string, string>;
}

const LabelsManageModal: React.FC<LabelsManageModalProps> = ({
  open,
  onClose,
  onSave,
  initialLabels = {},
  initialSelector = {},
}) => {
  const [form] = Form.useForm();

  useEffect(() => {
    if (open) {
      // 将对象格式的标签和选择器转换为数组格式
      const labelsArray = Object.entries(initialLabels || {}).map(([key, value]) => ({ key, value }));
      const selectorArray = Object.entries(initialSelector || {}).map(([key, value]) => ({ key, value }));
      
      form.setFieldsValue({
        labelsArray,
        selectorArray,
      });
    }
  }, [open, initialLabels, initialSelector, form]);

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      
      // 将数组格式的标签和选择器转换回对象格式
      const labels: Record<string, string> = {};
      const selector: Record<string, string> = {};
      
      (values.labelsArray || []).forEach((item: { key: string; value: string }) => {
        if (item && item.key) {
          labels[item.key] = item.value || '';
        }
      });
      
      (values.selectorArray || []).forEach((item: { key: string; value: string }) => {
        if (item && item.key) {
          selector[item.key] = item.value || '';
        }
      });
      
      onSave(labels, selector);
      onClose();
    } catch (error) {
      console.error('验证失败:', error);
    }
  };

  return (
    <Modal
      title="管理标签和选择器"
      open={open}
      onCancel={onClose}
      onOk={handleSubmit}
      width={600}
    >
      <Form
        form={form}
        layout="vertical"
      >
        <Form.Item label="标签" required>
          <Form.List name="labelsArray">
            {(fields, { add, remove }) => (
              <>
                {fields.map(({ key, name, ...restField }) => (
                  <Space key={key} style={{ display: 'flex', marginBottom: 8 }} align="baseline">
                    <Form.Item
                      {...restField}
                      name={[name, 'key']}
                      rules={[{ required: true, message: '请输入键' }]}
                    >
                      <Input placeholder="键" />
                    </Form.Item>
                    <Form.Item
                      {...restField}
                      name={[name, 'value']}
                      rules={[{ required: true, message: '请输入值' }]}
                    >
                      <Input placeholder="值" />
                    </Form.Item>
                    <Button 
                      type="text" 
                      icon={<MinusCircleOutlined />} 
                      onClick={() => remove(name)}
                      danger
                    />
                  </Space>
                ))}
                <Button 
                  type="dashed" 
                  onClick={() => add({ key: '', value: '' })} 
                  block 
                  icon={<PlusOutlined />}
                >
                  添加标签
                </Button>
              </>
            )}
          </Form.List>
        </Form.Item>
        
        <Form.Item label="选择器" required>
          <Form.List name="selectorArray">
            {(fields, { add, remove }) => (
              <>
                {fields.map(({ key, name, ...restField }) => (
                  <Space key={key} style={{ display: 'flex', marginBottom: 8 }} align="baseline">
                    <Form.Item
                      {...restField}
                      name={[name, 'key']}
                      rules={[{ required: true, message: '请输入键' }]}
                    >
                      <Input placeholder="键" />
                    </Form.Item>
                    <Form.Item
                      {...restField}
                      name={[name, 'value']}
                      rules={[{ required: true, message: '请输入值' }]}
                    >
                      <Input placeholder="值" />
                    </Form.Item>
                    <Button 
                      type="text" 
                      icon={<MinusCircleOutlined />} 
                      onClick={() => remove(name)}
                      danger
                    />
                  </Space>
                ))}
                <Button 
                  type="dashed" 
                  onClick={() => add({ key: '', value: '' })} 
                  block 
                  icon={<PlusOutlined />}
                >
                  添加选择器
                </Button>
              </>
            )}
          </Form.List>
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default LabelsManageModal;