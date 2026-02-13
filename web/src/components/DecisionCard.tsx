import { useState } from 'react'
import type { DecisionRecord, DecisionAction } from '../types'
import { t, type Language } from '../i18n/translations'

interface DecisionCardProps {
  decision: DecisionRecord
  language: Language
  onSymbolClick?: (symbol: string) => void
  onDelete?: (decisionId: number, traderId: string) => void
}

// Action type configuration
const ACTION_CONFIG: Record<string, { color: string; bg: string; icon: string; label: string }> = {
  open_long: { color: '#0ECB81', bg: 'rgba(14, 203, 129, 0.15)', icon: '📈', label: 'Open Long' },
  open_short: { color: '#F6465D', bg: 'rgba(246, 70, 93, 0.15)', icon: '📉', label: 'Open Short' },
  close_long: { color: '#F0B90B', bg: 'rgba(240, 185, 11, 0.15)', icon: '💰', label: 'Close Long' },
  close_short: { color: '#F0B90B', bg: 'rgba(240, 185, 11, 0.15)', icon: '💰', label: 'Close Short' },
  hold: { color: '#848E9C', bg: 'rgba(132, 142, 156, 0.15)', icon: '⏸️', label: 'Hold' },
  wait: { color: '#848E9C', bg: 'rgba(132, 142, 156, 0.15)', icon: '⏳', label: 'Wait' },
  partial_close: { color: '#9C27B0', bg: 'rgba(156, 39, 176, 0.15)', icon: '📉', label: 'Partial Close' },
  update_stop_loss: { color: '#FF9800', bg: 'rgba(255, 152, 0, 0.15)', icon: '🛡️', label: 'Update Stop Loss' },
  update_take_profit: { color: '#4CAF50', bg: 'rgba(76, 175, 80, 0.15)', icon: '🎯', label: 'Update Take Profit' },
  trailing_stop: { color: '#2196F3', bg: 'rgba(33, 150, 243, 0.15)', icon: '👣', label: 'Trailing Stop' },
  dynamic_take_profit: { color: '#00BCD4', bg: 'rgba(0, 188, 212, 0.15)', icon: '🚀', label: 'Dynamic Take Profit' },
  add_position: { color: '#795548', bg: 'rgba(121, 85, 72, 0.15)', icon: '➕', label: 'Add Position' },
  full_close: { color: '#607D8B', bg: 'rgba(96, 125, 139, 0.15)', icon: '🔚', label: 'Full Close' },
}

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

// 新增：渲染数值变化信息（阶段1）
function renderValueChanges(action: DecisionAction, positions?: any[]) {
  console.log('🔍 renderValueChanges called with action:', action);
  
  const changes = [];
  
  // 根据不同动作类型显示相应的数值变化
  switch(action.action) {
    case 'update_stop_loss':
      console.log('📊 Processing update_stop_loss, new_stop_loss:', action.new_stop_loss);
      if (action.new_stop_loss !== undefined && action.new_stop_loss !== null) {
        changes.push({
          label: 'New Stop Loss',
          value: formatPrice(action.new_stop_loss),
          color: '#FF9800'
        });
      }
      break;
      
    case 'update_take_profit':
      console.log('📊 Processing update_take_profit, new_take_profit:', action.new_take_profit);
      if (action.new_take_profit !== undefined && action.new_take_profit !== null) {
        changes.push({
          label: 'New Take Profit',
          value: formatPrice(action.new_take_profit),
          color: '#4CAF50'
        });
        
        // 获取当前持仓的止盈价格用于对比
        const currentPosition = positions?.find((pos: any) => pos.symbol === action.symbol);
        const currentTakeProfit = currentPosition?.take_profit;
        
        // 显示相对变化（如果没有原值则显示绝对值）
        if (currentTakeProfit !== undefined && currentTakeProfit !== null) {
          const change = action.new_take_profit - currentTakeProfit;
          const changePercent = (change / currentTakeProfit) * 100;
          const isIncrease = change > 0;
          changes.push({
            label: 'Change',
            value: `${isIncrease ? '+' : ''}${changePercent.toFixed(2)}% (${isIncrease ? '+' : ''}${formatPrice(change)})`,
            color: isIncrease ? '#4CAF50' : '#F44336'
          });
        } else if (action.take_profit !== undefined && action.take_profit !== null) {
          // 如果没有持仓数据，使用action中的原始止盈价格
          const change = action.new_take_profit - action.take_profit;
          const changePercent = (change / action.take_profit) * 100;
          const isIncrease = change > 0;
          changes.push({
            label: 'Change',
            value: `${isIncrease ? '+' : ''}${changePercent.toFixed(2)}% (${isIncrease ? '+' : ''}${formatPrice(change)})`,
            color: isIncrease ? '#4CAF50' : '#F44336'
          });
        }
        
        // 如果有止损价格，计算新的风险收益比
        if (action.stop_loss !== undefined && action.stop_loss !== null && 
            action.price !== undefined && action.price !== null) {
          const risk = Math.abs(action.price - action.stop_loss);
          const reward = Math.abs(action.new_take_profit - action.price);
          const riskRewardRatio = risk > 0 ? (reward / risk) : 0;
          const ratioColor = riskRewardRatio >= 3 ? '#4CAF50' : riskRewardRatio >= 2 ? '#FFC107' : '#F44336';
          
          changes.push({
            label: 'New R/R Ratio',
            value: `1:${riskRewardRatio.toFixed(2)}`,
            color: ratioColor
          });
        }
        
        // 显示距离当前市场价格的百分比
        if (action.price !== undefined && action.price !== null) {
          const distancePercent = ((action.new_take_profit - action.price) / action.price) * 100;
          const distanceColor = distancePercent > 0 ? '#4CAF50' : '#F44336';
          changes.push({
            label: 'Distance from Price',
            value: `${distancePercent > 0 ? '+' : ''}${distancePercent.toFixed(2)}%`,
            color: distanceColor
          });
        }
      }
      break;
      
    case 'partial_close':
      console.log('📊 Processing partial_close, close_percentage:', action.close_percentage);
      if (action.close_percentage !== undefined && action.close_percentage !== null) {
        changes.push({
          label: 'Close Percentage',
          value: `${action.close_percentage}%`,
          color: '#9C27B0'
        });
      }
      break;
      
    case 'trailing_stop':
      console.log('📊 Processing trailing_stop, trail_percentage:', action.trail_percentage, 'callback_rate:', action.callback_rate);
      if (action.trail_percentage !== undefined && action.trail_percentage !== null) {
        changes.push({
          label: 'Trail %',
          value: `${action.trail_percentage}%`,
          color: '#2196F3'
        });
      }
      if (action.callback_rate !== undefined && action.callback_rate !== null) {
        changes.push({
          label: 'Callback Rate',
          value: `${action.callback_rate}%`,
          color: '#2196F3'
        });
      }
      break;
      
    case 'dynamic_take_profit':
      console.log('📊 Processing dynamic_take_profit, target_roi:', action.target_roi, 'max_roi:', action.max_roi);
      if (action.target_roi !== undefined && action.target_roi !== null) {
        changes.push({
          label: 'Target ROI',
          value: `${action.target_roi}%`,
          color: '#00BCD4'
        });
      }
      if (action.max_roi !== undefined && action.max_roi !== null) {
        changes.push({
          label: 'Max ROI',
          value: `${action.max_roi}%`,
          color: '#00BCD4'
        });
      }
      break;
      
    case 'add_position':
      console.log('📊 Processing add_position, additional_position_size_usd:', action.additional_position_size_usd);
      if (action.additional_position_size_usd !== undefined && action.additional_position_size_usd !== null) {
        changes.push({
          label: 'Add Amount',
          value: `${action.additional_position_size_usd} USDT`,
          color: '#795548'
        });
      }
      break;
  }
  
  console.log('📋 Changes to display:', changes);
  
  // 如果没有变化信息，不显示
  if (changes.length === 0) {
    console.log('⚠️ No changes to display');
    return null;
  }
  
  return (
    <div className="mt-3 pt-3" style={{ borderTop: '1px solid #2B3139' }}>
      <div className="text-xs mb-2" style={{ color: '#848E9C' }}>Value Changes:</div>
      <div className="flex flex-wrap gap-2">
        {changes.map((change, index) => (
          <div 
            key={index}
            className="px-2 py-1 rounded text-xs font-mono"
            style={{ 
              background: `${change.color}22`,
              border: `1px solid ${change.color}55`,
              color: change.color
            }}
          >
            <span className="font-semibold">{change.label}:</span> {change.value}
          </div>
        ))}
      </div>
    </div>
  );
}

// Single Action Card Component
function ActionCard({ action, language, onSymbolClick, positions }: { action: DecisionAction; language: Language; onSymbolClick?: (symbol: string) => void; positions?: any[] }) {
  const config = ACTION_CONFIG[action.action] || ACTION_CONFIG.wait
  const isLong = action.action.includes('long')
  const isOpen = action.action.includes('open')

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
          <span className="text-xl">{config.icon}</span>
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
      {isOpen && (
        <div className="grid grid-cols-4 gap-3 mt-3 pt-3" style={{ borderTop: '1px solid #2B3139' }}>
          {/* Entry Price */}
          <div className="text-center">
            <div className="text-xs mb-1" style={{ color: '#848E9C' }}>
              {t('entryPrice', language)}
            </div>
            <div className="font-mono font-semibold" style={{ color: '#EAECEF' }}>
              {formatPrice(action.price)}
            </div>
          </div>

          {/* Stop Loss */}
          <div className="text-center">
            <div className="text-xs mb-1" style={{ color: '#F6465D' }}>
              {t('stopLoss', language)}
            </div>
            <div className="font-mono font-semibold" style={{ color: '#F6465D' }}>
              {formatPrice(action.stop_loss)}
            </div>
            {action.stop_loss && action.price && (
              <div className="text-xs mt-0.5" style={{ color: '#848E9C' }}>
                {calcPctChange(action.price, action.stop_loss, isLong)}
              </div>
            )}
          </div>

          {/* Take Profit */}
          <div className="text-center">
            <div className="text-xs mb-1" style={{ color: '#0ECB81' }}>
              {t('takeProfit', language)}
            </div>
            <div className="font-mono font-semibold" style={{ color: '#0ECB81' }}>
              {formatPrice(action.take_profit)}
            </div>
            {action.take_profit && action.price && (
              <div className="text-xs mt-0.5" style={{ color: '#848E9C' }}>
                {calcPctChange(action.price, action.take_profit, isLong)}
              </div>
            )}
          </div>

          {/* Leverage */}
          <div className="text-center">
            <div className="text-xs mb-1" style={{ color: '#848E9C' }}>
              {t('leverage', language)}
            </div>
            <div className="font-mono font-semibold" style={{ color: '#F0B90B' }}>
              {action.leverage}x
            </div>
          </div>
        </div>
      )}

      {/* 新增：数值变化显示（阶段1） */}
      {renderValueChanges(action, positions)}

      {/* Risk/Reward Ratio for open positions */}
      {isOpen && action.stop_loss && action.take_profit && action.price && (
        <div className="mt-3 pt-3 flex items-center justify-between" style={{ borderTop: '1px solid #2B3139' }}>
          <span className="text-xs" style={{ color: '#848E9C' }}>{t('riskReward', language)}</span>
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
          className="mt-3 rounded p-2 text-xs"
          style={{
            background: 'rgba(246, 70, 93, 0.1)',
            border: '1px solid rgba(246, 70, 93, 0.3)',
            color: '#F6465D',
          }}
        >
          ❌ {action.error}
        </div>
      )}
    </div>
  )
}

export function DecisionCard({ decision, language, onSymbolClick, onDelete }: DecisionCardProps) {
  const [showSystemPrompt, setShowSystemPrompt] = useState(false)
  const [showInputPrompt, setShowInputPrompt] = useState(false)
  const [showCoT, setShowCoT] = useState(false)
  const [showJSON, setShowJSON] = useState(false) // 新增：JSON显示状态


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
            className="px-4 py-1.5 rounded-full text-xs font-bold tracking-wider"
            style={
              decision.success
                ? { background: 'rgba(14, 203, 129, 0.15)', color: '#0ECB81', border: '1px solid rgba(14, 203, 129, 0.3)' }
                : { background: 'rgba(246, 70, 93, 0.15)', color: '#F6465D', border: '1px solid rgba(246, 70, 93, 0.3)' }
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
            <ActionCard key={`${action.symbol}-${index}`} action={action} language={language} onSymbolClick={onSymbolClick} positions={decision.positions} />
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
                onClick={() => setShowSystemPrompt(!showSystemPrompt)}
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
                onClick={() => setShowInputPrompt(!showInputPrompt)}
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
                onClick={() => setShowJSON(!showJSON)}
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
              onClick={() => setShowCoT(!showCoT)}
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

      {/* Error Message */}
      {decision.error_message && (
        <div
          className="rounded-lg p-3 mt-4 text-sm"
          style={{
            background: 'rgba(246, 70, 93, 0.1)',
            border: '1px solid rgba(246, 70, 93, 0.4)',
            color: '#F6465D',
          }}
        >
          ❌ {decision.error_message}
        </div>
      )}
    </div>
  )
}
