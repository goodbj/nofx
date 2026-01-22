import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

const TestPage: React.FC = () => {
  const { token } = useAuth();
  const [apiCredentials, setApiCredentials] = useState({
    apiKey: '53zcowwoNVUiRWyg46RQb2nRwSQhUNwYeBEHetyJsjB0MLms1rqxABxWjXClkYoA1',
    secretKey: 'ZkvgnmrJiyOjuix0QxSa1eeQBwvlBENcwY5xBVp9uS085oDUpQHQcxHNfk0dppuj1',
    apiUrl: 'https://testnet.binancefuture.com',
  });
  const [testResults, setTestResults] = useState<string>('');
  const [batchResults, setBatchResults] = useState<string>('');
  const [loading, setLoading] = useState(false);
  const [batchLoading, setBatchLoading] = useState(false);
  const [selectedTrader, setSelectedTrader] = useState<string>('');
  const [batchCommands, setBatchCommands] = useState<string>(`[
  { "symbol": "ETHUSDT", "action": "update_stop_loss", "new_stop_loss": 2992.00, "confidence": 75 },
  { "symbol": "BTCUSDT", "action": "partial_close", "close_percentage": 50, "confidence": 80 },
  { "symbol": "BTCUSDT", "action": "update_stop_loss", "new_stop_loss": 89000.00, "confidence": 70 }
]`);

  const [traders, setTraders] = useState<any[]>([]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setApiCredentials(prev => ({
      ...prev,
      [name]: value
    }));
  };

  // 获取交易者列表
  const fetchTraders = async () => {
    if (!token) return;
    
    try {
      const response = await fetch('/api/my-traders', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });
      
      if (response.ok) {
        const data = await response.json();
        setTraders(data);
        if (data.length > 0 && !selectedTrader) {
          setSelectedTrader(data[0].trader_id);
        }
      }
    } catch (error) {
      console.error('获取交易者列表失败:', error);
    }
  };

  // 页面加载时获取交易者列表
  React.useEffect(() => {
    fetchTraders();
  }, [token]);

  const runTest = async () => {
    if (!token) {
      setTestResults('Error: Not authenticated. Please log in first.');
      return;
    }
    
    setLoading(true);
    setTestResults('');
    
    try {
      const response = await fetch('/api/test/run', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({ 
          command: 'go run test/test_binance_testnet.go',
          apiKey: apiCredentials.apiKey,
          secretKey: apiCredentials.secretKey,
          apiUrl: apiCredentials.apiUrl,
        }),
      });

      const data = await response.json();
      setTestResults(`Status: ${data.status}\nOutput: ${data.output}`);
    } catch (error) {
      setTestResults(`Error: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  const runBatchCommand = async () => {
    if (!token) {
      setBatchResults('Error: Not authenticated. Please log in first.');
      return;
    }
    
    if (!selectedTrader) {
      setBatchResults('Error: Please select a trader first.');
      return;
    }
    
    try {
      setBatchLoading(true);
      setBatchResults('');
      
      // 尝试解析JSON
      let parsedCommands;
      try {
        parsedCommands = JSON.parse(batchCommands);
      } catch (parseError: any) {
        setBatchResults(`JSON解析错误: ${parseError.message}`);
        return;
      }
      
      const response = await fetch(`/api/traders/${selectedTrader}/execute-multiple-decisions`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify(parsedCommands),
      });

      const data = await response.json();
      setBatchResults(JSON.stringify(data, null, 2));
    } catch (error) {
      setBatchResults(`Error: ${error}`);
    } finally {
      setBatchLoading(false);
    }
  };

  const [activeTab, setActiveTab] = useState('api-test'); // 'api-test', 'batch-trade', 'templates'

  // 定义交易动作模板
  const tradeTemplates = [
    {
      id: 'open_long',
      name: '开多 (BTC)',
      description: '买入BTC做多',
      template: `[
  {
    "symbol": "BTCUSDT",
    "action": "open_long",
    "quantity": 0.001,
    "price": 60000.00,
    "confidence": 85
  }
]`
    },
    {
      id: 'open_short',
      name: '开空 (BTC)',
      description: '卖空BTC',
      template: `[
  {
    "symbol": "BTCUSDT",
    "action": "open_short",
    "quantity": 0.001,
    "price": 60000.00,
    "confidence": 85
  }
]`
    },
    {
      id: 'close_long',
      name: '平多 (BTC)',
      description: '平掉BTC多头仓位',
      template: `[
  {
    "symbol": "BTCUSDT",
    "action": "close_long",
    "quantity": 0.001,
    "confidence": 85
  }
]`
    },
    {
      id: 'close_short',
      name: '平空 (BTC)',
      description: '平掉BTC空头仓位',
      template: `[
  {
    "symbol": "BTCUSDT",
    "action": "close_short",
    "quantity": 0.001,
    "confidence": 85
  }
]`
    },
    {
      id: 'hold',
      name: '持有',
      description: '保持当前仓位',
      template: `[
  {
    "symbol": "BTCUSDT",
    "action": "hold",
    "reason": "市场趋势不明朗，暂时持有",
    "confidence": 60
  }
]`
    },
    {
      id: 'wait',
      name: '观望',
      description: '等待更好入场时机',
      template: `[
  {
    "symbol": "BTCUSDT",
    "action": "wait",
    "reason": "等待价格回调至支撑位",
    "confidence": 50
  }
]`
    },
    {
      id: 'update_stop_loss',
      name: '更新止损',
      description: '修改BTC止损价',
      template: `[
  {
    "symbol": "BTCUSDT",
    "action": "update_stop_loss",
    "new_stop_loss": 59000.00,
    "confidence": 75
  }
]`
    },
    {
      id: 'update_take_profit',
      name: '更新止盈',
      description: '修改BTC止盈价',
      template: `[
  {
    "symbol": "BTCUSDT",
    "action": "update_take_profit",
    "new_take_profit": 65000.00,
    "confidence": 75
  }
]`
    },
    {
      id: 'stop_take_combined',
      name: '止盈止损组合',
      description: '同时设置止盈和止损',
      template: `[
  {
    "symbol": "BTCUSDT",
    "action": "update_stop_loss",
    "new_stop_loss": 59000.00,
    "confidence": 75
  },
  {
    "symbol": "BTCUSDT",
    "action": "update_take_profit",
    "new_take_profit": 65000.00,
    "confidence": 80
  }
]`
    },
    {
      id: 'partial_close',
      name: '部分平仓',
      description: '平掉部分仓位',
      template: `[
  {
    "symbol": "BTCUSDT",
    "action": "partial_close",
    "close_percentage": 50,
    "confidence": 80
  }
]`
    },
    {
      id: 'oco_order',
      name: 'OCO订单',
      description: '一键设置止盈止损（OTO）',
      template: `[
  {
    "symbol": "BTCUSDT",
    "action": "oco_order",
    "quantity": 0.001,
    "limit_price": 65000.00,
    "stop_price": 59000.00,
    "stop_limit_price": 58900.00,
    "confidence": 80
  }
]`
    },
    {
      id: 'bracket_order',
      name: '括号订单',
      description: '同时下单开仓和配套的止盈止损',
      template: `[
  {
    "symbol": "BTCUSDT",
    "action": "bracket_order",
    "order_action": "open_long",
    "quantity": 0.001,
    "entry_price": 60000.00,
    "take_profit_price": 65000.00,
    "stop_loss_price": 58000.00,
    "confidence": 85
  }
]`
    }
  ];

  // 插入模板到指令框
  const insertTemplate = (template: string) => {
    setBatchCommands(template);
  };

  return (
    <div 
      className="min-h-screen" 
      style={{ background: '#0B0E11', color: '#EAECEF' }}
    >
      <div className="max-w-6xl mx-auto px-4 py-8">
        <div 
          className="rounded-lg p-6" 
          style={{ 
            background: '#181A20', 
            border: '1px solid #2B3139'
          }}
        >
          <h1 className="text-2xl font-bold mb-6" style={{ color: '#EAECEF' }}>
            API测试页面
          </h1>
          
          {/* Tab 标签页 */}
          <div className="mb-6 border-b border-gray-700">
            <nav className="flex space-x-2">
              <button
                className={`px-4 py-2 font-medium text-sm rounded-t-lg transition-colors ${activeTab === 'api-test' ? 'text-blue-400 border-b-2 border-blue-400 bg-gray-800' : 'text-gray-400 hover:text-gray-200'}`}
                onClick={() => setActiveTab('api-test')}
                style={activeTab === 'api-test' ? {
                  color: '#93C5FD',
                  borderBottom: '2px solid #93C5FD',
                  background: '#2D3748'
                } : {}}
              >
                API测试
              </button>
              <button
                className={`px-4 py-2 font-medium text-sm rounded-t-lg transition-colors ${activeTab === 'batch-trade' ? 'text-blue-400 border-b-2 border-blue-400 bg-gray-800' : 'text-gray-400 hover:text-gray-200'}`}
                onClick={() => setActiveTab('batch-trade')}
                style={activeTab === 'batch-trade' ? {
                  color: '#93C5FD',
                  borderBottom: '2px solid #93C5FD',
                  background: '#2D3748'
                } : {}}
              >
                批量交易
              </button>
              <button
                className={`px-4 py-2 font-medium text-sm rounded-t-lg transition-colors ${activeTab === 'templates' ? 'text-blue-400 border-b-2 border-blue-400 bg-gray-800' : 'text-gray-400 hover:text-gray-200'}`}
                onClick={() => setActiveTab('templates')}
                style={activeTab === 'templates' ? {
                  color: '#93C5FD',
                  borderBottom: '2px solid #93C5FD',
                  background: '#2D3748'
                } : {}}
              >
                模板库
              </button>
            </nav>
          </div>
          
          {/* Tab 内容 */}
          <div className="tab-content">
            {/* API测试标签页 */}
            {activeTab === 'api-test' && (
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium mb-1" style={{ color: '#EAECEF' }}>
                    API密钥:
                  </label>
                  <input
                    type="password"
                    name="apiKey"
                    value={apiCredentials.apiKey}
                    onChange={handleInputChange}
                    className="w-full px-3 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="输入API密钥"
                    style={{
                      background: '#2B3139',
                      border: '1px solid #3D444D',
                      color: '#EAECEF'
                    }}
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium mb-1" style={{ color: '#EAECEF' }}>
                    密钥:
                  </label>
                  <input
                    type="password"
                    name="secretKey"
                    value={apiCredentials.secretKey}
                    onChange={handleInputChange}
                    className="w-full px-3 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="输入密钥"
                    style={{
                      background: '#2B3139',
                      border: '1px solid #3D444D',
                      color: '#EAECEF'
                    }}
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium mb-1" style={{ color: '#EAECEF' }}>
                    API URL:
                  </label>
                  <input
                    type="text"
                    name="apiUrl"
                    value={apiCredentials.apiUrl}
                    onChange={handleInputChange}
                    className="w-full px-3 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="输入API URL"
                    style={{
                      background: '#2B3139',
                      border: '1px solid #3D444D',
                      color: '#EAECEF'
                    }}
                  />
                </div>
                
                <button
                  onClick={runTest}
                  disabled={loading}
                  className="px-4 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
                  style={{
                    background: '#2B3139',
                    color: '#EAECEF',
                    border: '1px solid #3D444D'
                  }}
                  onMouseEnter={(e) => {
                    if (!loading) {
                      e.currentTarget.style.background = '#3D444D';
                    }
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.background = '#2B3139';
                  }}
                >
                  {loading ? '测试中...' : '运行测试'}
                </button>
                
                {testResults && (
                  <div className="mt-6">
                    <h2 className="text-lg font-semibold mb-2" style={{ color: '#EAECEF' }}>
                      测试结果:
                    </h2>
                    <pre 
                      className="p-4 rounded-md text-sm overflow-auto max-h-96 font-mono" 
                      style={{ 
                        background: '#1E293B', 
                        border: '1px solid #334155',
                        color: '#E2E8F0',
                        fontFamily: 'monospace',
                        lineHeight: '1.5'
                      }}
                    >
                      {testResults}
                    </pre>
                  </div>
                )}
              </div>
            )}
            
            {/* 批量交易标签页 */}
            {activeTab === 'batch-trade' && (
              <div>
                <div className="mb-4">
                  <label className="block text-sm font-medium mb-1" style={{ color: '#EAECEF' }}>
                    选择交易者:
                  </label>
                  <select
                    value={selectedTrader}
                    onChange={(e) => setSelectedTrader(e.target.value)}
                    className="w-full px-3 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    style={{
                      background: '#2B3139',
                      border: '1px solid #3D444D',
                      color: '#EAECEF'
                    }}
                  >
                    <option value="">请选择交易者</option>
                    {traders.map((trader) => (
                      <option key={trader.trader_id} value={trader.trader_id}>
                        {trader.trader_name} ({trader.trader_id})
                      </option>
                    ))}
                  </select>
                </div>
                
                <div className="mb-4">
                  <label className="block text-sm font-medium mb-1" style={{ color: '#EAECEF' }}>
                    批量交易指令 (JSON格式):
                  </label>
                  <textarea
                    value={batchCommands}
                    onChange={(e) => setBatchCommands(e.target.value)}
                    rows={10}
                    className="w-full px-3 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono"
                    style={{
                      background: '#2B3139',
                      border: '1px solid #3D444D',
                      color: '#EAECEF',
                      fontFamily: 'monospace',
                      fontSize: '12px',
                      lineHeight: '1.5'
                    }}
                  />
                  <div className="text-xs mt-1 text-gray-400">
                    示例: 更新止损、部分平仓等指令
                  </div>
                </div>
                
                <button
                  onClick={runBatchCommand}
                  disabled={batchLoading || !selectedTrader}
                  className="px-4 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
                  style={{
                    background: '#2B3139',
                    color: '#EAECEF',
                    border: '1px solid #3D444D'
                  }}
                  onMouseEnter={(e) => {
                    if (!batchLoading && selectedTrader) {
                      e.currentTarget.style.background = '#3D444D';
                    }
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.background = '#2B3139';
                  }}
                >
                  {batchLoading ? '执行中...' : '执行批量指令'}
                </button>
                
                {batchResults && (
                  <div className="mt-6">
                    <h2 className="text-lg font-semibold mb-2" style={{ color: '#EAECEF' }}>
                      批量执行结果:
                    </h2>
                    <pre 
                      className="p-4 rounded-md text-sm overflow-auto max-h-96 font-mono" 
                      style={{ 
                        background: '#1E293B', 
                        border: '1px solid #334155',
                        color: '#E2E8F0',
                        fontFamily: 'monospace',
                        lineHeight: '1.5'
                      }}
                    >
                      {batchResults}
                    </pre>
                  </div>
                )}
              </div>
            )}
            
            {/* 模板库标签页 */}
            {activeTab === 'templates' && (
              <div>
                <h3 className="text-lg font-semibold mb-4" style={{ color: '#EAECEF' }}>交易动作模板</h3>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {tradeTemplates.map((template) => (
                    <div 
                      key={template.id}
                      className="border border-gray-700 rounded-lg p-4 cursor-pointer hover:bg-gray-800 transition-colors"
                      style={{
                        background: '#2D3748',
                      }}
                      onClick={() => insertTemplate(template.template)}
                    >
                      <h4 className="font-medium mb-1" style={{ color: '#E2E8F0' }}>{template.name}</h4>
                      <p className="text-sm text-gray-400 mb-2">{template.description}</p>
                      <button 
                        className="text-xs px-2 py-1 rounded bg-blue-600 hover:bg-blue-700"
                        style={{
                          background: '#3182CE',
                        }}
                        onClick={(e) => {
                          e.stopPropagation();
                          insertTemplate(template.template);
                        }}
                      >
                        加载模板
                      </button>
                    </div>
                  ))}
                </div>
                
                <div className="mt-6">
                  <h3 className="text-lg font-semibold mb-2" style={{ color: '#EAECEF' }}>当前模板内容:</h3>
                  <textarea
                    value={batchCommands}
                    onChange={(e) => setBatchCommands(e.target.value)}
                    rows={10}
                    className="w-full px-3 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono"
                    style={{
                      background: '#2B3139',
                      border: '1px solid #3D444D',
                      color: '#EAECEF',
                      fontFamily: 'monospace',
                      fontSize: '12px',
                      lineHeight: '1.5'
                    }}
                  />
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default TestPage;