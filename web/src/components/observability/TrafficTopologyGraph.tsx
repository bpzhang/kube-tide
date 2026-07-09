import React, { useMemo, useState } from 'react';
import { Checkbox, Empty, Space, Typography } from 'antd';
import { useTranslation } from 'react-i18next';
import { TopologyEdge, TopologyNode, TrafficTopology } from '@/api/traffic_topology';

const { Text } = Typography;

const NODE_WIDTH = 148;
const NODE_HEIGHT = 44;
const COL_GAP = 130;
const ROW_GAP = 18;
const PADDING = 48;

const NODE_COLORS: Record<string, { fill: string; stroke: string }> = {
  ingress: { fill: '#f9f0ff', stroke: '#722ed1' },
  service: { fill: '#e6f4ff', stroke: '#1677ff' },
  deployment: { fill: '#f6ffed', stroke: '#52c41a' },
  statefulset: { fill: '#f6ffed', stroke: '#389e0d' },
  daemonset: { fill: '#f6ffed', stroke: '#237804' },
  job: { fill: '#fff7e6', stroke: '#fa8c16' },
  replicaset: { fill: '#fffbe6', stroke: '#d4b106' },
  default: { fill: '#e6fffb', stroke: '#13c2c2' },
};

const EDGE_STYLES: Record<string, { stroke: string; dash?: string }> = {
  routes: { stroke: '#1677ff' },
  selects: { stroke: '#52c41a' },
  calls: { stroke: '#fa8c16', dash: '6 4' },
  policy_allow: { stroke: '#9254de', dash: '3 3' },
};

const WORKLOAD_TYPES = new Set(['deployment', 'statefulset', 'daemonset', 'job', 'replicaset']);

interface PositionedNode extends TopologyNode {
  x: number;
  y: number;
}

interface TrafficTopologyGraphProps {
  topology: TrafficTopology | null;
  loading?: boolean;
}

const truncate = (text: string, max = 16) => (text.length > max ? `${text.slice(0, max - 1)}…` : text);

const TrafficTopologyGraph: React.FC<TrafficTopologyGraphProps> = ({ topology, loading }) => {
  const { t } = useTranslation();
  const [showRoutes, setShowRoutes] = useState(true);
  const [showSelects, setShowSelects] = useState(true);
  const [showCalls, setShowCalls] = useState(true);
  const [showPolicy, setShowPolicy] = useState(true);

  const { positionedNodes, positions, width, height } = useMemo(() => {
    if (!topology?.nodes.length) {
      return { positionedNodes: [], positions: new Map<string, { x: number; y: number }>(), width: 800, height: 320 };
    }

    const columns: TopologyNode[][] = [[], [], []];
    for (const node of topology.nodes) {
      if (node.type === 'ingress') columns[0].push(node);
      else if (node.type === 'service') columns[1].push(node);
      else columns[2].push(node);
    }

    const maxRows = Math.max(columns[0].length, columns[1].length, columns[2].length, 1);
    const graphHeight = maxRows * (NODE_HEIGHT + ROW_GAP) - ROW_GAP + PADDING * 2;
    const graphWidth = PADDING * 2 + NODE_WIDTH * 3 + COL_GAP * 2;

    const posMap = new Map<string, { x: number; y: number }>();
    const placed: PositionedNode[] = [];

    columns.forEach((col, colIndex) => {
      const colX = PADDING + colIndex * (NODE_WIDTH + COL_GAP);
      const colHeight = col.length * (NODE_HEIGHT + ROW_GAP) - ROW_GAP;
      const offsetY = PADDING + (graphHeight - PADDING * 2 - colHeight) / 2;

      col.forEach((node, rowIndex) => {
        const x = colX;
        const y = offsetY + rowIndex * (NODE_HEIGHT + ROW_GAP);
        posMap.set(node.id, { x, y });
        placed.push({ ...node, x, y });
      });
    });

    return { positionedNodes: placed, positions: posMap, width: graphWidth, height: graphHeight };
  }, [topology]);

  const visibleEdges = useMemo(() => {
    if (!topology) return [];
    return topology.edges.filter((edge) => {
      if (edge.edgeType === 'routes') return showRoutes;
      if (edge.edgeType === 'selects') return showSelects;
      if (edge.edgeType === 'calls') return showCalls;
      if (edge.edgeType === 'policy_allow') return showPolicy;
      return true;
    });
  }, [topology, showRoutes, showSelects, showCalls, showPolicy]);

  if (!loading && (!topology || topology.nodes.length === 0)) {
    return <Empty description={t('trafficTopology.emptyGraph')} />;
  }

  const nodeColor = (type: string) => NODE_COLORS[type] || NODE_COLORS.default;

  const edgePath = (source: { x: number; y: number }, target: { x: number; y: number }) => {
    const forward = source.x < target.x;
    const x1 = forward ? source.x + NODE_WIDTH : source.x;
    const y1 = source.y + NODE_HEIGHT / 2;
    const x2 = forward ? target.x : target.x + NODE_WIDTH;
    const y2 = target.y + NODE_HEIGHT / 2;
    const curve = Math.max(Math.abs(x2 - x1) * 0.45, 48);
    if (forward) {
      const mx = (x1 + x2) / 2;
      return `M ${x1} ${y1} C ${mx} ${y1}, ${mx} ${y2}, ${x2} ${y2}`;
    }
    return `M ${x1} ${y1} C ${x1 - curve} ${y1}, ${x2 + curve} ${y2}, ${x2} ${y2}`;
  };

  const renderEdge = (edge: TopologyEdge, index: number) => {
    const source = positions.get(edge.source);
    const target = positions.get(edge.target);
    if (!source || !target) return null;

    const style = EDGE_STYLES[edge.edgeType] || EDGE_STYLES.routes;
    const label = edge.port || edge.evidence || edge.edgeType;

    return (
      <g key={`${edge.source}-${edge.target}-${edge.edgeType}-${index}`}>
        <path
          d={edgePath(source, target)}
          fill="none"
          stroke={style.stroke}
          strokeWidth={edge.edgeType === 'calls' ? 1.5 : 2}
          strokeDasharray={style.dash}
          markerEnd={`url(#arrow-${edge.edgeType})`}
          opacity={0.85}
        />
        {label && (
          <text
            x={(source.x + NODE_WIDTH + target.x) / 2}
            y={(source.y + target.y) / 2 + NODE_HEIGHT / 2 - 6}
            textAnchor="middle"
            fontSize={10}
            fill="#8c8c8c"
          >
            {truncate(label, 20)}
          </text>
        )}
      </g>
    );
  };

  return (
    <div>
      <Space wrap style={{ marginBottom: 12 }}>
        <Checkbox checked={showRoutes} onChange={(e) => setShowRoutes(e.target.checked)}>
          <Text style={{ color: EDGE_STYLES.routes.stroke }}>{t('trafficTopology.edgeRoutes')}</Text>
        </Checkbox>
        <Checkbox checked={showSelects} onChange={(e) => setShowSelects(e.target.checked)}>
          <Text style={{ color: EDGE_STYLES.selects.stroke }}>{t('trafficTopology.edgeSelects')}</Text>
        </Checkbox>
        <Checkbox checked={showCalls} onChange={(e) => setShowCalls(e.target.checked)}>
          <Text style={{ color: EDGE_STYLES.calls.stroke }}>{t('trafficTopology.edgeCalls')}</Text>
        </Checkbox>
        <Checkbox checked={showPolicy} onChange={(e) => setShowPolicy(e.target.checked)}>
          <Text style={{ color: EDGE_STYLES.policy_allow.stroke }}>{t('trafficTopology.edgePolicy')}</Text>
        </Checkbox>
      </Space>

      <div style={{ overflow: 'auto', border: '1px solid #f0f0f0', borderRadius: 8, background: '#fafafa' }}>
        <svg width={width} height={height} role="img" aria-label={t('trafficTopology.graphTitle')}>
          <defs>
            {Object.entries(EDGE_STYLES).map(([type, style]) => (
              <marker
                key={type}
                id={`arrow-${type}`}
                markerWidth="8"
                markerHeight="8"
                refX="7"
                refY="4"
                orient="auto"
              >
                <path d="M0,0 L8,4 L0,8 z" fill={style.stroke} />
              </marker>
            ))}
          </defs>

          {/* column headers */}
          {[t('trafficTopology.columns.ingress'), t('trafficTopology.columns.service'), t('trafficTopology.columns.workload')].map(
            (label, i) => (
              <text
                key={label}
                x={PADDING + i * (NODE_WIDTH + COL_GAP) + NODE_WIDTH / 2}
                y={20}
                textAnchor="middle"
                fontSize={12}
                fill="#595959"
                fontWeight={600}
              >
                {label}
              </text>
            ),
          )}

          {visibleEdges.map(renderEdge)}

          {positionedNodes.map((node) => {
            const colors = nodeColor(node.type);
            const isWorkload = WORKLOAD_TYPES.has(node.type);
            const title = `${node.type}/${node.namespace}/${node.name}`;
            const ports =
              node.extra?.ports && Array.isArray(node.extra.ports)
                ? (node.extra.ports as string[]).join(', ')
                : '';

            return (
              <g key={node.id} transform={`translate(${node.x}, ${node.y})`} style={{ cursor: 'default' }}>
                <title>
                  {title}
                  {ports ? ` | ${t('trafficTopology.columns.port')}: ${ports}` : ''}
                </title>
                <rect
                  width={NODE_WIDTH}
                  height={NODE_HEIGHT}
                  rx={6}
                  fill={colors.fill}
                  stroke={colors.stroke}
                  strokeWidth={1.5}
                />
                <text x={8} y={16} fontSize={10} fill="#8c8c8c">
                  {truncate(node.type, 12)}
                </text>
                <text x={8} y={32} fontSize={12} fill="#262626" fontWeight={600}>
                  {truncate(node.name)}
                </text>
                {isWorkload && (
                  <text x={NODE_WIDTH - 8} y={16} fontSize={9} fill="#8c8c8c" textAnchor="end">
                    {truncate(node.namespace, 10)}
                  </text>
                )}
              </g>
            );
          })}
        </svg>
      </div>
    </div>
  );
};

export default TrafficTopologyGraph;
