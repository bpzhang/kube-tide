import React from 'react';
import {
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { PrometheusChartSeries } from '@/utils/prometheus';

const COLORS = ['#1890ff', '#52c41a', '#fa8c16', '#eb2f96', '#722ed1', '#13c2c2'];

interface PrometheusChartProps {
  series: PrometheusChartSeries[];
  height?: number;
  yLabel?: string;
}

const PrometheusChart: React.FC<PrometheusChartProps> = ({ series, height = 300, yLabel }) => {
  if (!series.length) {
    return null;
  }

  const merged = series[0].data.map((point, idx) => {
    const row: Record<string, string | number> = { time: point.time };
    series.forEach((s, sIdx) => {
      row[`v${sIdx}`] = s.data[idx]?.value ?? 0;
    });
    return row;
  });

  return (
    <ResponsiveContainer width="100%" height={height}>
      <LineChart data={merged}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="time" />
        <YAxis label={yLabel ? { value: yLabel, angle: -90, position: 'insideLeft' } : undefined} />
        <Tooltip />
        <Legend />
        {series.map((s, idx) => (
          <Line
            key={s.name}
            type="monotone"
            dataKey={`v${idx}`}
            name={s.name}
            stroke={COLORS[idx % COLORS.length]}
            dot={false}
          />
        ))}
      </LineChart>
    </ResponsiveContainer>
  );
};

export default PrometheusChart;
