import { useState, useEffect } from 'react'
import type { DecisionRecord, DecisionAction } from '../types'
import { t, type Language } from '../i18n/translations'

// Add custom CSS for animations
const customStyles = `
  @keyframes pulse-once {
    0% { transform: scale(1); }
    50% { transform: scale(1.05); box-shadow: 0 0 20px rgba(14, 203, 129, 0.6); }
    100% { transform: scale(1); }
  }
  
  .animate-pulse-once {
    animation: pulse-once 0.5s ease-in-out;
  }
  
  .animate-pulse {
    animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  }
`;

// Inject styles
if (typeof document !== 'undefined') {
  const styleSheet = document.createElement('style');
  styleSheet.textContent = customStyles;
  document.head.appendChild(styleSheet);
}

interface DecisionCardProps {
  decision: DecisionRecord
  language: Language
  onSymbolClick?: (symbol: string) => void
  onDelete?: (decisionId: number, traderId: string) => void
  exchangeType?: string;
  exchangeCustomUrl?: string;
  exchanges?: Array<{
    id: string;
    exchange_type: string;
    customApiUrl?: string;
    name: string;
    enabled: boolean;
    testnet?: boolean;  // 添加testnet标识
  }>;
  traders?: Array<{
    trader_id: string;
    exchange_id?: string;
    [key: string]: any;
  }>;
  traderId?: string;
  expandedState?: {
    showSystemPrompt: boolean;
    showInputPrompt: boolean;
    showCoT: boolean;
    showJSON: boolean;
  };
  onUpdateExpansion?: (newState: Partial<{ showSystemPrompt: boolean, showInputPrompt: boolean, showCoT: boolean, showJSON: boolean }>) => void;
}

// Action type configuration
const ACTION_CONFIG = (language: Language): Record<string, { color: string; bg: string; icon: string; label: string }> => ({
  open_long: { color: '#0ECB81', bg: 'rgba(14, 203, 129, 0.15)', icon: '📈', label: t('openLong', language) },
  open_short: { color: '#F6465D', bg: 'rgba(246, 70, 93, 0.15)', icon: '📉', label: t('openShort', language) },
  close_long: { color: '#F0B90B', bg: 'rgba(240, 185, 11, 0.15)', icon: '💰', label: t('closeLong', language) },
  close_short: { color: '#F0B90B', bg: 'rgba(240, 185, 11, 0.15)', icon: '💰', label: t('closeShort', language) },
  hold: { color: '#848E9C', bg: 'rgba(132, 142, 156, 0.15)', icon: '⏸️', label: t('hold', language) },
  wait: { color: '#848E9C', bg: 'rgba(132, 142, 156, 0.15)', icon: '⏳', label: t('wait', language) },
  partial_close: { color: '#9C27B0', bg: 'rgba(156, 39, 176, 0.15)', icon: '📉', label: t('partialClose', language) },
  update_stop_loss: { color: '#FF9800', bg: 'rgba(255, 152, 0, 0.15)', icon: '🛡️', label: t('updateStopLoss', language) },
  update_take_profit: { color: '#4CAF50', bg: 'rgba(76, 175, 80, 0.15)', icon: '🎯', label: t('updateTakeProfit', language) },
  trailing_stop: { color: '#2196F3', bg: 'rgba(33, 150, 243, 0.15)', icon: '👣', label: t('trailingStop', language) },
  dynamic_take_profit: { color: '#00BCD4', bg: 'rgba(0, 188, 212, 0.15)', icon: '🚀', label: t('dynamicTakeProfit', language) },
  add_position: { color: '#795548', bg: 'rgba(121, 85, 72, 0.15)', icon: '➕', label: t('addPosition', language) },
  full_close: { color: '#607D8B', bg: 'rgba(96, 125, 139, 0.15)', icon: '🔚', label: t('fullClose', language) },
});

// Format price with proper decimals
function formatPrice(price: number | undefined): string {
  if (price === undefined || Number.isNaN(price)) return '-'
  if (price === 0) return '0'
  if (price >= 1000) return price.toFixed(2)
  if (price >= 1) return price.toFixed(4)
  return price.toFixed(6)
}

// Calculate percentage change
function calcPctChange(entry: number | undefined, target: number | undefined, isLong: boolean): string {
  if (!entry || !target || entry === 0) return '-'
  const pct = ((target - entry) / entry) * 100
  const adjustedPct = isLong ? pct : -pct
  return `${adjustedPct >= 0 ? '+' : ''}${adjustedPct.toFixed(2)}%`
}

// Get confidence color
function getConfidenceColor(confidence: number | undefined): string {
  if (!confidence) return '#848E9C'
  if (confidence >= 80) return '#0ECB81'
  if (confidence >= 60) return '#F0B90B'
  return '#F6465D'
}

// Helper function to open exchange link for a symbol
function openExchangeLink(symbol: string, exchangeType?: string, customUrl?: string, language?: string) {
  let exchangeUrl = '';
  
  // Determine language code for URL (default to 'en' if not specified or unsupported)
  const langCode = language === 'zh' ? 'zh-CN' : 'en';
  
  // If a custom URL is provided, use it as the base URL
  // But for binance_demo, always use the demo trading interface URL
  if (exchangeType?.toLowerCase() === 'binance_demo') {
    // Binance demo/virtual trading platform
    exchangeUrl = `https://demo.binance.com/${langCode}/futures/${symbol}`;
  } else if (customUrl) {
    // Check if the custom URL is a testnet URL by looking for testnet indicators
    const isTestnet = customUrl.toLowerCase().includes('test') || 
                      customUrl.toLowerCase().includes('sandbox') || 
                      customUrl.toLowerCase().includes('demo') ||
                      customUrl.toLowerCase().includes('futures-test') ||
                      customUrl.toLowerCase().includes('testnet');
      
    if (customUrl.includes('{symbol}')) {
      // Replace placeholder in custom URL if present
      exchangeUrl = customUrl.replace('{symbol}', symbol);
    } else if (customUrl.includes('{futures_symbol}')) {
      // Handle futures-specific symbol format (e.g. BTCUSDT for Binance Futures)
      exchangeUrl = customUrl.replace('{futures_symbol}', symbol);
    } else {
      // For known testnet URLs, use appropriate format
      if (isTestnet) {
        // For testnet environments, we need to handle different URL structures
        if (customUrl.includes('binance')) {
          // Binance testnet - use standard format
          exchangeUrl = `https://testnet.binancefuture.com/${langCode}/futures/${symbol}`;
        } else {
          // For other exchange testnets, append symbol to custom URL
          exchangeUrl = customUrl.endsWith('/') ? `${customUrl}${symbol}` : `${customUrl}/${symbol}`;
        }
      } else {
        // For non-testnet, append symbol to custom URL
        exchangeUrl = customUrl.endsWith('/') ? `${customUrl}${symbol}` : `${customUrl}/${symbol}`;
      }
    }
  } else {
    // Determine exchange URL based on exchange type
    
    // Check if we can infer testnet from other contextual clues
    const isLikelyTestnet = false; // We can't determine this without explicit config
    
    switch (exchangeType?.toLowerCase()) {
      case 'binance_demo':
        exchangeUrl = `https://demo.binance.com/${langCode}/futures/${symbol}`;
        break;
      case 'binance':
        exchangeUrl = `https://www.binance.com/${langCode}/futures/${symbol}`;
        break;
      case 'bybit':
        exchangeUrl = isLikelyTestnet 
          ? `https://testnet.bybit.com/trade/futures/${symbol.replace('USDT', '')}-USDT?l=${langCode}`
          : `https://www.bybit.com/${langCode}/trade/futures/${symbol.replace('USDT', '')}-USDT`;
        break;
      case 'okx':
        exchangeUrl = isLikelyTestnet 
          ? `https://www.okx.com/${langCode}/trade-futures-demo/${symbol.toLowerCase()}`
          : `https://www.okx.com/${langCode}/trade-futures/${symbol.toLowerCase()}`;
        break;
      case 'bitget':
        exchangeUrl = isLikelyTestnet 
          ? `https://www.bitget.com/futures/${symbol.replace('USDT', '')}_USDT?lng=${langCode}`
          : `https://www.bitget.com/futures/${symbol.replace('USDT', '')}_USDT?lng=${langCode}`;
        break;
      case 'hyperliquid':
        exchangeUrl = isLikelyTestnet 
          ? `https://app.hyperliquid.xyz/demo#/${symbol.replace('USDT', '')}?lang=${langCode}`
          : `https://app.hyperliquid.xyz/trade#${symbol.replace('USDT', '')}?lang=${langCode}`;
        break;
      case 'aster':
        exchangeUrl = isLikelyTestnet 
          ? `https://test.aster-trade.com/${langCode}/market/${symbol}`
          : `https://aster-trade.com/${langCode}/market/${symbol}`;
        break;
      case 'lighter':
        exchangeUrl = isLikelyTestnet 
          ? `https://test.lighter.trade/${langCode}/markets/${symbol}`
          : `https://lighter.trade/${langCode}/markets/${symbol}`;
        break;
      default:
        // Default to Binance futures mainnet as it's the most common exchange in the codebase
        exchangeUrl = `https://www.binance.com/${langCode}/futures/${symbol}`;
    }
  }
  
  window.open(exchangeUrl, '_blank', 'noopener,noreferrer');
}

// Helper function to get exchange info from exchanges list by traderId
function getExchangeInfoByTraderId(
  traderId: string,
  exchanges: Array<{
    id: string;
    exchange_type: string;
    customApiUrl?: string;
    name: string;
    enabled: boolean;
  }> | undefined,
  traders: Array<{
    trader_id: string;
    exchange_id?: string;
    [key: string]: any;
  }> | undefined
): { exchangeType?: string; customUrl?: string } {
  console.log('🔄 getExchangeInfoByTraderId调用详情:');
  console.log('traderId:', traderId);
  console.log('exchanges参数:', exchanges);
  console.log('traders参数:', traders);

  //增强参数验证
  if (!exchanges) {
    console.error('❌ exchanges参数为空');
    return {};
  }
  
  if (!traders) {
    console.error('❌ traders参数为空');
    return {};
  }
  
  if (!Array.isArray(traders)) {
    console.error('❌ traders不是数组类型:', typeof traders);
    return {};
  }
  
  if (traders.length === 0) {
    console.warn('⚠️ traders数组为空');
    return {};
  }

  // First, find the trader to get exchange_id
  console.log('🔍 查找trader...');
  const trader = traders.find(t => t.trader_id === traderId);
  console.log('找到的trader:', trader);

  if (!trader) {
    console.log('❌ 未找到匹配的trader');
    console.log('可用的trader_id列表:', traders.map(t => t.trader_id));
    return {};
  }

  // Then find the exchange using exchange_id
  console.log('🔍 查找exchange...');
  console.log('要查找的exchange_id:', trader.exchange_id);
  const exchange = exchanges.find(ex => ex.id === trader.exchange_id);
  console.log('找到的exchange:', exchange);

  if (exchange) {
    console.log('✅ 成功找到exchange信息');
    console.log('exchange_type:', exchange.exchange_type);
    console.log('customApiUrl:', exchange.customApiUrl);
    const result = {
      exchangeType: exchange.exchange_type,
      customUrl: exchange.customApiUrl
    };
    console.log('返回结果:', result);
    return result;
  }

  // If direct match not found, return empty object
  console.log('❌ 未找到匹配的exchange');
  console.log('可用的exchange id列表:', exchanges.map(e => ({id: e.id, type: e.exchange_type})));
  return {};
}

// 移除：渲染数值变化信息（阶段1）
// 此函数已被弃用，所有数值变化信息现在都在Trading Details Grid中统一显示

// Single Action Card Component
function ActionCard({ action, language, onSymbolClick, positions, exchangeType, exchangeCustomUrl, exchanges, traderId, traders }: { action: DecisionAction; language: Language; onSymbolClick?: (symbol: string) => void; positions?: any[]; exchangeType?: string; exchangeCustomUrl?: string; exchanges?: Array<{ id: string; exchange_type: string; customApiUrl?: string; name: string; enabled: boolean; }>; traderId?: string; traders?: Array<{ trader_id: string; exchange_id?: string; [key: string]: any; }> }) {
  const actionConfigs = ACTION_CONFIG(language);
  const config = actionConfigs[action.action] || actionConfigs.wait
  const isLong = action.action.includes('long')
  const isOpen = action.action.includes('open')
  const isClose = action.action.includes('close')
  const isPartialClose = action.action === 'partial_close'
  const isTrailingStop = action.action === 'trailing_stop'
  const isUpdateStopLoss = action.action === 'update_stop_loss'
  const isHold = action.action === 'hold'
  const isWait = action.action === 'wait'

  return (
    <div
      className="rounded-lg p-4 transition-all duration-200 hover:scale-[1.01]"
      style={{
        background: 'linear-gradient(135deg, #1E2329 0%, #181C21 100%)',
        border: `1px solid ${config.color}33`,
        boxShadow: `0 4px 12px rgba(0, 0, 0, 0.2), inset 0 1px 0 rgba(255, 255, 255, 0.03)`,
      }}
    >
      {/* Header Row */}
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-3">
          <span 
            className="text-xl cursor-pointer transition-all duration-200 hover:scale-110" 
            onClick={() => {
              //检查是否有直接传递的交易所信息
              if (exchangeType || exchangeCustomUrl) {
                openExchangeLink(action.symbol, exchangeType, exchangeCustomUrl, language);
                return;
              }
              
              //检查是否有通过traderId获取交易所信息的必要数据
              if (exchanges && traders && traderId && Array.isArray(traders) && traders.length > 0) {
                const exchangeInfo = getExchangeInfoByTraderId(traderId, exchanges, traders);
                openExchangeLink(action.symbol, exchangeInfo.exchangeType, exchangeInfo.customUrl, language);
                return;
              }
              
              // 最后的后备方案：使用traderId模式匹配
              const isDemoTrader = traderId?.includes('demo') || traderId?.includes('test') || traderId?.includes('虚拟');
              const fallbackExchangeType = isDemoTrader ? 'binance_demo' : 'binance';
              openExchangeLink(action.symbol, fallbackExchangeType, undefined, language);
            }}
            title={`Click to open ${action.symbol} on exchange`}
            style={{ transformOrigin: 'center' }}
          >
            {config.icon}
          </span>
          <span
            className="font-mono font-bold text-lg cursor-pointer transition-all duration-200 hover:scale-110"
            style={{ color: '#EAECEF' }}
            onClick={() => onSymbolClick?.(action.symbol)}
            title="Click to view chart"
          >
            {action.symbol.replace('USDT', '')}
          </span>
          <span
            className="px-3 py-1 rounded-full text-xs font-bold uppercase tracking-wider"
            style={{ background: config.bg, color: config.color, border: `1px solid ${config.color}55` }}
          >
            {config.label}
          </span>
        </div>

        {/* Status Badge */}
        <div className="flex items-center gap-2">
          {action.confidence !== undefined && action.confidence > 0 && (
            <div
              className="px-2 py-1 rounded text-xs font-semibold"
              style={{
                background: `${getConfidenceColor(action.confidence)}22`,
                color: getConfidenceColor(action.confidence)
              }}
            >
              {action.confidence.toFixed(0)}%
            </div>
          )}
          <div
            className="w-2 h-2 rounded-full"
            style={{ background: action.success ? '#0ECB81' : '#F6465D' }}
          />
        </div>
      </div>

      {/* Trading Details Grid */}
      {(isOpen || isClose || isPartialClose || isTrailingStop || isUpdateStopLoss || isHold || isWait) && (
        <div className="grid grid-cols-4 gap-3 mt-3 pt-3" style={{ borderTop: '1px solid #2B3139' }}>
          {/* Entry Price or Current Price or P&L Amount */}
          <div className="text-center">
            <div className="text-xs mb-1" style={{ color: '#848E9C' }}>
              {isOpen ? t('entryPrice', language) : isClose ? t('profitAmount', language) : isHold || isWait ? t('currentPrice', language) : t('currentPrice', language)}
            </div>
            <div className="font-mono font-semibold" style={{ color: isClose && action.realized_pnl !== undefined ? (action.realized_pnl >= 0 ? '#0ECB81' : '#F6465D') : '#EAECEF' }}>
              {isClose ? 
                (action.realized_pnl !== undefined ? `${action.realized_pnl >= 0 ? '+' : ''}${action.realized_pnl.toFixed(2)} USDT` : '-') : 
                formatPrice(action.price)}
            </div>
            {isClose && action.realized_pnl !== undefined && action.quantity !== undefined && (
              <div className="text-xs mt-0.5" style={{ color: action.realized_pnl >= 0 ? '#0ECB81' : '#F6465D' }}>
                {action.realized_pnl >= 0 ? t('profit', language) : t('loss', language)}
              </div>
            )}
          </div>

          {/* Stop Loss or Profit/Loss or Close Percentage or Wait Time */}
          <div className="text-center">
            {isClose ? (
              <>
                <div className="text-xs mb-1" style={{ color: '#0ECB81' }}>
                  {t('profitLossRatio', language)}
                </div>
                <div className="font-mono font-semibold" style={{ color: action.realized_pnl_percentage !== undefined && action.realized_pnl_percentage >= 0 ? '#0ECB81' : '#F6465D' }}>
                  {action.realized_pnl_percentage !== undefined ? 
                    `${action.realized_pnl_percentage >= 0 ? '+' : ''}${action.realized_pnl_percentage.toFixed(2)}%` : '-'}
                </div>
                {action.realized_pnl_percentage !== undefined && (
                  <div className="text-xs mt-0.5" style={{ color: '#848E9C' }}>
                    {Math.abs(action.realized_pnl_percentage).toFixed(2)}%{t('yieldRatio', language)}
                  </div>
                )}
              </>
            ) : isPartialClose ? (
              <>
                <div className="text-xs mb-1" style={{ color: '#9C27B0' }}>
                  {t('closePercentage', language)}
                </div>
                <div className="font-mono font-semibold" style={{ color: '#9C27B0' }}>
                  {action.close_percentage !== undefined ? `${action.close_percentage}%` : '-'}
                </div>
              </>
            ) : isTrailingStop ? (
              <>
                <div className="text-xs mb-1" style={{ color: '#2196F3' }}>
                  {t('trailPercent', language)}
                </div>
                <div className="font-mono font-semibold" style={{ color: '#2196F3' }}>
                  {action.trail_percentage !== undefined ? `${action.trail_percentage}%` : '-'}
                </div>
              </>
            ) : isHold || isWait ? (
              <>
                <div className="text-xs mb-1" style={{ color: '#848E9C' }}>
                  {isHold ? t('confidence', language) : t('waitingReason', language)}
                </div>
                <div className="font-mono font-semibold" style={{ color: getConfidenceColor(action.confidence) }}>
                  {isHold ? `${action.confidence || 0}%` : '-'}
                </div>
              </>
            ) : isUpdateStopLoss ? (
              <>
                <div className="text-xs mb-1" style={{ color: '#FF9800' }}>
                  {t('newStopLoss', language)}
                </div>
                <div className="font-mono font-semibold" style={{ color: '#FF9800' }}>
                  {action.new_stop_loss !== undefined ? formatPrice(action.new_stop_loss) : '-'}
                </div>
              </>
            ) : (
              <>
                <div className="text-xs mb-1" style={{ color: '#F6465D' }}>
                  {t('stopLoss', language)}
                </div>
                <div className="font-mono font-semibold" style={{ color: '#F6465D' }}>
                  {formatPrice(action.stop_loss)}
                </div>
                {action.stop_loss !== undefined && action.price !== undefined && (
                  <div className="text-xs mt-0.5" style={{ color: '#848E9C' }}>
                    {calcPctChange(action.price, action.stop_loss, isLong)}
                  </div>
                )}
              </>
            )}
          </div>

          {/* Take Profit or Entry Price or Callback Rate or Action Reason */}
          <div className="text-center">
            {isClose ? (
              <>
                <div className="text-xs mb-1" style={{ color: '#0ECB81' }}>
                  {t('entryPrice', language)}
                </div>
                <div className="font-mono font-semibold" style={{ color: '#EAECEF' }}>
                  {formatPrice(action.entry_price)}
                </div>
                {action.exit_price !== undefined && action.entry_price !== undefined && (
                  <div className="text-xs mt-0.5" style={{ color: action.realized_pnl !== undefined && action.realized_pnl >= 0 ? '#0ECB81' : '#F6465D' }}>
                    {calcPctChange(action.entry_price, action.exit_price, isLong)}
                  </div>
                )}
              </>
            ) : isTrailingStop ? (
              <>
                <div className="text-xs mb-1" style={{ color: '#2196F3' }}>
                  {t('callbackRate', language)}
                </div>
                <div className="font-mono font-semibold" style={{ color: '#2196F3' }}>
                  {action.callback_rate !== undefined ? `${action.callback_rate}%` : '-'}
                </div>
              </>
            ) : isHold || isWait ? (
              <>
                <div className="text-xs mb-1" style={{ color: '#848E9C' }}>
                  {isHold ? t('holdingTime', language) : t('waitingTime', language)}
                </div>
                <div className="font-mono font-semibold" style={{ color: '#848E9C' }}>
                  {action.timestamp ? `${Math.floor((new Date().getTime() - new Date(action.timestamp).getTime()) / 60000)} ${t('minutes', language)}` : '-'}
                </div>
              </>
            ) : (
              <>
                <div className="text-xs mb-1" style={{ color: '#0ECB81' }}>
                  {t('takeProfit', language)}
                </div>
                <div className="font-mono font-semibold" style={{ color: '#0ECB81' }}>
                  {formatPrice(action.take_profit)}
                </div>
                {action.take_profit !== undefined && action.price !== undefined && (
                  <div className="text-xs mt-0.5" style={{ color: '#848E9C' }}>
                    {calcPctChange(action.price, action.take_profit, isLong)}
                  </div>
                )}
              </>
            )}
          </div>

          {/* Leverage or Quantity or Confidence or Exit Price */}
          <div className="text-center">
            <div className="text-xs mb-1" style={{ color: '#848E9C' }}>
              {isOpen ? t('leverage', language) : isClose ? t('exitPrice', language) : isHold || isWait ? t('confidence', language) : isUpdateStopLoss ? t('adjustment', language) : t('quantity', language)}
            </div>
            <div className="font-mono font-semibold" style={{ color: isOpen ? '#F0B90B' : isClose ? '#EAECEF' : isHold || isWait ? getConfidenceColor(action.confidence) : isUpdateStopLoss ? '#FF9800' : '#EAECEF' }}>
              {isOpen ? `${action.leverage}x` : isClose ? formatPrice(action.exit_price) : isHold || isWait ? `${action.confidence !== undefined ? action.confidence : 0}%` : isUpdateStopLoss ? 
                (action.stop_loss !== undefined && action.new_stop_loss !== undefined ? 
                  `${action.new_stop_loss > action.stop_loss ? '↑' : '↓'}${Math.abs(((action.new_stop_loss - action.stop_loss) / action.stop_loss) * 100).toFixed(1)}%` : '-') : 
                (action.quantity !== undefined && action.quantity !== null && action.quantity > 0 ? formatPrice(action.quantity) : '-')}
            </div>
          </div>
        </div>
      )}

      {/* Risk/Reward Ratio for open positions */}
      {isOpen && action.stop_loss !== undefined && action.take_profit !== undefined && action.price !== undefined && (
        <div className="mt-3 pt-3 flex items-center justify-between" style={{ borderTop: '1px solid #2B3139' }}>
          <span className="text-xs" style={{ color: '#848E9C' }}>
            {t('riskReward', language)}
          </span>
          <div className="flex items-center gap-2">
            {(() => {
              const slDist = Math.abs(action.price - action.stop_loss)
              const tpDist = Math.abs(action.take_profit - action.price)
              const ratio = slDist > 0 ? (tpDist / slDist) : 0
              const ratioColor = ratio >= 3 ? '#0ECB81' : ratio >= 2 ? '#F0B90B' : '#F6465D'
              return (
                <>
                  <div className="flex gap-1">
                    <span style={{ color: '#F6465D' }}>1</span>
                    <span style={{ color: '#848E9C' }}>:</span>
                    <span style={{ color: '#0ECB81' }}>{ratio.toFixed(1)}</span>
                  </div>
                  <div
                    className="h-1.5 rounded-full"
                    style={{
                      width: '60px',
                      background: '#2B3139',
                    }}
                  >
                    <div
                      className="h-full rounded-full transition-all duration-300"
                      style={{
                        width: `${Math.min(ratio / 5 * 100, 100)}%`,
                        background: ratioColor
                      }}
                    />
                  </div>
                </>
              )
            })()}
          </div>
        </div>
      )}

      {/* Reasoning */}
      {action.reasoning && (
        <div className="mt-3 pt-3" style={{ borderTop: '1px solid #2B3139' }}>
          <div className="text-xs line-clamp-2" style={{ color: '#848E9C' }}>
            💡 {action.reasoning}
          </div>
        </div>
      )}

      {/* Error Message */}
      {action.error && (
        <div
          className="mt-3 rounded p-3 text-xs border-l-2 animate-pulse"
          style={{
            background: 'rgba(246, 70, 93, 0.15)',
            borderLeftColor: '#F6465D',
            border: '1px solid rgba(246, 70, 93, 0.3)',
            color: '#F6465D',
          }}
        >
          <div className="flex items-center gap-2">
            <span className="text-lg">⚠️</span>
            <span className="font-semibold">执行失败:</span>
          </div>
          <div className="mt-1 ml-6">
            {action.error}
          </div>
        </div>
      )}
    </div>
  )
}

export function DecisionCard({ decision, language, onSymbolClick, onDelete, exchangeType, exchangeCustomUrl, exchanges, traders, expandedState, onUpdateExpansion }: DecisionCardProps) {
  //调信息输出
  console.log('=== DecisionCard组件接收的props ===');
  console.log('decision.trader_id:', decision.trader_id);
  console.log('traders prop:', traders);
  console.log('traders类型:', typeof traders);
  console.log('traders是否为数组:', Array.isArray(traders));
  console.log('traders长度:', traders?.length);
  
  // 检查traders是否为undefined
  if (typeof traders === 'undefined') {
    console.error('❌ DecisionCard接收到的traders是undefined');
    console.trace('traders prop追踪');
  }
  
  // 使用传入的状态，如果没有则使用默认值
  const showSystemPrompt = expandedState?.showSystemPrompt || false;
  const showInputPrompt = expandedState?.showInputPrompt || false;
  const showCoT = expandedState?.showCoT || false;
  const showJSON = expandedState?.showJSON || false;


  // Copy text to clipboard
  const copyToClipboard = async (text: string, label: string) => {
    try {
      await navigator.clipboard.writeText(text)
      alert(`${label} copied!`)
    } catch (err) {
      console.error('Failed to copy:', err)
    }
  }

  // Download text as file
  const downloadAsFile = (text: string, filename: string) => {
    const blob = new Blob([text], { type: 'text/plain;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
  }

  return (
    <div
      className="rounded-xl p-5 transition-all duration-300 hover:translate-y-[-2px]"
      style={{
        border: '1px solid #2B3139',
        background: 'linear-gradient(180deg, #1E2329 0%, #181C21 100%)',
        boxShadow: '0 4px 16px rgba(0, 0, 0, 0.3)',
      }}
    >
      {/* Header */}
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-3">
          <div
            className="w-10 h-10 rounded-lg flex items-center justify-center"
            style={{ background: 'rgba(240, 185, 11, 0.15)' }}
          >
            <span className="text-xl">🤖</span>
          </div>
          <div>
            <div className="font-bold" style={{ color: '#EAECEF' }}>
              {t('cycle', language)} #{decision.cycle_number}
            </div>
            <div className="text-xs" style={{ color: '#848E9C' }}>
              {new Date(decision.timestamp).toLocaleString()}
            </div>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <div
            className={`px-4 py-1.5 rounded-full text-xs font-bold tracking-wider transition-all duration-300 ${
              decision.success 
                ? 'animate-pulse-once' 
                : 'animate-pulse border-2'
            }`}
            style={
              decision.success
                ? { 
                    background: 'rgba(14, 203, 129, 0.25)', 
                    color: '#0ECB81', 
                    border: '1px solid rgba(14, 203, 129, 0.5)',
                    boxShadow: '0 0 10px rgba(14, 203, 129, 0.3)'
                  }
                : { 
                    background: 'rgba(246, 70, 93, 0.25)', 
                    color: '#F6465D', 
                    border: '1px solid rgba(246, 70, 93, 0.7)',
                    boxShadow: '0 0 15px rgba(246, 70, 93, 0.4)'
                  }
            }
          >
            {t(decision.success ? 'success' : 'failed', language)}
          </div>
          {onDelete && (
            <button
              onClick={() => {
                if (window.confirm(`Are you sure you want to delete cycle #${decision.cycle_number}?`)) {
                  onDelete(decision.id, decision.trader_id);
                }
              }}
              className="p-1.5 rounded-lg hover:bg-red-500/20 transition-colors"
              style={{ color: '#F6465D' }}
              title="Delete this decision cycle"
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M3 6h18"></path>
                <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"></path>
                <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"></path>
              </svg>
            </button>
          )}
        </div>
      </div>

      {/* Decision Actions - Beautiful Grid */}
      {decision.decisions && decision.decisions.length > 0 && (
        <div className="space-y-3 mb-4">
          {decision.decisions.map((action, index) => (
            <ActionCard 
              key={`${action.symbol}-${index}`} 
              action={action} 
              language={language} 
              onSymbolClick={onSymbolClick} 
              positions={decision.positions} 
              exchangeType={exchangeType} 
              exchangeCustomUrl={exchangeCustomUrl}
              exchanges={exchanges}
              traderId={decision.trader_id}
              traders={traders}
            />
          ))}
        </div>
      )}

      {/* Collapsible Sections */}
      <div className="space-y-2">
        {/* System Prompt */}
        {decision.system_prompt && (
          <div>
            <div className="flex items-center gap-2 justify-between">
              <button
                onClick={() => onUpdateExpansion?.({ showSystemPrompt: !showSystemPrompt })}
                className="flex items-center gap-2 text-sm transition-colors w-full p-2 rounded hover:bg-white/5"
              >
                <div className="flex items-center gap-2">
                  <span className="text-base">⚙️</span>
                  <span className="font-semibold" style={{ color: '#a78bfa' }}>
                    System Prompt
                  </span>
                </div>
                <span
                  className="text-xs px-2 py-0.5 rounded"
                  style={{ background: 'rgba(167, 139, 250, 0.15)', color: '#a78bfa' }}
                >
                  {showSystemPrompt ? t('collapse', language) : t('expand', language)}
                </span>
              </button>
              <div className="flex items-center gap-2 ml-2">
                <button
                  onClick={(e) => {
                    e.stopPropagation()
                    copyToClipboard(decision.system_prompt, 'System Prompt')
                  }}
                  className="text-xs px-2.5 py-1 rounded hover:opacity-80 transition-opacity flex items-center gap-1"
                  style={{ background: 'rgba(167, 139, 250, 0.2)', color: '#a78bfa', border: '1px solid rgba(167, 139, 250, 0.3)' }}
                  title="Copy to clipboard"
                >
                  <span>📋</span>
                </button>
                <button
                  onClick={(e) => {
                    e.stopPropagation()
                    downloadAsFile(decision.system_prompt, `system-prompt-cycle-${decision.cycle_number}.txt`)
                  }}
                  className="text-xs px-2.5 py-1 rounded hover:opacity-80 transition-opacity flex items-center gap-1"
                  style={{ background: 'rgba(167, 139, 250, 0.2)', color: '#a78bfa', border: '1px solid rgba(167, 139, 250, 0.3)' }}
                  title="Download as file"
                >
                  <span>💾</span>
                </button>
              </div>
            </div>
            {showSystemPrompt && (
              <div
                className="mt-2 rounded-lg p-4 text-sm font-mono whitespace-pre-wrap max-h-96 overflow-y-auto"
                style={{
                  background: '#0B0E11',
                  border: '1px solid #2B3139',
                  color: '#EAECEF',
                }}
              >
                {decision.system_prompt}
              </div>
            )}
          </div>
        )}

        {/* User/Input Prompt */}
        {decision.input_prompt && (
          <div>
            <div className="flex items-center gap-2 justify-between">
              <button
                onClick={() => onUpdateExpansion?.({ showInputPrompt: !showInputPrompt })}
                className="flex items-center gap-2 text-sm transition-colors w-full p-2 rounded hover:bg-white/5"
              >
                <div className="flex items-center gap-2">
                  <span className="text-base">📥</span>
                  <span className="font-semibold" style={{ color: '#60a5fa' }}>
                    User Prompt
                  </span>
                </div>
                <span
                  className="text-xs px-2 py-0.5 rounded"
                  style={{ background: 'rgba(96, 165, 250, 0.15)', color: '#60a5fa' }}
                >
                  {showInputPrompt ? t('collapse', language) : t('expand', language)}
                </span>
              </button>
              <div className="flex items-center gap-2 ml-2">
                <button
                  onClick={(e) => {
                    e.stopPropagation()
                    copyToClipboard(decision.input_prompt, 'User Prompt')
                  }}
                  className="text-xs px-2.5 py-1 rounded hover:opacity-80 transition-opacity flex items-center gap-1"
                  style={{ background: 'rgba(96, 165, 250, 0.2)', color: '#60a5fa', border: '1px solid rgba(96, 165, 250, 0.3)' }}
                  title="Copy to clipboard"
                >
                  <span>📋</span>
                </button>
                <button
                  onClick={(e) => {
                    e.stopPropagation()
                    downloadAsFile(decision.input_prompt, `user-prompt-cycle-${decision.cycle_number}.txt`)
                  }}
                  className="text-xs px-2.5 py-1 rounded hover:opacity-80 transition-opacity flex items-center gap-1"
                  style={{ background: 'rgba(96, 165, 250, 0.2)', color: '#60a5fa', border: '1px solid rgba(96, 165, 250, 0.3)' }}
                  title="Download as file"
                >
                  <span>💾</span>
                </button>
              </div>
            </div>
            {showInputPrompt && (
              <div
                className="mt-2 rounded-lg p-4 text-sm font-mono whitespace-pre-wrap max-h-96 overflow-y-auto"
                style={{
                  background: '#0B0E11',
                  border: '1px solid #2B3139',
                  color: '#EAECEF',
                }}
              >
                {decision.input_prompt}
              </div>
            )}
          </div>
        )}

        {/* JSON Content - 显示决策JSON，使用绿色主题 */}
        {decision.decision_json && (
          <div className="mt-3">
            <div className="flex items-center gap-2 justify-between">
              <button
                onClick={() => onUpdateExpansion?.({ showJSON: !showJSON })}
                className="flex items-center gap-2 text-sm transition-colors w-full p-2 rounded hover:bg-white/5"
              >
                <div className="flex items-center gap-2">
                  <span className="text-base">📋</span>
                  <span className="font-semibold" style={{ color: '#10B981' }}>
                    JSON Content
                  </span>
                </div>
                <span
                  className="text-xs px-2 py-0.5 rounded"
                  style={{ background: 'rgba(16, 185, 129, 0.15)', color: '#10B981' }}
                >
                  {showJSON ? t('collapse', language) : t('expand', language)}
                </span>
              </button>
              <div className="flex items-center gap-2 ml-2">
                <button
                  onClick={(e) => {
                    e.stopPropagation()
                    copyToClipboard(decision.decision_json, 'JSON Content')
                  }}
                  className="text-xs px-2.5 py-1 rounded hover:opacity-80 transition-opacity flex items-center gap-1"
                  style={{ background: 'rgba(16, 185, 129, 0.2)', color: '#10B981', border: '1px solid rgba(16, 185, 129, 0.3)' }}
                  title="Copy to clipboard"
                >
                  <span>📋</span>
                </button>
                <button
                  onClick={(e) => {
                    e.stopPropagation()
                    downloadAsFile(decision.decision_json, `json-content-cycle-${decision.cycle_number}.json`)
                  }}
                  className="text-xs px-2.5 py-1 rounded hover:opacity-80 transition-opacity flex items-center gap-1"
                  style={{ background: 'rgba(16, 185, 129, 0.2)', color: '#10B981', border: '1px solid rgba(16, 185, 129, 0.3)' }}
                  title="Download as file"
                >
                  <span>💾</span>
                </button>
              </div>
            </div>
            {showJSON && (
              <div
                className="mt-2 rounded-lg p-4 text-sm font-mono whitespace-pre-wrap max-h-96 overflow-y-auto"
                style={{
                  background: '#0B0E11',
                  border: '1px solid #2B3139',
                  color: '#EAECEF',
                }}
              >
                {decision.decision_json}
              </div>
            )}
          </div>
        )}

        {/* AI Thinking - 显示思维链文本内容 */}
        {decision.cot_trace && (
          <div>
            <button
              onClick={() => onUpdateExpansion?.({ showCoT: !showCoT })}
              className="flex items-center gap-2 text-sm transition-colors w-full justify-between p-2 rounded hover:bg-white/5"
            >
              <div className="flex items-center gap-2">
                <span className="text-base">🧠</span>
                <span className="font-semibold" style={{ color: '#F0B90B' }}>
                  {t('aiThinking', language)}
                </span>
              </div>
              <span
                className="text-xs px-2 py-0.5 rounded"
                style={{ background: 'rgba(240, 185, 11, 0.15)', color: '#F0B90B' }}
              >
                {showCoT ? t('collapse', language) : t('expand', language)}
              </span>
            </button>
            {showCoT && (
              <div className="space-y-3 mt-2">
                <div
                  className="rounded-lg p-4 text-sm font-mono whitespace-pre-wrap max-h-96 overflow-y-auto"
                  style={{
                    background: '#0B0E11',
                    border: '1px solid #2B3139',
                    color: '#EAECEF',
                  }}
                >
                  {decision.cot_trace}
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Execution Log */}
      {decision.execution_log && decision.execution_log.length > 0 && (
        <div
          className="rounded-lg p-3 mt-4 text-xs font-mono space-y-1"
          style={{ background: '#0B0E11', border: '1px solid #2B3139' }}
        >
          {decision.execution_log.map((log, index) => (
            <div key={`${log}-${index}`} style={{ color: '#EAECEF' }}>
              {log}
            </div>
          ))}
        </div>
      )}

      {/* Error Message - Enhanced Display */}
      {(decision.error_message || decision.decisions?.some(d => d.error)) && (
        <div
          className="rounded-lg p-4 mt-4 border-l-4 animate-pulse"
          style={{
            background: 'rgba(246, 70, 93, 0.15)',
            borderLeftColor: '#F6465D',
            border: '1px solid rgba(246, 70, 93, 0.4)',
          }}
        >
          <div className="flex items-start gap-3">
            <div className="text-2xl">❌</div>
            <div className="flex-1">
              <div className="font-bold text-red-400 mb-2">
                {t('executionFailed', language) || '执行失败'}
              </div>
              {decision.error_message && (
                <div className="text-sm text-red-300 mb-3">
                  {decision.error_message}
                </div>
              )}
              {decision.decisions?.filter(d => d.error).length > 0 && (
                <div className="space-y-2">
                  <div className="text-xs text-red-200 font-semibold">
                    {t('failedActions', language) || '失败的操作:'}
                  </div>
                  {decision.decisions
                    .filter(d => d.error)
                    .map((failedAction, index) => (
                      <div 
                        key={index}
                        className="text-xs p-2 rounded bg-red-900/20 border border-red-800/30"
                      >
                        <div className="font-mono text-red-200">
                          {failedAction.symbol} {failedAction.action}
                        </div>
                        <div className="text-red-300 mt-1">
                          {failedAction.error}
                        </div>
                      </div>
                    ))
                  }
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
