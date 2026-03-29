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
    match: (pathname) => pathname.startsWith('/clusters'),
  },
  {
    key: 'nodes',
    icon: <CloudServerOutlined />,
    label: t('navigation.nodes'),
    path: '/nodes',
    match: (pathname) => pathname.startsWith('/nodes'),
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
