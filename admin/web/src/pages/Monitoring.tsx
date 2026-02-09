import React, { useState, useEffect } from 'react';
import {
  Card,
  Row,
  Col,
  Statistic,
  Table,
  Tabs,
  Space,
  Tag,
  Typography,
  Progress,
  Descriptions
} from 'antd';
import {
  DashboardOutlined,
  UserOutlined,
  LockOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  InfoCircleOutlined,
  ClockCircleOutlined,
  DatabaseOutlined,
  ApiOutlined,
  CloudServerOutlined,
  WarningOutlined,
  BugOutlined,
  HourglassOutlined,
  FieldTimeOutlined
} from '@ant-design/icons';
import { monitoringAPI, auditAPI, DashboardStats, AuditLog } from '../services/api';

const { TabPane } = Tabs;
const { Title, Text } = Typography;

const Monitoring: React.FC = () => {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [metrics, setMetrics] = useState<any>({});

  // 获取仪表板统计数据
  const fetchDashboardStats = async () => {
    try {
      const response = await monitoringAPI.getStatus();
      setStats(response.data);
    } catch (error) {
      console.error('获取仪表板统计数据失败:', error);
    }
  };

  // 获取监控指标
  const fetchMetrics = async () => {
    try {
      const response = await monitoringAPI.getMetrics();
      setMetrics(response.data);
    } catch (error) {
      console.error('获取监控指标失败:', error);
    }
  };

  // 获取审计日志
  const fetchAuditLogs = async (page = 1, limit = 10) => {
    setLoading(true);
    try {
      const response = await auditAPI.getLogs({
        page,
        limit,
      });
      setLogs(response.data.logs || []);
    } catch (error) {
      console.error('获取审计日志失败:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDashboardStats();
    fetchMetrics();
    fetchAuditLogs();
  }, []);

  // 系统状态指标
  const systemMetrics = metrics.system_status || {
    uptime: '0天 0小时 0分钟',
    cpu_usage: 0,
    memory_usage: 0,
    disk_usage: 0,
    network_io: { rx: 0, tx: 0 },
    db_connections: 0,
    active_users: 0,
  };

  // 审计日志表格列定义
  const logColumns = [
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      render: (date: string) => new Date(date).toLocaleString(),
    },
    {
      title: '用户名',
      dataIndex: 'username',
      key: 'username',
    },
    {
      title: 'IP地址',
      dataIndex: 'ip_address',
      key: 'ip_address',
    },
    {
      title: '操作',
      dataIndex: 'action',
      key: 'action',
    },
    {
      title: '资源',
      dataIndex: 'resource',
      key: 'resource',
    },
    {
      title: '详情',
      dataIndex: 'details',
      key: 'details',
      render: (details: Record<string, any>) => {
        return Object.entries(details).map(([key, value]) => (
          <Tag key={key} color="blue">{`${key}: ${value}`}</Tag>
        ));
      },
    },
  ];

  return (
    <div className="monitoring-page">
      <Title level={2} style={{ marginBottom: 24 }}>
        <DashboardOutlined /> 系统监控
      </Title>

      <Row gutter={[16, 16]}>
        <Col span={6}>
          <Card>
            <Statistic
              title="总用户数"
              value={stats?.total_users || 0}
              prefix={<UserOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="活跃用户"
              value={stats?.active_users || 0}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="锁定用户"
              value={stats?.locked_users || 0}
              prefix={<LockOutlined />}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="在线用户"
              value={systemMetrics.active_users || 0}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col span={12}>
          <Card title="系统状态">
            <Descriptions column={1}>
              <Descriptions.Item label="运行时间">
                <Text>{systemMetrics.uptime}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="CPU使用率">
                <Space>
                  <Progress
                    type="circle"
                    percent={systemMetrics.cpu_usage}
                    width={50}
                    status={systemMetrics.cpu_usage > 80 ? 'exception' : 'normal'}
                  />
                  <Text>{systemMetrics.cpu_usage}%</Text>
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label="内存使用率">
                <Space>
                  <Progress
                    type="circle"
                    percent={systemMetrics.memory_usage}
                    width={50}
                    status={systemMetrics.memory_usage > 80 ? 'exception' : 'normal'}
                  />
                  <Text>{systemMetrics.memory_usage}%</Text>
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label="磁盘使用率">
                <Space>
                  <Progress
                    type="circle"
                    percent={systemMetrics.disk_usage}
                    width={50}
                    status={systemMetrics.disk_usage > 80 ? 'exception' : 'normal'}
                  />
                  <Text>{systemMetrics.disk_usage}%</Text>
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label="数据库连接数">
                <Text>{systemMetrics.db_connections}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="网络IO">
                <Text>{`接收: ${(systemMetrics.network_io?.rx || 0).toFixed(2)}MB, 发送: ${(systemMetrics.network_io?.tx || 0).toFixed(2)}MB`}</Text>
              </Descriptions.Item>
            </Descriptions>
          </Card>
        </Col>
        <Col span={12}>
          <Card title="最近活动">
            <Table
              columns={[
                {
                  title: '时间',
                  dataIndex: 'timestamp',
                  key: 'timestamp',
                  render: (date: string) => new Date(date).toLocaleString(),
                },
                {
                  title: '用户',
                  dataIndex: 'username',
                  key: 'username',
                },
                {
                  title: '操作',
                  dataIndex: 'action',
                  key: 'action',
                  render: (action: string) => {
                    let color = 'default';
                    if (action.includes('login')) color = 'green';
                    else if (action.includes('logout')) color = 'orange';
                    else if (action.includes('create') || action.includes('add')) color = 'blue';
                    else if (action.includes('delete') || action.includes('remove')) color = 'red';
                    else if (action.includes('update') || action.includes('edit')) color = 'geekblue';
                    
                    return <Tag color={color}>{action}</Tag>;
                  },
                },
                {
                  title: '资源',
                  dataIndex: 'resource',
                  key: 'resource',
                },
              ]}
              dataSource={stats?.recent_logs || []}
              rowKey="id"
              pagination={false}
            />
          </Card>
        </Col>
      </Row>

      <Card title="审计日志" style={{ marginTop: 16 }}>
        <Table
          columns={logColumns}
          dataSource={logs}
          rowKey="id"
          loading={loading}
          pagination={{
            defaultPageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条记录`,
          }}
        />
      </Card>
    </div>
  );
};

export default Monitoring;