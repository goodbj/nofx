import React, { useState, useEffect } from 'react';
import { Card, Col, Row, Statistic, Table, Typography, Space } from 'antd';
import { ArrowUpOutlined, ArrowDownOutlined, UserOutlined, LockOutlined, SettingOutlined, AuditOutlined } from '@ant-design/icons';
import axios from 'axios';
import moment from 'moment';

const { Title } = Typography;

const Dashboard: React.FC = () => {
  const [dashboardData, setDashboardData] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchDashboardData();
  }, []);

  const fetchDashboardData = async () => {
    try {
      const token = localStorage.getItem('admin_token');
      const response = await axios.get('/api/dashboard', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      setDashboardData(response.data.data);
    } catch (error) {
      console.error('获取仪表板数据失败:', error);
    } finally {
      setLoading(false);
    }
  };

  // 模拟最近活动数据
  const recentActivities = [
    {
      key: '1',
      action: '用户登录',
      user: 'admin',
      time: moment().subtract(5, 'minutes').format('HH:mm:ss'),
      status: '成功',
    },
    {
      key: '2',
      action: '创建用户',
      user: 'superadmin',
      time: moment().subtract(1, 'hours').format('HH:mm:ss'),
      status: '成功',
    },
    {
      key: '3',
      action: '系统配置更新',
      user: 'system',
      time: moment().subtract(2, 'hours').format('HH:mm:ss'),
      status: '成功',
    },
    {
      key: '4',
      action: '权限分配',
      user: 'admin',
      time: moment().subtract(3, 'hours').format('HH:mm:ss'),
      status: '成功',
    },
    {
      key: '5',
      action: '用户锁定',
      user: 'moderator',
      time: moment().subtract(5, 'hours').format('HH:mm:ss'),
      status: '成功',
    },
  ];

  const columns = [
    {
      title: '操作',
      dataIndex: 'action',
      key: 'action',
    },
    {
      title: '用户',
      dataIndex: 'user',
      key: 'user',
    },
    {
      title: '时间',
      dataIndex: 'time',
      key: 'time',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
    },
  ];

  return (
    <div>
      <Title level={2}>管理员仪表板</Title>
      
      {dashboardData && (
        <Row gutter={16}>
          <Col span={6}>
            <Card>
              <Statistic
                title="总用户数"
                value={dashboardData.statistics.total_users}
                prefix={<UserOutlined />}
                valueStyle={{ color: '#3f8600' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="审计日志"
                value={dashboardData.statistics.total_audits}
                prefix={<AuditOutlined />}
                valueStyle={{ color: '#1890ff' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="活跃会话"
                value={dashboardData.statistics.active_sessions}
                prefix={<LockOutlined />}
                valueStyle={{ color: '#722ed1' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="系统健康"
                value="正常"
                prefix={<SettingOutlined />}
                valueStyle={{ color: '#52c41a' }}
              />
            </Card>
          </Col>
        </Row>
      )}

      <Row gutter={16} style={{ marginTop: 16 }}>
        <Col span={18}>
          <Card title="最近活动">
            <Table 
              columns={columns} 
              dataSource={recentActivities} 
              pagination={{ pageSize: 5 }}
              size="small"
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card title="快速操作">
            <Space direction="vertical" style={{ width: '100%' }}>
              <a>创建新用户</a>
              <a>系统配置</a>
              <a>查看日志</a>
              <a>权限管理</a>
              <a>备份数据</a>
            </Space>
          </Card>
        </Col>
      </Row>

      <Row gutter={16} style={{ marginTop: 16 }}>
        <Col span={24}>
          <Card title="系统信息">
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '16px' }}>
              <div>
                <h4>服务器状态</h4>
                <p>运行正常</p>
              </div>
              <div>
                <h4>数据库状态</h4>
                <p>连接正常</p>
              </div>
              <div>
                <h4>当前用户</h4>
                <p>{dashboardData?.user?.username || '-'}</p>
              </div>
              <div>
                <h4>用户角色</h4>
                <p>{dashboardData?.user?.role || '-'}</p>
              </div>
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;