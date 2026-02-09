import type {
  SystemStatus,
  AccountInfo,
  Position,
  DecisionRecord,
  Statistics,
  TraderInfo,
  TraderConfigData,
  AIModel,
  Exchange,
  CreateTraderRequest,
  CreateExchangeRequest,
  UpdateModelConfigRequest,
  UpdateExchangeConfigRequest,
  CompetitionData,
  BacktestRunsResponse,
  BacktestStartConfig,
  BacktestStatusPayload,
  BacktestEquityPoint,
  BacktestTradeEvent,
  BacktestMetrics,
  BacktestRunMetadata,
  BacktestKlinesResponse,
  Strategy,
  StrategyConfig,
  DebateSession,
  DebateSessionWithDetails,
  CreateDebateRequest,
  DebateMessage,
  DebateVote,
  DebatePersonalityInfo,
  PositionHistoryResponse,
  SystemConfig,
} from '../types'
import { CryptoService } from './crypto'
import { httpClient } from './httpClient'

const API_BASE = '/api'

// Helper function to get auth headers
function getAuthHeaders(): Record<string, string> {
  const token = localStorage.getItem('auth_token')
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }

  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  return headers
}

async function handleJSONResponse<T>(res: Response): Promise<T> {
  const text = await res.text()
  if (!res.ok) {
    let message = text || res.statusText
    try {
      const data = text ? JSON.parse(text) : null
      if (data && typeof data === 'object') {
        message = data.error || data.message || message
      }
    } catch {
      /* ignore JSON parse errors */
    }
    throw new Error(message || '请求失败')
  }
  if (!text) {
    return {} as T
  }
  return JSON.parse(text) as T
}

export const api = {
  // AI交易员管理接口
  async getTraders(): Promise<TraderInfo[]> {
    const result = await httpClient.get<TraderInfo[]>(`${API_BASE}/my-traders`)
    if (!result.success) throw new Error('获取trader列表失败')
    return Array.isArray(result.data) ? result.data : []
  },

  // 获取公开的交易员列表（无需认证）
  async getPublicTraders(): Promise<any[]> {
    const result = await httpClient.get<any[]>(`${API_BASE}/traders`)
    if (!result.success) throw new Error('获取公开trader列表失败')
    return result.data!
  },

  async createTrader(request: CreateTraderRequest): Promise<TraderInfo> {
    const result = await httpClient.post<TraderInfo>(
      `${API_BASE}/traders`,
      request
    )
    if (!result.success) throw new Error('创建交易员失败')
    return result.data!
  },

  async deleteTrader(traderId: string): Promise<void> {
    const result = await httpClient.delete(`${API_BASE}/traders/${traderId}`)
    if (!result.success) throw new Error('删除交易员失败')
  },

  async startTrader(traderId: string): Promise<void> {
    const result = await httpClient.post(
      `${API_BASE}/traders/${traderId}/start`
    )
    if (!result.success) throw new Error('启动交易员失败')
  },

  async stopTrader(traderId: string): Promise<void> {
    const result = await httpClient.post(`${API_BASE}/traders/${traderId}/stop`)
    if (!result.success) throw new Error('停止交易员失败')
  },

  async getTraderStatus(traderId: string): Promise<SystemStatus> {
    const result = await httpClient.get<SystemStatus>(
      `${API_BASE}/traders/${traderId}/status`
    )
    if (!result.success) throw new Error('获取交易员状态失败')
    return result.data!
  },

  async toggleCompetition(traderId: string, showInCompetition: boolean): Promise<void> {
    const result = await httpClient.put(
      `${API_BASE}/traders/${traderId}/competition`,
      { show_in_competition: showInCompetition }
    )
    if (!result.success) throw new Error('更新竞技场显示设置失败')
  },

  async closePosition(traderId: string, symbol: string, side: string): Promise<{ message: string }> {
    const result = await httpClient.post<{ message: string }>(
      `${API_BASE}/traders/${traderId}/close-position`,
      { symbol, side }
    )
    if (!result.success) throw new Error('平仓失败')
    return result.data!
  },

  async updateTraderPrompt(
    traderId: string,
    customPrompt: string
  ): Promise<void> {
    const result = await httpClient.put(
      `${API_BASE}/traders/${traderId}/prompt`,
      { custom_prompt: customPrompt }
    )
    if (!result.success) throw new Error('更新自定义策略失败')
  },

  async getTraderConfig(traderId: string): Promise<TraderConfigData> {
    const result = await httpClient.get<TraderConfigData>(
      `${API_BASE}/traders/${traderId}/config`
    )
    if (!result.success) throw new Error('获取交易员配置失败')
    return result.data!
  },

  async updateTrader(
    traderId: string,
    request: CreateTraderRequest
  ): Promise<TraderInfo> {
    const result = await httpClient.put<TraderInfo>(
      `${API_BASE}/traders/${traderId}`,
      request
    )
    if (!result.success) throw new Error('更新交易员失败')
    return result.data!
  },

  // AI模型配置接口
  async getModelConfigs(): Promise<AIModel[]> {
    const result = await httpClient.get<AIModel[]>(`${API_BASE}/models`)
    if (!result.success) throw new Error('获取模型配置失败')
    return Array.isArray(result.data) ? result.data : []
  },

  // 获取系统支持的AI模型列表（无需认证）
  async getSupportedModels(): Promise<AIModel[]> {
    const result = await httpClient.get<AIModel[]>(
      `${API_BASE}/supported-models`
    )
    if (!result.success) throw new Error('获取支持的模型失败')
    return result.data!
  },

  async getPromptTemplates(): Promise<string[]> {
    const res = await fetch(`${API_BASE}/prompt-templates`)
    if (!res.ok) throw new Error('获取提示词模板失败')
    const data = await res.json()
    if (Array.isArray(data.templates)) {
      return data.templates.map((item: { name: string }) => item.name)
    }
    return []
  },

  async updateModelConfigs(request: UpdateModelConfigRequest): Promise<void> {
    // 检查是否启用了传输加密
    const config = await CryptoService.fetchCryptoConfig()

    if (!config.transport_encryption) {
      // 传输加密禁用时，直接发送明文
      const result = await httpClient.put(`${API_BASE}/models`, request)
      if (!result.success) throw new Error('更新模型配置失败')
      return
    }

    // 获取RSA公钥
    const publicKey = await CryptoService.fetchPublicKey()

    // 初始化加密服务
    await CryptoService.initialize(publicKey)

    // 获取用户信息（从localStorage或其他地方）
    const userId = localStorage.getItem('user_id') || ''
    const sessionId = sessionStorage.getItem('session_id') || ''

    // 加密敏感数据
    const encryptedPayload = await CryptoService.encryptSensitiveData(
      JSON.stringify(request),
      userId,
      sessionId
    )

    // 发送加密数据
    const result = await httpClient.put(`${API_BASE}/models`, encryptedPayload)
    if (!result.success) throw new Error('更新模型配置失败')
  },

  // 删除AI模型配置
  async deleteModel(modelId: string): Promise<void> {
    const result = await httpClient.delete(`${API_BASE}/models/${modelId}`)
    if (!result.success) throw new Error('删除AI模型失败')
  },

  // 交易所配置接口
  async getExchangeConfigs(): Promise<Exchange[]> {
    const result = await httpClient.get<Exchange[]>(`${API_BASE}/exchanges`)
    if (!result.success) throw new Error('获取交易所配置失败')
    return result.data!
  },

  // 获取系统支持的交易所列表（无需认证）
  async getSupportedExchanges(): Promise<Exchange[]> {
    const result = await httpClient.get<Exchange[]>(
      `${API_BASE}/supported-exchanges`
    )
    if (!result.success) throw new Error('获取支持的交易所失败')
    return result.data!
  },

  async updateExchangeConfigs(
    request: UpdateExchangeConfigRequest
  ): Promise<void> {
    const result = await httpClient.put(`${API_BASE}/exchanges`, request)
    if (!result.success) throw new Error('更新交易所配置失败')
  },

  // 创建新的交易所账户
  async createExchange(request: CreateExchangeRequest): Promise<{ id: string }> {
    const result = await httpClient.post<{ id: string }>(`${API_BASE}/exchanges`, request)
    if (!result.success) throw new Error('创建交易所账户失败')
    return result.data!
  },

  // 创建新的交易所账户（加密传输）
  async createExchangeEncrypted(request: CreateExchangeRequest): Promise<{ id: string }> {
    // 检查是否启用了传输加密
    const config = await CryptoService.fetchCryptoConfig()

    if (!config.transport_encryption) {
      // 传输加密禁用时，直接发送明文
      const result = await httpClient.post<{ id: string }>(`${API_BASE}/exchanges`, request)
      if (!result.success) throw new Error('创建交易所账户失败')
      return result.data!
    }

    // 获取RSA公钥
    const publicKey = await CryptoService.fetchPublicKey()

    // 初始化加密服务
    await CryptoService.initialize(publicKey)

    // 获取用户信息
    const userId = localStorage.getItem('user_id') || ''
    const sessionId = sessionStorage.getItem('session_id') || ''

    // 加密敏感数据
    const encryptedPayload = await CryptoService.encryptSensitiveData(
      JSON.stringify(request),
      userId,
      sessionId
    )

    // 发送加密数据
    const result = await httpClient.post<{ id: string }>(
      `${API_BASE}/exchanges`,
      encryptedPayload
    )
    if (!result.success) throw new Error('创建交易所账户失败')
    return result.data!
  },

  // 删除交易所账户
  async deleteExchange(exchangeId: string): Promise<void> {
    const result = await httpClient.delete(`${API_BASE}/exchanges/${exchangeId}`)
    if (!result.success) throw new Error('删除交易所账户失败')
  },

  // 使用加密传输更新交易所配置（自动检测是否启用加密）
  async updateExchangeConfigsEncrypted(
    request: UpdateExchangeConfigRequest
  ): Promise<void> {
    // 检查是否启用了传输加密
    const config = await CryptoService.fetchCryptoConfig()

    if (!config.transport_encryption) {
      // 传输加密禁用时，直接发送明文
      const result = await httpClient.put(`${API_BASE}/exchanges`, request)
      if (!result.success) throw new Error('更新交易所配置失败')
      return
    }

    // 获取RSA公钥
    const publicKey = await CryptoService.fetchPublicKey()

    // 初始化加密服务
    await CryptoService.initialize(publicKey)

    // 获取用户信息（从localStorage或其他地方）
    const userId = localStorage.getItem('user_id') || ''
    const sessionId = sessionStorage.getItem('session_id') || ''

    // 加密敏感数据
    const encryptedPayload = await CryptoService.encryptSensitiveData(
      JSON.stringify(request),
      userId,
      sessionId
    )

    // 发送加密数据
    const result = await httpClient.put(
      `${API_BASE}/exchanges`,
      encryptedPayload
    )
    if (!result.success) throw new Error('更新交易所配置失败')
  },

  // 获取系统状态（支持trader_id）
  async getStatus(traderId?: string): Promise<SystemStatus> {
    const url = traderId
      ? `${API_BASE}/status?trader_id=${traderId}`
      : `${API_BASE}/status`
    const result = await httpClient.get<SystemStatus>(url)
    if (!result.success) throw new Error('获取系统状态失败')
    return result.data!
  },

  // 获取账户信息（支持trader_id）
  async getAccount(traderId?: string): Promise<AccountInfo> {
    const url = traderId
      ? `${API_BASE}/account?trader_id=${traderId}`
      : `${API_BASE}/account`
    const result = await httpClient.get<AccountInfo>(url)
    if (!result.success) throw new Error('获取账户信息失败')
    console.log('Account data fetched:', result.data)
    return result.data!
  },

  // 获取持仓列表（支持trader_id）
  async getPositions(traderId?: string): Promise<Position[]> {
    const url = traderId
      ? `${API_BASE}/positions?trader_id=${traderId}`
      : `${API_BASE}/positions`
    const result = await httpClient.get<Position[]>(url)
    if (!result.success) throw new Error('获取持仓列表失败')
    return result.data!
  },

  // 获取决策日志（支持trader_id）
  async getDecisions(traderId?: string): Promise<DecisionRecord[]> {
    const url = traderId
      ? `${API_BASE}/decisions?trader_id=${traderId}`
      : `${API_BASE}/decisions`
    const result = await httpClient.get<DecisionRecord[]>(url)
    if (!result.success) throw new Error('获取决策日志失败')
    return result.data!
  },

  // 获取最新决策（支持trader_id和limit参数）
  async getLatestDecisions(
    traderId?: string,
    limit: number = 5
  ): Promise<DecisionRecord[]> {
    const params = new URLSearchParams()
    if (traderId) {
      params.append('trader_id', traderId)
    }
    params.append('limit', limit.toString())

    const url = `${API_BASE}/decisions/latest?${params}`
    console.log('Fetching latest decisions from URL:', url) // 添加调试日志
    
    const result = await httpClient.get<DecisionRecord[]>(url)
    if (!result.success) {
      console.error('Failed to get latest decisions:', result.message) // 添加错误日志
      throw new Error(result.message || '获取最新决策失败')
    }
    return result.data!
  },

  // 获取统计信息（支持trader_id）
  async getStatistics(traderId?: string): Promise<Statistics> {
    const url = traderId
      ? `${API_BASE}/statistics?trader_id=${traderId}`
      : `${API_BASE}/statistics`
    const result = await httpClient.get<Statistics>(url)
    if (!result.success) throw new Error('获取统计信息失败')
    return result.data!
  },

  // 获取收益率历史数据（支持trader_id）
  async getEquityHistory(traderId?: string): Promise<any[]> {
    const url = traderId
      ? `${API_BASE}/equity-history?trader_id=${traderId}`
      : `${API_BASE}/equity-history`
    const result = await httpClient.get<any[]>(url)
    if (!result.success) throw new Error('获取历史数据失败')
    return result.data!
  },

  // 批量获取多个交易员的历史数据（无需认证）
  // hours: 可选参数，获取最近N小时的数据（0表示全部数据）
  // 常用值: 24=1天, 72=3天, 168=7天, 720=30天, 0=全部
  async getEquityHistoryBatch(traderIds: string[], hours?: number): Promise<any> {
    const result = await httpClient.post<any>(
      `${API_BASE}/equity-history-batch`,
      { trader_ids: traderIds, hours: hours || 0 }
    )
    if (!result.success) throw new Error('获取批量历史数据失败')
    return result.data!
  },

  // 获取前5名交易员数据（无需认证）
  async getTopTraders(): Promise<any[]> {
    const result = await httpClient.get<any[]>(`${API_BASE}/top-traders`)
    if (!result.success) throw new Error('获取前5名交易员失败')
    return result.data!
  },

  // 获取公开交易员配置（无需认证）
  async getPublicTraderConfig(traderId: string): Promise<any> {
    const result = await httpClient.get<any>(
      `${API_BASE}/trader/${traderId}/config`
    )
    if (!result.success) throw new Error('获取公开交易员配置失败')
    return result.data!
  },

  // 获取竞赛数据（无需认证）
  async getCompetition(): Promise<CompetitionData> {
    const result = await httpClient.get<CompetitionData>(
      `${API_BASE}/competition`
    )
    if (!result.success) throw new Error('获取竞赛数据失败')
    return result.data!
  },

  // 获取服务器IP（需要认证，用于白名单配置）
  async getServerIP(): Promise<{
    public_ip: string
    message: string
  }> {
    const result = await httpClient.get<{
      public_ip: string
      message: string
    }>(`${API_BASE}/server-ip`)
    if (!result.success) throw new Error('获取服务器IP失败')
    return result.data!
  },

  // Backtest APIs
  async getBacktestRuns(params?: {
    state?: string
    search?: string
    limit?: number
    offset?: number
  }): Promise<BacktestRunsResponse> {
    const query = new URLSearchParams()
    if (params?.state) query.set('state', params.state)
    if (params?.search) query.set('search', params.search)
    if (params?.limit) query.set('limit', String(params.limit))
    if (params?.offset) query.set('offset', String(params.offset))
    const url = `${API_BASE}/backtest/runs${query.toString() ? `?${query}` : ''}`
    const result = await httpClient.get<BacktestRunsResponse>(url)
    if (!result.success) throw new Error(result.message || '获取回测运行失败')
    return result.data!
  },

  async startBacktest(config: BacktestStartConfig): Promise<BacktestRunMetadata> {
    const result = await httpClient.post<BacktestRunMetadata>(`${API_BASE}/backtest/start`, { config })
    if (!result.success) throw new Error(result.message || '启动回测失败')
    return result.data!
  },

  async pauseBacktest(runId: string): Promise<BacktestRunMetadata> {
    const result = await httpClient.post<BacktestRunMetadata>(`${API_BASE}/backtest/pause`, { run_id: runId })
    if (!result.success) throw new Error(result.message || '暂停回测失败')
    return result.data!
  },

  async resumeBacktest(runId: string): Promise<BacktestRunMetadata> {
    const result = await httpClient.post<BacktestRunMetadata>(`${API_BASE}/backtest/resume`, { run_id: runId })
    if (!result.success) throw new Error(result.message || '恢复回测失败')
    return result.data!
  },

  async stopBacktest(runId: string): Promise<BacktestRunMetadata> {
    const result = await httpClient.post<BacktestRunMetadata>(`${API_BASE}/backtest/stop`, { run_id: runId })
    if (!result.success) throw new Error(result.message || '停止回测失败')
    return result.data!
  },

  async updateBacktestLabel(
    runId: string,
    label: string
  ): Promise<BacktestRunMetadata> {
    const result = await httpClient.post<BacktestRunMetadata>(`${API_BASE}/backtest/label`, { run_id: runId, label })
    if (!result.success) throw new Error(result.message || '更新回测标签失败')
    return result.data!
  },

  async deleteBacktestRun(runId: string): Promise<void> {
    const result = await httpClient.post(`${API_BASE}/backtest/delete`, { run_id: runId })
    if (!result.success) throw new Error(result.message || '删除回测运行失败')
  },

  async getBacktestStatus(runId: string): Promise<BacktestStatusPayload> {
    const result = await httpClient.get<BacktestStatusPayload>(`${API_BASE}/backtest/status`, { run_id: runId })
    if (!result.success) throw new Error(result.message || '获取回测状态失败')
    return result.data!
  },

  async getBacktestEquity(
    runId: string,
    timeframe?: string,
    limit?: number
  ): Promise<BacktestEquityPoint[]> {
    const query = new URLSearchParams({ run_id: runId })
    if (timeframe) query.set('tf', timeframe)
    if (limit) query.set('limit', String(limit))
    const url = `${API_BASE}/backtest/equity?${query}`
    const result = await httpClient.get<BacktestEquityPoint[]>(url)
    if (!result.success) throw new Error(result.message || '获取回测权益数据失败')
    return result.data!
  },

  async getBacktestTrades(
    runId: string,
    limit = 200
  ): Promise<BacktestTradeEvent[]> {
    const query = new URLSearchParams({
      run_id: runId,
      limit: String(limit),
    })
    const url = `${API_BASE}/backtest/trades?${query}`
    const result = await httpClient.get<BacktestTradeEvent[]>(url)
    if (!result.success) throw new Error(result.message || '获取回测交易数据失败')
    return result.data!
  },

  async getBacktestMetrics(runId: string): Promise<BacktestMetrics> {
    const url = `${API_BASE}/backtest/metrics?run_id=${runId}`
    const result = await httpClient.get<BacktestMetrics>(url)
    if (!result.success) throw new Error(result.message || '获取回测指标失败')
    return result.data!
  },

  async getBacktestKlines(
    runId: string,
    symbol: string,
    timeframe?: string
  ): Promise<BacktestKlinesResponse> {
    const query = new URLSearchParams({ run_id: runId, symbol })
    if (timeframe) query.set('timeframe', timeframe)
    const url = `${API_BASE}/backtest/klines?${query}`
    const result = await httpClient.get<BacktestKlinesResponse>(url)
    if (!result.success) throw new Error(result.message || '获取回测K线数据失败')
    return result.data!
  },

  async getBacktestTrace(
    runId: string,
    cycle?: number
  ): Promise<DecisionRecord> {
    const query = new URLSearchParams({ run_id: runId })
    if (cycle) query.set('cycle', String(cycle))
    const url = `${API_BASE}/backtest/trace?${query}`
    const result = await httpClient.get<DecisionRecord>(url)
    if (!result.success) throw new Error(result.message || '获取回测跟踪数据失败')
    return result.data!
  },

  async getBacktestDecisions(
    runId: string,
    limit = 20,
    offset = 0
  ): Promise<DecisionRecord[]> {
    const query = new URLSearchParams({
      run_id: runId,
      limit: String(limit),
      offset: String(offset),
    })
    const url = `${API_BASE}/backtest/decisions?${query}`
    const result = await httpClient.get<DecisionRecord[]>(url)
    if (!result.success) throw new Error(result.message || '获取回测决策数据失败')
    return result.data!
  },

  async exportBacktest(runId: string): Promise<Blob> {
    const url = `${API_BASE}/backtest/export?run_id=${runId}`
    const result = await httpClient.get<Blob>(url)
    if (!result.success) throw new Error(result.message || '导出回测失败')
    return result.data!
  },

  // Strategy APIs
  async getStrategies(): Promise<Strategy[]> {
    const result = await httpClient.get<{ strategies: Strategy[] }>(`${API_BASE}/strategies`)
    if (!result.success) throw new Error('获取策略列表失败')
    const strategies = result.data?.strategies
    return Array.isArray(strategies) ? strategies : []
  },

  async getStrategy(strategyId: string): Promise<Strategy> {
    const result = await httpClient.get<Strategy>(`${API_BASE}/strategies/${strategyId}`)
    if (!result.success) throw new Error('获取策略失败')
    return result.data!
  },

  async getActiveStrategy(): Promise<Strategy> {
    const result = await httpClient.get<Strategy>(`${API_BASE}/strategies/active`)
    if (!result.success) throw new Error('获取激活策略失败')
    return result.data!
  },

  async getDefaultStrategyConfig(): Promise<StrategyConfig> {
    const result = await httpClient.get<StrategyConfig>(`${API_BASE}/strategies/default-config`)
    if (!result.success) throw new Error('获取默认策略配置失败')
    return result.data!
  },

  async createStrategy(data: {
    name: string
    description: string
    config: StrategyConfig
  }): Promise<Strategy> {
    const result = await httpClient.post<Strategy>(`${API_BASE}/strategies`, data)
    if (!result.success) throw new Error('创建策略失败')
    return result.data!
  },

  async updateStrategy(
    strategyId: string,
    data: {
      name?: string
      description?: string
      config?: StrategyConfig
    }
  ): Promise<Strategy> {
    const result = await httpClient.put<Strategy>(`${API_BASE}/strategies/${strategyId}`, data)
    if (!result.success) throw new Error('更新策略失败')
    return result.data!
  },

  async deleteStrategy(strategyId: string): Promise<void> {
    const result = await httpClient.delete(`${API_BASE}/strategies/${strategyId}`)
    if (!result.success) throw new Error('删除策略失败')
  },

  async activateStrategy(strategyId: string): Promise<Strategy> {
    const result = await httpClient.post<Strategy>(`${API_BASE}/strategies/${strategyId}/activate`)
    if (!result.success) throw new Error('激活策略失败')
    return result.data!
  },

  async duplicateStrategy(strategyId: string): Promise<Strategy> {
    const result = await httpClient.post<Strategy>(`${API_BASE}/strategies/${strategyId}/duplicate`)
    if (!result.success) throw new Error('复制策略失败')
    return result.data!
  },

  // Debate Arena APIs
  async getDebates(): Promise<DebateSession[]> {
    const result = await httpClient.get<DebateSession[]>(`${API_BASE}/debates`)
    if (!result.success) throw new Error('获取辩论列表失败')
    return Array.isArray(result.data) ? result.data : []
  },

  async getDebate(debateId: string): Promise<DebateSessionWithDetails> {
    const result = await httpClient.get<DebateSessionWithDetails>(`${API_BASE}/debates/${debateId}`)
    if (!result.success) throw new Error('获取辩论详情失败')
    return result.data!
  },

  async createDebate(request: CreateDebateRequest): Promise<DebateSessionWithDetails> {
    const result = await httpClient.post<DebateSessionWithDetails>(`${API_BASE}/debates`, request)
    if (!result.success) throw new Error('创建辩论失败')
    return result.data!
  },

  async startDebate(debateId: string): Promise<void> {
    const result = await httpClient.post(`${API_BASE}/debates/${debateId}/start`)
    if (!result.success) throw new Error('启动辩论失败')
  },

  async cancelDebate(debateId: string): Promise<void> {
    const result = await httpClient.post(`${API_BASE}/debates/${debateId}/cancel`)
    if (!result.success) throw new Error('取消辩论失败')
  },

  async executeDebate(debateId: string, traderId: string): Promise<DebateSessionWithDetails> {
    const result = await httpClient.post<{ message: string; session: DebateSessionWithDetails }>(
      `${API_BASE}/debates/${debateId}/execute`,
      { trader_id: traderId }
    )
    if (!result.success) throw new Error('执行交易失败')
    return result.data!.session
  },

  async deleteDebate(debateId: string): Promise<void> {
    const result = await httpClient.delete(`${API_BASE}/debates/${debateId}`)
    if (!result.success) throw new Error('删除辩论失败')
  },

  async getDebateMessages(debateId: string): Promise<DebateMessage[]> {
    const result = await httpClient.get<DebateMessage[]>(`${API_BASE}/debates/${debateId}/messages`)
    if (!result.success) throw new Error('获取辩论消息失败')
    return result.data!
  },

  async getDebateVotes(debateId: string): Promise<DebateVote[]> {
    const result = await httpClient.get<DebateVote[]>(`${API_BASE}/debates/${debateId}/votes`)
    if (!result.success) throw new Error('获取辩论投票失败')
    return result.data!
  },

  async getDebatePersonalities(): Promise<DebatePersonalityInfo[]> {
    const result = await httpClient.get<DebatePersonalityInfo[]>(`${API_BASE}/debates/personalities`)
    if (!result.success) throw new Error('获取AI性格列表失败')
    return result.data!
  },

  // SSE stream for live debate updates
  createDebateStream(debateId: string): EventSource {
    // For SSE streams, we still need to pass the token as a query parameter
    // since headers cannot be sent with EventSource
    const token = localStorage.getItem('auth_token')
    if (!token) {
      throw new Error('Authentication token not found')
    }
    return new EventSource(`${API_BASE}/debates/${debateId}/stream?token=${token}`)
  },

  // Position History API
  async getPositionHistory(traderId: string, limit: number = 100): Promise<PositionHistoryResponse> {
    const result = await httpClient.get<PositionHistoryResponse>(
      `${API_BASE}/positions/history?trader_id=${traderId}&limit=${limit}`
    )
    if (!result.success) throw new Error('获取历史仓位失败')
    return result.data!
  },

  // Manual sync position history
  async syncPositionHistory(traderId: string): Promise<{ message: string; created: number; skipped: number }> {
    const result = await httpClient.post<{ message: string; created: number; skipped: number }>(
      `${API_BASE}/traders/${traderId}/sync-positions`
    )
    if (!result.success) throw new Error('同步历史仓位失败')
    return result.data!
  },

  // 手动触发AI决策
  async triggerDecision(traderId: string): Promise<{ message: string; result?: any; execution_time_ms?: number; execution_time_formatted?: string }> {
    const result = await httpClient.post(
      `${API_BASE}/traders/${traderId}/execute-decision`
    )
    if (!result.success) throw new Error(result.message || '手动触发决策失败')
    return result.data!
  },

  // 获取系统配置
  async getSystemConfig(): Promise<SystemConfig> {
    const result = await httpClient.get<SystemConfig>(`${API_BASE}/system-config`)
    if (!result.success) throw new Error('获取系统配置失败')
    return result.data!
  },

  // 生成完整的AI提示词（包含实时数据）
  async generateFullPrompt(traderId: string): Promise<{ 
    success: boolean, 
    data?: {
      system_prompt: string
      user_prompt?: string
      success: boolean
    },
    message?: string 
  }> {
    const result = await httpClient.post<any>(
      `${API_BASE}/test/generate-full-prompt`,
      { trader_id: traderId }
    )
    return result
  },

  // 提交AI决策
  async submitAIDecision(traderId: string, decisionJson: string): Promise<{ 
    success: boolean, 
    data?: any,
    message?: string 
  }> {
    const result = await httpClient.post<any>(
      `${API_BASE}/test/submit-ai-decision`,
      { 
        trader_id: traderId,
        decision_json: decisionJson
      }
    )
    return result
  },

  // 删除单个决策记录
  async deleteDecision(decisionId: number, traderId: string): Promise<{ 
    success: boolean, 
    message?: string 
  }> {
    const result = await httpClient.delete(`${API_BASE}/decisions/${decisionId}?trader_id=${traderId}`)
    return result
  },
}
