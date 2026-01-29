import React, { useState } from 'react';
import Card from 'antd/lib/card';
import Button from 'antd/lib/button';
import Space from 'antd/lib/space';
import Typography from 'antd/lib/typography';
import message from 'antd/lib/message';
import TextArea from 'antd/lib/input/TextArea';

const { Title, Text, Paragraph } = Typography;

const GuardianTestPage: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [testResult, setTestResult] = useState('');
  const [aiContent, setAiContent] = useState('');
  const [predefinedPrompt, setPredefinedPrompt] = useState('');

  // 模拟Guardian浏览器自动化测试
  const testGuardianBrowser = async () => {
    setLoading(true);
    setTestResult('');
    
    try {
      // 发送请求到后端触发Guardian功能
      const response = await fetch('/api/test-guardian', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          prompt: predefinedPrompt || '默认测试提示词',
          mode: 'guardian-test' // 特殊模式，跳过交易所验证
        })
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      setTestResult(data.message || 'Guardian浏览器自动化测试已启动');
      // 如果返回了AI内容，也更新AI内容状态
      if (data.result) {
        setAiContent(data.result);
      } else if (data.analysis) {
        setAiContent(data.analysis);
      }
      
      message.success('Guardian测试已启动，请查看浏览器弹窗');
    } catch (error: any) {
      console.error('Guardian测试失败:', error);
      message.error('Guardian测试失败: ' + error.message);
      setTestResult('测试失败: ' + error.message);
    } finally {
      setLoading(false);
    }
  };

  // 模拟获取AI分析
  const testGuardianAIAnalysis = async () => {
    setLoading(true);
    setTestResult('');
    
    try {
      const response = await fetch('/api/test-guardian-analysis', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          prompt: predefinedPrompt || '默认测试提示词',
          testMode: true
        })
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      setTestResult(JSON.stringify(data, null, 2));
      
      // 如果返回了AI分析内容，也更新AI内容状态
      if (data.analysis) {
        setAiContent(data.analysis);
      } else if (data.result) {
        setAiContent(data.result);
      }
      
      message.success('AI分析测试完成');
    } catch (error: any) {
      console.error('AI分析测试失败:', error);
      message.error('AI分析测试失败: ' + error.message);
      setTestResult('AI分析测试失败: ' + error.message);
    } finally {
      setLoading(false);
    }
  };

  // 打开长时间保持的浏览器窗口
  const openLongLivedBrowser = async () => {
    setLoading(true);
    setTestResult('');
    
    try {
      const response = await fetch('/api/open-guardian-browser', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          url: 'https://chat.deepseek.com/'
        })
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      setTestResult(data.message || '长时间浏览器窗口已打开');
      
      message.success('浏览器窗口已打开，请进行设置操作');
    } catch (error: any) {
      console.error('打开浏览器失败:', error);
      message.error('打开浏览器失败: ' + error.message);
      setTestResult('打开浏览器失败: ' + error.message);
    } finally {
      setLoading(false);
    }
  };

  const predefinedPrompts = [
    "分析比特币未来24小时走势",
    "以太坊价格预测和交易建议",
    "当前市场趋势分析",
    "风险评估和投资建议"
  ];

  return (
    <div style={{ padding: '24px' }}>
      <Title level={2}>Guardian浏览器自动化测试</Title>
      <Paragraph type="secondary">
        此页面用于测试Guardian浏览器自动化功能，无需连接交易所
      </Paragraph>

      <Card style={{ marginBottom: '24px' }}>
        <Space direction="vertical" style={{ width: '100%' }}>
          <Text strong>预设提示词：</Text>
          <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
            {predefinedPrompts.map((prompt, index) => (
              <Button
                key={index}
                size="small"
                onClick={() => setPredefinedPrompt(prompt)}
                type={predefinedPrompt === prompt ? 'primary' : 'default'}
              >
                {prompt}
              </Button>
            ))}
          </div>
          
          <Text strong>自定义提示词：</Text>
          <TextArea
            rows={12}
            placeholder="输入自定义提示词..."
            value={predefinedPrompt}
            onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setPredefinedPrompt(e.target.value)}
            style={{ minHeight: 180, maxHeight: 500 }}
            maxLength={200000}
            showCount
          />
        </Space>
      </Card>

      <Card style={{ marginBottom: '24px' }}>
        <Space wrap>
          <Button
            type="primary"
            loading={loading}
            onClick={testGuardianBrowser}
          >
            测试Guardian浏览器自动化
          </Button>
          
          <Button
            type="default"
            loading={loading}
            onClick={testGuardianAIAnalysis}
          >
            测试AI分析功能
          </Button>
          
          <Button
            type="default"
            loading={loading}
            onClick={openLongLivedBrowser}
          >
            设置浏览器
          </Button>
        </Space>
      </Card>

      {testResult && (
        <Card title="测试结果">
          <pre style={{ 
            backgroundColor: '#f5f5f5', 
            padding: '16px', 
            borderRadius: '4px',
            maxHeight: '400px',
            overflow: 'auto'
          }}>
            {testResult}
          </pre>
        </Card>
      )}

      {aiContent && (
        <Card title="AI输出内容" style={{ marginTop: '24px' }}>
          <div style={{ 
            backgroundColor: '#fafafa', 
            padding: '16px', 
            borderRadius: '4px',
            maxHeight: '500px',
            overflow: 'auto',
            lineHeight: '1.6'
          }}>
            {aiContent.split('\n').map((line, index) => (
              <div key={index}>{line}</div>
            ))}
          </div>
        </Card>
      )}

      <Card title="AI内容展示区" style={{ marginTop: '24px' }}>
        <TextArea
          rows={20}
          value={aiContent}
          readOnly
          placeholder="AI输出内容将显示在这里..."
          style={{ 
            fontFamily: 'monospace',
            fontSize: '14px',
            backgroundColor: '#f9f9f9',
            minHeight: 300,
            maxHeight: 600 
          }}
          showCount
        />
      </Card>

      <Card title="使用说明" style={{ marginTop: '24px' }}>
        <Paragraph>
          <Text strong>1. 浏览器自动化测试：</Text>
          点击"测试Guardian浏览器自动化"按钮将启动Chrome浏览器，
          自动打开AI网站并输入您的提示词。
        </Paragraph>
        <Paragraph>
          <Text strong>2. AI分析测试：</Text>
          测试Guardian获取AI分析结果的功能。
        </Paragraph>
        <Paragraph>
          <Text strong>3. 设置浏览器：</Text>
          点击"设置浏览器"按钮将打开一个长时间保持的浏览器窗口，
          方便您进行登录等设置操作。
        </Paragraph>
        <Paragraph>
          <Text strong>注意：</Text>
          此测试不连接任何交易所，仅测试浏览器自动化功能。
        </Paragraph>
      </Card>
    </div>
  );
};

export default GuardianTestPage;