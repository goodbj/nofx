import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { useLanguage } from '../contexts/LanguageContext';
import { t as globalT } from '../i18n/translations';

const TestPage: React.FC = () => {
  const { token } = useAuth();
  const [apiCredentials, setApiCredentials] = useState({
    apiKey: '53zcowwoNVUiRWyg46RQb2nRwSQhUNwYeBEHetyJsjB0MLms1rqxABxWjXClkYoA1',
    secretKey: 'ZkvgnmrJiyOjuix0QxSa1eeQBwvlBENcwY5xBVp9uS085oDUpQHQcxHNfk0dppuj1',
    apiUrl: 'https://testnet.binancefuture.com',
  });
  
  const { language } = useLanguage();
  
  // 国际化文本
  const t = (key: string) => {
    return globalT(key, language);
  };
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
  const [currentPrices, setCurrentPrices] = useState<Record<string, number>>({});
  const [isFetchingPrice, setIsFetchingPrice] = useState(false);
  const [countdown, setCountdown] = useState(0);
  
  // 清理定时器
  useEffect(() => {
    return () => {
      // 组件卸载时清除可能存在的定时器
    };
  }, []);

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

  // 获取当前市场价格
  const fetchCurrentPrices = async () => {
    if (!token) return;
    
    try {
      // 尝试使用klines API获取最新价格
      const response = await fetch('/api/klines?symbol=BTCUSDT&interval=1m&limit=1', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });
      
      if (response.ok) {
        const klines = await response.json();
        // 从K线数据获取最新价格 (使用收盘价)
        const prices: Record<string, number> = {};
        
        if (Array.isArray(klines) && klines.length > 0) {
          // 获取BTCUSDT的最新价格
          const latestKline = klines[klines.length - 1];
          if (latestKline && latestKline.close !== undefined) {
            prices['BTCUSDT'] = parseFloat(latestKline.close);
          }
        }
        
        // 同时获取其他主要币种的价格
        const otherSymbols = ['ETHUSDT', 'SOLUSDT', 'BNBUSDT', 'ADAUSDT', 'XRPUSDT'];
        for (const symbol of otherSymbols) {
          try {
            const symbolResponse = await fetch(`/api/klines?symbol=${symbol}&interval=1m&limit=1`, {
              method: 'GET',
              headers: {
                'Authorization': `Bearer ${token}`,
              },
            });
            
            if (symbolResponse.ok) {
              const symbolKlines = await symbolResponse.json();
              if (Array.isArray(symbolKlines) && symbolKlines.length > 0) {
                const latestKline = symbolKlines[symbolKlines.length - 1];
                if (latestKline && latestKline.close !== undefined) {
                  prices[symbol] = parseFloat(latestKline.close);
                }
              }
            }
          } catch (err) {
            console.warn(`获取 ${symbol} 价格失败:`, err);
          }
        }
        
        // 如果从API获取不到价格数据，使用模拟价格
        if (Object.keys(prices).length === 0) {
          throw new Error('No price data available from API');
        }
        
        setCurrentPrices(prices);
        return prices;
      } else {
        // 如果klines API也不可用，使用模拟价格
        throw new Error('Klines API not available');
      }
    } catch (error) {
      console.warn('获取市场价格失败，使用模拟价格:', error);
      // 如果API不可用，使用模拟价格
      const mockPrices = {
        'BTCUSDT': 43560.42,
        'ETHUSDT': 2635.80,
        'SOLUSDT': 98.45,
        'BNBUSDT': 300.50,
        'ADAUSDT': 0.52,
        'XRPUSDT': 0.55,
        'DOTUSDT': 6.20,
        'AVAXUSDT': 35.80,
      };
      setCurrentPrices(mockPrices);
      return mockPrices;
    }
  };

  // 生成基于当前价格的开多模板
  const generateLongTemplate = (symbol: string = 'BTCUSDT') => {
    const currentPrice = currentPrices[symbol] || 43560.42;
    const stopLoss = currentPrice * 0.98; // -2%
    const takeProfit = currentPrice * 1.03; // +3%
    
    const template = `[
  {
    "symbol": "${symbol}",
    "action": "open_long",
    "leverage": 10,
    "position_size_usd": 201,
    "stop_loss": ${stopLoss.toFixed(2)},
    "take_profit": ${takeProfit.toFixed(2)},
    "confidence": 85,
    "risk_usd": 20
  }
]`;
    
    insertTemplate(template);
    return template;
  };

  // 生成基于当前价格的开空模板
  const generateShortTemplate = (symbol: string = 'BTCUSDT') => {
    const currentPrice = currentPrices[symbol] || 43560.42;
    const stopLoss = currentPrice * 1.02; // +2%
    const takeProfit = currentPrice * 0.97; // -3%
    
    const template = `[
  {
    "symbol": "${symbol}",
    "action": "open_short",
    "leverage": 10,
    "position_size_usd": 201,
    "stop_loss": ${stopLoss.toFixed(2)},
    "take_profit": ${takeProfit.toFixed(2)},
    "confidence": 85,
    "risk_usd": 20
  }
]`;
    
    insertTemplate(template);
    return template;
  };

  // 生成基于当前价格的更新止损模板
  const generateUpdateStopLossTemplate = (symbol: string = 'BTCUSDT') => {
    const currentPrice = currentPrices[symbol] || 43560.42;
    const newStopLoss = currentPrice * 0.99; // -1% (用于多单) 或 +1% (用于空单)
    
    const template = `[
  {
    "symbol": "${symbol}",
    "action": "update_stop_loss",
    "new_stop_loss": ${newStopLoss.toFixed(2)},
    "confidence": 75
  }
]`;
    
    insertTemplate(template);
    return template;
  };

  // 生成基于当前价格的更新止盈模板
  const generateUpdateTakeProfitTemplate = (symbol: string = 'BTCUSDT') => {
    const currentPrice = currentPrices[symbol] || 43560.42;
    const newTakeProfit = currentPrice * 1.02; // +2% (用于多单) 或 -2% (用于空单)
    
    const template = `[
  {
    "symbol": "${symbol}",
    "action": "update_take_profit",
    "new_take_profit": ${newTakeProfit.toFixed(2)},
    "confidence": 75
  }
]`;
    
    insertTemplate(template);
    return template;
  };

  // 获取并生成当前价格模板
  const loadCurrentPriceTemplates = async () => {
    if (countdown > 0) return; // 如果倒计时未结束，不执行操作
    
    setIsFetchingPrice(true);
    
    // 立即设置倒计时，20秒内禁止重复点击
    setCountdown(20);
    const interval = setInterval(() => {
      setCountdown(prev => {
        if (prev <= 1) {
          clearInterval(interval);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);
    
    try {
      await fetchCurrentPrices();
    } finally {
      setIsFetchingPrice(false);
    }
  };

  // 页面加载时获取交易者列表
  React.useEffect(() => {
    fetchTraders();
  }, [token]);

  const runTest = async () => {
    if (!token) {
      setTestResults(t('not_authenticated'));
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
      setTestResults(`${t('error_occurred')}${error}`);
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
      setBatchResults(t('please_select_trader_error'));
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
        setBatchResults(`${t('json_parse_error')}${parseError.message}`);
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
      setBatchResults(`${t('error_occurred')}${error}`);
    } finally {
      setBatchLoading(false);
    }
  };

  const [activeTab, setActiveTab] = useState('batch-trade'); // 'api-test', 'batch-trade'

  // 定义交易动作模板
  const tradeTemplates = [
    {
      id: 'open_long',
      name: t('open_long_btc'),
      description: t('buy_btc_long'),
      category: 'long',
      template: `[
  {
    "symbol": "BTCUSDT",
    "action": "open_long",
    "leverage": 10,
    "position_size_usd": 201,
    "stop_loss": 59000.00,
    "take_profit": 62000.00,
    "confidence": 85,
    "risk_usd": 20
  }
]`
    },
    {
      id: 'open_short',
      name: t('open_short_btc'),
      description: t('sell_btc_short'),
      category: 'short',
      template: `[
  {
    "symbol": "BTCUSDT",
    "action": "open_short",
    "leverage": 10,
    "position_size_usd": 201,
    "stop_loss": 61000.00,
    "take_profit": 58000.00,
    "confidence": 85,
    "risk_usd": 20
  }
]`
    },
    {
      id: 'close_long',
      name: t('close_long_btc'),
      description: t('close_btc_long'),
      category: 'long',
      template: `[
  {
    "symbol": "BTCUSDT",
    "action": "close_long",
    "confidence": 85,
    "risk_usd": 20
  }
]`
    },
    {
      id: 'close_short',
      name: t('close_short_btc'),
      description: t('close_btc_short'),
      category: 'short',
      template: `[
  {
    "symbol": "BTCUSDT",
    "action": "close_short",
    "confidence": 85,
    "risk_usd": 20
  }
]`
    },
    {
      id: 'hold',
      name: t('hold_position'),
      description: t('maintain_position'),
      category: 'general',
      template: `[
  {
    "symbol": "BTCUSDT",
    "action": "hold",
    "confidence": 60,
    "risk_usd": 20
  }
]`
    },
    {
      id: 'wait',
      name: t('wait_and_watch'),
      description: t('wait_for_opportunity'),
      category: 'general',
      template: `[
  {
    "symbol": "BTCUSDT",
    "action": "wait",
    "confidence": 50,
    "risk_usd": 20
  }
]`
    },
    {
      id: 'update_stop_loss',
      name: t('update_stop_loss'),
      description: t('modify_stop_loss'),
      category: 'general',
      template: `[
  {
    "symbol": "BTCUSDT",
    "action": "update_stop_loss",
    "new_stop_loss": 59000.00,
    "confidence": 75,
    "risk_usd": 20
  }
]`
    },
    {
      id: 'update_take_profit',
      name: t('update_take_profit'),
      description: t('modify_take_profit'),
      category: 'general',
      template: `[
  {
    "symbol": "BTCUSDT",
    "action": "update_take_profit",
    "new_take_profit": 65000.00,
    "confidence": 75,
    "risk_usd": 20
  }
]`
    },
    {
      id: 'stop_take_combined',
      name: t('stop_take_combo'),
      description: t('set_stop_take'),
      category: 'general',
      template: 'dynamic:stop_take_combined'
    },
    {
      id: 'partial_close',
      name: t('partial_close'),
      description: t('close_part_position'),
      category: 'general',
      template: `[
  {
    "symbol": "BTCUSDT",
    "action": "partial_close",
    "close_percentage": 50,
    "confidence": 80,
    "risk_usd": 20
  }
]`
    },
    {
      id: 'oco_order',
      name: t('oco_order'),
      description: t('set_stop_take_oto'),
      category: 'general',
      template: 'dynamic:oco_order'
    },
    {
      id: 'bracket_order',
      name: t('bracket_order'),
      description: t('place_bracket_order'),
      category: 'general',
      template: 'dynamic:bracket_order'
    },
    {
      id: 'trailing_stop',
      name: t('trailing_stop'),
      description: t('set_trailing_stop'),
      category: 'general',
      template: `[{  "symbol": "BTCUSDT",  "action": "trailing_stop",  "trail_percentage": 3.0,  "activation_price": 43560.42,  "callback_rate": 0.02,  "confidence": 80,  "risk_usd": 20}]`
    },
    {
      id: 'dynamic_take_profit',
      name: t('dynamic_take_profit'),
      description: t('adjust_take_profit_dynamic'),
      category: 'general',
      template: `[{  "symbol": "BTCUSDT",  "action": "dynamic_take_profit",  "target_roi": 15.0,  "max_roi": 25.0,  "time_limit_hours": 24,  "confidence": 75,  "risk_usd": 20}]`
    }
  ];

  // 插入模板到指令框
  const insertTemplate = (templateOrType: string) => {
    // 检查是否是模板类型而不是实际的模板字符串
    if (templateOrType.startsWith('dynamic:')) {
      const templateType = templateOrType.replace('dynamic:', '');
      
      // 根据类型生成动态模板
      switch (templateType) {
        case 'open_long':
          generateLongTemplate('BTCUSDT');
          return;
        case 'open_short':
          generateShortTemplate('BTCUSDT');
          return;
        case 'update_stop_loss':
          generateUpdateStopLossTemplate('BTCUSDT');
          return;
        case 'update_take_profit':
          generateUpdateTakeProfitTemplate('BTCUSDT');
          return;
        case 'stop_take_combined':
          // 组合模板：同时生成更新止损和更新止盈
          const currentPrice = currentPrices['BTCUSDT'] || 43560.42;
          const stopLoss = currentPrice * 0.99; // -1%
          const takeProfit = currentPrice * 1.02; // +2%
          const combinedTemplate = `[
  {
    "symbol": "BTCUSDT",
    "action": "update_stop_loss",
    "new_stop_loss": ${stopLoss.toFixed(2)},
    "confidence": 75,
    "risk_usd": 20
  },
  {
    "symbol": "BTCUSDT",
    "action": "update_take_profit",
    "new_take_profit": ${takeProfit.toFixed(2)},
    "confidence": 80,
    "risk_usd": 20
  }
]`;
          setBatchCommands(combinedTemplate);
          return;
        case 'oco_order':
          // OCO订单：基于当前价格生成止盈止损
          const ocoCurrentPrice = currentPrices['BTCUSDT'] || 43560.42;
          const ocoStopLoss = ocoCurrentPrice * 0.99; // -1%
          const ocoTakeProfit = ocoCurrentPrice * 1.02; // +2%
          const ocoTemplate = `[
  {
    "symbol": "BTCUSDT",
    "action": "oco_order",
    "stop_loss": ${ocoStopLoss.toFixed(2)},
    "take_profit": ${ocoTakeProfit.toFixed(2)},
    "order_quantity": 0.001,
    "confidence": 80,
    "risk_usd": 20
  }
]`;
          setBatchCommands(ocoTemplate);
          return;
        case 'bracket_order':
          // Bracket订单：基于当前价格生成止盈止损
          const bracketCurrentPrice = currentPrices['BTCUSDT'] || 43560.42;
          const bracketStopLoss = bracketCurrentPrice * 0.99; // -1%
          const bracketTakeProfit = bracketCurrentPrice * 1.02; // +2%
          const bracketTemplate = `[
  {
    "symbol": "BTCUSDT",
    "action": "bracket_order",
    "stop_loss": ${bracketStopLoss.toFixed(2)},
    "take_profit": ${bracketTakeProfit.toFixed(2)},
    "order_quantity": 0.001,
    "confidence": 80,
    "risk_usd": 20
  }
]`;
          setBatchCommands(bracketTemplate);
          return;
        default:
          break;
      }
    }
    
    // 如果不是动态模板类型，则使用原始模板字符串
    setBatchCommands(templateOrType);
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
            {t('test_page_title')}
          </h1>
          
          {/* Tab 标签页 */}
          <div className="mb-6 border-b border-gray-700">
            <nav className="flex space-x-2">
              <button
                className={`px-4 py-2 font-medium text-sm rounded-t-lg transition-colors ${activeTab === 'batch-trade' ? 'text-blue-400 border-b-2 border-blue-400 bg-gray-800' : 'text-gray-400 hover:text-gray-200'}`}
                onClick={() => setActiveTab('batch-trade')}
                style={activeTab === 'batch-trade' ? {
                  color: '#93C5FD',
                  borderBottom: '2px solid #93C5FD',
                  background: '#2D3748'
                } : {}}
              >
                {t('batch_trade_tab')}
              </button>
              <button
                className={`px-4 py-2 font-medium text-sm rounded-t-lg transition-colors ${activeTab === 'api-test' ? 'text-blue-400 border-b-2 border-blue-400 bg-gray-800' : 'text-gray-400 hover:text-gray-200'}`}
                onClick={() => setActiveTab('api-test')}
                style={activeTab === 'api-test' ? {
                  color: '#93C5FD',
                  borderBottom: '2px solid #93C5FD',
                  background: '#2D3748'
                } : {}}
              >
                {t('api_test_tab')}
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
                    {t('api_key_label')}:
                  </label>
                  <input
                    type="password"
                    name="apiKey"
                    value={apiCredentials.apiKey}
                    onChange={handleInputChange}
                    className="w-full px-3 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder={t('api_key_label')}
                    style={{
                      background: '#2B3139',
                      border: '1px solid #3D444D',
                      color: '#EAECEF'
                    }}
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium mb-1" style={{ color: '#EAECEF' }}>
                    {t('secret_key_label')}:
                  </label>
                  <input
                    type="password"
                    name="secretKey"
                    value={apiCredentials.secretKey}
                    onChange={handleInputChange}
                    className="w-full px-3 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder={t('secret_key_label')}
                    style={{
                      background: '#2B3139',
                      border: '1px solid #3D444D',
                      color: '#EAECEF'
                    }}
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium mb-1" style={{ color: '#EAECEF' }}>
                    {t('api_url_label')}:
                  </label>
                  <input
                    type="text"
                    name="apiUrl"
                    value={apiCredentials.apiUrl}
                    onChange={handleInputChange}
                    className="w-full px-3 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder={t('api_url_label')}
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
                  {loading ? t('running_test') : t('run_test')}
                </button>
                
                {testResults && (
                  <div className="mt-6">
                    <h2 className="text-lg font-semibold mb-2" style={{ color: '#EAECEF' }}>
                      {t('test_results')}:
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
                    {t('select_trader_label')}:
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
                    <option value="">{t('please_select_trader')}</option>
                    {traders.map((trader) => (
                      <option key={trader.trader_id} value={trader.trader_id}>
                        {trader.trader_name} ({trader.trader_id})
                      </option>
                    ))}
                  </select>
                </div>
                
                {/* 模板库 */}
                <div className="mb-6">
                  <div className="flex justify-between items-center mb-4">
                    <h3 className="text-lg font-semibold" style={{ color: '#EAECEF' }}>{t('trade_action_templates')}</h3>
                    <div className="flex items-center gap-2">
                      <span className="text-sm text-gray-400">
                        {t('current_btc_price')}{(currentPrices['BTCUSDT'] || 43560.42).toFixed(2)} USDT
                      </span>
                      <button 
                        onClick={loadCurrentPriceTemplates}
                        disabled={countdown > 0}
                        className="px-3 py-1 rounded text-sm disabled:opacity-50"
                        style={{
                          background: countdown > 0 ? '#718096' : '#3182CE',
                          color: 'white'
                        }}
                      >
                        {isFetchingPrice ? t('fetching') : countdown > 0 ? `${countdown}${t('seconds_remaining')}` : t('get_current_price')}
                      </button>
                    </div>
                  </div>
                  
                  <div className="flex flex-col lg:flex-row gap-6">
                    <div className="lg:w-1/2">
                      <div className="space-y-4">
                        {/* 多仓相关模板 */}
                        <div>
                          <h4 className="font-medium mb-2" style={{ color: '#E2E8F0', fontSize: '1rem' }}>{t('long_positions')}</h4>
                          <div className="grid grid-cols-2 md:grid-cols-3 gap-1">
                            {tradeTemplates
                              .filter(template => template.category === 'long')
                              .map((template) => (
                                <button
                                  key={template.id}
                                  className="border border-gray-700 rounded p-1 text-xs text-left break-words"
                                  style={{
                                    background: '#38A169',
                                    color: 'white',
                                    minHeight: 'auto',
                                    height: 'auto'
                                  }}
                                  onClick={() => insertTemplate(template.template)}
                                >
                                  BTC{template.name.replace('(BTC)', '')}
                                  <br/>{template.description.split(' ').join('\n')}
                                </button>
                              ))}
                          </div>
                        </div>
                        
                        {/* 空仓相关模板 */}
                        <div>
                          <h4 className="font-medium mb-2" style={{ color: '#E2E8F0', fontSize: '1rem' }}>{t('short_positions')}</h4>
                          <div className="grid grid-cols-2 md:grid-cols-3 gap-1">
                            {tradeTemplates
                              .filter(template => template.category === 'short')
                              .map((template) => (
                                <button
                                  key={template.id}
                                  className="border border-gray-700 rounded p-1 text-xs text-left break-words"
                                  style={{
                                    background: '#E53E3E',
                                    color: 'white',
                                    minHeight: 'auto',
                                    height: 'auto'
                                  }}
                                  onClick={() => insertTemplate(template.template)}
                                >
                                  BTC{template.name.replace('(BTC)', '')}
                                  <br/>{template.description.split(' ').join('\n')}
                                </button>
                              ))}
                          </div>
                        </div>
                        
                        {/* 通用操作模板 */}
                        <div>
                          <h4 className="font-medium mb-2" style={{ color: '#E2E8F0', fontSize: '1rem' }}>{t('general_operations')}</h4>
                          <div className="grid grid-cols-2 md:grid-cols-3 gap-1">
                            {tradeTemplates
                              .filter(template => template.category === 'general')
                              .map((template) => (
                                <button
                                  key={template.id}
                                  className="border border-gray-700 rounded p-1 text-xs text-left break-words"
                                  style={{
                                    background: '#4A5568',
                                    color: 'white',
                                    minHeight: 'auto',
                                    height: 'auto'
                                  }}
                                  onClick={() => insertTemplate(template.template)}
                                >
                                  BTC{template.name.replace('(BTC)', '')}
                                  <br/>{template.description.split(' ').join('\n')}
                                </button>
                              ))}
                          </div>
                        </div>
                      </div>
                    </div>
                    
                    <div className="lg:w-1/2">
                      <div className="mb-4">
                        <label className="block text-sm font-medium mb-1" style={{ color: '#EAECEF' }}>
                          {t('batch_commands_label')}
                        </label>
                        <textarea
                          value={batchCommands}
                          onChange={(e) => setBatchCommands(e.target.value)}
                          rows={16}
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
                          {t('example_instructions')}
                        </div>
                      </div>
                      
                      <button
                        onClick={runBatchCommand}
                        disabled={batchLoading || !selectedTrader}
                        className="w-full px-4 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
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
                        {batchLoading ? t('executing') : t('execute_batch')}
                      </button>
                      
                      {batchResults && (
                        <div className="mt-6">
                          <h2 className="text-lg font-semibold mb-2" style={{ color: '#EAECEF' }}>
                            {t('batch_execution_results')}:
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
                  </div>
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