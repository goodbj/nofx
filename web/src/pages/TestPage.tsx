import React, { useState, useEffect } from 'react';
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
  const [manualScanLoading, setManualScanLoading] = useState(false);
  const [syncBalanceLoading, setSyncBalanceLoading] = useState(false);
  const [traderActionLoading, setTraderActionLoading] = useState(false);
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

  const [activeTab, setActiveTab] = useState('batch-trade'); // 'api-test', 'batch-trade', 'test-scripts', 'ai-prompt'
  const [testScripts, setTestScripts] = useState<string[]>([]);
  const [fetchingScripts, setFetchingScripts] = useState(false);
  
  // AI Prompt 状态
  const [promptData, setPromptData] = useState<{
    system_prompt: string;
    user_prompt: string;
    timestamp: string;
    timestamp_ms: number;
    system_bytes: number;
    user_bytes: number;
    total_bytes: number;
    system_chars: number;
    user_chars: number;
    total_chars: number;
    estimated_tokens: number;
    actual_request_json?: string; // 实际发送的 JSON 请求体
  } | null>(null);
  const [promptLoading, setPromptLoading] = useState(false);
  const [promptError, setPromptError] = useState<string>('');
  const [systemConfig, setSystemConfig] = useState<any>(null);
  const [showTechnicalDetails, setShowTechnicalDetails] = useState(false); // 显示技术验证详情

  // 获取测试脚本列表
  const fetchTestScripts = async () => {
    if (!token) return;
    setFetchingScripts(true);
    try {
      const response = await fetch('/api/test/list', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      if (response.ok) {
        const data = await response.json();
        setTestScripts(data.tests || []);
      }
    } catch (error) {
      console.error('获取测试脚本列表失败:', error);
    } finally {
      setFetchingScripts(false);
    }
  };

  // 获取最近一次的 AI Prompt
  const fetchLastPrompt = async () => {
    if (!token) return;
    setPromptLoading(true);
    setPromptError('');
    try {
      const response = await fetch('/api/test/last-prompt', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      if (response.ok) {
        const data = await response.json();
        setPromptData(data);
      } else {
        const errorData = await response.json();
        setPromptError(errorData.error || '获取失败');
        setPromptData(null);
      }
    } catch (error: any) {
      setPromptError(`${t('error_occurred')}${error.message}`);
      setPromptData(null);
    } finally {
      setPromptLoading(false);
    }
  };

  useEffect(() => {
    if (activeTab === 'test-scripts') {
      fetchTestScripts();
    } else if (activeTab === 'ai-prompt') {
      fetchLastPrompt();
    }
  }, [activeTab, token]);
  
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
      const response = await fetch('/api/klines?symbol=BTCUSDT&interval=1m&limit=1', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });
      
      if (response.ok) {
        const klines = await response.json();
        const prices: Record<string, number> = {};
        
        if (Array.isArray(klines) && klines.length > 0) {
          const latestKline = klines[klines.length - 1];
          if (latestKline && latestKline.close !== undefined) {
            prices['BTCUSDT'] = parseFloat(latestKline.close);
          }
        }
        
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
        
        if (Object.keys(prices).length === 0) {
          throw new Error('No price data available from API');
        }
        
        setCurrentPrices(prices);
        return prices;
      } else {
        throw new Error('Klines API not available');
      }
    } catch (error) {
      console.warn('获取市场价格失败，使用模拟价格:', error);
      const mockPrices = {
        'BTCUSDT': 43560.42,
        'ETHUSDT': 2635.80,
        'SOLUSDT': 98.45,
        'BNBUSDT': 300.50,
        'ADAUSDT': 0.52,
        'XRPUSDT': 0.55,
      };
      setCurrentPrices(mockPrices);
      return mockPrices;
    }
  };

  const generateLongTemplate = (symbol: string = 'BTCUSDT') => {
    const currentPrice = currentPrices[symbol] || 43560.42;
    const stopLoss = currentPrice * 0.98;
    const takeProfit = currentPrice * 1.03;
    const template = `[
  {
    "symbol": "${symbol}",
    "action": "open_long",
    "leverage": 10,
    "position_size_usd": 200,
    "stop_loss": ${stopLoss.toFixed(2)},
    "take_profit": ${takeProfit.toFixed(2)},
    "confidence": 85,
    "risk_usd": 20
  }
]`;
    insertTemplate(template);
    return template;
  };

  const generateShortTemplate = (symbol: string = 'BTCUSDT') => {
    const currentPrice = currentPrices[symbol] || 43560.42;
    const stopLoss = currentPrice * 1.02;
    const takeProfit = currentPrice * 0.97;
    const template = `[
  {
    "symbol": "${symbol}",
    "action": "open_short",
    "leverage": 10,
    "position_size_usd": 200,
    "stop_loss": ${stopLoss.toFixed(2)},
    "take_profit": ${takeProfit.toFixed(2)},
    "confidence": 85,
    "risk_usd": 20
  }
]`;
    insertTemplate(template);
    return template;
  };

  const generateUpdateStopLossTemplate = (symbol: string = 'BTCUSDT') => {
    const currentPrice = currentPrices[symbol] || 43560.42;
    const newStopLoss = currentPrice * 0.99;
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

  const generateUpdateTakeProfitTemplate = (symbol: string = 'BTCUSDT') => {
    const currentPrice = currentPrices[symbol] || 43560.42;
    const newTakeProfit = currentPrice * 1.02;
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

  const loadCurrentPriceTemplates = async () => {
    if (countdown > 0) return;
    setIsFetchingPrice(true);
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

  useEffect(() => {
    fetchTraders();
    fetchSystemConfig();
  }, [token]);

  // 获取系统配置
  const fetchSystemConfig = async () => {
    try {
      const response = await fetch('/api/config');
      if (response.ok) {
        const data = await response.json();
        setSystemConfig(data);
      }
    } catch (error) {
      console.error('获取系统配置失败:', error);
    }
  };

  // 计算费用
  const calculateCost = (tokens: number, modelName: string = 'default'): { usd: number; cny: number } => {
    if (!systemConfig?.model_pricing) {
      return { usd: 0, cny: 0 };
    }

    // 匹配模型定价
    const lowerModelName = modelName.toLowerCase();
    let pricing = systemConfig.model_pricing[lowerModelName] || systemConfig.model_pricing['default'];

    // 尝试模糊匹配
    if (!pricing) {
      const keys = Object.keys(systemConfig.model_pricing);
      for (const key of keys) {
        if (lowerModelName.includes(key) || key.includes(lowerModelName)) {
          pricing = systemConfig.model_pricing[key];
          break;
        }
      }
    }

    if (!pricing) {
      pricing = { input_price: 0.5, output_price: 2.0 };
    }

    // 价格是每百万 tokens，需要转换
    const costUSD = (tokens / 1000000) * pricing.input_price;
    const exchangeRate = systemConfig.usd_to_cny_rate || 7.2;
    const costCNY = costUSD * exchangeRate;

    return { usd: costUSD, cny: costCNY };
  };

  // 格式化费用显示
  const formatCost = (cost: { usd: number; cny: number }): string => {
    const displayUnit = systemConfig?.cost_display_unit || 'BOTH';
    
    if (cost.usd === 0) return '免费 (本地模型)';

    const formatNumber = (num: number): string => {
      if (num >= 1) return num.toFixed(4);
      if (num >= 0.001) return num.toFixed(6);
      return num.toExponential(2);
    };

    switch (displayUnit) {
      case 'USD':
        return `$${formatNumber(cost.usd)}`;
      case 'CNY':
        return `¥${formatNumber(cost.cny)}`;
      case 'BOTH':
      default:
        return `$${formatNumber(cost.usd)} (¥${formatNumber(cost.cny)})`;
    }
  };

  // 复制到剪贴板
  const copyToClipboard = async (text: string, label: string) => {
    try {
      await navigator.clipboard.writeText(text);
      // 可以添加一个toast提示
      alert(`${label} 已复制到剪贴板！`);
    } catch (error) {
      console.error('复制失败:', error);
      alert('复制失败，请手动复制');
    }
  };

  const runTest = async (reqCommand?: string) => {
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
          command: reqCommand || 'go run test/test_binance_testnet.go',
          apiKey: apiCredentials.apiKey,
          secretKey: apiCredentials.secretKey,
          apiUrl: apiCredentials.apiUrl,
        }),
      });

      const data = await response.json();
      if (!response.ok) {
        setTestResults(`Error (${response.status}): ${data.error || data.output || JSON.stringify(data)}`);
      } else {
        setTestResults(`Status: ${data.status}\nOutput: ${data.output}`);
      }
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
      if (!response.ok) {
        setBatchResults(`Error (${response.status}): ${data.error || JSON.stringify(data, null, 2)}`);
      } else {
        setBatchResults(JSON.stringify(data, null, 2));
      }
    } catch (error) {
      setBatchResults(`${t('error_occurred')}${error}`);
    } finally {
      setBatchLoading(false);
    }
  };

  const handleManualScan = async () => {
    if (!token) return;
    if (!selectedTrader) {
      setBatchResults(t('please_select_trader_error'));
      return;
    }
    setManualScanLoading(true);
    setBatchResults('');
    try {
      const response = await fetch(`/api/traders/${selectedTrader}/execute-decision`, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}` }
      });
      const data = await response.json();
      setBatchResults(JSON.stringify(data, null, 2));
    } catch (error) {
      setBatchResults(`${t('error_occurred')}${error}`);
    } finally {
      setManualScanLoading(false);
    }
  };

  const handleSyncBalance = async () => {
    if (!token) return;
    if (!selectedTrader) {
      setBatchResults(t('please_select_trader_error'));
      return;
    }
    setSyncBalanceLoading(true);
    setBatchResults('');
    try {
      const response = await fetch(`/api/traders/${selectedTrader}/sync-balance`, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}` }
      });
      const data = await response.json();
      setBatchResults(JSON.stringify(data, null, 2));
    } catch (error) {
      setBatchResults(`${t('error_occurred')}${error}`);
    } finally {
      setSyncBalanceLoading(false);
    }
  };

  const handleStartTrader = async () => {
    if (!token || !selectedTrader) return;
    setTraderActionLoading(true);
    try {
      const response = await fetch(`/api/traders/${selectedTrader}/start`, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}` }
      });
      const data = await response.json();
      setBatchResults(JSON.stringify(data, null, 2));
      fetchTraders(); // Refresh status
    } catch (error) {
      setBatchResults(`${t('error_occurred')}${error}`);
    } finally {
      setTraderActionLoading(false);
    }
  };

  const handleStopTrader = async () => {
    if (!token || !selectedTrader) return;
    setTraderActionLoading(true);
    try {
      const response = await fetch(`/api/traders/${selectedTrader}/stop`, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}` }
      });
      const data = await response.json();
      setBatchResults(JSON.stringify(data, null, 2));
      fetchTraders(); // Refresh status
    } catch (error) {
      setBatchResults(`${t('error_occurred')}${error}`);
    } finally {
      setTraderActionLoading(false);
    }
  };

  const handleCloseAllPositions = async () => {
    if (!token || !selectedTrader) return;
    setTraderActionLoading(true);
    try {
      // Fetch current positions first
      const posResponse = await fetch(`/api/positions?trader_id=${selectedTrader}`, {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      const positions = await posResponse.json();
      
      if (!Array.isArray(positions) || positions.length === 0) {
        setBatchResults('No active positions to close.');
        setTraderActionLoading(false);
        return;
      }

      setBatchResults(`Closing ${positions.length} positions...`);
      
      const results = [];
      for (const pos of positions) {
        const closeRes = await fetch(`/api/traders/${selectedTrader}/close-position`, {
          method: 'POST',
          headers: { 
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}` 
          },
          body: JSON.stringify({ symbol: pos.symbol, side: pos.side.toUpperCase() })
        });
        const data = await closeRes.json();
        results.push({ symbol: pos.symbol, result: data });
      }
      
      setBatchResults(JSON.stringify(results, null, 2));
    } catch (error) {
      setBatchResults(`${t('error_occurred')}${error}`);
    } finally {
      setTraderActionLoading(false);
    }
  };

  const tradeTemplates = [
    { id: 'open_long_btc', name: t('open_long_btc'), description: t('buy_btc_long'), category: 'long', template: 'dynamic:open_long:BTCUSDT' },
    { id: 'close_long_btc', name: t('close_long_btc'), description: t('close_btc_long'), category: 'long', template: '[\n  {\n    "symbol": "BTCUSDT",\n    "action": "close_long",\n    "confidence": 100,\n    "reasoning": "Manual test"\n  }\n]' },
    
    { id: 'open_short_btc', name: t('open_short_btc'), description: t('sell_btc_short'), category: 'short', template: 'dynamic:open_short:BTCUSDT' },
    { id: 'close_short_btc', name: t('close_short_btc'), description: t('close_btc_short'), category: 'short', template: '[\n  {\n    "symbol": "BTCUSDT",\n    "action": "close_short",\n    "confidence": 100,\n    "reasoning": "Manual test"\n  }\n]' },

    { id: 'hold', name: t('hold_position'), description: t('maintain_position'), category: 'general', template: '[\n  {\n    "symbol": "BTCUSDT",\n    "action": "hold",\n    "confidence": 60,\n    "risk_usd": 20,\n    "reasoning": "Manual test"\n  }\n]' },
    { id: 'wait', name: t('wait_and_watch'), description: t('wait_for_opportunity'), category: 'general', template: '[\n  {\n    "symbol": "BTCUSDT",\n    "action": "wait",\n    "confidence": 50,\n    "risk_usd": 20,\n    "reasoning": "Manual test"\n  }\n]' },
    { id: 'update_stop_loss_btc', name: t('modify_stop_loss'), description: 'SL', category: 'general', template: 'dynamic:update_stop_loss:BTCUSDT' },
    { id: 'update_take_profit_btc', name: t('modify_take_profit'), description: 'TP', category: 'general', template: 'dynamic:update_take_profit:BTCUSDT' },
    { id: 'stop_take_combo_btc', name: t('stop_take_combo'), description: 'SL/TP', category: 'general', template: 'dynamic:stop_take_combo:BTCUSDT' },
    { id: 'partial_close_btc', name: t('partial_close'), description: 'Part Close', category: 'general', template: 'dynamic:partial_close:BTCUSDT' },
    { id: 'trailing_stop_btc', name: t('trailing_stop'), description: 'Trail SL', category: 'advanced', template: 'dynamic:trailing_stop:BTCUSDT' },
    { id: 'dynamic_take_profit_btc', name: t('dynamic_take_profit'), description: 'Dyn TP', category: 'advanced', template: 'dynamic:dynamic_take_profit:BTCUSDT' },
    { id: 'oco_order_btc', name: t('oco_order'), description: 'OCO', category: 'advanced', template: 'dynamic:oco_order:BTCUSDT' },
    { id: 'bracket_order_btc', name: t('bracket_order'), description: 'Bracket', category: 'advanced', template: 'dynamic:bracket_order:BTCUSDT' },
    { id: 'add_to_position_btc', name: t('add_to_position'), description: 'Add Pos', category: 'advanced', template: 'dynamic:add_to_position:BTCUSDT' },
  ];

  const insertTemplate = (templateOrType: string) => {
    if (templateOrType.startsWith('dynamic:')) {
      const parts = templateOrType.split(':');
      const templateType = parts[1];
      const symbol = parts[2] || 'BTCUSDT';
      
      switch (templateType) {
        case 'open_long': generateLongTemplate(symbol); break;
        case 'open_short': generateShortTemplate(symbol); break;
        case 'update_stop_loss': generateUpdateStopLossTemplate(symbol); break;
        case 'update_take_profit': generateUpdateTakeProfitTemplate(symbol); break;
        case 'stop_take_combo':
          const cp = currentPrices[symbol] || 43560.42;
          const isLong = true; // Assume testing long stop/take
          setBatchCommands(`[
  {
    "symbol": "${symbol}",
    "action": "update_stop_loss",
    "new_stop_loss": ${(isLong ? cp * 0.98 : cp * 1.02).toFixed(2)},
    "confidence": 85,
    "reasoning": "Manual test SL"
  },
  {
    "symbol": "${symbol}",
    "action": "update_take_profit",
    "new_take_profit": ${(isLong ? cp * 1.05 : cp * 0.95).toFixed(2)},
    "confidence": 85,
    "reasoning": "Manual test TP"
  }
]`);
          break;
        case 'partial_close':
          setBatchCommands(`[
  {
    "symbol": "${symbol}",
    "action": "partial_close",
    "close_percentage": 50,
    "confidence": 80,
    "risk_usd": 20,
    "reasoning": "Manual test"
  }
]`);
          break;
        case 'trailing_stop':
          const cpTrail = currentPrices[symbol] || 43560.42;
          const trailPercentage = 2.0; // 2% callback rate
          // Note: Binance API expects callback_rate as percentage value (0.1 to 5.0)
          // NOT as decimal (0.001 to 0.05)
          setBatchCommands(`[
  {
    "symbol": "${symbol}",
    "action": "trailing_stop",
    "trail_percentage": ${trailPercentage},
    "callback_rate": ${trailPercentage},
    "activation_price": ${(cpTrail * 1.02).toFixed(2)},
    "confidence": 80,
    "reasoning": "测试追踪止损：回调率${trailPercentage}%，当价格回撤${trailPercentage}%时触发止损"
  }
]`);
          break;
        case 'dynamic_take_profit':
          setBatchCommands(`[
  {
    "symbol": "${symbol}",
    "action": "dynamic_take_profit",
    "target_roi": 5.0,
    "max_roi": 10.0,
    "time_limit_hours": 24,
    "confidence": 75,
    "reasoning": "测试动态止盈：目标收益5%，最大10%，24小时内自动调整。⚠️ 需要先有持仓才能执行"
  }
]`);
          break;
        case 'oco_order':
          const cpOco = currentPrices[symbol] || 43560.42;
          setBatchCommands(`[
  {
    "symbol": "${symbol}",
    "action": "oco_order",
    "stop_loss": ${(cpOco * 0.97).toFixed(2)},
    "take_profit": ${(cpOco * 1.05).toFixed(2)},
    "confidence": 85,
    "reasoning": "测试OCO订单：止损-3%，止盈+5%。⚠️ 需要先有持仓才能执行（新开仓不支持OCO）"
  }
]`);
          break;
        case 'bracket_order':
          const cpBracket = currentPrices[symbol] || 43560.42;
          setBatchCommands(`[
  {
    "symbol": "${symbol}",
    "action": "bracket_order",
    "leverage": 3,
    "position_size_usd": 200,
    "stop_loss": ${(cpBracket * 0.98).toFixed(2)},
    "take_profit": ${(cpBracket * 1.06).toFixed(2)},
    "confidence": 85,
    "reasoning": "测试括号订单：开仓${symbol}，3倍杠杆，100 USDT，止损-2%，止盈+6%，分步执行"
  }
]`);
          break;
        case 'add_to_position':
          setBatchCommands(`[
  {
    "symbol": "${symbol}",
    "action": "add_to_position",
    "additional_position_size_usd": 50,
    "add_position_type": "long",
    "confidence": 80,
    "reasoning": "测试加仓：向多仓追加50 USDT。⚠️ 仅在已有盈利仓位时使用，永远不要追亏损！"
  }
]`);
          break;
      }
    } else {
      setBatchCommands(templateOrType);
    }
  };

  return (
    <div className="min-h-screen" style={{ background: '#0B0E11', color: '#EAECEF' }}>
      <div className="max-w-6xl mx-auto px-4 py-8">
        <div className="rounded-lg p-6" style={{ background: '#181A20', border: '1px solid #2B3139' }}>
          <h1 className="text-2xl font-bold mb-6">{t('test_page_title')}</h1>
          
          <div className="mb-6 border-b border-gray-700">
            <nav className="flex space-x-2">
              <button
                className={`px-4 py-2 font-medium text-sm rounded-t-lg transition-colors ${activeTab === 'batch-trade' ? 'text-blue-400 border-b-2 border-blue-400 bg-gray-800' : 'text-gray-400 hover:text-gray-200'}`}
                onClick={() => setActiveTab('batch-trade')}
                style={activeTab === 'batch-trade' ? { color: '#93C5FD', borderBottom: '2px solid #93C5FD', background: '#2D3748' } : {}}
              >
                {t('batch_trade_tab')}
              </button>
              <button
                className={`px-4 py-2 font-medium text-sm rounded-t-lg transition-colors ${activeTab === 'api-test' ? 'text-blue-400 border-b-2 border-blue-400 bg-gray-800' : 'text-gray-400 hover:text-gray-200'}`}
                onClick={() => setActiveTab('api-test')}
                style={activeTab === 'api-test' ? { color: '#93C5FD', borderBottom: '2px solid #93C5FD', background: '#2D3748' } : {}}
              >
                {t('api_test_tab')}
              </button>
              <button
                className={`px-4 py-2 font-medium text-sm rounded-t-lg transition-colors ${activeTab === 'test-scripts' ? 'text-blue-400 border-b-2 border-blue-400 bg-gray-800' : 'text-gray-400 hover:text-gray-200'}`}
                onClick={() => setActiveTab('test-scripts')}
                style={activeTab === 'test-scripts' ? { color: '#93C5FD', borderBottom: '2px solid #93C5FD', background: '#2D3748' } : {}}
              >
                {t('test_scripts_tab')}
              </button>
              <button
                className={`px-4 py-2 font-medium text-sm rounded-t-lg transition-colors ${activeTab === 'ai-prompt' ? 'text-blue-400 border-b-2 border-blue-400 bg-gray-800' : 'text-gray-400 hover:text-gray-200'}`}
                onClick={() => setActiveTab('ai-prompt')}
                style={activeTab === 'ai-prompt' ? { color: '#93C5FD', borderBottom: '2px solid #93C5FD', background: '#2D3748' } : {}}
              >
                📄 AI Prompt
              </button>
            </nav>
          </div>
          
          <div className="tab-content">
            {activeTab === 'api-test' && (
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium mb-1">{t('api_key_label')}:</label>
                  <input type="password" name="apiKey" value={apiCredentials.apiKey} onChange={handleInputChange} className="w-full px-3 py-2 rounded-md bg-[#2B3139] border border-[#3D444D]" />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">{t('secret_key_label')}:</label>
                  <input type="password" name="secretKey" value={apiCredentials.secretKey} onChange={handleInputChange} className="w-full px-3 py-2 rounded-md bg-[#2B3139] border border-[#3D444D]" />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">{t('api_url_label')}:</label>
                  <input type="text" name="apiUrl" value={apiCredentials.apiUrl} onChange={handleInputChange} className="w-full px-3 py-2 rounded-md bg-[#2B3139] border border-[#3D444D]" />
                </div>
                <div className="flex gap-2">
                  <button onClick={() => runTest()} disabled={loading} className="px-4 py-2 rounded-md bg-[#3182CE] text-white hover:bg-[#2B6CB0] disabled:bg-gray-600 transition-colors">
                    {loading ? t('running_test') : t('run_test')}
                  </button>
                  <button onClick={() => setTestResults('')} className="px-4 py-2 rounded-md border border-[#4A5568] text-gray-300 hover:bg-[#2D3748] transition-colors">
                    🗑️ {t('clear')}
                  </button>
                </div>
                {testResults && (
                  <div className="mt-6">
                    <h2 className="text-lg font-semibold mb-2">{t('test_results')}:</h2>
                    <pre className="p-4 rounded-md text-sm overflow-auto max-h-96 font-mono bg-[#1E293B] border border-[#334155]">{testResults}</pre>
                  </div>
                )}
              </div>
            )}
            
            {activeTab === 'batch-trade' && (
              <div>
                <div className="mb-4 flex flex-col md:flex-row gap-4">
                  <div className="flex-1">
                    <label className="block text-sm font-medium mb-1">{t('select_trader_label')}:</label>
                    <div className="flex gap-2">
                      <select value={selectedTrader} onChange={(e) => setSelectedTrader(e.target.value)} className="flex-1 px-3 py-2 rounded-md bg-[#2B3139] border border-[#3D444D]">
                        <option value="">{t('please_select_trader')}</option>
                        {traders.map((trader) => (
                          <option key={trader.trader_id} value={trader.trader_id}>
                            {trader.trader_name} ({trader.trader_id.slice(0, 8)}) {trader.is_running ? '🟢' : '🔴'}
                          </option>
                        ))}
                      </select>
                      <button onClick={fetchTraders} className="px-3 py-2 rounded-md bg-[#2B3139] border border-[#3D444D] hover:bg-[#3D444D]" title={t('refresh_traders')}>
                        🔄
                      </button>
                    </div>
                  </div>
                  
                  <div className="flex items-end gap-2">
                    <button onClick={handleStartTrader} disabled={traderActionLoading || !selectedTrader} className="px-4 py-2 rounded-md bg-[#38A169] text-white hover:bg-[#2F855A] disabled:bg-gray-600 transition-colors flex items-center gap-1">
                      {traderActionLoading ? '⏳' : '▶️'} {t('start_trader')}
                    </button>
                    <button onClick={handleStopTrader} disabled={traderActionLoading || !selectedTrader} className="px-4 py-2 rounded-md bg-[#E53E3E] text-white hover:bg-[#C53030] disabled:bg-gray-600 transition-colors flex items-center gap-1">
                      {traderActionLoading ? '⏳' : '⏹'} {t('stop_trader')}
                    </button>
                  </div>
                </div>
                
                <div className="mb-6">
                  <div className="flex justify-between items-center mb-4">
                    <h3 className="text-lg font-semibold">{t('trade_action_templates')}</h3>
                    <div className="flex items-center gap-2">
                      <span className="text-sm text-gray-400">{t('current_btc_price')}{(currentPrices['BTCUSDT'] || 43560.42).toFixed(2)} USDT</span>
                      <button onClick={loadCurrentPriceTemplates} disabled={countdown > 0} className={`px-3 py-1 rounded text-sm ${countdown > 0 ? 'bg-[#718096]' : 'bg-[#3182CE]'} text-white`}>
                        {isFetchingPrice ? t('fetching') : countdown > 0 ? `${countdown}${t('seconds_remaining')}` : t('get_current_price')}
                      </button>
                    </div>
                  </div>
                  
                  <div className="flex flex-col lg:flex-row gap-6">
                    <div className="lg:w-1/2">
                      <div className="space-y-4">
                        <div>
                          <h4 className="font-medium mb-2 text-[#E2E8F0]">{t('long_positions')}</h4>
                          <div className="grid grid-cols-2 md:grid-cols-3 gap-1">
                            {tradeTemplates.filter(v => v.category === 'long').map(v => (
                              <button key={v.id} onClick={() => insertTemplate(v.template)} className="bg-[#38A169] text-white p-1 text-[10px] text-left rounded hover:bg-[#2F855A] transition-colors h-10 overflow-hidden">
                                {v.name}
                              </button>
                            ))}
                          </div>
                        </div>
                        <div>
                          <h4 className="font-medium mb-2 text-[#E2E8F0]">{t('short_positions')}</h4>
                          <div className="grid grid-cols-2 md:grid-cols-3 gap-1">
                            {tradeTemplates.filter(v => v.category === 'short').map(v => (
                              <button key={v.id} onClick={() => insertTemplate(v.template)} className="bg-[#E53E3E] text-white p-1 text-[10px] text-left rounded hover:bg-[#C53030] transition-colors h-10 overflow-hidden">
                                {v.name}
                              </button>
                            ))}
                          </div>
                        </div>
                        <div>
                          <h4 className="font-medium mb-2 text-[#E2E8F0]">{t('general_operations')}</h4>
                          <div className="grid grid-cols-2 md:grid-cols-3 gap-1">
                            {tradeTemplates.filter(v => v.category === 'general').map(v => (
                              <button key={v.id} onClick={() => insertTemplate(v.template)} className="bg-[#4A5568] text-white p-1 text-[10px] text-left rounded hover:bg-[#2D3748] transition-colors h-10 overflow-hidden">
                                {v.name}
                              </button>
                            ))}
                          </div>
                        </div>
                        <div>
                          <h4 className="font-medium mb-2 text-[#E2E8F0]">{t('advanced_operations')}</h4>
                          <div className="grid grid-cols-2 md:grid-cols-3 gap-1">
                            {tradeTemplates.filter(v => v.category === 'advanced').map(v => (
                              <button key={v.id} onClick={() => insertTemplate(v.template)} className="bg-[#805AD5] text-white p-1 text-[10px] text-left rounded hover:bg-[#6B46C1] transition-colors h-10 overflow-hidden">
                                {v.name}
                              </button>
                            ))}
                          </div>
                        </div>
                      </div>
                    </div>
                    
                    <div className="lg:w-1/2">
                      <label className="block text-sm font-medium mb-1">{t('batch_commands_label')}</label>
                      <textarea value={batchCommands} onChange={(e) => setBatchCommands(e.target.value)} rows={12} className="w-full px-3 py-2 rounded-md bg-[#2B3139] border border-[#3D444D] font-mono text-xs" />
                      <div className="grid grid-cols-2 md:grid-cols-3 gap-2 mt-4">
                        <button onClick={runBatchCommand} disabled={batchLoading || !selectedTrader} className="px-4 py-2 rounded-md bg-[#3182CE] text-white hover:bg-[#2B6CB0] disabled:bg-gray-600 transition-colors">
                          {batchLoading ? t('executing') : t('execute_batch')}
                        </button>
                        <button onClick={handleManualScan} disabled={manualScanLoading || !selectedTrader} className="px-4 py-2 rounded-md bg-[#805AD5] text-white hover:bg-[#6B46C1] disabled:bg-gray-600 transition-colors flex items-center justify-center gap-2">
                          {manualScanLoading ? (
                            <>
                              <span className="animate-spin text-lg">⏳</span>
                              {t('scanning')}
                            </>
                          ) : (
                            <>
                              <span>🔍</span>
                              {t('manual_scan')}
                            </>
                          )}
                        </button>
                        <button onClick={handleSyncBalance} disabled={syncBalanceLoading || !selectedTrader} className="px-4 py-2 rounded-md bg-[#319795] text-white hover:bg-[#285E61] disabled:bg-gray-600 transition-colors flex items-center justify-center gap-2">
                          {syncBalanceLoading ? (
                            <>
                              <span className="animate-spin text-lg">🔄</span>
                              {t('syncing')}
                            </>
                          ) : (
                            <>
                              <span>💎</span>
                              {t('sync_balance')}
                            </>
                          )}
                        </button>
                        <button onClick={handleCloseAllPositions} disabled={traderActionLoading || !selectedTrader} className="px-4 py-2 rounded-md bg-[#D69E2E] text-white hover:bg-[#B7791F] disabled:bg-gray-600 transition-colors flex items-center justify-center gap-2">
                          {traderActionLoading ? '⏳' : '🚫'} {t('close_all_positions')}
                        </button>
                        <button onClick={() => setBatchCommands('[]')} className="px-4 py-2 rounded-md border border-[#4A5568] text-gray-300 hover:bg-[#2D3748] transition-colors">
                          🗑️ {t('clear')}
                        </button>
                      </div>
                      
                      {batchResults && (
                        <div className="mt-6">
                          <h2 className="text-lg font-semibold mb-2">{t('batch_execution_results')}:</h2>
                          <pre className="p-4 rounded-md text-sm overflow-auto max-h-96 font-mono bg-[#1E293B] border border-[#334155]">{batchResults}</pre>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            )}

            {activeTab === 'test-scripts' && (
              <div className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div className="bg-gray-800/30 p-4 rounded-xl border border-white/5">
                    <div className="flex justify-between items-center mb-4">
                      <h3 className="text-lg font-bold text-nofx-gold">{t('available_test_scripts')}</h3>
                      <button onClick={fetchTestScripts} className="text-xs text-blue-400 hover:underline flex items-center gap-1">
                        🔄 {t('refresh_list')}
                      </button>
                    </div>
                    {fetchingScripts ? <p className="text-gray-400">{t('loading')}...</p> : (
                      <div className="space-y-2 max-h-[400px] overflow-y-auto pr-2">
                        {testScripts.length > 0 ? testScripts.map(script => (
                          <div key={script} className="flex items-center justify-between p-2 rounded-lg bg-gray-900/50 border border-gray-700 hover:border-gray-500 transition-colors">
                            <span className="font-mono text-[10px] truncate flex-1 mr-2">{script}</span>
                            <button onClick={() => runTest(`go run test/${script}`)} disabled={loading} className="px-2 py-1 rounded text-[10px] bg-nofx-gold text-black font-bold hover:bg-yellow-500 transition-colors whitespace-nowrap">
                              {loading ? t('running') : t('run')}
                            </button>
                          </div>
                        )) : <p className="text-gray-500">{t('no_scripts_found')}</p>}
                      </div>
                    )}
                  </div>
                  <div className="bg-gray-800/30 p-4 rounded-xl border border-white/5">
                    <div className="flex justify-between items-center mb-4">
                      <h3 className="text-lg font-bold text-nofx-gold">{t('execution_output')}</h3>
                      <button onClick={() => setTestResults('')} className="text-xs text-gray-400 hover:text-white flex items-center gap-1">
                        🗑️ {t('clear')}
                      </button>
                    </div>
                    <pre className="p-4 rounded-md text-[10px] overflow-auto h-[400px] font-mono bg-[#1E293B] border border-[#334155]">{testResults || t('no_output_yet')}</pre>
                  </div>
                </div>
              </div>
            )}

            {activeTab === 'ai-prompt' && (
              <div className="space-y-4">
                <div className="flex justify-between items-center mb-4">
                  <div>
                    <h2 className="text-xl font-bold">📄 最近一次 AI Prompt</h2>
                    <p className="text-sm text-gray-400 mt-1">
                      查看系统实际发送给 AI 的完整提示词文档（包含K线数据、技术指标、持仓信息等）
                    </p>
                  </div>
                  <button 
                    onClick={fetchLastPrompt} 
                    disabled={promptLoading}
                    className="px-4 py-2 rounded-md bg-[#3182CE] text-white hover:bg-[#2B6CB0] disabled:bg-gray-600 transition-colors flex items-center gap-2"
                  >
                    {promptLoading ? (
                      <>
                        <span className="animate-spin">🔄</span>
                        加载中...
                      </>
                    ) : (
                      <>
                        🔄 刷新
                      </>
                    )}
                  </button>
                </div>

                {promptError && (
                  <div className="p-4 rounded-lg bg-red-900/30 border border-red-700 text-red-200">
                    <p className="font-semibold">❌ 错误</p>
                    <p className="text-sm mt-1">{promptError}</p>
                  </div>
                )}

                {promptData && (
                  <div className="space-y-4">
                    {/* 数据说明 */}
                    <div className="p-4 rounded-lg bg-indigo-900/30 border border-indigo-700">
                      <p className="text-sm text-indigo-200 font-semibold mb-2">📌 此文档包含的数据：</p>
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-2 text-xs text-indigo-100">
                        <div>✅ 完整K线数据表格（OHLCV + 时间戳）</div>
                        <div>✅ 技术指标（EMA、MACD、RSI、BOLL等）</div>
                        <div>✅ 当前持仓详情（入场价、盈亏、止盈止损）</div>
                        <div>✅ 账户余额和可用保证金</div>
                        <div>✅ 候选币种市场数据</div>
                        <div>✅ 开仓持仓排名、资金流向</div>
                      </div>
                      <p className="text-xs text-indigo-300 mt-2">
                        💡 这是调用 <code className="bg-indigo-800 px-1 rounded">CallWithMessages()</code> 前一刻的完整文档，与AI实际接收的内容100%一致
                      </p>
                    </div>

                    {/* 统计信息卡片 */}
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                      <div className="p-3 rounded-lg bg-blue-900/30 border border-blue-700">
                        <p className="text-xs text-blue-300 mb-1">⏰ 生成时间</p>
                        <p className="text-sm font-mono text-blue-100">
                          {new Date(promptData.timestamp).toLocaleString('zh-CN', { 
                            year: 'numeric', 
                            month: '2-digit', 
                            day: '2-digit', 
                            hour: '2-digit', 
                            minute: '2-digit', 
                            second: '2-digit' 
                          })}
                        </p>
                      </div>

                      <div className="p-3 rounded-lg bg-purple-900/30 border border-purple-700">
                        <p className="text-xs text-purple-300 mb-1">📊 数据统计</p>
                        <div className="text-sm font-mono text-purple-100 space-y-1">
                          <p>总字节: {promptData.total_bytes.toLocaleString()}</p>
                          <p>总字符: {promptData.total_chars.toLocaleString()}</p>
                          <p>估算Tokens: {promptData.estimated_tokens.toLocaleString()}</p>
                        </div>
                      </div>

                      <div className="p-3 rounded-lg bg-green-900/30 border border-green-700">
                        <p className="text-xs text-green-300 mb-1">💰 预估费用</p>
                        <p className="text-sm font-mono text-green-100">
                          {formatCost(calculateCost(promptData.estimated_tokens))}
                        </p>
                        <p className="text-xs text-green-400 mt-1">
                          (仅输入token，输出需额外计算)
                        </p>
                      </div>
                    </div>

                    {/* System Prompt */}
                    <div className="bg-gray-800/30 p-4 rounded-xl border border-white/5">
                      <div className="flex justify-between items-center mb-3">
                        <div className="flex items-center gap-3">
                          <h3 className="text-lg font-bold text-green-400">🤖 System Prompt</h3>
                          <span className="text-xs text-gray-400 font-mono">
                            {promptData.system_bytes.toLocaleString()} bytes / {promptData.system_chars.toLocaleString()} chars
                          </span>
                        </div>
                        <button 
                          onClick={() => copyToClipboard(promptData.system_prompt, 'System Prompt')}
                          className="px-3 py-1 text-xs rounded bg-gray-700 hover:bg-gray-600 transition-colors flex items-center gap-1"
                          title="复制 System Prompt"
                        >
                          📋 复制
                        </button>
                      </div>
                      <pre className="p-4 rounded-md text-xs overflow-auto max-h-96 font-mono bg-[#1E293B] border border-[#334155] whitespace-pre-wrap">
                        {promptData.system_prompt}
                      </pre>
                    </div>

                    {/* User Prompt */}
                    <div className="bg-gray-800/30 p-4 rounded-xl border border-white/5">
                      <div className="flex justify-between items-center mb-3">
                        <div className="flex items-center gap-3">
                          <h3 className="text-lg font-bold text-yellow-400">💬 User Prompt</h3>
                          <span className="text-xs text-gray-400 font-mono">
                            {promptData.user_bytes.toLocaleString()} bytes / {promptData.user_chars.toLocaleString()} chars
                          </span>
                        </div>
                        <button 
                          onClick={() => copyToClipboard(promptData.user_prompt, 'User Prompt')}
                          className="px-3 py-1 text-xs rounded bg-gray-700 hover:bg-gray-600 transition-colors flex items-center gap-1"
                          title="复制 User Prompt"
                        >
                          📋 复制
                        </button>
                      </div>
                      <pre className="p-4 rounded-md text-xs overflow-auto max-h-96 font-mono bg-[#1E293B] border border-[#334155] whitespace-pre-wrap">
                        {promptData.user_prompt}
                      </pre>
                    </div>

                    {/* 全文复制按钮 */}
                    <div className="flex justify-center">
                      <button 
                        onClick={() => copyToClipboard(
                          `=== System Prompt ===

${promptData.system_prompt}

=== User Prompt ===

${promptData.user_prompt}`,
                          '完整 Prompt'
                        )}
                        className="px-6 py-2 rounded-md bg-[#805AD5] text-white hover:bg-[#6B46C1] transition-colors flex items-center gap-2"
                      >
                        📋 复制完整 Prompt (System + User)
                      </button>
                    </div>

                    {/* 技术验证区域（可折叠） */}
                    {promptData.actual_request_json && (
                      <div className="border-t border-gray-700 pt-4">
                        <button
                          onClick={() => setShowTechnicalDetails(!showTechnicalDetails)}
                          className="w-full flex items-center justify-between p-3 rounded-lg bg-gray-800/50 hover:bg-gray-700/50 transition-colors"
                        >
                          <span className="text-sm font-semibold text-gray-300">
                            🔧 技术验证：查看实际发送给AI的JSON请求体
                          </span>
                          <span className="text-gray-400">{showTechnicalDetails ? '▼' : '▶'}</span>
                        </button>
                        
                        {showTechnicalDetails && (
                          <div className="mt-4 p-4 rounded-lg bg-gray-900 border border-gray-700">
                            <p className="text-xs text-yellow-300 mb-3">
                              ⚠️ 这是mcp.Client.buildMCPRequestBody()生成的JSON结构，与实际HTTP请求完全一致。
                              您可以对比 messages[0].content (system) 和 messages[1].content (user) 确认内容无修改。
                            </p>
                            <pre className="p-4 rounded-md text-xs overflow-auto max-h-96 font-mono bg-black text-green-400 border border-green-700">
                              {promptData.actual_request_json}
                            </pre>
                            <button 
                              onClick={() => copyToClipboard(promptData.actual_request_json || '', 'JSON 请求体')}
                              className="mt-3 px-3 py-1 text-xs rounded bg-green-700 hover:bg-green-600 transition-colors"
                            >
                              📋 复制 JSON 请求体
                            </button>
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                )}

                {!promptData && !promptError && !promptLoading && (
                  <div className="text-center py-12 text-gray-400">
                    <p className="text-lg mb-2">📭 暂无 Prompt 数据</p>
                    <p className="text-sm">请先执行一次手动扫描，或等待自动决策周期执行</p>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default TestPage;
