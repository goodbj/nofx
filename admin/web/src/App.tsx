import React, { useState, useEffect } from 'react';
import { Layout, Menu, theme } from 'antd';
import {
  DashboardOutlined,
  UserOutlined,
  LockOutlined,
  SettingOutlined,
  BarChartOutlined,
  AuditOutlined,
  TeamOutlined,
} from '@ant-design/icons';
import { useLocation, useNavigate } from 'react-router-dom';

import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import Users from './pages/Users';
import Permissions from './pages/Permissions';
import SystemConfig from './pages/SystemConfig';
import Monitoring from './pages/Monitoring';
import AuditLogs from './pages/AuditLogs';
import Profile from './pages/Profile';

const { Header, Sider, Content } = Layout;

const App: React.FC = () => {
  const [collapsed, setCollapsed] = useState(false);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [user, setUser] = useState<any>(null);
  const {
    token: { colorBgContainer },
  } = theme.useToken();
  
  const location = useLocation();
  const navigate = useNavigate();

  // 检查认证状态
  useEffect(() => {
    const token = localStorage.getItem('admin_token');
    if (token) {
      setIsAuthenticated(true);
      // 获取用户信息
      fetchUserProfile();
    } else if (location.pathname !== '/login') {
      navigate('/login');
    }
  }, [location.pathname]);

  const fetchUserProfile = async () => {
    try {
      const response = await fetch('/api/profile', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('admin_token')}`,
        },
      });
      
      if (response.ok) {
        const data = await response.json();
        setUser(data.user);
      } else {
        localStorage.removeItem('admin_token');
        setIsAuthenticated(false);
        navigate('/login');
      }
    } catch (error) {
      console.error('获取用户信息失败:', error);
      localStorage.removeItem('admin_token');
      setIsAuthenticated(false);
      navigate('/login');
    }
  };

  const handleLogout = () => {
    localStorage.removeItem('admin_token');
    setIsAuthenticated(false);
    setUser(null);
    navigate('/login');
  };

  // 菜单项
  const menuItems = [
    {
      key: '/',
      icon: <DashboardOutlined />,
      label: '仪表板',
    },
    {
      key: '/users',
      icon: <UserOutlined />,
      label: '用户管理',
    },
    {
      key: '/permissions',
      icon: <LockOutlined />,
      label: '权限管理',
    },
    {
      key: '/system-config',
      icon: <SettingOutlined />,
      label: '系统配置',
    },
    {
      key: '/monitoring',
      icon: <BarChartOutlined />,
      label: '系统监控',
    },
    {
      key: '/audit-logs',
      icon: <AuditOutlined />,
      label: '审计日志',
    },
    {
      key: '/profile',
      icon: <TeamOutlined />,
      label: '个人资料',
    },
  ];

  // 未认证时显示登录页
  if (!isAuthenticated && location.pathname !== '/login') {
    return <Login onLogin={() => setIsAuthenticated(true)} />;
  }

  // 登录页面
  if (location.pathname === '/login') {
    return <Login onLogin={() => setIsAuthenticated(true)} />;
  }

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider collapsible collapsed={collapsed} onCollapse={(value) => setCollapsed(value)}>
        <div className="demo-logo-vertical" />
        <Menu
          theme="dark"
          defaultSelectedKeys={['/']}
          mode="inline"
          items={menuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <Layout>
        <Header style={{ padding: 0, background: colorBgContainer }}>
          <div style={{ float: 'right', marginRight: 20, color: '#fff' }}>
            欢迎, {user?.username} ({user?.role}) 
            <a onClick={handleLogout} style={{ marginLeft: 20, color: '#fff' }}>退出</a>
          </div>
        </Header>
        <Content style={{ margin: '24px 16px 0', overflow: 'initial' }}>
          <div style={{ padding: 24, textAlign: 'center' }}>
            {location.pathname === '/' && <Dashboard />}
            {location.pathname === '/users' && <Users />}
            {location.pathname === '/permissions' && <Permissions />}
            {location.pathname === '/system-config' && <SystemConfig />}
            {location.pathname === '/monitoring' && <Monitoring />}
            {location.pathname === '/audit-logs' && <AuditLogs />}
            {location.pathname === '/profile' && <Profile />}
          </div>
        </Content>
      </Layout>
    </Layout>
  );
};

export default App;