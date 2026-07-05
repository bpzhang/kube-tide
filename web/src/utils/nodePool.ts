const NODE_POOL_LABEL_KEYS = [
  'k8s.io/pool-name',
  'alibabacloud.com/nodepool-id',
  'node.alibabacloud.com/nodepool-id',
  'eks.amazonaws.com/nodegroup',
  'cloud.google.com/gke-nodepool',
];

export function getNodePoolName(labels?: Record<string, string>): string | undefined {
  if (!labels) return undefined;
  for (const key of NODE_POOL_LABEL_KEYS) {
    const value = labels[key];
    if (value) return value;
  }
  return undefined;
}

export function getNodePoolDisplayName(
  poolId: string | undefined,
  nodePools: Array<{ name: string; displayName?: string }>
): string | undefined {
  if (!poolId) return undefined;
  const pool = nodePools.find((item) => item.name === poolId);
  return pool?.displayName || poolId;
}

export function resolveNodePoolLabel(
  pool: { name: string; displayName?: string; source?: string }
): { title: string; subtitle?: string } {
  if (pool.displayName) {
    return {
      title: pool.displayName,
      subtitle: pool.source === 'discovered' ? pool.name : undefined,
    };
  }
  return { title: pool.name };
}
