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

// 移除：渲染数值变化信息（阶段1）
// 此函数已被弃用，所有数值变化信息现在都在Trading Details Grid中统一显示

// Single Action Card Component
function ActionCard({ action, language, onSymbolClick, positions }: { action: DecisionAction; language: Language; onSymbolClick?: (symbol: string) => void; positions?: any[] }) {
  const config = ACTION_CONFIG[action.action] || ACTION_CONFIG.wait
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
      {(isOpen || isClose || isPartialClose || isTrailingStop || isUpdateStopLoss || isHold || isWait) && (
        <div className="grid grid-cols-4 gap-3 mt-3 pt-3" style={{ borderTop: '1px solid #2B3139' }}>
          {/* Entry Price or Current Price */}
          <div className="text-center">
            <div className="text-xs mb-1" style={{ color: '#848E9C' }}>
              {isOpen ? t('entryPrice', language) : isHold || isWait ? t('currentPrice', language) : t('currentPrice', language)}
            </div>
            <div className="font-mono font-semibold" style={{ color: '#EAECEF' }}>
              {formatPrice(action.price)}
            </div>
          </div>

          {/* Stop Loss or Profit/Loss or Close Percentage or Wait Time */}
          <div className="text-center">
            {isClose ? (
              <>
                <div className="text-xs mb-1" style={{ color: '#0ECB81' }}>
                  盈亏金额
                </div>
                <div className="font-mono font-semibold" style={{ color: action.pnl && action.pnl >= 0 ? '#0ECB81' : '#F6465D' }}>
                  {action.pnl !== undefined ? `${action.pnl >= 0 ? '+' : ''}${action.pnl.toFixed(2)} USDT` : '-'}
                </div>
                {action.pnl && action.quantity && (
                  <div className="text-xs mt-0.5" style={{ color: '#848E9C' }}>
                    {action.pnl >= 0 ? '盈利' : '亏损'}
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
                  Trail %
                </div>
                <div className="font-mono font-semibold" style={{ color: '#2196F3' }}>
                  {action.trail_percentage !== undefined ? `${action.trail_percentage}%` : '-'}
                </div>
              </>
            ) : isHold || isWait ? (
              <>
                <div className="text-xs mb-1" style={{ color: '#848E9C' }}>
                  {isHold ? '置信度' : '等待原因'}
                </div>
                <div className="font-mono font-semibold" style={{ color: getConfidenceColor(action.confidence) }}>
                  {isHold ? `${action.confidence || 0}%` : action.reasoning?.substring(0, 15) || '-'}
                </div>
              </>
            ) : isUpdateStopLoss ? (
              <>
                <div className="text-xs mb-1" style={{ color: '#FF9800' }}>
                  新止损价
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
                {action.stop_loss && action.price && (
                  <div className="text-xs mt-0.5" style={{ color: '#848E9C' }}>
                    {calcPctChange(action.price, action.stop_loss, isLong)}
                  </div>
                )}
              </>
            )}
          </div>

          {/* Take Profit or PnL Percentage or Callback Rate or Action Reason */}
          <div className="text-center">
            {isClose ? (
              <>
                <div className="text-xs mb-1" style={{ color: '#0ECB81' }}>
                  盈亏比例
                </div>
                <div className="font-mono font-semibold" style={{ color: action.pnl && action.pnl >= 0 ? '#0ECB81' : '#F6465D' }}>
                  {action.pnl !== undefined && action.price && action.quantity ? 
                    `${action.pnl >= 0 ? '+' : ''}${((action.pnl / (action.price * action.quantity)) * 100).toFixed(2)}%` : '-'}
                </div>
                {action.pnl && action.price && action.quantity && (
                  <div className="text-xs mt-0.5" style={{ color: '#848E9C' }}>
                    {Math.abs((action.pnl / (action.price * action.quantity)) * 100).toFixed(2)}%收益率
                  </div>
                )}
              </>
            ) : isTrailingStop ? (
              <>
                <div className="text-xs mb-1" style={{ color: '#2196F3' }}>
                  Callback Rate
                </div>
                <div className="font-mono font-semibold" style={{ color: '#2196F3' }}>
                  {action.callback_rate !== undefined ? `${action.callback_rate}%` : '-'}
                </div>
              </>
            ) : isHold || isWait ? (
              <>
                <div className="text-xs mb-1" style={{ color: '#848E9C' }}>
                  {isHold ? '持有时间' : '等待时间'}
                </div>
                <div className="font-mono font-semibold" style={{ color: '#848E9C' }}>
                  {action.timestamp ? `${Math.floor((new Date().getTime() - new Date(action.timestamp).getTime()) / 60000)}分钟` : '-'}
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
                {action.take_profit && action.price && (
                  <div className="text-xs mt-0.5" style={{ color: '#848E9C' }}>
                    {calcPctChange(action.price, action.take_profit, isLong)}
                  </div>
                )}
              </>
            )}
          </div>

          {/* Leverage or Quantity or Confidence */}
          <div className="text-center">
            <div className="text-xs mb-1" style={{ color: '#848E9C' }}>
              {isOpen ? t('leverage', language) : isHold || isWait ? '置信度' : isUpdateStopLoss ? '调整幅度' : t('quantity', language)}
            </div>
            <div className="font-mono font-semibold" style={{ color: isOpen ? '#F0B90B' : isHold || isWait ? getConfidenceColor(action.confidence) : isUpdateStopLoss ? '#FF9800' : '#EAECEF' }}>
              {isOpen ? `${action.leverage}x` : isHold || isWait ? `${action.confidence || 0}%` : isUpdateStopLoss ? 
                (action.stop_loss && action.new_stop_loss ? 
                  `${action.new_stop_loss > action.stop_loss ? '↑' : '↓'}${Math.abs(((action.new_stop_loss - action.stop_loss) / action.stop_loss) * 100).toFixed(1)}%` : '-') : 
                (action.quantity !== undefined && action.quantity !== null && action.quantity > 0 ? formatPrice(action.quantity) : '-')}
            </div>
          </div>
        </div>
      )}

      {/* Risk/Reward Ratio for open positions */}
      {isOpen && action.stop_loss && action.take_profit && action.price && (
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
