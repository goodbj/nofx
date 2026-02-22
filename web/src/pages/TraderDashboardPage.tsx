import { useEffect, useState, useRef } from 'react'
import { mutate } from 'swr'
import { useAuth } from '../contexts/AuthContext'
import { api } from '../lib/api'
import { ChartTabs } from '../components/ChartTabs'
import { DecisionCard } from '../components/DecisionCard'
import { PositionHistory } from '../components/PositionHistory'

import { TableExporter, ExportFormat } from '../utils/tableExporter'
import { PunkAvatar, getTraderAvatar } from '../components/PunkAvatar'
import { confirmToast, notify } from '../lib/notify'
import { t, type Language } from '../i18n/translations'
import {
  LogOut,
  Loader2,
  Eye,
  EyeOff,
  Copy,
  Check,
  Bot,
  Zap,
  Code,
  Clipboard,
  ClipboardPaste,
  Send,
  RefreshCw,
} from 'lucide-react'
import { DeepVoidBackground } from '../components/DeepVoidBackground'
import type {
  SystemStatus,
  AccountInfo,
  Position,
  DecisionRecord,
  Statistics,
  TraderInfo,
  Exchange,
  HistoricalPosition,
} from '../types'

// --- Helper Functions ---

// 获取友好的AI模型名称
function getModelDisplayName(modelId: string): string {
  switch (modelId.toLowerCase()) {
    case 'deepseek':
      return 'DeepSeek'
    case 'qwen':
      return 'Qwen'
    case 'claude':
      return 'Claude'
    default:
      return modelId.toUpperCase()
  }
}

// Helper function to get exchange display name from exchange ID (UUID)
function getExchangeDisplayNameFromList(
  exchangeId: string | undefined,
  exchanges: Exchange[] | undefined
): string {
  if (!exchangeId) return 'Unknown'
  const exchange = exchanges?.find((e) => e.id === exchangeId)
  if (!exchange) return exchangeId.substring(0, 8).toUpperCase() + '...'
  const typeName = exchange.exchange_type?.toUpperCase() || exchange.name
  return exchange.account_name
    ? `${typeName} - ${exchange.account_name}`
    : typeName
}

// Helper function to get exchange type from exchange ID (UUID) - for kline charts
function getExchangeTypeFromList(
  exchangeId: string | undefined,
  exchanges: Exchange[] | undefined
): string {
  if (!exchangeId) return 'binance'
  const exchange = exchanges?.find((e) => e.id === exchangeId)
  if (!exchange) return 'binance' // Default to binance for charts
  return exchange.exchange_type?.toLowerCase() || 'binance'
}

// Helper function to check if exchange is a perp-dex type (wallet-based)
function isPerpDexExchange(exchangeType: string | undefined): boolean {
  if (!exchangeType) return false
  const perpDexTypes = ['hyperliquid', 'lighter', 'aster']
  return perpDexTypes.includes(exchangeType.toLowerCase())
}

// Helper function to get wallet address for perp-dex exchanges
function getWalletAddress(exchange: Exchange | undefined): string | undefined {
  if (!exchange) return undefined
  const type = exchange.exchange_type?.toLowerCase()
  switch (type) {
    case 'hyperliquid':
      return exchange.hyperliquidWalletAddr
    case 'lighter':
      return exchange.lighterWalletAddr
    case 'aster':
      return exchange.asterSigner
    default:
      return undefined
  }
}

// Helper function to truncate wallet address for display
function truncateAddress(address: string, startLen = 6, endLen = 4): string {
  if (address.length <= startLen + endLen + 3) return address
  return `${address.slice(0, startLen)}...${address.slice(-endLen)}`
}

// --- Components ---

interface TraderDashboardPageProps {
  selectedTrader?: TraderInfo
  traders?: TraderInfo[]
  tradersError?: Error
  selectedTraderId?: string
  onTraderSelect: (traderId: string) => void
  onNavigateToTraders: () => void
  status?: SystemStatus
  account?: AccountInfo
  positions?: Position[]
  decisions?: DecisionRecord[]
  decisionsLimit: number
  onDecisionsLimitChange: (limit: number) => void
  stats?: Statistics
  lastUpdate: string
  language: Language
  exchanges?: Exchange[]
}

export function TraderDashboardPage({
  selectedTrader,
  status,
  account,
  positions,
  decisions,
  decisionsLimit,
  onDecisionsLimitChange,
  lastUpdate,
  language,
  traders,
  tradersError,
  selectedTraderId,
  onTraderSelect,
  onNavigateToTraders,
  exchanges,
}: TraderDashboardPageProps) {
  const { token } = useAuth();
  const [closingPosition, setClosingPosition] = useState<string | null>(null)
  const [selectedChartSymbol, setSelectedChartSymbol] = useState<
    string | undefined
  >(undefined)
  const [chartUpdateKey, setChartUpdateKey] = useState<number>(0)
  const chartSectionRef = useRef<HTMLDivElement>(null)
  const [showWalletAddress, setShowWalletAddress] = useState<boolean>(false)
  const [copiedAddress, setCopiedAddress] = useState<boolean>(false)
  const [startDate, setStartDate] = useState<string>('')
  const [endDate, setEndDate] = useState<string>('')
  const [isManualDecisionLoading, setIsManualDecisionLoading] =
    useState<boolean>(false)
  const [manualScanCooldown, setManualScanCooldown] = useState<boolean>(false)
  const [startButtonCooldown, setStartButtonCooldown] = useState<boolean>(false)
  const [isSystemScanning, setIsSystemScanning] = useState<boolean>(false)
  const [nextScanCountdown, setNextScanCountdown] = useState<number>(0) // 🔥 新增：下次扫描倒计时（秒）
  const [isDelayedByManual, setIsDelayedByManual] = useState<boolean>(false) // 🔥 新增：是否被手动扫描延迟
  
  // 无限滚动相关状态
  const [allDecisions, setAllDecisions] = useState<DecisionRecord[]>(decisions || []);
  const [decisionsPage, setDecisionsPage] = useState<number>(0);
  const [hasMoreDecisions, setHasMoreDecisions] = useState<boolean>(true);
  const [loadingMoreDecisions, setLoadingMoreDecisions] = useState<boolean>(false);
  const decisionsContainerRef = useRef<HTMLDivElement>(null);

  // AI Semi-Auto Workflow States
  const [manualAIDecision, setManualAIDecision] = useState<string>('')
  const [submitAILoading, setSubmitAILoading] = useState(false)
  const [submitAIResult, setSubmitAIResult] = useState<string>('')
  const [promptPreview, setPromptPreview] = useState<{
    system_prompt: string
    user_prompt?: string
    prompt_variant: string
    config_summary: Record<string, unknown>
  } | null>(null)
  const [isLoadingPrompt, setIsLoadingPrompt] = useState(false)

  // 获取 AI 提示词预览
  const fetchPromptPreview = async () => {
    console.log('fetchPromptPreview called, selectedTraderId:', selectedTraderId)
    
    if (!selectedTraderId) {
      console.log('No selectedTraderId, showing warning')
      notify.warning(
        language === 'zh'
          ? '请先选择一个交易员'
          : 'Please select a trader first'
      )
      return
    }

    if (!token) {
      console.log('No token found from auth context')
      notify.error(
        language === 'zh'
          ? '认证令牌不存在，请重新登录'
          : 'Authentication token not found, please log in again'
      )
      return
    }
    console.log('Token found from auth context, proceeding with API call')

    setIsLoadingPrompt(true)
    try {
      const response = await api.generateFullPrompt(selectedTraderId)
      
      console.log('API response received, status:', response.success ? 200 : 400)

      if (response.success) {
        const data = response.data!;
        console.log('API response data:', data);
        setPromptPreview({
          system_prompt: data.system_prompt || '',
          user_prompt: data.user_prompt || '',
          prompt_variant: 'balanced',
          config_summary: { trader: selectedTrader?.trader_name },
        });
        notify.success(
          language === 'zh'
            ? 'Prompt 生成成功（含实时数据）'
            : 'Real-time prompt generated successfully'
        );
      } else {
        notify.error(response.message || 'Failed to generate prompt');
      }
    } catch (err) {
      console.error('Error fetching prompt preview:', err)
      notify.error(
        language === 'zh' ? '生成Prompt时出错' : 'Error generating prompt'
      )
    } finally {
      setIsLoadingPrompt(false)
    }
  }

  // 提交 AI 决策
  const handleSubmitAIDecision = async () => {
    if (!selectedTraderId) {
      notify.warning(
        language === 'zh' ? '请选择交易员' : 'Please select trader'
      )
      return
    }

    if (!token) {
      notify.error(
        language === 'zh'
          ? '认证令牌不存在，请重新登录'
          : 'Authentication token not found, please log in again'
      )
      return
    }

    setSubmitAILoading(true)
    setSubmitAIResult('')

    try {
      const response = await api.submitAIDecision(selectedTraderId, manualAIDecision)

      if (!response.success) {
        setSubmitAIResult(
          `❌ Error: ${response.message || JSON.stringify(response.data, null, 2)}`
        )
        notify.error(language === 'zh' ? '提交失败' : 'Submission failed')
      } else {
        setSubmitAIResult(
          `✅ ${language === 'zh' ? '执行完成！' : 'Execution completed!'}\n\n${JSON.stringify(response.data, null, 2)}`
        )
        notify.success(
          language === 'zh' ? '执行完成！' : 'Execution completed!'
        )

        // 刷新相关数据，确保最新决策出现在列表中
        await Promise.all([
          mutate(`positions-${selectedTraderId}`),
          mutate(`account-${selectedTraderId}`),
          mutate(`decisions-${selectedTraderId}`), // 这会刷新最近决策列表
          mutate(`position-history-${selectedTraderId}`),
        ])
      }
    } catch (error: unknown) {
      const errorMessage = 
        error instanceof Error 
          ? error.message 
          : language === 'zh' 
            ? '网络错误或未知错误' 
            : 'Network error or unknown error';
            
      setSubmitAIResult(
        `❌ ${language === 'zh' ? '错误：' : 'Error: '}${errorMessage}`
      )
      notify.error(errorMessage)
    } finally {
      setSubmitAILoading(false)
    }
  }

  // 删除决策记录
  const handleDeleteDecision = async (decisionId: number, traderId: string) => {
    if (!token) {
      notify.error(
        language === 'zh'
          ? '认证令牌不存在，请重新登录'
          : 'Authentication token not found, please log in again'
      )
      return
    }

    try {
      const response = await api.deleteDecision(decisionId, traderId)
      
      if (!response.success) {
        notify.error(response.message || (language === 'zh' ? '删除决策失败' : 'Failed to delete decision'))
        return
      }

      notify.success(language === 'zh' ? '决策删除成功' : 'Decision deleted successfully')
      
      // 刷新决策列表
      await mutate(`decisions-${selectedTraderId}`)
    } catch (error: unknown) {
      const errorMessage = 
        error instanceof Error 
          ? error.message 
          : language === 'zh' 
            ? '网络错误或未知错误' 
            : 'Network error or unknown error';
            
      notify.error(errorMessage)
    }
  }

  // 加载更多决策记录
  const loadMoreDecisions = async () => {
    if (loadingMoreDecisions || !hasMoreDecisions || !selectedTraderId) return;
    
    setLoadingMoreDecisions(true);
    
    try {
      const nextPage = decisionsPage + 1;
      const moreDecisions = await api.getLatestDecisions(
        selectedTraderId,
        decisionsLimit,
        nextPage * decisionsLimit
      );
      
      if (moreDecisions && moreDecisions.length > 0) {
        setAllDecisions(prev => [...prev, ...moreDecisions]);
        setDecisionsPage(nextPage);
        // 如果返回的数量小于请求的数量，说明没有更多数据了
        if (moreDecisions.length < decisionsLimit) {
          setHasMoreDecisions(false);
        }
      } else {
        setHasMoreDecisions(false);
      }
    } catch (error) {
      console.error('加载更多决策记录失败:', error);
      notify.error(
        language === 'zh' ? '加载更多决策记录失败' : 'Failed to load more decisions'
      );
    } finally {
      setLoadingMoreDecisions(false);
    }
  };

  // 处理滚动事件以实现无限滚动
  useEffect(() => {
    const container = decisionsContainerRef.current;
    if (!container) return;

    const handleScroll = () => {
      const { scrollTop, scrollHeight, clientHeight } = container;
      // 当滚动到底部附近时（距离底部100px内）加载更多
      if (scrollHeight - scrollTop <= clientHeight + 100) {
        loadMoreDecisions();
      }
    };

    container.addEventListener('scroll', handleScroll);
    return () => container.removeEventListener('scroll', handleScroll);
  }, [selectedTraderId, decisionsPage, hasMoreDecisions, loadingMoreDecisions]);

  // 当交易员切换时重置分页状态
  useEffect(() => {
    setAllDecisions(decisions || []);
    setDecisionsPage(0);
    setHasMoreDecisions(true);
  }, [selectedTraderId, decisions]);

  // Current positions pagination
  const [positionsPageSize, setPositionsPageSize] = useState<number>(20)
  const [positionsCurrentPage, setPositionsCurrentPage] = useState<number>(1)

  // Calculate paginated positions
  const totalPositions = positions?.length || 0
  const totalPositionPages = Math.ceil(totalPositions / positionsPageSize)
  const paginatedPositions =
    positions?.slice(
      (positionsCurrentPage - 1) * positionsPageSize,
      positionsCurrentPage * positionsPageSize
    ) || []

  // Reset page when positions change
  useEffect(() => {
    setPositionsCurrentPage(1)
  }, [selectedTraderId, positionsPageSize])

  // 轮询AI分析状态（每500ms检查一次is_executing）
  useEffect(() => {
    if (!selectedTraderId || !status?.is_running) {
      setIsSystemScanning(false)
      setNextScanCountdown(0)
      setIsDelayedByManual(false)
      return
    }

    const pollInterval = setInterval(async () => {
      try {
        // 只在trader运行中时轮询
        const currentStatus = await api.getTraderStatus(selectedTraderId)

        // 检查 is_executing 字段（系统正在执行决策）
        if (currentStatus && 'is_executing' in currentStatus) {
          setIsSystemScanning(currentStatus.is_executing === true)
        }

        // 🔥 新增：更新倒计时信息
        if (currentStatus && 'seconds_until_next_scan' in currentStatus) {
          setNextScanCountdown(currentStatus.seconds_until_next_scan || 0)
        }

        // 🔥 新增：更新延迟状态
        if (currentStatus && 'is_delayed_by_manual' in currentStatus) {
          setIsDelayedByManual(currentStatus.is_delayed_by_manual === true)
        }
      } catch (error) {
        // 静默失败，不影响用户体验
        console.debug('Status poll failed:', error)
      }
    }, 500) // 每500ms轮询一次

    return () => clearInterval(pollInterval)
  }, [selectedTraderId, status?.is_running])

  // Get current exchange info for perp-dex wallet display
  const currentExchange = exchanges?.find(
    (e) => e.id === selectedTrader?.exchange_id
  )
  const walletAddress = getWalletAddress(currentExchange)
  const isPerpDex = isPerpDexExchange(currentExchange?.exchange_type)

  // Copy wallet address to clipboard
  const handleCopyAddress = async () => {
    if (!walletAddress) return
    try {
      await navigator.clipboard.writeText(walletAddress)
      setCopiedAddress(true)
      setTimeout(() => setCopiedAddress(false), 2000)
    } catch (err) {
      console.error('Failed to copy address:', err)
    }
  }

  // Handle symbol click from Decision Card
  const handleSymbolClick = (symbol: string) => {
    // Set the selected symbol
    setSelectedChartSymbol(symbol)
    // Scroll to chart section
    setTimeout(() => {
      chartSectionRef.current?.scrollIntoView({
        behavior: 'smooth',
        block: 'start',
      })
    }, 100)
  }

  // 平仓操作
  const handleClosePosition = async (symbol: string, side: string) => {
    if (!selectedTraderId) return

    const confirmMsg =
      language === 'zh'
        ? `确定要平仓 ${symbol} ${side === 'LONG' ? '多仓' : '空仓'} 吗？`
        : `Are you sure you want to close ${symbol} ${side === 'LONG' ? 'LONG' : 'SHORT'} position?`

    const confirmed = await confirmToast(confirmMsg, {
      title: language === 'zh' ? '确认平仓' : 'Confirm Close',
      okText: language === 'zh' ? '确认' : 'Confirm',
      cancelText: language === 'zh' ? '取消' : 'Cancel',
    })

    if (!confirmed) return

    setClosingPosition(symbol)
    try {
      await api.closePosition(selectedTraderId, symbol, side)
      notify.success(
        language === 'zh' ? '平仓成功' : 'Position closed successfully'
      )
      // 使用 SWR mutate 刷新数据而非重新加载页面
      await Promise.all([
        mutate(`positions-${selectedTraderId}`),
        mutate(`account-${selectedTraderId}`),
        mutate(`position-history-${selectedTraderId}`),
      ])
    } catch (err: unknown) {
      const errorMsg =
        err instanceof Error
          ? err.message
          : language === 'zh'
            ? '平仓失败'
            : 'Failed to close position'
      notify.error(errorMsg)
    } finally {
      setClosingPosition(null)
    }
  }

  // If API failed with error, show empty state (likely backend not running)
  if (tradersError) {
    return (
      <div className="flex items-center justify-center min-h-[60vh] relative z-10">
        <div className="text-center max-w-md mx-auto px-6">
          <div
            className="w-24 h-24 mx-auto mb-6 rounded-full flex items-center justify-center nofx-glass"
            style={{
              background: 'rgba(240, 185, 11, 0.1)',
              borderColor: 'rgba(240, 185, 11, 0.3)',
            }}
          >
            <svg
              className="w-12 h-12 text-nofx-gold"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
              />
            </svg>
          </div>
          <h2 className="text-2xl font-bold mb-3 text-nofx-text-main">
            {language === 'zh' ? '无法连接到服务器' : 'Connection Failed'}
          </h2>
          <p className="text-base mb-6 text-nofx-text-muted">
            {language === 'zh'
              ? '请确认后端服务已启动。'
              : 'Please check if the backend service is running.'}
          </p>
          <button
            onClick={() => window.location.reload()}
            className="px-6 py-3 rounded-lg font-semibold transition-all hover:scale-105 active:scale-95 nofx-glass border border-nofx-gold/30 text-nofx-gold hover:bg-nofx-gold/10"
          >
            {language === 'zh' ? '重试' : 'Retry'}
          </button>
        </div>
      </div>
    )
  }

  // If traders is loaded and empty, show empty state
  if (traders && traders.length === 0) {
    return (
      <div className="flex items-center justify-center min-h-[60vh] relative z-10">
        <div className="text-center max-w-md mx-auto px-6">
          <div
            className="w-24 h-24 mx-auto mb-6 rounded-full flex items-center justify-center nofx-glass"
            style={{
              background: 'rgba(240, 185, 11, 0.1)',
              borderColor: 'rgba(240, 185, 11, 0.3)',
            }}
          >
            <svg
              className="w-12 h-12 text-nofx-gold"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
              />
            </svg>
          </div>
          <h2 className="text-2xl font-bold mb-3 text-nofx-text-main">
            {t('dashboardEmptyTitle', language)}
          </h2>
          <p className="text-base mb-6 text-nofx-text-muted">
            {t('dashboardEmptyDescription', language)}
          </p>
          <button
            onClick={onNavigateToTraders}
            className="px-6 py-3 rounded-lg font-semibold transition-all hover:scale-105 active:scale-95 nofx-glass border border-nofx-gold/30 text-nofx-gold hover:bg-nofx-gold/10"
          >
            {t('goToTradersPage', language)}
          </button>
        </div>
      </div>
    )
  }

  // If traders is still loading or selectedTrader is not ready, show skeleton
  if (!selectedTrader) {
    return (
      <div className="space-y-6 relative z-10">
        <div className="nofx-glass p-6 animate-pulse">
          <div className="h-8 w-48 mb-3 bg-nofx-bg/50 rounded"></div>
          <div className="flex gap-4">
            <div className="h-4 w-32 bg-nofx-bg/50 rounded"></div>
            <div className="h-4 w-24 bg-nofx-bg/50 rounded"></div>
            <div className="h-4 w-28 bg-nofx-bg/50 rounded"></div>
          </div>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="nofx-glass p-5 animate-pulse">
              <div className="h-4 w-24 mb-3 bg-nofx-bg/50 rounded"></div>
              <div className="h-8 w-32 bg-nofx-bg/50 rounded"></div>
            </div>
          ))}
        </div>
        <div className="nofx-glass p-6 animate-pulse">
          <div className="h-6 w-40 mb-4 bg-nofx-bg/50 rounded"></div>
          <div className="h-64 w-full bg-nofx-bg/50 rounded"></div>
        </div>
      </div>
    )
  }

  return (
    <DeepVoidBackground className="min-h-screen pb-12" disableAnimation>
      <div className="w-full px-4 md:px-8 relative z-10 pt-6">
        {/* Trader Header */}
        <div
          className="mb-6 rounded-lg p-6 animate-scale-in nofx-glass group"
          style={{
            background:
              'linear-gradient(135deg, rgba(15, 23, 42, 0.6) 0%, rgba(15, 23, 42, 0.4) 100%)',
          }}
        >
          <div className="flex items-start justify-between mb-4">
            <h2 className="text-2xl font-bold flex items-center gap-4 text-nofx-text-main">
              <div className="relative">
                <PunkAvatar
                  seed={getTraderAvatar(
                    selectedTrader.trader_id,
                    selectedTrader.trader_name
                  )}
                  size={56}
                  className="rounded-xl border-2 border-nofx-gold/30 shadow-[0_0_15px_rgba(240,185,11,0.2)]"
                />
                <div className="absolute -bottom-1 -right-1 w-4 h-4 bg-nofx-green rounded-full border-2 border-[#0B0E11] shadow-[0_0_8px_rgba(14,203,129,0.8)] animate-pulse" />
              </div>
              <div className="flex flex-col">
                <span className="text-3xl tracking-tight text-nofx-text font-semibold">
                  {selectedTrader.trader_name}
                </span>
                <span className="text-xs font-mono text-nofx-text-muted opacity-60 flex items-center gap-2">
                  <div className="w-1.5 h-1.5 bg-nofx-gold rounded-full" />
                  ID: {selectedTrader.trader_id.slice(0, 8)}...
                </span>
              </div>
            </h2>

            <div className="flex items-center gap-4">
              {/* Trader Selector */}
              {traders && traders.length > 0 && (
                <div className="flex items-center gap-2 nofx-glass px-1 py-1 rounded-lg border border-white/5">
                  <select
                    value={selectedTraderId}
                    onChange={(e) => onTraderSelect(e.target.value)}
                    className="bg-transparent text-sm font-medium cursor-pointer transition-colors text-nofx-text-main focus:outline-none px-2 py-1"
                  >
                    {traders.map((trader) => (
                      <option
                        key={trader.trader_id}
                        value={trader.trader_id}
                        className="bg-[#0B0E11]"
                      >
                        {trader.trader_name} ({trader.trader_id.slice(0, 8)}) {trader.is_running ? '🟢' : '🔴'}
                      </option>
                    ))}
                  </select>
                </div>
              )}

              {/* Wallet Address Display for Perp-DEX */}
              {exchanges && isPerpDex && (
                <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg nofx-glass border border-nofx-gold/20">
                  {walletAddress ? (
                    <>
                      <span className="text-xs font-mono text-nofx-gold">
                        {showWalletAddress
                          ? walletAddress
                          : truncateAddress(walletAddress)}
                      </span>
                      <button
                        type="button"
                        onClick={() => setShowWalletAddress(!showWalletAddress)}
                        className="p-1 rounded hover:bg-white/10 transition-colors"
                        title={
                          showWalletAddress
                            ? language === 'zh'
                              ? '隐藏地址'
                              : 'Hide address'
                            : language === 'zh'
                              ? '显示完整地址'
                              : 'Show full address'
                        }
                      >
                        {showWalletAddress ? (
                          <EyeOff className="w-3.5 h-3.5 text-nofx-text-muted" />
                        ) : (
                          <Eye className="w-3.5 h-3.5 text-nofx-text-muted" />
                        )}
                      </button>
                      <button
                        type="button"
                        onClick={handleCopyAddress}
                        className="p-1 rounded hover:bg-white/10 transition-colors"
                        title={language === 'zh' ? '复制地址' : 'Copy address'}
                      >
                        {copiedAddress ? (
                          <Check className="w-3.5 h-3.5 text-nofx-green" />
                        ) : (
                          <Copy className="w-3.5 h-3.5 text-nofx-text-muted" />
                        )}
                      </button>
                    </>
                  ) : (
                    <span className="text-xs text-nofx-text-muted">
                      {language === 'zh'
                        ? '未配置地址'
                        : 'No address configured'}
                    </span>
                  )}
                </div>
              )}
            </div>
          </div>
          <div className="flex items-center gap-6 text-sm flex-wrap text-nofx-text-muted font-mono pl-2">
            <span className="flex items-center gap-2">
              <span className="opacity-60">AI Model:</span>
              <span
                className="font-bold px-2 py-0.5 rounded text-xs tracking-wide"
                style={{
                  background: selectedTrader.ai_model.includes('qwen')
                    ? 'rgba(192, 132, 252, 0.15)'
                    : 'rgba(96, 165, 250, 0.15)',
                  color: selectedTrader.ai_model.includes('qwen')
                    ? '#c084fc'
                    : '#60a5fa',
                  border: `1px solid ${selectedTrader.ai_model.includes('qwen') ? '#c084fc' : '#60a5fa'}40`,
                }}
              >
                {getModelDisplayName(
                  selectedTrader.ai_model.split('_').pop() ||
                    selectedTrader.ai_model
                )}
              </span>
            </span>
            <span className="w-px h-3 bg-white/10 hidden md:block" />
            <span className="flex items-center gap-2">
              <span className="opacity-60">Exchange:</span>
              <span className="text-nofx-text-main font-semibold">
                {getExchangeDisplayNameFromList(
                  selectedTrader.exchange_id,
                  exchanges
                )}
              </span>
            </span>
            <span className="w-px h-3 bg-white/10 hidden md:block" />
            <span className="flex items-center gap-2">
              <span className="opacity-60">Strategy:</span>
              <span className="text-nofx-gold font-semibold tracking-wide">
                {selectedTrader.strategy_name || 'No Strategy'}
              </span>
            </span>
            {status && (
              <div className="hidden md:contents">
                <span className="w-px h-3 bg-white/10" />
                <span>
                  Cycles:{' '}
                  <span className="text-nofx-text-main">
                    {status.call_count}
                  </span>
                </span>
                <span className="w-px h-3 bg-white/10" />
                <span>
                  Runtime:{' '}
                  <span className="text-nofx-text-main">
                    {status.runtime_minutes} min
                  </span>
                </span>
              </div>
            )}
          </div>
        </div>

        {/* Debug Info */}
        {account && (
          <div className="mb-4 px-3 py-1.5 rounded bg-black/40 border border-white/5 text-[10px] font-mono text-nofx-text-muted flex justify-between items-center opacity-60 hover:opacity-100 transition-opacity">
            <span>SYSTEM_STATUS::ONLINE</span>
            <div className="flex gap-4">
              <span>LAST_UPDATE::{lastUpdate}</span>
              <span>EQ::{account?.total_equity?.toFixed(2)}</span>
              <span>PNL::{account?.total_pnl?.toFixed(2)}</span>
            </div>
          </div>
        )}

        {/* Account Overview */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
          <StatCard
            title={t('totalEquity', language)}
            value={`${account?.total_equity?.toFixed(2) || '0.00'}`}
            unit="USDT"
            change={account?.total_pnl_pct || 0}
            positive={(account?.total_pnl ?? 0) > 0}
            icon="💰"
          />
          <StatCard
            title={t('availableBalance', language)}
            value={`${account?.available_balance?.toFixed(2) || '0.00'}`}
            unit="USDT"
            subtitle={`${account?.available_balance && account?.total_equity ? ((account.available_balance / account.total_equity) * 100).toFixed(1) : '0.0'}% ${t('free', language)}`}
            icon="💳"
          />
          <StatCard
            title={t('totalPnL', language)}
            value={`${account?.total_pnl !== undefined && account.total_pnl >= 0 ? '+' : ''}${account?.total_pnl?.toFixed(2) || '0.00'}`}
            unit="USDT"
            change={account?.total_pnl_pct || 0}
            positive={(account?.total_pnl ?? 0) >= 0}
            icon="📈"
          />
          <StatCard
            title={t('positions', language)}
            value={`${account?.position_count || 0}`}
            unit="ACTIVE"
            subtitle={`${t('margin', language)}: ${account?.margin_used_pct?.toFixed(1) || '0.0'}%`}
            icon="📊"
          />
        </div>

        {/* Main Content Area */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
          {/* Left Column: Charts + Positions + AI Semi-Auto Workflow */}
          <div className="space-y-6">
            {/* Chart Tabs (Equity / K-line) */}
            <div
              ref={chartSectionRef}
              className="chart-container animate-slide-in scroll-mt-32 backdrop-blur-sm"
              style={{ animationDelay: '0.1s' }}
            >
              <ChartTabs
                traderId={selectedTrader.trader_id}
                selectedSymbol={selectedChartSymbol}
                updateKey={chartUpdateKey}
                exchangeId={getExchangeTypeFromList(
                  selectedTrader.exchange_id,
                  exchanges
                )}
              />
            </div>

            {/* Current Positions */}
            <div
              className="nofx-glass p-6 animate-slide-in relative overflow-hidden group"
              style={{ animationDelay: '0.15s' }}
            >
              <div className="absolute top-0 right-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
                <div className="w-24 h-24 rounded-full bg-blue-500 blur-3xl" />
              </div>
              <div className="flex items-center justify-between mb-5 relative z-10">
                <h2 className="text-lg font-bold flex items-center gap-2 text-nofx-text-main uppercase tracking-wide">
                  <span className="text-blue-500">◈</span>{' '}
                  {t('currentPositions', language)}
                </h2>
                {positions && positions.length > 0 && (
                  <div className="text-xs px-2 py-1 rounded bg-nofx-gold/10 text-nofx-gold border border-nofx-gold/20 font-mono shadow-[0_0_10px_rgba(240,185,11,0.1)]">
                    {positions.length} {t('active', language)}
                  </div>
                )}
              </div>
              {positions && positions.length > 0 ? (
                <div>
                  <div className="overflow-x-auto">
                    <table className="w-full text-xs">
                      <thead className="text-left border-b border-white/5">
                        <tr>
                          <th className="px-1 pb-3 font-semibold text-nofx-text-muted whitespace-nowrap text-left">
                            {t('symbol', language)}
                          </th>
                          <th className="px-1 pb-3 font-semibold text-nofx-text-muted whitespace-nowrap text-center">
                            {t('side', language)}
                          </th>
                          <th className="px-1 pb-3 font-semibold text-nofx-text-muted whitespace-nowrap text-center">
                            {language === 'zh' ? '操作' : 'Action'}
                          </th>
                          <th
                            className="px-1 pb-3 font-semibold text-nofx-text-muted whitespace-nowrap text-right hidden md:table-cell"
                            title={t('entryPrice', language)}
                          >
                            {language === 'zh' ? '入场价' : 'Entry'}
                          </th>
                          <th
                            className="px-1 pb-3 font-semibold text-nofx-text-muted whitespace-nowrap text-right hidden md:table-cell"
                            title={t('markPrice', language)}
                          >
                            {language === 'zh' ? '标记价' : 'Mark'}
                          </th>
                          <th
                            className="px-1 pb-3 font-semibold text-nofx-text-muted whitespace-nowrap text-right"
                            title={t('quantity', language)}
                          >
                            {language === 'zh' ? '数量' : 'Qty'}
                          </th>
                          <th
                            className="px-1 pb-3 font-semibold text-nofx-text-muted whitespace-nowrap text-right hidden md:table-cell"
                            title={t('positionValue', language)}
                          >
                            {language === 'zh' ? '价值' : 'Value'}
                          </th>
                          <th
                            className="px-1 pb-3 font-semibold text-nofx-text-muted whitespace-nowrap text-center hidden md:table-cell"
                            title={t('leverage', language)}
                          >
                            {language === 'zh' ? '杠杆' : 'Lev.'}
                          </th>
                          <th
                            className="px-1 pb-3 font-semibold text-nofx-text-muted whitespace-nowrap text-right"
                            title={t('unrealizedPnL', language)}
                          >
                            {language === 'zh'
                              ? '盈亏（回报率）'
                              : 'PnL (Return)'}
                          </th>
                          <th
                            className="px-1 pb-3 font-semibold text-nofx-text-muted whitespace-nowrap text-right hidden md:table-cell"
                            title={t('liqPrice', language)}
                          >
                            {language === 'zh' ? '强平价' : 'Liq.'}
                          </th>
                        </tr>
                      </thead>
                      <tbody>
                        {paginatedPositions.map((pos, i) => (
                          <tr
                            key={i}
                            className="border-b border-white/5 last:border-0 transition-all hover:bg-white/5 cursor-pointer group/row"
                            onClick={() => {
                              setSelectedChartSymbol(pos.symbol)
                              setChartUpdateKey(Date.now())
                              if (chartSectionRef.current) {
                                chartSectionRef.current.scrollIntoView({
                                  behavior: 'smooth',
                                  block: 'start',
                                })
                              }
                            }}
                          >
                            <td className="px-1 py-3 font-mono font-semibold whitespace-nowrap text-left text-nofx-text-main group-hover/row:text-white transition-colors">
                              {pos.symbol}
                            </td>
                            <td className="px-1 py-3 whitespace-nowrap text-center">
                              <span
                                className={`px-1.5 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider ${pos.side === 'long' ? 'bg-nofx-green/10 text-nofx-green shadow-[0_0_8px_rgba(14,203,129,0.2)]' : 'bg-nofx-red/10 text-nofx-red shadow-[0_0_8px_rgba(246,70,93,0.2)]'}`}
                              >
                                {t(
                                  pos.side === 'long' ? 'long' : 'short',
                                  language
                                )}
                              </span>
                            </td>
                            <td className="px-1 py-3 whitespace-nowrap text-center">
                              <button
                                type="button"
                                onClick={(e) => {
                                  e.stopPropagation()
                                  handleClosePosition(
                                    pos.symbol,
                                    pos.side.toUpperCase()
                                  )
                                }}
                                disabled={closingPosition === pos.symbol}
                                className="inline-flex items-center gap-1 px-2 py-1 rounded text-[10px] font-semibold transition-all hover:scale-105 disabled:opacity-50 disabled:cursor-not-allowed mx-auto bg-nofx-red/10 text-nofx-red border border-nofx-red/30 hover:bg-nofx-red/20"
                                title={
                                  language === 'zh' ? '平仓' : 'Close Position'
                                }
                              >
                                {closingPosition === pos.symbol ? (
                                  <Loader2 className="w-3 h-3 animate-spin" />
                                ) : (
                                  <LogOut className="w-3 h-3" />
                                )}
                                {language === 'zh' ? '平仓' : 'Close'}
                              </button>
                            </td>
                            <td className="px-1 py-3 font-mono whitespace-nowrap text-right text-nofx-text-main hidden md:table-cell">
                              {pos.entry_price.toFixed(4)}
                            </td>
                            <td className="px-1 py-3 font-mono whitespace-nowrap text-right text-nofx-text-main hidden md:table-cell">
                              {pos.mark_price.toFixed(4)}
                            </td>
                            <td className="px-1 py-3 font-mono whitespace-nowrap text-right text-nofx-text-main">
                              {pos.quantity.toFixed(4)}
                            </td>
                            <td className="px-1 py-3 font-mono font-bold whitespace-nowrap text-right text-nofx-text-main hidden md:table-cell">
                              {(pos.quantity * pos.mark_price).toFixed(2)}
                            </td>
                            <td className="px-1 py-3 font-mono whitespace-nowrap text-center text-nofx-gold hidden md:table-cell">
                              {pos.leverage}x
                            </td>
                            <td className="px-1 py-3 font-mono whitespace-nowrap text-right">
                              <div className="flex items-center gap-1">
                                <span
                                  className={`font-bold ${pos.unrealized_pnl >= 0 ? 'text-nofx-green shadow-nofx-green' : 'text-nofx-red shadow-nofx-red'}`}
                                  style={{
                                    textShadow:
                                      pos.unrealized_pnl >= 0
                                        ? '0 0 10px rgba(14,203,129,0.3)'
                                        : '0 0 10px rgba(246,70,93,0.3)',
                                  }}
                                >
                                  {pos.unrealized_pnl >= 0 ? '+' : ''}
                                  {pos.unrealized_pnl.toFixed(2)}
                                </span>
                                <span className="text-[10px] text-nofx-text-muted">
                                  {pos.entry_price &&
                                  pos.mark_price &&
                                  pos.entry_price !== 0
                                    ? `(${(((pos.mark_price - pos.entry_price) / pos.entry_price) * 100).toFixed(2)}%)`
                                    : '(0.00%)'}
                                </span>
                              </div>
                            </td>
                            <td className="px-1 py-3 font-mono whitespace-nowrap text-right text-nofx-text-muted hidden md:table-cell">
                              {pos.liquidation_price.toFixed(4)}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                  {/* Pagination footer */}
                  {totalPositions > 10 && (
                    <div className="flex flex-wrap items-center justify-between gap-3 pt-4 mt-4 text-xs border-t border-white/5 text-nofx-text-muted">
                      <span>
                        {language === 'zh'
                          ? `显示 ${paginatedPositions.length} / ${totalPositions} 个持仓`
                          : `Showing ${paginatedPositions.length} of ${totalPositions} positions`}
                      </span>
                      <div className="flex items-center gap-3">
                        <div className="flex items-center gap-2">
                          <span>
                            {language === 'zh' ? '每页' : 'Per page'}:
                          </span>
                          <select
                            value={positionsPageSize}
                            onChange={(e) =>
                              setPositionsPageSize(Number(e.target.value))
                            }
                            className="bg-black/40 border border-white/10 rounded px-2 py-1 text-xs text-nofx-text-main focus:outline-none focus:border-nofx-gold/50 transition-colors"
                          >
                            <option value={20}>20</option>
                            <option value={50}>50</option>
                            <option value={100}>100</option>
                          </select>
                        </div>
                        {totalPositionPages > 1 && (
                          <div className="flex items-center gap-1">
                            {[
                              '«',
                              '‹',
                              `${positionsCurrentPage} / ${totalPositionPages}`,
                              '›',
                              '»',
                            ].map((label, idx) => {
                              const isText = idx === 2
                              const isFirst = idx === 0
                              const isPrev = idx === 1
                              const isNext = idx === 3
                              const isLast = idx === 4
                              if (isText)
                                return (
                                  <span
                                    key={idx}
                                    className="px-3 text-nofx-text-main"
                                  >
                                    {label}
                                  </span>
                                )

                              let onClick = () => {}
                              let disabled = false

                              if (isFirst) {
                                onClick = () => setPositionsCurrentPage(1)
                                disabled = positionsCurrentPage === 1
                              }
                              if (isPrev) {
                                onClick = () =>
                                  setPositionsCurrentPage((p) =>
                                    Math.max(1, p - 1)
                                  )
                                disabled = positionsCurrentPage === 1
                              }
                              if (isNext) {
                                onClick = () =>
                                  setPositionsCurrentPage((p) =>
                                    Math.min(totalPositionPages, p + 1)
                                  )
                                disabled =
                                  positionsCurrentPage === totalPositionPages
                              }
                              if (isLast) {
                                onClick = () =>
                                  setPositionsCurrentPage(totalPositionPages)
                                disabled =
                                  positionsCurrentPage === totalPositionPages
                              }

                              return (
                                <button
                                  key={idx}
                                  onClick={onClick}
                                  disabled={disabled}
                                  className={`px-2 py-1 rounded transition-colors ${disabled ? 'opacity-30 cursor-not-allowed' : 'hover:bg-white/10 text-nofx-text-main bg-white/5'}`}
                                >
                                  {label}
                                </button>
                              )
                            })}
                          </div>
                        )}
                      </div>
                    </div>
                  )}
                </div>
              ) : (
                <div className="text-center py-16 text-nofx-text-muted opacity-60">
                  <div className="text-6xl mb-4 opacity-50 grayscale">📊</div>
                  <div className="text-lg font-semibold mb-2">
                    {t('noPositions', language)}
                  </div>
                  <div className="text-sm">
                    {t('noActivePositions', language)}
                  </div>
                </div>
              )}
            </div>

            {/* AI Semi-Auto Workflow Section */}
            <div
              className="nofx-glass p-6 animate-slide-in border-2 border-blue-500/30 rounded-xl"
              style={{ animationDelay: '0.25s' }}
            >
              <div className="flex items-center justify-between mb-5">
                <h2 className="text-lg font-bold flex items-center gap-2 text-nofx-text-main uppercase tracking-wide">
                  <Bot className="w-5 h-5 text-blue-400" />{' '}
                  {language === 'zh'
                    ? 'AI 半自动工作流'
                    : 'AI Semi-Auto Workflow'}
                </h2>
                <div className="text-xs px-3 py-1 rounded bg-nofx-gold/10 text-nofx-gold border border-nofx-gold/20 font-mono shadow-[0_0_10px_rgba(240,185,11,0.1)]">
                  {selectedTraderId
                    ? `${selectedTrader?.trader_name || 'Trader'}`
                    : 'No Trader Selected'}
                </div>
              </div>

              <div className="space-y-6">
                {/* Instructions */}
                <div className="bg-blue-900/20 p-4 rounded-lg border border-blue-700/50">
                  <p className="text-sm font-semibold text-blue-300 mb-2 flex items-center gap-2">
                    <Zap className="w-4 h-4" />{' '}
                    {language === 'zh' ? '操作说明' : 'Instructions'}:
                  </p>
                  <ol className="text-xs text-blue-100 space-y-1 list-decimal list-inside">
                    <li>
                      {language === 'zh'
                        ? '此功能允许您手动触发AI决策，适用于节省API Token的场景'
                        : 'This feature allows you to manually trigger AI decisions, suitable for saving API tokens'}
                    </li>
                    <li>
                      {language === 'zh'
                        ? '点击"获取提示词数据"获取包含实时数据的提示词'
                        : 'Click "Get Prompt Data" to get real-time data'}
                    </li>
                    <li>
                      {language === 'zh'
                        ? '将 Prompt 复制到外部AI服务（如DeepSeek网页版）并获取决策结果'
                        : 'Copy the Prompt to external AI service and get decision result'}
                    </li>
                    <li>
                      {language === 'zh'
                        ? '将AI返回的决策JSON粘贴到下方并提交'
                        : 'Paste AI decision JSON below and submit'}
                    </li>
                  </ol>
                </div>

                {/* Step 1: Prompt Display and Copy - Fixed height */}
                <div className="bg-gray-800/30 p-4 rounded-lg border border-gray-700">
                  <div className="flex justify-between items-center mb-2">
                    <label className="text-sm font-bold text-blue-400 flex items-center gap-2">
                      <Code className="w-4 h-4" />
                      {language === 'zh'
                        ? '完整 Prompt (System + User)'
                        : 'Full Prompt (System + User)'}
                    </label>
                    <div className="flex gap-2">
                      <div className="text-xs text-gray-400 flex items-center">
                        {language === 'zh'
                          ? `字符数: ${promptPreview ? (promptPreview.system_prompt + (promptPreview.user_prompt || '')).length : 0}`
                          : `Chars: ${promptPreview ? (promptPreview.system_prompt + (promptPreview.user_prompt || '')).length : 0}`}
                      </div>
                      <button
                        onClick={() => {
                          if (!promptPreview) return
                          const fullPrompt = `System Prompt:
${promptPreview.system_prompt}

User Prompt (包含实时行情、持仓、指标数据):
${promptPreview.user_prompt}`
                          navigator.clipboard.writeText(fullPrompt)
                          notify.success(
                            language === 'zh'
                              ? '已复制完整 Prompt'
                              : 'Full prompt copied'
                          )
                        }}
                        className={`text-xs px-3 py-1.5 rounded ${
                          promptPreview
                            ? 'bg-blue-600/30 text-blue-300 hover:bg-blue-600/50 border border-blue-500/30'
                            : 'bg-gray-600/30 text-gray-400 cursor-not-allowed'
                        } flex items-center gap-1 transition-colors`}
                        disabled={!promptPreview}
                      >
                        <Clipboard className="w-3.5 h-3.5" />
                        {language === 'zh' ? '复制' : 'Copy'}
                      </button>
                    </div>
                  </div>
                  <pre className="text-[10px] sm:text-[11px] text-gray-300 whitespace-pre-wrap font-mono bg-black/30 p-3 rounded border border-gray-700 overflow-auto max-h-[120px] min-h-[100px]">
                    {promptPreview
                      ? `System Prompt:
${promptPreview.system_prompt}

User Prompt (包含实时行情、持仓、指标数据):
${promptPreview.user_prompt}`
                      : language === 'zh'
                        ? '点击"获取提示词数据"按钮以获取完整提示词'
                        : 'Click "Get Prompt Data" to get full prompt'}
                  </pre>
                  {promptPreview && (
                    <div className="mt-2 text-xs text-gray-500 flex justify-between flex-wrap gap-2">
                      <span>
                        {language === 'zh'
                          ? '提示: 大多数AI模型上下文长度限制在 32K-128K tokens 之间'
                          : 'Note: Most AI models have context length limits of 32K-128K tokens'}
                      </span>
                      <span
                        className={`${(promptPreview.system_prompt + (promptPreview.user_prompt || '')).length > 50000 ? 'text-amber-400' : 'text-green-400'}`}
                      >
                        {(
                          promptPreview.system_prompt +
                          (promptPreview.user_prompt || '')
                        ).length > 50000
                          ? language === 'zh'
                            ? '⚠️ 字符数较多，可能超出模型限制'
                            : '⚠️ High char count, may exceed model limits'
                          : language === 'zh'
                            ? '✅ 字符数正常'
                            : '✅ Normal char count'}
                      </span>
                    </div>
                  )}
                  {/* Generate Prompt Button moved here */}
                  <div className="mt-3 flex justify-end gap-2">
                    <button
                      onClick={fetchPromptPreview}
                      disabled={isLoadingPrompt || !selectedTraderId}
                      title={
                        !selectedTraderId
                          ? language === 'zh'
                            ? '请先选择一个交易员'
                            : 'Please select a trader first'
                          : ''
                      }
                      className="px-4 py-2.5 rounded-lg bg-gradient-to-r from-blue-600 to-indigo-600 text-white font-medium disabled:from-gray-600 disabled:to-gray-600 disabled:cursor-not-allowed transition-all flex items-center justify-center gap-2 min-w-[150px]"
                    >
                      {isLoadingPrompt ? (
                        <>
                          <Loader2 className="w-4 h-4 animate-spin" />
                          {language === 'zh' ? '生成中...' : 'Generating...'}
                        </>
                      ) : (
                        <>
                          <RefreshCw className="w-4 h-4" />
                          {language === 'zh'
                            ? '获取提示词数据'
                            : 'Get Prompt Data'}
                        </>
                      )}
                    </button>
                  </div>
                </div>

                {/* Step 3: Decision Submission */}
                <div className="bg-gray-800/30 p-4 rounded-lg border border-gray-700">
                  <div className="flex justify-between items-center mb-3">
                    <label className="text-sm font-medium text-gray-200">
                      3.{' '}
                      {language === 'zh'
                        ? '粘贴AI决策JSON'
                        : 'Paste AI Decision JSON'}
                    </label>
                    <div className="flex gap-2">
                      <button
                        onClick={async () => {
                          try {
                            const text = await navigator.clipboard.readText()
                            setManualAIDecision(text)
                            notify.success(
                              language === 'zh'
                                ? '已从剪贴板粘贴内容'
                                : 'Content pasted from clipboard'
                            )
                          } catch (err) {
                            console.error(
                              'Failed to read clipboard contents: ',
                              err
                            )
                            notify.error(
                              language === 'zh'
                                ? '无法访问剪贴板，请检查权限'
                                : 'Could not access clipboard'
                            )
                          }
                        }}
                        className="text-xs px-2 py-1 text-gray-400 hover:text-white flex items-center gap-1"
                        title={
                          language === 'zh'
                            ? '从剪贴板粘贴'
                            : 'Paste from clipboard'
                        }
                      >
                        <ClipboardPaste className="w-3 h-3" />
                        {language === 'zh' ? '粘贴' : 'Paste'}
                      </button>
                      <button
                        onClick={() => setManualAIDecision('')}
                        className="text-xs px-2 py-1 text-gray-400 hover:text-white"
                      >
                        {language === 'zh' ? '清空' : 'Clear'}
                      </button>
                    </div>
                  </div>
                  <textarea
                    value={manualAIDecision}
                    onChange={(e) => setManualAIDecision(e.target.value)}
                    placeholder={
                      language === 'zh'
                        ? '粘贴 AI 返回的 JSON，例如: { "decisions": [...] }'
                        : 'Paste AI returned JSON, e.g.: { "decisions": [...] }'
                    }
                    className="w-full p-3 rounded bg-black/40 border border-white/10 text-white font-mono text-sm focus:ring-2 focus:ring-emerald-500 h-[120px]"
                  />
                  <button
                    onClick={handleSubmitAIDecision}
                    disabled={
                      submitAILoading ||
                      !selectedTraderId
                    }
                    title={
                      !selectedTraderId
                        ? language === 'zh'
                          ? '请先选择一个交易员'
                          : 'Please select a trader first'
                        : ''
                    }
                    className="mt-4 w-full px-4 py-3 rounded-lg bg-gradient-to-r from-emerald-600 to-green-600 text-white font-bold hover:from-emerald-700 hover:to-green-700 disabled:from-gray-700 disabled:to-gray-700 disabled:cursor-not-allowed transition-all flex items-center justify-center gap-2"
                  >
                    {submitAILoading ? (
                      <Loader2 className="w-5 h-5 animate-spin" />
                    ) : (
                      <Send className="w-5 h-5" />
                    )}
                    {language === 'zh' ? '提交决策' : 'Submit Decision'}
                  </button>
                </div>

                {/* Execution Result */}
                {submitAIResult && (
                  <div className="bg-black/40 p-4 rounded-lg border border-gray-700">
                    <p className="text-xs font-bold text-gray-400 mb-2 uppercase tracking-wider">
                      {language === 'zh' ? '执行结果' : 'Execution Result'}:
                    </p>
                    <pre className="text-xs text-emerald-400 whitespace-pre-wrap font-mono overflow-auto max-h-[150px]">
                      {submitAIResult}
                    </pre>
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Right Column: Recent Decisions */}
          <div
            className="nofx-glass p-6 animate-slide-in h-fit lg:sticky lg:top-24 lg:max-h-[calc(400vh-120px)] flex flex-col"
            style={{ animationDelay: '0.2s' }}
          >
            {/* Header */}
            <div className="flex items-center justify-between gap-3 mb-3 pb-2 border-b border-white/5 shrink-0">
              <div className="flex items-center gap-3">
                <div
                  className="w-10 h-10 rounded-xl flex items-center justify-center text-xl shadow-[0_4px_14px_rgba(99,102,241,0.4)]"
                  style={{
                    background:
                      'linear-gradient(135deg, #6366F1 0%, #8B5CF6 100%)',
                  }}
                >
                  🧠
                </div>
                <div>
                  <h2 className="text-xl font-bold text-nofx-text-main">
                    {t('recentDecisions', language)}
                  </h2>
                  {decisions && decisions.length > 0 && (
                    <div className="text-xs text-nofx-text-muted">
                      {t('lastCycles', language, { count: decisions.length })}
                    </div>
                  )}
                </div>
              </div>
              {/* AI分析状态指示器 */}
              {isSystemScanning && (
                <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-purple-500/10 border border-purple-500/30 mr-2">
                  <Loader2 className="w-4 h-4 animate-spin text-purple-400" />
                  <span className="text-sm text-purple-300">
                    {language === 'zh' ? 'AI分析中...' : 'AI Analyzing...'}
                  </span>
                </div>
              )}
              {/* 🔥 新增：下次扫描倒计时 */}
              {!isSystemScanning &&
                status?.is_running &&
                nextScanCountdown > 0 && (
                  <div
                    className={`flex items-center gap-2 px-3 py-1.5 rounded-lg border mr-2 ${
                      isDelayedByManual
                        ? 'bg-amber-500/10 border-amber-500/30'
                        : 'bg-blue-500/10 border-blue-500/30'
                    }`}
                  >
                    <span
                      className={`text-sm font-medium ${
                        isDelayedByManual ? 'text-amber-300' : 'text-blue-300'
                      }`}
                    >
                      {language === 'zh'
                        ? `AI分析将在 ${nextScanCountdown} 秒后开始`
                        : `AI analysis in ${nextScanCountdown}s`}
                    </span>
                    {isDelayedByManual && (
                      <span className="text-xs text-amber-400/70">
                        {language === 'zh' ? '(手动延迟)' : '(Manual Delay)'}
                      </span>
                    )}
                  </div>
                )}
              <button
                onClick={async () => {
                  if (!selectedTraderId) {
                    notify.error(
                      language === 'zh'
                        ? '请选择交易员'
                        : 'Please select a trader'
                    )
                    return
                  }

                  // 🔥 新增：AI分析中时拒绝手动扫盘
                  if (isSystemScanning) {
                    notify.warning(
                      language === 'zh'
                        ? 'AI正在分析中，请等待完成后再手动扫盘（避免给AI增加负担）'
                        : 'AI is analyzing, please wait for completion before manual scan (to avoid AI burden)'
                    )
                    return
                  }

                  if (manualScanCooldown) {
                    notify.error(
                      language === 'zh'
                        ? '操作过于频繁，请稍后再试'
                        : 'Action too frequent, please try again later'
                    )
                    return
                  }

                  // 检查交易员状态
                  if (status && !status.is_running) {
                    notify.error(
                      language === 'zh'
                        ? '交易员未运行，无法手动触发扫盘'
                        : 'Trader is not running, cannot trigger manual scan'
                    )
                    return
                  }

                  // 注意：AccountInfo 没有 account_status 字段，跳过此项检查

                  setIsManualDecisionLoading(true)
                  const startTime = Date.now()
                  notify.info(
                    language === 'zh'
                      ? '正在触发手动扫盘...'
                      : 'Triggering manual scan...'
                  )

                  try {
                    const result = await api.triggerDecision(selectedTraderId)

                    // 计算执行时间
                    const endTime = Date.now()
                    const executionTime = endTime - startTime

                    // 检查结果并提供更详细的反馈
                    if (result && result.message) {
                      // 🔥 根据延迟时长动态显示提示
                      const delayMessage =
                        language === 'zh'
                          ? 'AI分析已智能延迟'
                          : 'AI analysis intelligently delayed'

                      // 如果后端返回了执行时间，优先使用后端的时间
                      if (result.execution_time_formatted) {
                        notify.success(
                          language === 'zh'
                            ? `✅ 手动扫盘已完成！耗时: ${result.execution_time_formatted}\n⏰ ${delayMessage}`
                            : `✅ Manual scan completed! Duration: ${result.execution_time_formatted}\n⏰ ${delayMessage}`
                        )
                      } else {
                        notify.success(
                          language === 'zh'
                            ? `✅ 手动扫盘已成功触发！客户端耗时: ${executionTime}ms\n⏰ ${delayMessage}`
                            : `✅ Manual scan triggered successfully! Client duration: ${executionTime}ms\n⏰ ${delayMessage}`
                        )
                      }
                    }

                    // 刷新相关数据 - 使用正确的SWR key确保决策列表及时更新
                    await Promise.all([
                      mutate(`positions-${selectedTraderId}`),
                      mutate(`account-${selectedTraderId}`),
                      mutate(`decisions/latest-${selectedTraderId}-${decisionsLimit}`), // 使用正确的决策数据key
                      mutate(`position-history-${selectedTraderId}`),
                    ])

                    // 设置冷却时间（固定20秒）
                    const cooldownTime = 20000 // 固定20秒冷却时间

                    setManualScanCooldown(true)
                    setTimeout(
                      () => {
                        setManualScanCooldown(false)
                      },
                      Math.max(cooldownTime, 10000)
                    ) // 最少10秒冷却时间
                  } catch (error: any) {
                    // 计算执行时间（即使失败）
                    const endTime = Date.now()
                    const executionTime = endTime - startTime

                    console.error('手动扫盘失败:', error)

                    // 提供更详细的错误反馈
                    let errorMessage = error.message || '未知错误'

                    // 检查具体的错误类型
                    if (error.message?.includes('already executing')) {
                      notify.warning(
                        language === 'zh'
                          ? 'AI决策已在执行中，请等待完成后再试'
                          : 'AI decision is already executing, please wait for completion'
                      )
                    } else if (error.message?.includes('not running')) {
                      notify.error(
                        language === 'zh'
                          ? '交易员未运行，无法执行手动扫盘'
                          : 'Trader is not running, cannot execute manual scan'
                      )
                    } else if (error.message?.includes('Network error')) {
                      // 在开发模式下显示更详细的错误信息
                      if (process.env.NODE_ENV === 'development') {
                        console.error('Manual scan network error details:', {
                          message: error.message,
                          stack: error.stack,
                          config: error.config,
                          url: error.config?.url,
                          method: error.config?.method,
                        })

                        notify.error(
                          language === 'zh'
                            ? `网络连接错误，请检查服务是否正常运行: ${error.message || 'Connection failed'} (耗时: ${executionTime}ms)`
                            : `Network connection error, please check if service is running: ${error.message || 'Connection failed'} (duration: ${executionTime}ms)`
                        )
                      } else {
                        notify.error(
                          language === 'zh'
                            ? '网络连接错误，请检查服务是否正常运行'
                            : 'Network connection error, please check if service is running'
                        )
                      }
                    } else if (
                      error.message?.includes('connectex') ||
                      error.message?.includes('connection failed')
                    ) {
                      // 在开发模式下显示更详细的错误信息
                      if (process.env.NODE_ENV === 'development') {
                        console.error('Manual scan connection error details:', {
                          message: error.message,
                          stack: error.stack,
                          config: error.config,
                          url: error.config?.url,
                          method: error.config?.method,
                        })

                        notify.warning(
                          language === 'zh'
                            ? `网络连接问题，但AI分析可能仍在运行: ${error.message || 'Connection failed'} (耗时: ${executionTime}ms)`
                            : `Network connection issue, but AI analysis may still run: ${error.message || 'Connection failed'} (duration: ${executionTime}ms)`
                        )
                      } else {
                        notify.warning(
                          language === 'zh'
                            ? '网络连接问题，但AI分析可能仍在运行'
                            : 'Network connection issue, but AI analysis may still run'
                        )
                      }
                    } else {
                      // 在开发模式下显示更详细的错误信息
                      if (process.env.NODE_ENV === 'development') {
                        console.error('Manual scan error details:', {
                          message: error.message,
                          stack: error.stack,
                          config: error.config,
                          url: error.config?.url,
                          method: error.config?.method,
                          data: error.response?.data,
                          status: error.response?.status,
                        })

                        notify.error(
                          language === 'zh'
                            ? `手动扫盘失败: ${errorMessage} (耗时: ${executionTime}ms)\n状态码: ${error.response?.status || 'N/A'}\nURL: ${error.config?.url || 'N/A'}`
                            : `Manual scan failed: ${errorMessage} (duration: ${executionTime}ms)\nStatus: ${error.response?.status || 'N/A'}\nURL: ${error.config?.url || 'N/A'}`
                        )
                      } else {
                        notify.error(
                          language === 'zh'
                            ? `手动扫盘失败: ${errorMessage} (耗时: ${executionTime}ms)`
                            : `Manual scan failed: ${errorMessage} (duration: ${executionTime}ms)`
                        )
                      }
                    }
                  } finally {
                    setIsManualDecisionLoading(false)
                  }
                }}
                disabled={
                  !selectedTraderId ||
                  isManualDecisionLoading ||
                  manualScanCooldown ||
                  isSystemScanning
                }
                className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-all hover:scale-105 active:scale-95 nofx-glass border text-sm ${
                  !selectedTraderId || isSystemScanning
                    ? 'border-nofx-gray/30 text-nofx-gray/50 cursor-not-allowed'
                    : 'border-nofx-blue/30 text-nofx-blue hover:bg-nofx-blue/10'
                } flex items-center gap-1 mr-2`}
                title={
                  !selectedTraderId
                    ? '请先选择交易员'
                    : isSystemScanning
                      ? 'AI正在分析，请稍候...'
                      : manualScanCooldown
                        ? '冷却中，请稍后再试'
                        : '手动触发AI扫盘决策'
                }
              >
                {isManualDecisionLoading ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin" />
                    {language === 'zh' ? '扫盘中...' : 'Scanning...'}
                  </>
                ) : manualScanCooldown ? (
                  <>
                    <span>⏳</span>
                    {language === 'zh' ? '冷却中...' : 'Cooldown...'}
                  </>
                ) : (
                  <>
                    <span>🔍</span>
                    {language === 'zh' ? '手动扫盘' : 'Manual Scan'}
                  </>
                )}
              </button>
              {/* Stop/Start Trader Button */}
              <button
                onClick={async () => {
                  if (!selectedTraderId) {
                    notify.error(
                      language === 'zh'
                        ? '请选择交易员'
                        : 'Please select a trader'
                    )
                    return
                  }

                  // 如果是启动操作且按钮处于冷却状态，阻止重复点击
                  if (!status?.is_running && startButtonCooldown) {
                    notify.info(
                      language === 'zh'
                        ? '请稍后再试，避免重复点击'
                        : 'Please wait, avoiding duplicate clicks'
                    )
                    return
                  }

                  try {
                    if (status?.is_running) {
                      // 停止交易员
                      await api.stopTrader(selectedTraderId)
                      notify.success(
                        language === 'zh'
                          ? '交易员已停止'
                          : 'Trader stopped successfully'
                      )
                    } else {
                      // 设置启动按钮冷却状态，防止5-10秒内的重复点击
                      setStartButtonCooldown(true)
                      setTimeout(() => {
                        setStartButtonCooldown(false)
                      }, 8000) // 8秒冷却时间

                      // 启动交易员
                      await api.startTrader(selectedTraderId)
                      notify.success(
                        language === 'zh'
                          ? '交易员已启动'
                          : 'Trader started successfully'
                      )
                    }

                    // 刷新相关数据
                    await Promise.all([
                      mutate(`positions-${selectedTraderId}`),
                      mutate(`account-${selectedTraderId}`),
                      mutate(`decisions-${selectedTraderId}`),
                      mutate(`position-history-${selectedTraderId}`),
                    ])
                  } catch (error: any) {
                    // 如果启动失败，清除冷却状态
                    if (!status?.is_running) {
                      setStartButtonCooldown(false)
                    }
                    console.error('切换交易员状态失败:', error)
                    const errorMsg =
                      error.message ||
                      (status?.is_running
                        ? language === 'zh'
                          ? '停止交易员失败'
                          : 'Failed to stop trader'
                        : language === 'zh'
                          ? '启动交易员失败'
                          : 'Failed to start trader')
                    notify.error(errorMsg)
                  }
                }}
                disabled={
                  !selectedTraderId ||
                  (startButtonCooldown && !status?.is_running)
                }
                className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-all hover:scale-105 active:scale-95 nofx-glass border text-sm ${!selectedTraderId ? 'border-nofx-gray/30 text-nofx-gray/50' : status?.is_running ? 'border-nofx-red/30 text-nofx-red hover:bg-nofx-red/10' : startButtonCooldown ? 'border-nofx-gray/30 text-nofx-gray/50 cursor-not-allowed' : 'border-nofx-green/30 text-nofx-green hover:bg-nofx-green/10'} flex items-center gap-1`}
                title={
                  !selectedTraderId
                    ? '请先选择交易员'
                    : status?.is_running
                      ? language === 'zh'
                        ? '停止交易员'
                        : 'Stop Trader'
                      : startButtonCooldown
                        ? language === 'zh'
                          ? '启动中，请稍候...'
                          : 'Starting, please wait...'
                        : language === 'zh'
                          ? '启动交易员'
                          : 'Start Trader'
                }
              >
                {status?.is_running ? (
                  <>
                    <span>⏹</span>
                    {language === 'zh' ? '停止' : 'Stop'}
                  </>
                ) : startButtonCooldown ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin" />
                    {language === 'zh' ? '启动中...' : 'Starting...'}
                  </>
                ) : (
                  <>
                    <span>▶️</span>
                    {language === 'zh' ? '启动' : 'Start'}
                  </>
                )}
              </button>
            </div>

            {/* Decisions List - Scrollable */}
            <div
              ref={decisionsContainerRef}
              className="space-y-4 overflow-y-auto pr-2 custom-scrollbar"
              style={{ maxHeight: 'calc(240vh - 280px)' }}
            >
              {allDecisions && allDecisions.length > 0 ? (
                allDecisions.map((decision) => (
                  <DecisionCard
                    key={`${decision.trader_id}-${decision.cycle_number}-${decision.timestamp}`}
                    decision={decision}
                    language={language}
                    onSymbolClick={handleSymbolClick}
                    onDelete={handleDeleteDecision}
                  />
                ))
              ) : (
                <div className="py-16 text-center text-nofx-text-muted opacity-60">
                  <div className="text-6xl mb-4 opacity-30 grayscale">🧠</div>
                  <div className="text-lg font-semibold mb-2 text-nofx-text-main">
                    {t('noDecisionsYet', language)}
                  </div>
                  <div className="text-sm">
                    {t('aiDecisionsWillAppear', language)}
                  </div>
                </div>
              )}
              {loadingMoreDecisions && (
                <div className="text-center py-4">
                  <div className="inline-block animate-spin rounded-full h-6 w-6 border-t-2 border-b-2 border-nofx-accent"></div>
                  <p className="mt-2 text-nofx-text-muted text-sm">
                    {language === 'zh' ? '加载更多周期...' : 'Loading more cycles...'}
                  </p>
                </div>
              )}
            </div>
            {/* Controls Row */}
            <div className="flex items-center gap-2 mb-3">
              <div className="flex gap-2 flex-wrap">
                <input
                  type="date"
                  value={startDate}
                  onChange={(e) => setStartDate(e.target.value)}
                  className="px-3 py-1.5 rounded-lg text-sm font-medium cursor-pointer transition-all bg-black/40 text-nofx-text-main border border-white/10 hover:border-nofx-accent focus:outline-none"
                  title="Start date"
                />
                <input
                  type="date"
                  value={endDate}
                  onChange={(e) => setEndDate(e.target.value)}
                  className="px-3 py-1.5 rounded-lg text-sm font-medium cursor-pointer transition-all bg-black/40 text-nofx-text-main border border-white/10 hover:border-nofx-accent focus:outline-none"
                  title="End date"
                />
                <select
                  value={decisionsLimit}
                  onChange={(e) =>
                    onDecisionsLimitChange(Number(e.target.value))
                  }
                  className="px-3 py-1.5 rounded-lg text-sm font-medium cursor-pointer transition-all bg-black/40 text-nofx-text-main border border-white/10 hover:border-nofx-accent focus:outline-none"
                >
                  <option value={5}>5</option>
                  <option value={10}>10</option>
                  <option value={20}>20</option>
                  <option value={50}>50</option>
                  <option value={100}>100</option>
                </select>
                <button
                  onClick={() => {
                    // 导出决策历史为CSV，按时间范围过滤
                    const filteredDecisions = filterDecisionsByDate(
                      decisions || [],
                      startDate,
                      endDate
                    )
                    exportDecisionHistoryToCSV(
                      filteredDecisions,
                      selectedTrader?.trader_name || 'trader'
                    )
                  }}
                  className="px-3 py-1.5 rounded-lg text-sm font-medium transition-all hover:scale-105 active:scale-95 nofx-glass border border-nofx-gold/30 text-nofx-gold hover:bg-nofx-gold/10"
                  title="Export decision history to CSV"
                >
                  📥 {t('positionHistory.exportDecisions', language)}
                </button>
                <button
                  onClick={() => {
                    // 导出决策历史为高级格式，包含System Prompt和User Prompt
                    const filteredDecisions = filterDecisionsByDate(
                      decisions || [],
                      startDate,
                      endDate
                    )
                    exportAdvancedDecisionHistoryToCSV(
                      filteredDecisions,
                      selectedTrader?.trader_name || 'trader'
                    )
                  }}
                  className="px-3 py-1.5 rounded-lg text-sm font-medium transition-all hover:scale-105 active:scale-95 nofx-glass border border-nofx-gold/30 text-nofx-gold hover:bg-nofx-gold/10"
                  title="Export advanced decision history with System/User Prompts and Chain of Thought"
                >
                  📤 {t('positionHistory.exportAdvancedDecisions', language)}
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Position History Section */}
        {selectedTraderId && (
          <div
            className="nofx-glass p-6 animate-slide-in"
            style={{ animationDelay: '0.25s' }}
          >
            <div className="flex items-center justify-between mb-5">
              <h2 className="text-xl font-bold flex items-center gap-2 text-nofx-text-main">
                <span className="text-2xl">📜</span>
                {t('positionHistory.title', language)}
              </h2>
              <div className="flex gap-2">
                <input
                  type="date"
                  value={startDate}
                  onChange={(e) => setStartDate(e.target.value)}
                  className="px-3 py-1.5 rounded-lg text-sm font-medium cursor-pointer transition-all bg-black/40 text-nofx-text-main border border-white/10 hover:border-nofx-accent focus:outline-none"
                  title="Start date"
                />
                <input
                  type="date"
                  value={endDate}
                  onChange={(e) => setEndDate(e.target.value)}
                  className="px-3 py-1.5 rounded-lg text-sm font-medium cursor-pointer transition-all bg-black/40 text-nofx-text-main border border-white/10 hover:border-nofx-accent focus:outline-none"
                  title="End date"
                />
                <button
                  onClick={() => {
                    // 导出交易历史为CSV，按时间范围过滤
                    exportPositionHistoryToCSV(
                      selectedTraderId,
                      selectedTrader?.trader_name || 'trader',
                      startDate,
                      endDate
                    )
                  }}
                  className="px-4 py-2 rounded-lg font-medium text-sm transition-all hover:scale-105 active:scale-95 nofx-glass border border-nofx-gold/30 text-nofx-gold hover:bg-nofx-gold/10"
                  title="Export position history to CSV"
                >
                  📥 {t('positionHistory.export', language)}
                </button>
              </div>
            </div>
            <PositionHistory
              traderId={selectedTraderId}
              onExport={exportPositionHistoryToCSV}
              traderName={selectedTrader?.trader_name}
            />
          </div>
        )}
      </div>
    </DeepVoidBackground>
  )
}

// 导出交易历史为CSV文件
async function exportPositionHistoryToCSV(
  traderId: string,
  traderName: string,
  startDate?: string,
  endDate?: string
) {
  try {
    // 获取交易历史数据
    const data = await api.getPositionHistory(traderId, 1000)
    let positions = data.positions || []

    // 按日期范围过滤位置
    if (startDate || endDate) {
      positions = positions.filter((position) => {
        const exitTime = position.exit_time
          ? new Date(position.exit_time)
          : null

        if (!exitTime || isNaN(exitTime.getTime())) {
          return true // 如果退出时间无效，则保留记录
        }

        const exitDateString = exitTime.toISOString().split('T')[0]

        if (startDate && endDate) {
          return exitDateString >= startDate && exitDateString <= endDate
        } else if (startDate) {
          return exitDateString >= startDate
        } else if (endDate) {
          return exitDateString <= endDate
        }

        return true
      })
    }

    if (positions.length === 0) {
      notify.error('No position history to export')
      return
    }

    // 定义CSV头部
    const headers = [
      '序号',
      '交易对',
      '方向 (多/空)',
      '开仓时间',
      '平仓时间',
      '持仓时长',
      '开仓价格',
      '平仓价格',
      '仓位大小 (USD)',
      '杠杆倍数',
      '止损价',
      '止盈价',
      '盈亏 (USD)',
      '盈亏 (%)',
      '开仓理由 (对照DP_V18哪条规则)',
    ]

    // 转换数据为CSV格式
    const csvRows = []
    csvRows.push(headers.join(','))

    positions.forEach((position: HistoricalPosition, index: number) => {
      const entryTime = position.entry_time
        ? new Date(position.entry_time).toLocaleString()
        : ''
      const exitTime = position.exit_time
        ? new Date(position.exit_time).toLocaleString()
        : ''

      // 计算持仓时长
      const entryTimeObj = position.entry_time
        ? new Date(position.entry_time).getTime()
        : 0
      const exitTimeObj = position.exit_time
        ? new Date(position.exit_time).getTime()
        : 0
      const holdingMinutes =
        entryTimeObj && exitTimeObj && exitTimeObj > entryTimeObj
          ? (exitTimeObj - entryTimeObj) / 60000
          : 0
      const holdingDuration = formatDurationForCSV(holdingMinutes)

      // 计算盈亏百分比
      const entryPrice = position.entry_price || 0
      const exitPrice = position.exit_price || 0
      let pnlPct = 0
      if (entryPrice > 0) {
        const isLong = (position.side || '').toUpperCase() === 'LONG'
        if (isLong) {
          pnlPct = ((exitPrice - entryPrice) / entryPrice) * 100
        } else {
          pnlPct = ((entryPrice - exitPrice) / entryPrice) * 100
        }
      }

      // 计算仓位大小 (USD)
      const entryQuantity = position.entry_quantity || position.quantity || 0
      const positionValue = entryPrice * entryQuantity

      const row = [
        index + 1, // 序号
        position.symbol, // 交易对
        (position.side || '').toUpperCase(), // 方向
        entryTime, // 开仓时间
        exitTime, // 平仓时间
        holdingDuration, // 持仓时长
        entryPrice, // 开仓价格
        exitPrice, // 平仓价格
        positionValue.toFixed(2), // 仓位大小 (USD)
        position.leverage || 1, // 杠杆倍数
        '', // 止损价 (数据库中未存储)
        '', // 止盈价 (数据库中未存储)
        (position.realized_pnl || 0).toFixed(2), // 盈亏 (USD)
        pnlPct.toFixed(2), // 盈亏 (%)
        '', // 开仓理由 (如果数据库中有存储的话)
      ].map((field) => {
        // 处理包含逗号或引号的字段
        const fieldStr = String(field)
        if (
          fieldStr.includes(',') ||
          fieldStr.includes('"') ||
          fieldStr.includes('\n')
        ) {
          return `"${fieldStr.replace(/"/g, '""')}"`
        }
        return fieldStr
      })

      csvRows.push(row.join(','))
    })

    // 创建CSV内容
    const csvContent = csvRows.join('\n')

    // 添加 BOM 以支持中文字符
    const BOM = '\uFEFF'
    const blob = new Blob([BOM, csvContent], {
      type: 'text/csv;charset=utf-8;',
    })
    const link = document.createElement('a')
    const fileName = `${traderName}_position_history_${new Date().toISOString().split('T')[0]}.csv`
    link.setAttribute('href', URL.createObjectURL(blob))
    link.setAttribute('download', fileName)
    link.style.visibility = 'hidden'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)

    notify.success(`Position history exported to ${fileName}`)
  } catch (error) {
    console.error('Error exporting position history:', error)
    notify.error('Failed to export position history')
  }
}

// 格式化持续时间用于CSV输出
function formatDurationForCSV(minutes: number): string {
  if (!minutes || minutes <= 0) return ''
  if (minutes < 60) return `${Math.floor(minutes)}m`
  if (minutes < 1440)
    return `${Math.floor(minutes / 60)}h${Math.floor(minutes % 60)}m`
  return `${Math.floor(minutes / 1440)}d${Math.floor((minutes % 1440) / 60)}h`
}

// 按日期范围过滤决策记录
function filterDecisionsByDate(
  decisions: DecisionRecord[],
  startDate: string,
  endDate: string
): DecisionRecord[] {
  if (!startDate && !endDate) return decisions

  return decisions.filter((decision) => {
    // 尝试解析时间戳 - 可能是 ISO 格式或时间戳数字
    let decisionDate: Date | null = null

    if (typeof decision.timestamp === 'number') {
      // 如果是数字时间戳（毫秒）
      decisionDate = new Date(decision.timestamp)
    } else if (typeof decision.timestamp === 'string') {
      // 如果是字符串格式的时间
      decisionDate = new Date(decision.timestamp)
    }

    if (!decisionDate || isNaN(decisionDate.getTime())) {
      return true // 如果日期无效，则保留该记录
    }

    const decisionDateString = decisionDate.toISOString().split('T')[0]

    if (startDate && endDate) {
      return decisionDateString >= startDate && decisionDateString <= endDate
    } else if (startDate) {
      return decisionDateString >= startDate
    } else if (endDate) {
      return decisionDateString <= endDate
    }

    return true
  })
}

// 导出高级决策历史为CSV文件，包含System Prompt和User Prompt等信息
async function exportAdvancedDecisionHistoryToCSV(
  decisions: DecisionRecord[],
  traderName: string
) {
  try {
    if (!decisions || decisions.length === 0) {
      notify.error('No decision history to export')
      return
    }

    // 使用TableExporter工具
    const headers = [
      '序号',
      '时间戳',
      '循环编号',
      '系统提示(System Prompt)',
      '用户提示(User Prompt)',
      'AI链式思维(CoT)',
      '决策JSON',
      '账户状态',
      '持仓',
      '候选币种',
      '决策动作',
      '执行日志',
      '是否成功',
      '错误信息',
    ]

    const rows = decisions.map((decision, index) => [
      index + 1, // 序号
      decision.timestamp, // 时间戳
      decision.cycle_number, // 循环编号
      decision.system_prompt
        ? JSON.stringify(decision.system_prompt).replace(/,/g, ';')
        : '', // 系统提示(System Prompt)
      decision.input_prompt
        ? JSON.stringify(decision.input_prompt).replace(/,/g, ';')
        : '', // 用户提示(User Prompt)
      decision.cot_trace
        ? JSON.stringify(decision.cot_trace).replace(/,/g, ';')
        : '', // AI链式思维
      decision.decision_json, // 决策JSON
      decision.account_state ? JSON.stringify(decision.account_state) : '', // 账户状态
      decision.positions ? JSON.stringify(decision.positions) : '', // 持仓
      decision.candidate_coins ? JSON.stringify(decision.candidate_coins) : '', // 候选币种
      decision.decisions ? JSON.stringify(decision.decisions) : '', // 决策动作
      decision.execution_log ? JSON.stringify(decision.execution_log) : '', // 执行日志
      decision.success ? 'true' : 'false', // 是否成功
      decision.error_message || '', // 错误信息
    ])

    const tableData = {
      headers,
      rows,
      title: 'Advanced Decision History',
    }

    const filename = `${traderName}_advanced_decision_history_${new Date().toISOString().split('T')[0]}`

    // 使用我们创建的TableExporter进行导出
    TableExporter.export(tableData, filename, ExportFormat.CSV)

    notify.success(`Advanced decision history exported to ${filename}`)
  } catch (error) {
    console.error('Error exporting advanced decision history:', error)
    notify.error('Failed to export advanced decision history')
  }
}

// 导出决策历史为CSV文件
async function exportDecisionHistoryToCSV(
  decisions: DecisionRecord[],
  traderName: string
) {
  try {
    if (!decisions || decisions.length === 0) {
      notify.error('No decision history to export')
      return
    }

    // 使用TableExporter工具
    const headers = [
      '序号',
      '时间戳',
      '循环编号',
      '系统提示(System Prompt)',
      '用户提示(User Prompt)',
      'AI链式思维(CoT)',
      '决策JSON',
      '账户状态',
      '持仓',
      '候选币种',
      '决策动作',
      '执行日志',
      '是否成功',
      '错误信息',
    ]

    const rows = decisions.map((decision, index) => [
      index + 1, // 序号
      decision.timestamp, // 时间戳
      decision.cycle_number, // 循环编号
      decision.system_prompt
        ? JSON.stringify(decision.system_prompt).replace(/,/g, ';')
        : '', // 系统提示(System Prompt)
      decision.input_prompt
        ? JSON.stringify(decision.input_prompt).replace(/,/g, ';')
        : '', // 用户提示(User Prompt)
      decision.cot_trace
        ? JSON.stringify(decision.cot_trace).replace(/,/g, ';')
        : '', // AI链式思维
      decision.decision_json, // 决策JSON
      decision.account_state ? JSON.stringify(decision.account_state) : '', // 账户状态
      decision.positions ? JSON.stringify(decision.positions) : '', // 持仓
      decision.candidate_coins ? JSON.stringify(decision.candidate_coins) : '', // 候选币种
      decision.decisions ? JSON.stringify(decision.decisions) : '', // 决策动作
      decision.execution_log ? JSON.stringify(decision.execution_log) : '', // 执行日志
      decision.success ? 'true' : 'false', // 是否成功
      decision.error_message || '', // 错误信息
    ])

    const tableData = {
      headers,
      rows,
      title: 'Decision History',
    }

    const filename = `${traderName}_decision_history_${new Date().toISOString().split('T')[0]}`

    // 使用我们创建的TableExporter进行导出
    TableExporter.export(tableData, filename, ExportFormat.CSV)

    notify.success(`Decision history exported to ${filename}`)
  } catch (error) {
    console.error('Error exporting decision history:', error)
    notify.error('Failed to export decision history')
  }
}

// Stat Card Component - Deep Void Style
function StatCard({
  title,
  value,
  unit,
  change,
  positive,
  subtitle,
  icon,
}: {
  title: string
  value: string
  unit?: string
  change?: number
  positive?: boolean
  subtitle?: string
  icon?: string
}) {
  return (
    <div className="group nofx-glass p-5 rounded-lg transition-all duration-300 hover:bg-white/5 hover:translate-y-[-2px] border border-white/5 hover:border-nofx-gold/20 relative overflow-hidden">
      <div className="absolute top-0 right-0 p-4 opacity-5 group-hover:opacity-10 transition-opacity text-4xl grayscale group-hover:grayscale-0">
        {icon}
      </div>
      <div className="text-xs mb-2 font-mono uppercase tracking-wider text-nofx-text-muted flex items-center gap-2">
        {title}
      </div>
      <div className="flex items-baseline gap-1 mb-1">
        <div className="text-2xl font-bold font-mono text-nofx-text-main tracking-tight group-hover:text-white transition-colors">
          {value}
        </div>
        {unit && (
          <span className="text-xs font-mono text-nofx-text-muted opacity-60">
            {unit}
          </span>
        )}
      </div>

      {change !== undefined && (
        <div className="flex items-center gap-1">
          <div
            className={`text-sm mono font-bold flex items-center gap-1 ${positive ? 'text-nofx-green' : 'text-nofx-red'}`}
          >
            <span>{positive ? '▲' : '▼'}</span>
            <span>
              {positive ? '+' : ''}
              {change.toFixed(2)}%
            </span>
          </div>
        </div>
      )}
      {subtitle && (
        <div className="text-xs mt-2 mono text-nofx-text-muted opacity-80">
          {subtitle}
        </div>
      )}
    </div>
  )
}

export default TraderDashboardPage
