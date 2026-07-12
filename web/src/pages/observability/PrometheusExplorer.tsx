import React, { useEffect, useState } from 'react';
import { Card, Form, Input, Select, Button, Radio, Table, Alert, Space, message, Tag } from 'antd';
import { SearchOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useClusterNamespace } from '@/hooks/useClusterNamespace';
import ClusterNamespaceToolbar from '@/components/k8s/common/ClusterNamespaceToolbar';
import { queryPrometheusRange, queryPrometheusInstant } from '@/api/prometheus';
import { getPrometheusInfo } from '@/api/observability';
import PrometheusChart from '@/components/observability/PrometheusChart';
import {
  buildRangeParams,
  parsePrometheusInstantTable,
  parsePrometheusRange,
  PROMQL_PRESETS,
  PrometheusRangeResponse,
} from '@/utils/prometheus';

const PrometheusExplorer: React.FC = () => {
  const { t } = useTranslation();
  const { selectedCluster, setSelectedCluster, clusters, clustersLoading } = useClusterNamespace(t);
  const [form] = Form.useForm();
  const [queryType, setQueryType] = useState<'range' | 'instant'>('range');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [chartSeries, setChartSeries] = useState<ReturnType<typeof parsePrometheusRange>>([]);
  const [tableData, setTableData] = useState<ReturnType<typeof parsePrometheusInstantTable>>([]);
  const [promInfo, setPromInfo] = useState<{ healthy: boolean; source: string; url?: string } | null>(null);

  useEffect(() => {
    if (!selectedCluster) return;
    getPrometheusInfo(selectedCluster).then((res) => {
      if (res.data.code === 0) {
        setPromInfo(res.data.data.prometheus);
      }
    });
  }, [selectedCluster]);

  const handleQuery = async (values: { query: string; hours?: number; step?: string }) => {
    if (!selectedCluster) return;
    setLoading(true);
    setError(null);
    setChartSeries([]);
    setTableData([]);
    try {
      if (queryType === 'range') {
        const params = buildRangeParams(values.query, values.hours || 1, values.step || '60');
        const res = await queryPrometheusRange(selectedCluster, params);
        const body = res.data as PrometheusRangeResponse;
        if (body.status !== 'success') {
          setError(body.error || t('prometheus.queryFailed'));
          return;
        }
        setChartSeries(parsePrometheusRange(body));
        if (!body.data?.result?.length) {
          setError(t('prometheus.noData'));
        }
      } else {
        const res = await queryPrometheusInstant(selectedCluster, { query: values.query });
        const body = res.data as PrometheusRangeResponse;
        if (body.status !== 'success') {
          setError(body.error || t('prometheus.queryFailed'));
          return;
        }
        setTableData(parsePrometheusInstantTable(body));
        if (!body.data?.result?.length) {
          setError(t('prometheus.noData'));
        }
      }
    } catch {
      setError(t('prometheus.queryFailed'));
    } finally {
      setLoading(false);
    }
  };

  const applyPreset = (key: string) => {
    const preset = PROMQL_PRESETS.find((p) => p.key === key);
    if (preset) {
      form.setFieldsValue({ query: preset.query });
    }
  };

  const instantColumns = [
    {
      title: t('prometheus.explorer.metric'),
      key: 'metric',
      render: (_: unknown, row: { metric: Record<string, string> }) =>
        Object.entries(row.metric)
          .map(([k, v]) => `${k}=${v}`)
          .join(' '),
    },
    { title: t('prometheus.explorer.value'), dataIndex: 'value', key: 'value' },
  ];

  return (
    <Card title={t('prometheus.explorer.title')}>
      <ClusterNamespaceToolbar
        clusters={clusters}
        selectedCluster={selectedCluster}
        onClusterChange={setSelectedCluster}
        namespace="all"
        onNamespaceChange={() => {}}
        loading={clustersLoading}
        showNamespace={false}
      />

      {promInfo && (
        <Alert
          style={{ marginBottom: 16 }}
          type={promInfo.healthy ? 'success' : 'warning'}
          showIcon
          message={
            <Space>
              <span>{t('prometheus.explorer.status')}</span>
              <Tag color={promInfo.healthy ? 'green' : 'orange'}>{promInfo.source}</Tag>
              {promInfo.url && <span style={{ fontSize: 12, color: '#888' }}>{promInfo.url}</span>}
            </Space>
          }
        />
      )}

      <Form
        form={form}
        layout="vertical"
        onFinish={handleQuery}
        initialValues={{ hours: 1, step: '60', query: PROMQL_PRESETS[0].query }}
      >
        <Form.Item label={t('prometheus.explorer.queryType')}>
          <Radio.Group value={queryType} onChange={(e) => setQueryType(e.target.value)}>
            <Radio.Button value="range">{t('prometheus.explorer.range')}</Radio.Button>
            <Radio.Button value="instant">{t('prometheus.explorer.instant')}</Radio.Button>
          </Radio.Group>
        </Form.Item>

        <Form.Item label={t('prometheus.explorer.presets')}>
          <Select
            style={{ width: 320 }}
            placeholder={t('prometheus.explorer.selectPreset')}
            onChange={applyPreset}
            allowClear
          >
            {PROMQL_PRESETS.map((p) => (
              <Select.Option key={p.key} value={p.key}>
                {t(`prometheus.explorer.presetsList.${p.key}`)}
              </Select.Option>
            ))}
          </Select>
        </Form.Item>

        <Form.Item name="query" label="PromQL" rules={[{ required: true }]}>
          <Input.TextArea rows={3} style={{ fontFamily: 'monospace' }} />
        </Form.Item>

        {queryType === 'range' && (
          <Space>
            <Form.Item name="hours" label={t('prometheus.explorer.timeRange')}>
              <Select style={{ width: 120 }}>
                <Select.Option value={1}>1h</Select.Option>
                <Select.Option value={6}>6h</Select.Option>
                <Select.Option value={24}>24h</Select.Option>
              </Select>
            </Form.Item>
            <Form.Item name="step" label={t('prometheus.explorer.step')}>
              <Select style={{ width: 120 }}>
                <Select.Option value="30">30s</Select.Option>
                <Select.Option value="60">60s</Select.Option>
                <Select.Option value="300">5m</Select.Option>
              </Select>
            </Form.Item>
          </Space>
        )}

        <Form.Item>
          <Button type="primary" htmlType="submit" icon={<SearchOutlined />} loading={loading}>
            {t('prometheus.explorer.run')}
          </Button>
        </Form.Item>
      </Form>

      {error && <Alert type="info" showIcon message={error} style={{ marginBottom: 16 }} />}

      {queryType === 'range' && chartSeries.length > 0 && (
        <Card type="inner" title={t('prometheus.explorer.chart')}>
          <PrometheusChart series={chartSeries} />
        </Card>
      )}

      {queryType === 'instant' && tableData.length > 0 && (
        <Table
          size="small"
          rowKey={(_, idx) => String(idx)}
          dataSource={tableData}
          columns={instantColumns}
          pagination={{ pageSize: 15 }}
        />
      )}
    </Card>
  );
};

export default PrometheusExplorer;
