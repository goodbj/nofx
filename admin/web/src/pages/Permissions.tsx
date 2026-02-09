import React, { useState, useEffect } from 'react';
import { 
  Table, 
  Card, 
  Button, 
  Modal, 
  Form, 
  Select, 
  message, 
  Space,
  Tag,
  Tooltip,
  Popconfirm
} from 'antd';
import { 
  PlusOutlined, 
  EditOutlined, 
  DeleteOutlined,
  UserOutlined,
  KeyOutlined,
  ReloadOutlined
} from '@ant-design/icons';
import { permissionAPI, userAPI } from '../services/api';

const { Option } = Select;

const Permissions: React.FC = () => {
  const [permissions, setPermissions] = useState<any[]>([]);
  const [users, setUsers] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingPermission, setEditingPermission] = useState<any>(null);
  const [form] = Form.useForm();
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });

  // 权限选项
  const permissionOptions = [
    { value: 'user:read', label: '用户读取', group: '用户管理' },
    { value: 'user:write', label: '用户写入', group: '用户管理' },
    { value: 'user:delete', label: '用户删除', group: '用户管理' },
    { value: 'user:manage', label: '用户管理', group: '用户管理' },
    { value: 'system:read', label: '系统读取', group: '系统管理' },
    { value: 'system:write', label: '系统写入', group: '系统管理' },
    { value: 'system:manage', label: '系统管理', group: '系统管理' },
    { value: 'monitor:read', label: '监控读取', group: '监控管理' },
    { value: 'monitor:write', label: '监控写入', group: '监控管理' },
    { value: 'audit:read', label: '审计读取', group: '审计管理' },
    { value: 'audit:write', label: '审计写入', group: '审计管理' },
    { value: 'audit:manage', label: '审计管理', group: '审计管理' },
    { value: '*', label: '全部权限', group: '超级权限' },
  ];

  // 获取权限列表
  const fetchPermissions = async (page = 1, pageSize = 10) => {
    setLoading(true);
    try {
      const response = await permissionAPI.getPermissions({
        page,
        limit: pageSize,
      });
      
      setPermissions(response.data.permissions);
      setPagination({
        current: page,
        pageSize,
        total: response.data.pagination.total,
      });
    } catch (error) {
      console.error('获取权限列表失败:', error);
      message.error('获取权限列表失败');
    } finally {
      setLoading(false);
    }
  };

  // 获取用户列表
  const fetchUsers = async () => {
    try {
      const response = await userAPI.getUsers();
      setUsers(response.data.users);
    } catch (error) {
      console.error('获取用户列表失败:', error);
      message.error('获取用户列表失败');
    }
  };

  // 组件挂载时获取数据
  useEffect(() => {
    fetchPermissions();
    fetchUsers();
  }, []);

  // 表格列定义
  const columns = [
    {
      title: '用户',
      dataIndex: ['user', 'username'],
      key: 'username',
      render: (username: string, record: any) => (
        <Space>
          <UserOutlined />
          {username || record.user_id}
        </Space>
      ),
    },
    {
      title: '权限',
      dataIndex: 'permission',
      key: 'permission',
      render: (permission: string) => {
        const permOption = permissionOptions.find(opt => opt.value === permission);
        return (
          <Tag color={permission === '*' ? 'red' : 'blue'}>
            {permOption?.label || permission}
          </Tag>
        );
      },
    },
    {
      title: '权限组',
      dataIndex: 'permission',
      key: 'group',
      render: (permission: string) => {
        const permOption = permissionOptions.find(opt => opt.value === permission);
        return permOption?.group || '-';
      },
    },
    {
      title: '授予者',
      dataIndex: 'granted_by',
      key: 'granted_by',
      render: (grantedBy: string) => grantedBy || '-',
    },
    {
      title: '授予时间',
      dataIndex: 'granted_at',
      key: 'granted_at',
      render: (date: string) => date ? new Date(date).toLocaleString() : '-',
    },
    {
      title: '撤销时间',
      dataIndex: 'revoked_at',
      key: 'revoked_at',
      render: (date: string) => date ? new Date(date).toLocaleString() : '-',
    },
    {
      title: '备注',
      dataIndex: 'reason',
      key: 'reason',
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: any, record: any) => (
        <Space>
          {record.revoked_at ? null : (
            <Tooltip title="撤销权限">
              <Popconfirm
                title="确认撤销此权限?"
                onConfirm={() => handleRevoke(record.id)}
                okText="确认"
                cancelText="取消"
              >
                <Button type="link" danger icon={<DeleteOutlined />} />
              </Popconfirm>
            </Tooltip>
          )}
        </Space>
      ),
    },
  ];

  // 处理撤销权限
  const handleRevoke = async (permissionId: string) => {
    try {
      await permissionAPI.revokePermission(permissionId);
      message.success('权限撤销成功');
      fetchPermissions(pagination.current, pagination.pageSize);
    } catch (error) {
      console.error('撤销权限失败:', error);
      message.error('撤销权限失败');
    }
  };

  // 提交表单
  const handleSubmit = async (values: any) => {
    try {
      await permissionAPI.grantPermission(values);
      message.success('权限授予成功');
      setModalVisible(false);
      form.resetFields();
      fetchPermissions(pagination.current, pagination.pageSize);
    } catch (error) {
      console.error('授予权限失败:', error);
      message.error('授予权限失败');
    }
  };

  // 打开授予权限模态框
  const handleGrantPermission = () => {
    form.resetFields();
    setModalVisible(true);
  };

  return (
    <div>
      <Card 
        title="权限管理"
        extra={
          <Space>
            <Button 
              icon={<ReloadOutlined />} 
              onClick={() => fetchPermissions(pagination.current, pagination.pageSize)}
            >
              刷新
            </Button>
            <Button 
              type="primary" 
              icon={<KeyOutlined />} 
              onClick={handleGrantPermission}
            >
              授予权限
            </Button>
          </Space>
        }
      >
        <Table
          columns={columns}
          dataSource={permissions}
          rowKey="id"
          loading={loading}
          pagination={{
            ...pagination,
            onChange: (page, pageSize) => {
              fetchPermissions(page, pageSize || 10);
            },
          }}
        />
      </Card>

      {/* 授予权限模态框 */}
      <Modal
        title="授予权限"
        open={modalVisible}
        onCancel={() => {
          setModalVisible(false);
          form.resetFields();
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
            name="user_id"
            label="用户"
            rules={[{ required: true, message: '请选择用户!' }]}
          >
            <Select 
              placeholder="请选择用户" 
              showSearch
              optionFilterProp="children"
            >
              {users.map(user => (
                <Option key={user.id} value={user.id}>
                  {user.username} ({user.email})
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            name="permission"
            label="权限"
            rules={[{ required: true, message: '请选择权限!' }]}
          >
            <Select placeholder="请选择权限">
              {permissionOptions.map(option => (
                <Option key={option.value} value={option.value}>
                  <Space>
                    <Tag color={option.value === '*' ? 'red' : 'blue'}>{option.label}</Tag>
                    <small>({option.group})</small>
                  </Space>
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            name="reason"
            label="备注"
          >
            <Select placeholder="选择备注或输入自定义原因">
              <Option value="系统配置需求">系统配置需求</Option>
              <Option value="临时权限">临时权限</Option>
              <Option value="业务需求">业务需求</Option>
              <Option value="其他">其他</Option>
            </Select>
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                授予权限
              </Button>
              <Button onClick={() => {
                setModalVisible(false);
                form.resetFields();
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

export default Permissions;