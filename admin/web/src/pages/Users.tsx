import React, { useState, useEffect } from 'react';
import { 
  Table, 
  Card, 
  Button, 
  Modal, 
  Form, 
  Input, 
  Select, 
  message, 
  Tag,
  Space,
  Tooltip,
  Popconfirm
} from 'antd';
import { 
  UserAddOutlined, 
  EditOutlined, 
  DeleteOutlined, 
  LockOutlined, 
  UnlockOutlined,
  ReloadOutlined
} from '@ant-design/icons';
import { userAPI, AdminUser } from '../services/api';

const { Option } = Select;

const Users: React.FC = () => {
  const [users, setUsers] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingUser, setEditingUser] = useState<any>(null);
  const [form] = Form.useForm();
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });

  // 角色选项
  const roleOptions = [
    { value: 'super_admin', label: '超级管理员', color: 'red' },
    { value: 'system_admin', label: '系统管理员', color: 'orange' },
    { value: 'user_admin', label: '用户管理员', color: 'blue' },
    { value: 'monitor_admin', label: '监控管理员', color: 'purple' },
    { value: 'view_only', label: '只读用户', color: 'gray' },
  ];

  // 状态选项
  const statusOptions = [
    { value: 'active', label: '活跃', color: 'green' },
    { value: 'inactive', label: '非活跃', color: 'default' },
    { value: 'locked', label: '锁定', color: 'red' },
    { value: 'suspended', label: '暂停', color: 'volcano' },
  ];

  // 获取用户列表
  const fetchUsers = async (page = 1, pageSize = 10) => {
    setLoading(true);
    try {
      const response = await userAPI.getUsers({
        page,
        limit: pageSize,
      });
      
      setUsers(response.data.users);
      setPagination({
        current: page,
        pageSize,
        total: response.data.pagination.total,
      });
    } catch (error) {
      console.error('获取用户列表失败:', error);
      message.error('获取用户列表失败');
    } finally {
      setLoading(false);
    }
  };

  // 组件挂载时获取用户列表
  useEffect(() => {
    fetchUsers();
  }, []);

  // 表格列定义
  const columns = [
    {
      title: '用户名',
      dataIndex: 'username',
      key: 'username',
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      key: 'email',
    },
    {
      title: '角色',
      dataIndex: 'role',
      key: 'role',
      render: (role: string) => {
        const roleOption = roleOptions.find(opt => opt.value === role);
        return (
          <Tag color={roleOption?.color || 'default'}>
            {roleOption?.label || role}
          </Tag>
        );
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const statusOption = statusOptions.find(opt => opt.value === status);
        return (
          <Tag color={statusOption?.color || 'default'}>
            {statusOption?.label || status}
          </Tag>
        );
      },
    },
    {
      title: '最后登录',
      dataIndex: 'last_login_at',
      key: 'last_login_at',
      render: (date: string) => date ? new Date(date).toLocaleString() : '-',
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => new Date(date).toLocaleString(),
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: any, record: any) => (
        <Space>
          <Tooltip title="编辑">
            <Button 
              type="link" 
              icon={<EditOutlined />} 
              onClick={() => handleEdit(record)}
            />
          </Tooltip>
          <Tooltip title="删除">
            <Popconfirm
              title="确认删除用户?"
              onConfirm={() => handleDelete(record.id)}
              okText="确认"
              cancelText="取消"
            >
              <Button type="link" danger icon={<DeleteOutlined />} />
            </Popconfirm>
          </Tooltip>
          {record.status === 'locked' ? (
            <Tooltip title="解锁">
              <Button 
                type="link" 
                icon={<UnlockOutlined />} 
                onClick={() => handleUnlock(record.id)}
              />
            </Tooltip>
          ) : (
            <Tooltip title="锁定">
              <Button 
                type="link" 
                icon={<LockOutlined />} 
                onClick={() => handleLock(record.id)}
              />
            </Tooltip>
          )}
        </Space>
      ),
    },
  ];

  // 处理编辑用户
  const handleEdit = (user: any) => {
    setEditingUser(user);
    form.setFieldsValue({
      username: user.username,
      email: user.email,
      role: user.role,
      status: user.status,
      password: '', // 不显示现有密码
    });
    setModalVisible(true);
  };

  // 处理删除用户
  const handleDelete = async (userId: string) => {
    try {
      await userAPI.deleteUser(userId);
      message.success('用户删除成功');
      fetchUsers(pagination.current, pagination.pageSize);
    } catch (error) {
      console.error('删除用户失败:', error);
      message.error('删除用户失败');
    }
  };

  // 处理锁定用户
  const handleLock = async (userId: string) => {
    try {
      await userAPI.lockUser(userId);
      message.success('用户锁定成功');
      fetchUsers(pagination.current, pagination.pageSize);
    } catch (error) {
      console.error('锁定用户失败:', error);
      message.error('锁定用户失败');
    }
  };

  // 处理解锁用户
  const handleUnlock = async (userId: string) => {
    try {
      await userAPI.unlockUser(userId);
      message.success('用户解锁成功');
      fetchUsers(pagination.current, pagination.pageSize);
    } catch (error) {
      console.error('解锁用户失败:', error);
      message.error('解锁用户失败');
    }
  };

  // 提交表单
  const handleSubmit = async (values: any) => {
    try {
      if (editingUser) {
        // 更新用户
        await userAPI.updateUser(editingUser.id, values);
        message.success('用户更新成功');
      } else {
        // 创建用户
        await userAPI.createUser(values);
        message.success('用户创建成功');
      }
      
      setModalVisible(false);
      form.resetFields();
      setEditingUser(null);
      fetchUsers(pagination.current, pagination.pageSize);
    } catch (error) {
      console.error('操作失败:', error);
      message.error('操作失败');
    }
  };

  // 打开创建用户模态框
  const handleCreate = () => {
    setEditingUser(null);
    form.resetFields();
    setModalVisible(true);
  };

  return (
    <div>
      <Card 
        title="用户管理"
        extra={
          <Space>
            <Button 
              icon={<ReloadOutlined />} 
              onClick={() => fetchUsers(pagination.current, pagination.pageSize)}
            >
              刷新
            </Button>
            <Button 
              type="primary" 
              icon={<UserAddOutlined />} 
              onClick={handleCreate}
            >
              新增用户
            </Button>
          </Space>
        }
      >
        <Table
          columns={columns}
          dataSource={users}
          rowKey="id"
          loading={loading}
          pagination={{
            ...pagination,
            onChange: (page, pageSize) => {
              fetchUsers(page, pageSize || 10);
            },
          }}
        />
      </Card>

      {/* 用户编辑/创建模态框 */}
      <Modal
        title={editingUser ? '编辑用户' : '创建用户'}
        open={modalVisible}
        onCancel={() => {
          setModalVisible(false);
          form.resetFields();
          setEditingUser(null);
        }}
        footer={null}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
        >
          <Form.Item
            name="username"
            label="用户名"
            rules={[{ required: true, message: '请输入用户名!' }, { min: 3, max: 50 }]}
          >
            <Input placeholder="请输入用户名" />
          </Form.Item>

          <Form.Item
            name="email"
            label="邮箱"
            rules={[{ required: true, type: 'email', message: '请输入有效的邮箱!' }]}
          >
            <Input placeholder="请输入邮箱" />
          </Form.Item>

          <Form.Item
            name="password"
            label="密码"
            rules={[
              editingUser ? { required: false } : { required: true, min: 6, message: '密码至少6位!' }
            ]}
          >
            <Input.Password placeholder={editingUser ? "留空则不修改密码" : "请输入密码"} />
          </Form.Item>

          <Form.Item
            name="role"
            label="角色"
            rules={[{ required: true, message: '请选择角色!' }]}
          >
            <Select placeholder="请选择角色">
              {roleOptions.map(option => (
                <Option key={option.value} value={option.value}>
                  <Tag color={option.color}>{option.label}</Tag>
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            name="status"
            label="状态"
            rules={[{ required: true, message: '请选择状态!' }]}
          >
            <Select placeholder="请选择状态">
              {statusOptions.map(option => (
                <Option key={option.value} value={option.value}>
                  <Tag color={option.color}>{option.label}</Tag>
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                {editingUser ? '更新' : '创建'}
              </Button>
              <Button onClick={() => {
                setModalVisible(false);
                form.resetFields();
                setEditingUser(null);
              }}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default Users;