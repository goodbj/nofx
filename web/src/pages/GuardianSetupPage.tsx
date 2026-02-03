import React, { useState } from 'react';
import { Card, Button, Space, Typography, message, Select } from 'antd';

const { Title, Text, Paragraph } = Typography;
const { Option } = Select;

const GuardianSetupPage: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [selectedProvider, setSelectedProvider] = useState('deepseek-browser');

  // 打开长时间保持的浏览器窗口，用于登录设置
  const openLongLivedBrowser = async () => {
    setLoading(true);

    try {
      // Import the guardian API and get the target URL
      const guardianModule = await import('../api/guardian');
      const targetUrl = guardianModule.guardianAPI.getTargetUrl(selectedProvider);

      const response = await fetch('/api/open-guardian-browser', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          url: targetUrl
        })
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      message.success('浏览器窗口已打开，请完成登录设置操作');
    } catch (error: any) {
      console.error('打开浏览器失败:', error);
      message.error('打开浏览器失败: ' + error.message);
    } finally {
      setLoading(false);
    }
  };

  // 检查登录状态
  const checkLoginStatus = async () => {
    setLoading(true);

    try {
      const result = await import('../api/guardian').then(mod => 
        mod.guardianAPI.checkLoginStatus({ provider: selectedProvider })
      );
      
      if (result.isLoggedIn) {
        message.success('✅ 您已登录');
      } else {
        message.warning('❌ 您未登录，请先完成登录设置');
      }
    } catch (error: any) {
      console.error('检查登录状态失败:', error);
      message.error('检查登录状态失败: ' + error.message);
    } finally {
      setLoading(false);
    }
  };

  const providers = [
    { value: 'deepseek-browser', label: 'DeepSeek' },
    { value: 'chatgpt-browser', label: 'ChatGPT' },
    { value: 'claude-browser', label: 'Claude' },
    { value: 'qwen-browser', label: '通义千问' },
    { value: 'gemini-browser', label: 'Gemini' },
    { value: 'grok-browser', label: 'Grok' },
    { value: 'kimi-browser', label: 'Kimi' },
    { value: 'ollama-browser', label: 'Ollama' },
    { value: 'guardian-ai', label: 'Guardian AI' },
  ];

  return (
    <div style={{ padding: '24px' }}>
      <Title level={2}>Guardian浏览器自动化设置</Title>
      <Paragraph type="secondary">
        此页面用于设置Guardian浏览器自动化功能所需的登录凭据
      </Paragraph>

      <Card style={{ marginBottom: '24px' }}>
        <Space direction="vertical" style={{ width: '100%' }}>
          <Text strong>选择AI服务提供商：</Text>
          <Select
            value={selectedProvider}
            onChange={setSelectedProvider}
            style={{ width: '100%' }}
          >
            {providers.map(provider => (
              <Option key={provider.value} value={provider.value}>
                {provider.label}
              </Option>
            ))}
          </Select>
        </Space>
      </Card>

      <Card style={{ marginBottom: '24px' }}>
        <Space wrap>
          <Button
            type="primary"
            loading={loading}
            onClick={openLongLivedBrowser}
          >
            打开浏览器进行设置
          </Button>
          
          <Button
            type="default"
            loading={loading}
            onClick={checkLoginStatus}
          >
            检查登录状态
          </Button>
        </Space>
      </Card>

      <Card title="使用说明">
        <Paragraph>
          <Text strong>1. 选择AI服务提供商：</Text>
          从下拉菜单中选择您要设置的AI服务（如DeepSeek、ChatGPT等）。
        </Paragraph>
        <Paragraph>
          <Text strong>2. 打开浏览器进行设置：</Text>
          点击"打开浏览器进行设置"按钮，将打开一个长时间保持的浏览器窗口，
          请在此窗口中完成登录和其他必要设置。
        </Paragraph>
        <Paragraph>
          <Text strong>3. 检查登录状态：</Text>
          点击"检查登录状态"按钮，确认是否已成功登录。
        </Paragraph>
        <Paragraph>
          <Text strong>注意：</Text>
          登录信息将被保存在浏览器配置文件中，下次使用时无需重新登录。
        </Paragraph>
      </Card>
    </div>
  );
};

export default GuardianSetupPage;