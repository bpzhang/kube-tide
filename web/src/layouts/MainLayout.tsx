import React from 'react';
import { Layout, Menu } from 'antd';
import { Link, Outlet } from 'react-router-dom';
import {
  DashboardOutlined,
  ClusterOutlined,
  CloudServerOutlined,
  AppstoreOutlined,
  BlockOutlined,
  ApartmentOutlined,
  ProductOutlined,
} from '@ant-design/icons';

const { Header, Sider, Content } = Layout;

const MainLayout: React.FC = () => {
  const menuItems = [
    {
      key: 'dashboard',
      icon: <DashboardOutlined />,
      label: <Link to="/dashboard">仪表盘</Link>,
    },
    {
      key: 'clusters',
      icon: <ClusterOutlined />,
      label: <Link to="/clusters">集群管理</Link>,
    },
    {
      key: 'nodes',
      icon: <CloudServerOutlined />,
      label: <Link to="/nodes">节点管理</Link>,
    },
    {
      key: 'workloads',
      icon: <AppstoreOutlined />,
      label: '工作负载',
      children: [
        {
          key: 'deployments',
          icon: <BlockOutlined />,
          label: <Link to="/workloads/deployments">Deployment 管理</Link>,
        },
        {
          key: 'pods',
          icon: <ProductOutlined />,
          label: <Link to="/workloads/pods">Pod 管理</Link>,
        },
        {
          key: 'services',
          icon: <ApartmentOutlined />,
          label: <Link to="/workloads/services">Service 管理</Link>,
        },
      ],
    },
  ];

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header style={{ padding: 0, background: '#fff' }}>
        <div style={{ fontSize: '20px', fontWeight: 'bold', padding: '0 24px' }}>
          Kubernetes 平台
        </div>
      </Header>
      <Layout>
        <Sider width={200} style={{ background: '#fff' }}>
          <Menu
            mode="inline"
            defaultSelectedKeys={['dashboard']}
            style={{ height: '100%', borderRight: 0 }}
            items={menuItems}
          />
        </Sider>
        <Layout style={{ padding: '24px' }}>
          <Content style={{ background: '#fff', padding: 24, margin: 0, minHeight: 280 }}>
            <Outlet />
          </Content>
        </Layout>
      </Layout>
    </Layout>
  );
};

export default MainLayout;