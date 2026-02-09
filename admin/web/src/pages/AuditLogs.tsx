import React, { useState, useEffect } from 'react';
import {
  Card,
  Table,
  DatePicker,
  Input,
  Select,
  Button,
  Space,
  Tag,
  Typography
} from 'antd';
import {
  AuditOutlined,
  SearchOutlined,
  ReloadOutlined,
  UserOutlined,
  HistoryOutlined
} from '@ant-design/icons';
import { auditAPI, AuditLog } from '../services/api';

const { RangePicker } = DatePicker;
const { Title, Text } = Typography;
const { Option } = Select;

const AuditLogs: React.FC = () => {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });
  const [filters, setFilters] = useState({
    start_date: '',
    end_date: '',
    user_id: '',
    username: '',
    action: '',
    ip_address: '',
  });

  // 获取审计日志
  const fetchAuditLogs = async (page = 1, limit = 10) => {
    setLoading(true);
    try {
      const params = {
        ...filters,
        page,
        limit,
      };
      
      const response = await auditAPI.getLogs(params);
      setLogs(response.data.logs || []);
      setPagination({
        current: page,
        pageSize: limit,
        total: response.data.pagination?.total || 0,
      });
    } catch (error) {
      console.error('获取审计日志失败:', error);
    } finally {
      setLoading(false);
    }
  };

  // 组件挂载时获取日志
  useEffect(() => {
    fetchAuditLogs();
  }, []);

  // 重置过滤器
  const resetFilters = () => {
    setFilters({
      start_date: '',
      end_date: '',
      user_id: '',
      username: '',
      action: '',
      ip_address: '',
    });
    fetchAuditLogs();
  };

  // 表格列定义
  const columns = [
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      render: (date: string) => new Date(date).toLocaleString(),
      sorter: true,
    },
    {
      title: '用户名',
      dataIndex: 'username',
      key: 'username',
    },
    {
      title: '用户ID',
      dataIndex: 'user_id',
      key: 'user_id',
    },
    {
      title: 'IP地址',
      dataIndex: 'ip_address',
      key: 'ip_address',
    },
    {
      title: '用户代理',
      dataIndex: 'user_agent',
      key: 'user_agent',
      render: (agent: string) => agent?.substring(0, 50) + (agent?.length > 50 ? '...' : ''),
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
        else if (action.includes('view') || action.includes('get')) color = 'cyan';
        
        return <Tag color={color}>{action}</Tag>;
      },
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
        if (!details || typeof details !== 'object') return '-';
        
        return Object.entries(details).slice(0, 2).map(([key, value], index) => (
          <Tag key={key} color="blue" style={{ marginBottom: '2px' }}>
            {key}: {typeof value === 'object' ? JSON.stringify(value) : String(value)}
          </Tag>
        ));
      },
    },
  ];

  return (
    <div className="audit-logs-page">
      <Title level={2} style={{ marginBottom: 24 }}>
        <AuditOutlined /> 审计日志
      </Title>

      <Card 
        title="查询条件"
        style={{ marginBottom: 16 }}
      >
        <Space wrap style={{ width: '100%' }}>
          <DatePicker
            placeholder="开始日期"
            onChange={(date) => setFilters({
              ...filters,
              start_date: date ? date.format('YYYY-MM-DD') : ''
            })}
          />
          <DatePicker
            placeholder="结束日期"
            onChange={(date) => setFilters({
              ...filters,
              end_date: date ? date.format('YYYY-MM-DD') : ''
            })}
          />
          <Input
            placeholder="用户名"
            value={filters.username}
            onChange={(e) => setFilters({ ...filters, username: e.target.value })}
            style={{ width: 150 }}
          />
          <Input
            placeholder="用户ID"
            value={filters.user_id}
            onChange={(e) => setFilters({ ...filters, user_id: e.target.value })}
            style={{ width: 150 }}
          />
          <Select
            placeholder="操作类型"
            value={filters.action}
            onChange={(value) => setFilters({ ...filters, action: value })}
            style={{ width: 150 }}
            allowClear
          >
            <Option value="user_create">用户创建</Option>
            <Option value="user_update">用户更新</Option>
            <Option value="user_delete">用户删除</Option>
            <Option value="user_lock">用户锁定</Option>
            <Option value="user_unlock">用户解锁</Option>
            <Option value="login">登录</Option>
            <Option value="logout">登出</Option>
            <Option value="permission_grant">权限授予</Option>
            <Option value="permission_revoke">权限撤销</Option>
            <Option value="config_update">配置更新</Option>
          </Select>
          <Input
            placeholder="IP地址"
            value={filters.ip_address}
            onChange={(e) => setFilters({ ...filters, ip_address: e.target.value })}
            style={{ width: 150 }}
          />
          <Button 
            type="primary" 
            icon={<SearchOutlined />} 
            onClick={() => fetchAuditLogs()}
          >
            查询
          </Button>
          <Button 
            icon={<ReloadOutlined />} 
            onClick={resetFilters}
          >
            重置
          </Button>
        </Space>
      </Card>

      <Card>
        <Table
          columns={columns}
          dataSource={logs}
          rowKey="id"
          loading={loading}
          pagination={{
            ...pagination,
            onChange: (page, pageSize) => {
              fetchAuditLogs(page, pageSize || 10);
            },
          }}
        />
      </Card>
    </div>
  );
};

export default AuditLogs;