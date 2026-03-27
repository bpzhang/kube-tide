import React, { useEffect, useMemo, useState } from 'react';
import { Layout, Menu } from 'antd';
import { Outlet, useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { MenuFoldOutlined, MenuUnfoldOutlined } from '@ant-design/icons';
import LanguageSwitcher from '../components/common/LanguageSwitcher';
import { buildMenuItems, createMenuConfig, resolveMenuState } from './menuConfig';

const { Header, Sider, Content } = Layout;

const MainLayout: React.FC = () => {
  const { t } = useTranslation();
  const [collapsed, setCollapsed] = useState(false);
  const location = useLocation();

  const toggleCollapsed = () => {
    setCollapsed(!collapsed);
  };

  const menuConfig = useMemo(() => createMenuConfig(t), [t]);

  const menuItems = useMemo(() => buildMenuItems(menuConfig), [menuConfig]);

  const { selectedKeys, routeOpenKeys } = useMemo(
    () => resolveMenuState(location.pathname, menuConfig),
    [location.pathname, menuConfig]
  );

  const [openKeys, setOpenKeys] = useState<string[]>(routeOpenKeys);

  useEffect(() => {
    if (routeOpenKeys.length > 0) {
      setOpenKeys(routeOpenKeys);
    }
  }, [routeOpenKeys]);

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
            selectedKeys={selectedKeys}
            openKeys={collapsed ? [] : openKeys}
            onOpenChange={(keys) => setOpenKeys(keys as string[])}
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