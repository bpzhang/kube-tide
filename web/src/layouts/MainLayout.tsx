import React from 'react';
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
} from '@ant-design/icons';
import LanguageSwitcher from '../components/common/LanguageSwitcher';

const { Header, Sider, Content } = Layout;

const MainLayout: React.FC = () => {
  const { t } = useTranslation();

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
        <div style={{ fontSize: '20px', fontWeight: 'bold', padding: '0 24px' }}>
          {t('app.title')}
        </div>
        <div style={{ padding: '0 24px' }}>
          <LanguageSwitcher />
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