import React, { useState, useEffect } from 'react'
import { Button, Card, message, Select, Space, List, Spin } from 'antd'
import { guardianAPI } from '../api/guardian'
import { useAuth } from '../contexts/AuthContext'
import { api } from '../lib/api'

interface TraderInfo {
  trader_id: string
  trader_name: string
  ai_model: string
  is_running: boolean
  // 其他可能的字段...
}

interface TraderLoginConfig {
  traderId: string
  traderName: string
  provider: string
  targetUrl: string
  status: 'pending' | 'logged-in' | 'failed'
  lastLogin?: string
}

const { Option } = Select

const AutoLoginTestPage: React.FC = () => {
  const { user, token } = useAuth()
  const [loading, setLoading] = useState(false)
  const [tradersLoading, setTradersLoading] = useState(false)
  const [selectedProvider, setSelectedProvider] = useState('deepseek')
  const [selectedTraderId, setSelectedTraderId] = useState<string>('')
  const [traders, setTraders] = useState<TraderInfo[]>([])
  const [loginConfigs, setLoginConfigs] = useState<TraderLoginConfig[]>([])

  // 获取交易员列表
  const fetchTraders = async () => {
    if (!token) return
    
    setTradersLoading(true)
    try {
      const response = await fetch('/api/my-traders', {
        headers: { Authorization: `Bearer ${token}` },
      })
      
      if (response.ok) {
        const data = await response.json()
        setTraders(data)
        
        // 如果有交易员且没有选中任何交易员，默认选中第一个
        if (data.length > 0 && !selectedTraderId) {
          setSelectedTraderId(data[0].trader_id)
        }
      }
    } catch (error) {
      console.error('获取交易员列表失败:', error)
      message.error('获取交易员列表失败')
    } finally {
      setTradersLoading(false)
    }
  }

  // 组件加载时获取交易员列表
  useEffect(() => {
    fetchTraders()
  }, [token])

  // 当选中的交易员改变时，更新登录配置
  useEffect(() => {
    if (selectedTraderId && traders.length > 0) {
      const trader = traders.find(t => t.trader_id === selectedTraderId)
      if (trader) {
        // 检查是否已存在该交易员的配置
        const existingConfig = loginConfigs.find(config => config.traderId === selectedTraderId)
        if (!existingConfig) {
          // 添加新的配置
          const newConfig: TraderLoginConfig = {
            traderId: selectedTraderId,
            traderName: trader.trader_name,
            provider: selectedProvider,
            targetUrl: guardianAPI.getTargetUrl(selectedProvider),
            status: 'pending'
          }
          setLoginConfigs([...loginConfigs, newConfig])
        }
      }
    }
  }, [selectedTraderId, traders, selectedProvider, loginConfigs])

  // 检查登录状态
  const checkLoginStatus = async (config: TraderLoginConfig) => {
    setLoading(true)
    try {
      const response = await guardianAPI.checkLoginStatus({
        provider: config.provider,
        targetUrl: config.targetUrl,
        traderId: config.traderId
      })
      
      // 更新配置状态
      const updatedConfigs = loginConfigs.map(item => 
        item.traderId === config.traderId 
          ? { ...item, status: response.isLoggedIn ? 'logged-in' as const : 'pending' as const }
          : item
      )
      
      setLoginConfigs(updatedConfigs)
      
      if (response.isLoggedIn) {
        message.success(`✅ ${config.traderName} 已登录`) 
      } else {
        message.warning(`❌ ${config.traderName} 未登录`) 
      }
    } catch (error: any) {
      console.error('检查登录状态失败:', error)
      message.error(`${config.traderName} 检查登录状态失败: ` + error.message)
      
      // 标记为失败状态
      const updatedConfigs = loginConfigs.map(item => 
        item.traderId === config.traderId 
          ? { ...item, status: 'failed' as const }
          : item
      )
      setLoginConfigs(updatedConfigs)
    } finally {
      setLoading(false)
    }
  }

  // 自动登录
  const autoLogin = async (config: TraderLoginConfig) => {
    setLoading(true)
    try {
      const response = await guardianAPI.autoLogin({
        provider: config.provider,
        targetUrl: config.targetUrl,
        aiService: config.provider,
        traderId: config.traderId
      })
      
      if (response.success) {
        message.success(`🚀 ${config.traderName} 自动登录已启动，请在浏览器中完成登录`)
        
        // 更新状态为进行中
        const updatedConfigs = loginConfigs.map(item => 
          item.traderId === config.traderId 
            ? { ...item, status: 'pending' as const }
            : item
        )
        setLoginConfigs(updatedConfigs)
        
        // 3秒后重新检查登录状态
        setTimeout(() => {
          checkLoginStatus(config)
        }, 3000)
      } else {
        message.warning(`${config.traderName} ` + response.description)
        
        // 标记为失败状态
        const updatedConfigs = loginConfigs.map(item => 
          item.traderId === config.traderId 
            ? { ...item, status: 'failed' as const }
            : item
        )
        setLoginConfigs(updatedConfigs)
      }
    } catch (error: any) {
      console.error('自动登录失败:', error)
      message.error(`${config.traderName} 自动登录失败: ` + error.message)
      
      // 标记为失败状态
      const updatedConfigs = loginConfigs.map(item => 
        item.traderId === config.traderId 
          ? { ...item, status: 'failed' as const }
          : item
      )
      setLoginConfigs(updatedConfigs)
    } finally {
      setLoading(false)
    }
  }

  // 重置配置状态
  const resetConfigStatus = (traderId: string) => {
    const updatedConfigs = loginConfigs.map(item => 
      item.traderId === traderId 
        ? { ...item, status: 'pending' as const }
        : item
    )
    setLoginConfigs(updatedConfigs)
    message.info('状态已重置')
  }

  const providers = [
    { value: 'deepseek', label: 'DeepSeek' },
    { value: 'openai', label: 'ChatGPT' },
    { value: 'claude', label: 'Claude' },
    { value: 'qwen', label: 'Qwen (通义千问)' },
    { value: 'gemini', label: 'Google Gemini' },
  ]

  // 获取状态颜色
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'logged-in': return '#52c41a'
      case 'pending': return '#faad14'
      case 'failed': return '#ff4d4f'
      default: return '#8c8c8c'
    }
  }

  // 获取状态文本
  const getStatusText = (status: string) => {
    switch (status) {
      case 'logged-in': return '✅ 已登录'
      case 'pending': return '⏳ 待登录'
      case 'failed': return '❌ 登录失败'
      default: return '❓ 未知状态'
    }
  }

  return (
    <div style={{ padding: '24px', maxWidth: '1000px', margin: '0 auto' }}>
      <Card title="🤖 交易员首次登录设置工具" style={{ marginBottom: '24px' }}>
        <Space direction="vertical" style={{ width: '100%' }} size="large">
          
          {/* 选择交易员和AI服务 */}
          <Card size="small" title="配置交易员登录">
            <Space>
              <div>
                <div style={{ marginBottom: '8px' }}><strong>选择交易员:</strong></div>
                <Select
                  value={selectedTraderId}
                  onChange={setSelectedTraderId}
                  style={{ width: '250px' }}
                  loading={tradersLoading}
                  placeholder="请选择交易员"
                  showSearch
                  optionFilterProp="children"
                >
                  {traders.map(trader => (
                    <Option key={trader.trader_id} value={trader.trader_id}>
                      {trader.trader_name} ({trader.trader_id.substring(0, 8)}...)
                    </Option>
                  ))}
                </Select>
              </div>
              
              <div>
                <div style={{ marginBottom: '8px' }}><strong>AI服务提供商:</strong></div>
                <Select
                  value={selectedProvider}
                  onChange={setSelectedProvider}
                  style={{ width: '150px' }}
                >
                  {providers.map(provider => (
                    <Option key={provider.value} value={provider.value}>
                      {provider.label}
                    </Option>
                  ))}
                </Select>
              </div>
              
              <div style={{ paddingTop: '22px' }}>
                <Button 
                  type="primary"
                  onClick={fetchTraders}
                  loading={tradersLoading}
                >
                  刷新列表
                </Button>
              </div>
            </Space>
          </Card>

          {/* 登录操作区域 */}
          <Card size="small" title="交易员登录操作">
            {selectedTraderId ? (
              <div>
                <div style={{ marginBottom: '20px', padding: '16px', backgroundColor: '#f5f5f5', borderRadius: '6px' }}>
                  <h3 style={{ margin: '0 0 12px 0' }}>当前选择的交易员</h3>
                  <div><strong>名称:</strong> {traders.find(t => t.trader_id === selectedTraderId)?.trader_name}</div>
                  <div><strong>ID:</strong> {selectedTraderId}</div>
                  <div><strong>AI服务:</strong> {providers.find(p => p.value === selectedProvider)?.label}</div>
                </div>
                
                <Space>
                  <Button 
                    type="primary" 
                    size="large"
                    onClick={() => {
                      const config = loginConfigs.find(c => c.traderId === selectedTraderId)
                      if (config) checkLoginStatus(config)
                    }}
                    loading={loading}
                  >
                    🔍 检查登录状态
                  </Button>
                  
                  <Button 
                    type="primary" 
                    size="large"
                    danger
                    onClick={() => {
                      const config = loginConfigs.find(c => c.traderId === selectedTraderId)
                      if (config) autoLogin(config)
                    }}
                    loading={loading}
                  >
                    🚀 首次登录
                  </Button>
                  
                  <Button 
                    type="default" 
                    size="large"
                    onClick={() => {
                      resetConfigStatus(selectedTraderId)
                    }}
                  >
                    🔄 重置状态
                  </Button>
                </Space>
                
                {/* 状态显示 */}
                {loginConfigs.some(c => c.traderId === selectedTraderId) && (
                  <div style={{ marginTop: '20px', padding: '12px', backgroundColor: '#fffbe6', borderRadius: '6px' }}>
                    <strong>当前状态:</strong>
                    <span style={{ 
                      marginLeft: '12px', 
                      fontSize: '16px',
                      color: getStatusColor(loginConfigs.find(c => c.traderId === selectedTraderId)?.status || 'pending')
                    }}>
                      {getStatusText(loginConfigs.find(c => c.traderId === selectedTraderId)?.status || 'pending')}
                    </span>
                  </div>
                )}
              </div>
            ) : (
              <div style={{ textAlign: 'center', padding: '40px', color: '#8c8c8c' }}>
                请先从上方下拉列表中选择一个交易员
              </div>
            )}
          </Card>

          {/* 使用说明 */}
          <Card size="small" title="📘 使用说明">
            <ul>
              <li><strong>选择交易员</strong>: 从下拉列表中选择需要登录的现有交易员</li>
              <li><strong>选择AI服务</strong>: 为该交易员选择对应的AI服务提供商</li>
              <li><strong>检查状态</strong>: 查看该交易员当前的登录状态</li>
              <li><strong>首次登录</strong>: 为未登录的交易员启动自动登录流程</li>
              <li><strong>状态指示</strong>: 
                <span style={{ color: '#52c41a' }}>✅ 已登录</span>、
                <span style={{ color: '#faad14' }}>⏳ 待登录</span>、
                <span style={{ color: '#ff4d4f' }}>❌ 登录失败</span>
              </li>
              <li><strong>独立会话</strong>: 每个交易员有独立的浏览器会话</li>
              <li><strong>自动保存</strong>: 登录状态会自动持久化保存</li>
            </ul>
          </Card>

        </Space>
      </Card>
    </div>
  )
}

export default AutoLoginTestPage