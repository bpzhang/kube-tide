import React, { useState } from 'react';
import { Layout, Menu } from 'antd';
import { Link, Outlet } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  DashboardOutlined,
  ClusterOutlined,
  CloudServerOutlined,
  AppstoreOutlined,
  BlockOutlined,
  ApartmentOutlined,
  ProductOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
} from '@ant-design/icons';
import LanguageSwitcher from '../components/common/LanguageSwitcher';

const { Header, Sider, Content } = Layout;

const MainLayout: React.FC = () => {
  const { t } = useTranslation();
  const [collapsed, setCollapsed] = useState(false);

  const toggleCollapsed = () => {
    setCollapsed(!collapsed);
  };

  const menuItems = [
    {
      key: 'dashboard',
      icon: <DashboardOutlined />,
      label: <Link to="/dashboard">{t('navigation.dashboard')}</Link>,
    },
    {
      key: 'clusters',
      icon: <ClusterOutlined />,
      label: <Link to="/clusters">{t('navigation.clusters')}</Link>,
    },
    {
      key: 'nodes',
      icon: <CloudServerOutlined />,
      label: <Link to="/nodes">{t('navigation.nodes')}</Link>,
    },
    {
      key: 'workloads',
      icon: <AppstoreOutlined />,
      label: t('navigation.workloads'),
      children: [
        {
          key: 'deployments',
          icon: <BlockOutlined />,
          label: <Link to="/workloads/deployments">{t('navigation.deployments')}</Link>,
        },
        {
          key: 'statefulsets',
          icon: <BlockOutlined />,
          label: <Link to="/workloads/statefulsets">{t('navigation.statefulsets')}</Link>,
        },
        {
          key: 'pods',
          icon: <ProductOutlined />,
          label: <Link to="/workloads/pods">{t('navigation.pods')}</Link>,
        },
        {
          key: 'services',
          icon: <ApartmentOutlined />,
          label: <Link to="/workloads/services">{t('navigation.services')}</Link>,
        },
      ],
    },
  ];

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header style={{ padding: 0, background: '#fff', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div style={{ fontSize: '20px', fontWeight: 'bold', padding: '0 24px', display: 'flex', alignItems: 'center' }}>
          {t('app.title')}
          <div 
            onClick={toggleCollapsed}
            style={{ marginLeft: '20px', cursor: 'pointer', fontSize: '16px' }}
          >
            {collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
          </div>
        </div>
        <div style={{ padding: '0 24px' }}>
          <LanguageSwitcher />
        </div>
      </Header>
      <Layout>
        <Sider 
          width={200} 
          collapsible 
          collapsed={collapsed} 
          onCollapse={setCollapsed} 
          trigger={null}
          style={{ background: '#fff' }}
        >
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