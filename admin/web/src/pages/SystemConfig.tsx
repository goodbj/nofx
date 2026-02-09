import React, { useState, useEffect } from 'react';
import {
  Card,
  Form,
  Input,
  Button,
  message,
  Switch,
  Select,
  InputNumber,
  Divider,
  Space,
  Typography
} from 'antd';
import {
  SettingOutlined,
  SaveOutlined,
  ReloadOutlined,
  CloudServerOutlined,
  SafetyCertificateOutlined,
  ApiOutlined
} from '@ant-design/icons';
import { systemAPI, SystemConfig } from '../services/api';

const { Title } = Typography;
const { Option } = Select;

const SystemConfigPage: React.FC = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState<boolean>(true);
  const [saving, setSaving] = useState<boolean>(false);

  // 加载系统配置
  const loadConfig = async () => {
    setLoading(true);
    try {
      const response = await systemAPI.getConfig();
      const configData = response.data.config || {};
      
      // 将配置数组转换为表单对象
      const formValues: Record<string, any> = {};
      configData.forEach((item: SystemConfig) => {
        formValues[item.key] = item.value;
      });
      
      form.setFieldsValue(formValues);
      message.success('配置加载成功');
    } catch (error) {
      console.error('加载系统配置失败:', error);
      message.error('加载系统配置失败');
    } finally {
      setLoading(false);
    }
  };

  // 保存系统配置
  const saveConfig = async () => {
    try {
      const values = form.getFieldsValue();
      const configArray = Object.keys(values).map(key => ({
        key,
        value: values[key].toString()
      }));

      setSaving(true);
      await systemAPI.updateConfig({ config: configArray });
      message.success('系统配置保存成功');
    } catch (error) {
      console.error('保存系统配置失败:', error);
      message.error('保存系统配置失败');
    } finally {
      setSaving(false);
    }
  };

  // 组件挂载时加载配置
  useEffect(() => {
    loadConfig();
  }, []);

  return (
    <div className="system-config-page">
      <Title level={2} style={{ marginBottom: 24 }}>
        <SettingOutlined /> 系统配置
      </Title>

      <Card
        title="系统参数配置"
        extra={
          <Space>
            <Button
              icon={<ReloadOutlined />}
              onClick={loadConfig}
              disabled={loading}
            >
              刷新
            </Button>
            <Button
              type="primary"
              icon={<SaveOutlined />}
              onClick={saveConfig}
              loading={saving}
            >
              保存配置
            </Button>
          </Space>
        }
      >
        <Form
          form={form}
          layout="vertical"
          initialValues={{
            'app.name': 'NOFX Admin System',
            'app.version': '1.0.0',
            'app.description': 'NOFX Trading Platform Admin System',
            'app.debug': 'false',
            'security.rate_limit.enabled': 'true',
            'security.rate_limit.requests_per_minute': '100',
            'security.login_attempts.max': '5',
            'security.session.timeout_minutes': '60',
            'database.connection_timeout': '30',
            'logging.level': 'info',
            'logging.max_size_mb': '100',
            'backup.enabled': 'true',
            'backup.schedule': '0 2 * * *', // 每天凌晨2点备份
            'backup.retention_days': '30',
            'cache.enabled': 'true',
            'cache.ttl_seconds': '3600',
          }}
        >
          <Divider orientation="left">应用配置</Divider>
          
          <Form.Item
            name="app.name"
            label="应用名称"
            tooltip="应用程序的显示名称"
          >
            <Input placeholder="请输入应用名称" />
          </Form.Item>
          
          <Form.Item
            name="app.version"
            label="应用版本"
            tooltip="应用程序的版本号"
          >
            <Input placeholder="请输入应用版本" />
          </Form.Item>
          
          <Form.Item
            name="app.description"
            label="应用描述"
            tooltip="应用程序的简短描述"
          >
            <Input.TextArea rows={3} placeholder="请输入应用描述" />
          </Form.Item>
          
          <Form.Item
            name="app.debug"
            label="调试模式"
            valuePropName="checked"
            tooltip="是否开启调试模式"
          >
            <Switch checkedChildren="开启" unCheckedChildren="关闭" />
          </Form.Item>
          
          <Divider orientation="left">安全配置</Divider>
          
          <Form.Item
            name="security.rate_limit.enabled"
            label="启用速率限制"
            valuePropName="checked"
            tooltip="是否启用API请求速率限制"
          >
            <Switch checkedChildren="开启" unCheckedChildren="关闭" />
          </Form.Item>
          
          <Form.Item
            name="security.rate_limit.requests_per_minute"
            label="每分钟最大请求数"
            tooltip="每个IP地址每分钟允许的最大请求数"
          >
            <InputNumber min={1} max={10000} style={{ width: '100%' }} />
          </Form.Item>
          
          <Form.Item
            name="security.login_attempts.max"
            label="最大登录尝试次数"
            tooltip="连续登录失败后账户锁定前的最大尝试次数"
          >
            <InputNumber min={1} max={100} style={{ width: '100%' }} />
          </Form.Item>
          
          <Form.Item
            name="security.session.timeout_minutes"
            label="会话超时时间(分钟)"
            tooltip="用户会话自动过期的时间(分钟)"
          >
            <InputNumber min={1} max={1440} style={{ width: '100%' }} />
          </Form.Item>
          
          <Divider orientation="left">数据库配置</Divider>
          
          <Form.Item
            name="database.connection_timeout"
            label="连接超时时间(秒)"
            tooltip="数据库连接超时时间(秒)"
          >
            <InputNumber min={1} max={300} style={{ width: '100%' }} />
          </Form.Item>
          
          <Divider orientation="left">日志配置</Divider>
          
          <Form.Item
            name="logging.level"
            label="日志级别"
            tooltip="系统日志输出级别"
          >
            <Select placeholder="选择日志级别">
              <Option value="debug">调试(debug)</Option>
              <Option value="info">信息(info)</Option>
              <Option value="warn">警告(warn)</Option>
              <Option value="error">错误(error)</Option>
            </Select>
          </Form.Item>
          
          <Form.Item
            name="logging.max_size_mb"
            label="日志文件最大大小(MB)"
            tooltip="单个日志文件的最大大小(MB)"
          >
            <InputNumber min={1} max={1000} style={{ width: '100%' }} />
          </Form.Item>
          
          <Divider orientation="left">备份配置</Divider>
          
          <Form.Item
            name="backup.enabled"
            label="启用自动备份"
            valuePropName="checked"
            tooltip="是否启用自动备份功能"
          >
            <Switch checkedChildren="开启" unCheckedChildren="关闭" />
          </Form.Item>
          
          <Form.Item
            name="backup.schedule"
            label="备份计划"
            tooltip="使用cron表达式设置备份计划"
          >
            <Input placeholder="例如: 0 2 * * * (每天凌晨2点)" />
          </Form.Item>
          
          <Form.Item
            name="backup.retention_days"
            label="备份保留天数"
            tooltip="自动删除超过指定天数的旧备份"
          >
            <InputNumber min={1} max={365} style={{ width: '100%' }} />
          </Form.Item>
          
          <Divider orientation="left">缓存配置</Divider>
          
          <Form.Item
            name="cache.enabled"
            label="启用缓存"
            valuePropName="checked"
            tooltip="是否启用系统缓存"
          >
            <Switch checkedChildren="开启" unCheckedChildren="关闭" />
          </Form.Item>
          
          <Form.Item
            name="cache.ttl_seconds"
            label="缓存过期时间(秒)"
            tooltip="缓存项的生存时间(秒)"
          >
            <InputNumber min={60} max={86400} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default SystemConfigPage;