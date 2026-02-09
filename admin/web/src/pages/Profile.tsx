import React, { useState, useEffect } from 'react';
import {
  Card,
  Form,
  Input,
  Button,
  message,
  Divider,
  Typography,
  Space
} from 'antd';
import {
  UserOutlined,
  MailOutlined,
  LockOutlined,
  SaveOutlined,
  ReloadOutlined
} from '@ant-design/icons';
import { dashboardAPI } from '../services/api';

const { Title } = Typography;

interface UserProfile {
  id: string;
  username: string;
  email: string;
  role: string;
  created_at: string;
  updated_at: string;
}

interface ChangePasswordForm {
  old_password: string;
  new_password: string;
  confirm_new_password: string;
}

const Profile: React.FC = () => {
  const [profileForm] = Form.useForm();
  const [passwordForm] = Form.useForm();
  const [loading, setLoading] = useState<boolean>(true);
  const [saving, setSaving] = useState<boolean>(false);
  const [profile, setProfile] = useState<UserProfile | null>(null);

  // 加载用户资料
  const loadProfile = async () => {
    setLoading(true);
    try {
      const response = await dashboardAPI.getProfile();
      const userData = response.data.user;
      setProfile(userData);
      
      profileForm.setFieldsValue({
        email: userData.email,
      });
      
      message.success('用户资料加载成功');
    } catch (error) {
      console.error('加载用户资料失败:', error);
      message.error('加载用户资料失败');
    } finally {
      setLoading(false);
    }
  };

  // 更新用户资料
  const updateProfile = async (values: { email: string }) => {
    try {
      setSaving(true);
      await dashboardAPI.updateProfile(values);
      message.success('用户资料更新成功');
      loadProfile(); // 重新加载更新后的资料
    } catch (error) {
      console.error('更新用户资料失败:', error);
      message.error('更新用户资料失败');
    } finally {
      setSaving(false);
    }
  };

  // 修改密码
  const changePassword = async (values: ChangePasswordForm) => {
    if (values.new_password !== values.confirm_new_password) {
      message.error('新密码与确认密码不一致');
      return;
    }

    try {
      setSaving(true);
      await dashboardAPI.changePassword({
        old_password: values.old_password,
        new_password: values.new_password
      });
      message.success('密码修改成功');
      passwordForm.resetFields();
    } catch (error) {
      console.error('修改密码失败:', error);
      message.error('修改密码失败：原密码可能不正确');
    } finally {
      setSaving(false);
    }
  };

  // 组件挂载时加载资料
  useEffect(() => {
    loadProfile();
  }, []);

  return (
    <div className="profile-page">
      <Title level={2} style={{ marginBottom: 24 }}>
        <UserOutlined /> 个人资料
      </Title>

      <Card title="基本信息" style={{ marginBottom: 16 }}>
        <Form
          form={profileForm}
          layout="vertical"
          onFinish={updateProfile}
          initialValues={{ email: profile?.email }}
        >
          <Form.Item
            name="username"
            label="用户名"
          >
            <Input disabled value={profile?.username} />
          </Form.Item>
          
          <Form.Item
            name="email"
            label="邮箱"
            rules={[
              { required: true, type: 'email', message: '请输入正确的邮箱地址!' }
            ]}
          >
            <Input prefix={<MailOutlined />} placeholder="请输入邮箱地址" />
          </Form.Item>
          
          <Form.Item
            name="role"
            label="角色"
          >
            <Input disabled value={profile?.role} />
          </Form.Item>
          
          <Form.Item
            name="created_at"
            label="创建时间"
          >
            <Input disabled value={profile?.created_at ? new Date(profile.created_at).toLocaleString() : ''} />
          </Form.Item>
          
          <Form.Item>
            <Space>
              <Button
                type="primary"
                icon={<SaveOutlined />}
                htmlType="submit"
                loading={saving}
              >
                保存资料
              </Button>
              <Button
                icon={<ReloadOutlined />}
                onClick={loadProfile}
                disabled={loading}
              >
                刷新
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>

      <Card title="修改密码">
        <Form
          form={passwordForm}
          layout="vertical"
          onFinish={changePassword}
        >
          <Form.Item
            name="old_password"
            label="原密码"
            rules={[{ required: true, message: '请输入原密码!' }]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="请输入原密码" />
          </Form.Item>
          
          <Form.Item
            name="new_password"
            label="新密码"
            rules={[
              { required: true, message: '请输入新密码!' },
              { min: 6, message: '密码长度不能少于6位!' }
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="请输入新密码" />
          </Form.Item>
          
          <Form.Item
            name="confirm_new_password"
            label="确认新密码"
            dependencies={['new_password']}
            rules={[
              { required: true, message: '请确认新密码!' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('new_password') === value) {
                    return Promise.resolve();
                  }
                  return Promise.reject(new Error('两次输入的密码不一致!'));
                },
              }),
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="请再次输入新密码" />
          </Form.Item>
          
          <Form.Item>
            <Button
              type="primary"
              icon={<LockOutlined />}
              htmlType="submit"
              loading={saving}
            >
              修改密码
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default Profile;