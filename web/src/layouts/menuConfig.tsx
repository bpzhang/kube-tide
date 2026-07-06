import type { ReactNode } from 'react';
import type { MenuProps } from 'antd';
import { Link } from 'react-router-dom';
import {
  DashboardOutlined,
  ClusterOutlined,
  CloudServerOutlined,
  AppstoreOutlined,
  BlockOutlined,
  ApartmentOutlined,
  ProductOutlined,
  SettingOutlined,
  FolderOutlined,
  GlobalOutlined,
  DatabaseOutlined,
  SafetyCertificateOutlined,
  LineChartOutlined,
  ScheduleOutlined,
  ThunderboltOutlined,
  FileProtectOutlined,
  KeyOutlined,
} from '@ant-design/icons';

export type MenuConfigItem = {
  key: string;
  label: string;
  icon: ReactNode;
  path?: string;
  match?: (pathname: string) => boolean;
  children?: MenuConfigItem[];
};

export type RouteMenuState = {
  selectedKeys: string[];
  routeOpenKeys: string[];
};

export const createMenuConfig = (
  t: (key: string) => string
): MenuConfigItem[] => [
  {
    key: 'dashboard',
    icon: <DashboardOutlined />,
    label: t('navigation.dashboard'),
    path: '/dashboard',
    match: (pathname) => pathname === '/' || pathname.startsWith('/dashboard'),
  },
  {
    key: 'clusters',
    icon: <ClusterOutlined />,
    label: t('navigation.clusters'),
    path: '/clusters',
    match: (pathname) => pathname.startsWith('/clusters') && !pathname.includes('/deployments'),
  },
  {
    key: 'nodes',
    icon: <CloudServerOutlined />,
    label: t('navigation.nodes'),
    path: '/nodes',
    match: (pathname) => pathname.startsWith('/nodes'),
  },
  {
    key: 'namespaces',
    icon: <FolderOutlined />,
    label: t('navigation.namespaces'),
    path: '/namespaces',
    match: (pathname) => pathname.startsWith('/namespaces'),
  },
  {
    key: 'config',
    icon: <SettingOutlined />,
    label: t('navigation.config'),
    children: [
      {
        key: 'configmaps',
        icon: <FileProtectOutlined />,
        label: t('navigation.configMaps'),
        path: '/config/configmaps',
        match: (pathname) => pathname.startsWith('/config/configmaps'),
      },
      {
        key: 'secrets',
        icon: <KeyOutlined />,
        label: t('navigation.secrets'),
        path: '/config/secrets',
        match: (pathname) => pathname.startsWith('/config/secrets'),
      },
    ],
  },
  {
    key: 'workloads',
    icon: <AppstoreOutlined />,
    label: t('navigation.workloads'),
    children: [
      {
        key: 'deployments',
        icon: <BlockOutlined />,
        label: t('navigation.deployments'),
        path: '/workloads/deployments',
        match: (pathname) => pathname.startsWith('/workloads/deployments') || pathname.includes('/deployments'),
      },
      {
        key: 'statefulsets',
        icon: <BlockOutlined />,
        label: t('navigation.statefulsets'),
        path: '/workloads/statefulsets',
        match: (pathname) => pathname.startsWith('/workloads/statefulsets'),
      },
      {
        key: 'daemonsets',
        icon: <ThunderboltOutlined />,
        label: t('navigation.daemonSets'),
        path: '/workloads/daemonsets',
        match: (pathname) => pathname.startsWith('/workloads/daemonsets'),
      },
      {
        key: 'hpas',
        icon: <LineChartOutlined />,
        label: t('navigation.hpas'),
        path: '/workloads/hpas',
        match: (pathname) => pathname.startsWith('/workloads/hpas'),
      },
      {
        key: 'jobs',
        icon: <ScheduleOutlined />,
        label: t('navigation.jobs'),
        path: '/workloads/jobs',
        match: (pathname) => pathname.startsWith('/workloads/jobs'),
      },
      {
        key: 'cronjobs',
        icon: <ScheduleOutlined />,
        label: t('navigation.cronJobs'),
        path: '/workloads/cronjobs',
        match: (pathname) => pathname.startsWith('/workloads/cronjobs'),
      },
      {
        key: 'pods',
        icon: <ProductOutlined />,
        label: t('navigation.pods'),
        path: '/workloads/pods',
        match: (pathname) => pathname.startsWith('/workloads/pods'),
      },
    ],
  },
  {
    key: 'network',
    icon: <ApartmentOutlined />,
    label: t('navigation.network'),
    children: [
      {
        key: 'services',
        icon: <ApartmentOutlined />,
        label: t('navigation.services'),
        path: '/workloads/services',
        match: (pathname) => pathname.startsWith('/workloads/services'),
      },
      {
        key: 'ingresses',
        icon: <GlobalOutlined />,
        label: t('navigation.ingresses'),
        path: '/network/ingresses',
        match: (pathname) => pathname.startsWith('/network/ingresses'),
      },
      {
        key: 'networkpolicies',
        icon: <SafetyCertificateOutlined />,
        label: t('navigation.networkPolicies'),
        path: '/network/networkpolicies',
        match: (pathname) => pathname.startsWith('/network/networkpolicies'),
      },
    ],
  },
  {
    key: 'storage',
    icon: <DatabaseOutlined />,
    label: t('navigation.storage'),
    children: [
      {
        key: 'pvcs',
        icon: <DatabaseOutlined />,
        label: t('navigation.pvcs'),
        path: '/storage/pvcs',
        match: (pathname) => pathname.startsWith('/storage/pvcs'),
      },
      {
        key: 'pvs',
        icon: <DatabaseOutlined />,
        label: t('navigation.pvs'),
        path: '/storage/pvs',
        match: (pathname) => pathname.startsWith('/storage/pvs'),
      },
      {
        key: 'storageclasses',
        icon: <DatabaseOutlined />,
        label: t('navigation.storageClasses'),
        path: '/storage/storageclasses',
        match: (pathname) => pathname.startsWith('/storage/storageclasses'),
      },
    ],
  },
  {
    key: 'governance',
    icon: <SafetyCertificateOutlined />,
    label: t('navigation.governance'),
    children: [
      {
        key: 'resourcequotas',
        icon: <SafetyCertificateOutlined />,
        label: t('navigation.resourceQuotas'),
        path: '/governance/resourcequotas',
        match: (pathname) => pathname.startsWith('/governance/resourcequotas'),
      },
      {
        key: 'limitranges',
        icon: <SafetyCertificateOutlined />,
        label: t('navigation.limitRanges'),
        path: '/governance/limitranges',
        match: (pathname) => pathname.startsWith('/governance/limitranges'),
      },
      {
        key: 'pdbs',
        icon: <SafetyCertificateOutlined />,
        label: t('navigation.pdbs'),
        path: '/governance/pdbs',
        match: (pathname) => pathname.startsWith('/governance/pdbs'),
      },
      {
        key: 'rbac',
        icon: <KeyOutlined />,
        label: t('navigation.rbac'),
        path: '/governance/rbac',
        match: (pathname) => pathname.startsWith('/governance/rbac'),
      },
    ],
  },
  {
    key: 'observability',
    icon: <LineChartOutlined />,
    label: t('navigation.observability'),
    children: [
      {
        key: 'label-logs',
        icon: <LineChartOutlined />,
        label: t('navigation.labelLogs'),
        path: '/observability/label-logs',
        match: (pathname) => pathname.startsWith('/observability/label-logs'),
      },
    ],
  },
];

export const buildMenuItems = (items: MenuConfigItem[]): MenuProps['items'] => {
  return items.map((item) => ({
    key: item.key,
    icon: item.icon,
    label: item.path ? <Link to={item.path}>{item.label}</Link> : item.label,
    children: item.children ? buildMenuItems(item.children) : undefined,
  }));
};

export const resolveMenuState = (
  pathname: string,
  items: MenuConfigItem[],
  parentKeys: string[] = []
): RouteMenuState => {
  for (const item of items) {
    if (item.match?.(pathname)) {
      return {
        selectedKeys: [item.key],
        routeOpenKeys: parentKeys,
      };
    }

    if (item.children) {
      const childState = resolveMenuState(pathname, item.children, [...parentKeys, item.key]);
      if (childState.selectedKeys.length > 0) {
        return childState;
      }
    }
  }

  return {
    selectedKeys: [],
    routeOpenKeys: [],
  };
};
