import React from 'react';
import { Select } from 'antd';

const { Option } = Select;

/**
 * PortNameSelect
 * functional component for selecting port names
 * supports search and clear
 * @param {string} value - selected value
 * @param {(value: string) => void} onChange - callback function when value changes
 * @param {string} placeholder - placeholder text
 * @param {React.CSSProperties} style - custom styles
 */
interface PortNameSelectProps {
    value?: string;
    onChange?: (value: string) => void;
    placeholder?: string;
    style?: React.CSSProperties;
}


const PortNameSelect: React.FC<PortNameSelectProps> = ({
    value,
    onChange,
    placeholder = '选择通信协议',
    style
}) => {
    return (
        <Select 
            value={value} 
            onChange={onChange} 
            placeholder={placeholder}
            allowClear
            style={style}
            showSearch
            filterOption={(input, option) => 
                (option?.children as unknown as string).toLowerCase().includes(input.toLowerCase()) ||
                (option?.value as string).toLowerCase().includes(input.toLowerCase())
            }
            optionFilterProp="children"
        >
            <Option value="http">HTTP</Option>
            <Option value="https">HTTPS</Option>
            <Option value="grpc">gRPC</Option>
            <Option value="tcp">TCP</Option>
            <Option value="udp">UDP</Option>
            <Option value="ws">WebSocket</Option>
            <Option value="wss">WebSocket Secure</Option>
            <Option value="dns">DNS</Option>
            <Option value="mongo">MongoDB</Option>
            <Option value="mysql">MySQL</Option>
            <Option value="redis">Redis</Option>
            <Option value="postgres">PostgreSQL</Option>
            <Option value="kafka">Kafka</Option>
            <Option value="amqp">AMQP</Option>
            <Option value="mqtt">MQTT</Option>
        </Select>
    );
};

export default PortNameSelect;