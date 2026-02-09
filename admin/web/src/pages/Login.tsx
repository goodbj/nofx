import React, { useState } from 'react';
import { LockOutlined, UserOutlined } from '@ant-design/icons';
import { Button, Form, Input, Card, Typography, message } from 'antd';
import axios from 'axios';

const { Title } = Typography;

interface LoginProps {
  onLogin: () => void;
}

const Login: React.FC<LoginProps> = ({ onLogin }) => {
  const [loading, setLoading] = useState(false);

  const onFinish = async (values: { username: string; password: string }) => {
    setLoading(true);
    try {
      const response = await axios.post('/api/auth/login', {
        username: values.username,
        password: values.password,
      });

      if (response.status === 200) {
        // 保存令牌到本地存储
        localStorage.setItem('admin_token', response.data.token);
        
        // 显示成功消息
        message.success('登录成功！');
        
        // 调用父组件的登录回调
        onLogin();
      }
    } catch (error: any) {
      console.error('登录失败:', error);
      if (error.response) {
        message.error(error.response.data.error || '登录失败');
      } else {
        message.error('网络错误，请稍后重试');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-container">
      <Card className="login-form">
        <div style={{ textAlign: 'center', marginBottom: 24 }}>
          <Title level={2} style={{ color: '#1890ff' }}>NOFX 管理员</Title>
          <p>请使用管理员账号登录</p>
        </div>
        
        <Form
          name="login"
          initialValues={{ remember: true }}
          onFinish={onFinish}
        >
          <Form.Item
            name="username"
            rules={[{ required: true, message: '请输入用户名!' }]}
          >
            <Input prefix={<UserOutlined />} placeholder="用户名或邮箱" />
          </Form.Item>
          
          <Form.Item
            name="password"
            rules={[{ required: true, message: '请输入密码!' }]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="密码" />
          </Form.Item>
          
          <Form.Item>
            <Button 
              type="primary" 
              htmlType="submit" 
              block 
              loading={loading}
            >
              登录
            </Button>
          </Form.Item>
        </Form>
        
        <div style={{ marginTop: 16, fontSize: '12px', color: '#999' }}>
          <p>默认管理员账户:</p>
          <p>用户名: superadmin</p>
          <p>密码: SuperAdmin123!</p>
          <p style={{ marginTop: 8 }}><strong>注意: 请在生产环境中修改默认密码</strong></p>
        </div>
      </Card>
    </div>
  );
};

export default Login;