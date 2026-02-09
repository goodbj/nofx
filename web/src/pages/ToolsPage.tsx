import { useState, useEffect } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { useLanguage } from '../contexts/LanguageContext'
import { Zap, Settings, Activity, BarChart3, Shield, Target, FileText, Globe, Eye, Play, Loader2, RefreshCw, Clock, Bot, Terminal, Code, Send, Download, Upload, Clipboard, ClipboardCheck, ClipboardPaste, UserPlus } from 'lucide-react'
import type { Strategy, StrategyConfig, AIModel } from '../types'
import { confirmToast, notify } from '../lib/notify'
import { DeepVoidBackground } from '../components/DeepVoidBackground'
import AutoLoginTestPage from './AutoLoginTestPage'

const API_BASE = import.meta.env.VITE_API_BASE || ''

export function ToolsPage() {
  const { token } = useAuth()
  const { language } = useLanguage()
  
  const [activeTool, setActiveTool] = useState<'ai-semi-auto' | 'auto-login'>('ai-semi-auto')
  const [aiModels, setAiModels] = useState<AIModel[]>([])
  const [selectedModelId, setSelectedModelId] = useState<string>('')
  
  // AI Semi-Auto states
  const [manualAIDecision, setManualAIDecision] = useState<string>('')
  const [submitAILoading, setSubmitAILoading] = useState(false)
  const [submitAIResult, setSubmitAIResult] = useState<string>('')
  const [selectedSemiAutoTraderId, setSelectedSemiAutoTraderId] = useState<string>('')
  const [traders, setTraders] = useState<any[]>([])
  
  // Prompt preview states
  const [promptPreview, setPromptPreview] = useState<{
    system_prompt: string
    user_prompt?: string
    prompt_variant: string
    config_summary: Record<string, unknown>
  } | null>(null)
  const [isLoadingPrompt, setIsLoadingPrompt] = useState(false)

  // Manual scan data states
  const [selectedScanDataTraderId, setSelectedScanDataTraderId] = useState<string>('')
  const [scanData, setScanData] = useState<any>(null)
  const [isLoadingScanData, setIsLoadingScanData] = useState(false)

  const t = (key: string) => {
    const translations: Record<string, Record<string, string>> = {
      tools: { zh: '工具', en: 'Tools' },
      utilities: { zh: '实用工具', en: 'Utilities' },
      aiSemiAuto: { zh: 'AI 半自动', en: 'AI Semi-Auto' },
      autoLogin: { zh: '交易员登录', en: 'Trader Login' },
      otherTool: { zh: '其他工具', en: 'Other Tools' },
      aiManualWorkflow: { zh: 'AI 手动工作流', en: 'AI Manual Workflow' },
      aiManualWorkflowDesc: { zh: '手动复制Prompt到AI服务，粘贴决策结果', en: 'Manually copy prompt to AI service, paste decision result' },
      generatePrompt: { zh: '生成 Prompt', en: 'Generate Prompt' },
      submitDecision: { zh: '提交决策', en: 'Submit Decision' },
      loadPrompt: { zh: '生成 Prompt', en: 'Generate Prompt' },
      refreshPrompt: { zh: '刷新', en: 'Refresh' },
      promptVariant: { zh: '风格', en: 'Style' },
      balanced: { zh: '平衡', en: 'Balanced' },
      aggressive: { zh: '激进', en: 'Aggressive' },
      conservative: { zh: '保守', en: 'Conservative' },
      selectTrader: { zh: '选择交易员', en: 'Select Trader' },
      pasteDecisionJson: { zh: '粘贴AI决策JSON', en: 'Paste AI Decision JSON' },
      clear: { zh: '清空', en: 'Clear' },
      paste: { zh: '粘贴', en: 'Paste' },
      executionResult: { zh: '执行结果', en: 'Execution Result' },
      trader: { zh: '交易员', en: 'Trader' },
      pleaseSelectTrader: { zh: '请选择交易员', en: 'Please select trader' },
      pleasePasteDecisionJson: { zh: '请粘贴AI决策JSON', en: 'Please paste AI decision JSON' },
      executionCompleted: { zh: '执行完成！', en: 'Execution completed!' },
      executing: { zh: '执行中...', en: 'Executing...' },
      generating: { zh: '生成中...', en: 'Generating...' },
      fullPrompt: { zh: '完整 Prompt (System + User)', en: 'Full Prompt (System + User)' },
      copy: { zh: '复制', en: 'Copy' },
      estimatedCost: { zh: '预估成本', en: 'Estimated Cost' },
      potentialSavings: { zh: '潜在节省', en: 'Potential Savings' },
      optimizePromptForCost: { zh: '优化Prompt以降低成本', en: 'Optimize prompt for cost' },
    }
    return translations[key]?.[language] || key
  }

  // Fetch AI Models
  const fetchAiModels = async () => {
    if (!token) return
    try {
      const response = await fetch(`${API_BASE}/api/models`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (response.ok) {
        const data = await response.json()
        const allModels = Array.isArray(data) ? data : (data.models || [])
        const enabledModels = allModels.filter((m: AIModel) => m.enabled)
        setAiModels(enabledModels)
        if (enabledModels.length > 0 && !selectedModelId) {
          setSelectedModelId(enabledModels[0].id)
        }
      }
    } catch (err) {
      console.error('Failed to fetch AI models:', err)
    }
  }

  // Fetch traders
  const fetchTraders = async () => {
    if (!token) return
    try {
      const response = await fetch(`${API_BASE}/api/my-traders`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (response.ok) {
        const data = await response.json()
        setTraders(data)
      } else {
        console.error('Failed to fetch traders:', response.status, response.statusText)
      }
    } catch (err) {
      console.error('Error fetching traders:', err)
    }
  }



  // Fetch trader config by ID
  const getTraderConfig = async (traderId: string) => {
    if (!token) return null
    
    try {
      const response = await fetch(`${API_BASE}/api/traders/${traderId}/config`, {
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
      })
      
      if (response.ok) {
        const data = await response.json()
        return data
      } else {
        console.error('Failed to fetch trader config:', response.status, response.statusText)
        return null
      }
    } catch (err) {
      console.error('Error fetching trader config:', err)
      return null
    }
  }

  // Fetch prompt preview
  const fetchPromptPreview = async () => {
    if (!selectedSemiAutoTraderId) {
      notify.warning(language === 'zh' ? '请先选择一个交易员' : 'Please select a trader first')
      return
    }
    
    if (!token) return
    
    setIsLoadingPrompt(true)
    try {
      const response = await fetch(`${API_BASE}/api/test/generate-full-prompt`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          trader_id: selectedSemiAutoTraderId,
        }),
      })
      
      if (response.ok) {
        const data = await response.json()
        if (data.success) {
          setPromptPreview({
            system_prompt: data.system_prompt || '',
            user_prompt: data.user_prompt || '',
            prompt_variant: 'balanced',
            config_summary: { trader: data.trader_name }
          })
          notify.success(language === 'zh' ? 'Prompt 生成成功（含真实数据）' : 'Real-time prompt generated successfully')
        } else {
          notify.error(data.error || 'Failed to generate prompt')
        }
      } else {
        const errorData = await response.json()
        notify.error(`Error: ${errorData.error || 'Failed to generate prompt'}`)
      }
    } catch (err) {
      console.error('Error fetching prompt preview:', err)
      notify.error(language === 'zh' ? '生成Prompt时出错' : 'Error generating prompt')
    } finally {
      setIsLoadingPrompt(false)
    }
  }

  useEffect(() => {
    fetchAiModels()
    fetchTraders()
  }, [token])

  return (
    <DeepVoidBackground className="h-[calc(100vh-64px)] flex flex-col bg-nofx-bg relative overflow-hidden">
      {/* Header */}
      <div className="flex-shrink-0 px-4 py-3 border-b border-nofx-gold/20 bg-nofx-bg/60 backdrop-blur-md z-10">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-gradient-to-br from-nofx-gold to-yellow-500">
              <Zap className="w-5 h-5 text-black" />
            </div>
            <div>
              <h1 className="text-lg font-bold text-nofx-text">{language === 'zh' ? 'AI 手动工作流工具' : 'AI Manual Workflow Tools'}</h1>
              <p className="text-xs text-nofx-text-muted">{t('utilities')}</p>
            </div>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex overflow-hidden">
        {/* Tool Selection Sidebar */}
        <div className="w-48 flex-shrink-0 border-r border-nofx-gold/20 overflow-y-auto bg-nofx-bg/30 backdrop-blur-sm z-10">
          <div className="p-2">
            <div className="space-y-1">
              <button
                onClick={() => setActiveTool('ai-semi-auto')}
                className={`w-full flex items-center gap-2 px-3 py-2.5 rounded-lg text-sm transition-colors ${
                  activeTool === 'ai-semi-auto'
                    ? 'ring-1 ring-blue-500/50 bg-blue-500/10 text-blue-400'
                    : 'hover:bg-white/5 text-nofx-text-muted'
                }`}
              >
                <Zap className="w-4 h-4" />
                <span>{t('aiSemiAuto')}</span>
              </button>
              <button
                onClick={() => setActiveTool('auto-login')}
                className={`w-full flex items-center gap-2 px-3 py-2.5 rounded-lg text-sm transition-colors ${
                  activeTool === 'auto-login'
                    ? 'ring-1 ring-blue-500/50 bg-blue-500/10 text-blue-400'
                    : 'hover:bg-white/5 text-nofx-text-muted'
                }`}
              >
                <UserPlus className="w-4 h-4" />
                <span>{t('autoLogin')}</span>
              </button>
            </div>
          </div>
        </div>

        {/* Tool Content Area */}
        <div className="flex-1 overflow-y-auto">
          {activeTool === 'ai-semi-auto' && (
            <div className="p-4">
              <div className="bg-gradient-to-r from-blue-900/30 to-indigo-900/30 p-6 rounded-xl border border-blue-700/50 max-w-5xl mx-auto">
                <h3 className="text-xl font-bold text-blue-300 mb-4 flex items-center gap-2">
                  <Bot className="w-6 h-6" />
                  {t('aiManualWorkflow')}
                </h3>
                        
                <div className="space-y-6">
                  {/* 使用说明 */}
                  <div className="bg-indigo-900/40 p-4 rounded-lg border border-indigo-700/50">
                    <p className="text-sm font-semibold text-indigo-200 mb-2">📝 {language === 'zh' ? '操作步骤' : 'Instructions'}:</p>
                    <ol className="text-xs text-indigo-100 space-y-1 list-decimal list-inside">
                      <li>{language === 'zh' ? '选择一个交易员' : 'Select a trader'}</li>
                      <li>{language === 'zh' ? '点击"生成 Prompt"获取包含实时数据的提示词' : 'Click "Generate Prompt" to get real-time data'}</li>
                      <li>{language === 'zh' ? '分别复制 System 和 User Prompt 到 AI 服务（如 DeepSeek 网页版）' : 'Copy System and User Prompts to AI service'}</li>
                      <li>{language === 'zh' ? '将 AI 返回的决策 JSON 粘贴到最下方' : 'Paste AI decision JSON below'}</li>
                      <li>{language === 'zh' ? '点击"提交决策"由交易员执行' : 'Click "Submit Decision" to execute'}</li>
                    </ol>
                  </div>
        
                  {/* 第一步：交易员选择 */}
                  <div className="bg-gray-800/50 p-4 rounded-lg border border-gray-700">
                    <label className="block text-sm font-medium mb-3 text-gray-200">
                      1. {t('selectTrader')}
                    </label>
                    <div className="flex gap-3">
                      <select
                        value={selectedSemiAutoTraderId}
                        onChange={(e) => setSelectedSemiAutoTraderId(e.target.value)}
                        className="flex-1 p-2.5 rounded bg-gray-900 border border-gray-600 text-white focus:ring-2 focus:ring-blue-500"
                      >
                        <option value="">{t('pleaseSelectTrader')}</option>
                        {traders.map((trader) => (
                          <option key={trader.trader_id} value={trader.trader_id}>
                            {trader.trader_name} ({trader.exchange_id})
                          </option>
                        ))}
                      </select>
                      <button
                        onClick={fetchPromptPreview}
                        disabled={isLoadingPrompt || !selectedSemiAutoTraderId}
                        className="px-6 py-2 rounded bg-blue-600 hover:bg-blue-500 text-white font-semibold disabled:bg-gray-700 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
                      >
                        {isLoadingPrompt ? <Loader2 className="w-4 h-4 animate-spin" /> : <RefreshCw className="w-4 h-4" />}
                        {t('generatePrompt')}
                      </button>
                    </div>
                  </div>
        
                  {/* 第二步：Prompt 展示与复制 - 始终显示，固定高度约40px */}
                  <div className="bg-gray-800/50 p-4 rounded-lg border border-gray-700">
                    <div className="flex justify-between items-center mb-2">
                      <label className="text-sm font-bold text-blue-400 flex items-center gap-2">
                        <Code className="w-4 h-4" />
                        完整 Prompt (System + User)
                      </label>
                      <div className="flex gap-2">
                        <div className="text-xs text-gray-400 flex items-center">
                          {`字符数: ${(promptPreview ? (promptPreview.system_prompt + promptPreview.user_prompt).length : 0)}`}
                        </div>
                        <button
                          onClick={() => {
                            if (!promptPreview) return;
                            const fullPrompt = `System Prompt:
${promptPreview.system_prompt}

User Prompt (包含实时行情、持仓、指标数据):
${promptPreview.user_prompt}`;
                            navigator.clipboard.writeText(fullPrompt);
                            notify.success(language === 'zh' ? '已复制完整 Prompt' : 'Full prompt copied');
                          }}
                          className={`text-xs px-3 py-1.5 rounded ${
                            promptPreview 
                              ? 'bg-blue-600/30 text-blue-300 hover:bg-blue-600/50 border border-blue-500/30' 
                              : 'bg-gray-600/30 text-gray-400 cursor-not-allowed'
                          } flex items-center gap-1 transition-colors`}
                          disabled={!promptPreview}
                        >
                          <Clipboard className="w-3.5 h-3.5" />
                          {t('copy')}
                        </button>
                      </div>
                    </div>
                    <pre className="text-[11px] text-gray-300 whitespace-pre-wrap font-mono bg-black/50 p-3 rounded border border-gray-700 overflow-auto max-h-[40px] min-h-[40px]">
                      {promptPreview 
                        ? `System Prompt:
${promptPreview.system_prompt}

User Prompt (包含实时行情、持仓、指标数据):
${promptPreview.user_prompt}`
                        : '点击"生成 Prompt"按钮以获取完整提示词'}
                    </pre>
                    {promptPreview && (
                      <div className="mt-2 text-xs text-gray-500 flex justify-between">
                        <span>提示: 大多数AI模型上下文长度限制在 32K-128K tokens 之间，1个中文字符约等于1-2个tokens</span>
                        <span className={`${(promptPreview.system_prompt + promptPreview.user_prompt).length > 50000 ? 'text-red-400' : 'text-green-400'}`}>
                          {(promptPreview.system_prompt + promptPreview.user_prompt).length > 50000 ? '⚠️ 字符数较多，可能超出模型限制' : '✅ 字符数正常'}
                        </span>
                      </div>
                    )}
                  </div>
        
                  {/* 第三步：决策提交 */}
                  <div className="bg-gray-800/50 p-4 rounded-lg border border-gray-700">
                    <div className="flex justify-between items-center mb-3">
                      <label className="text-sm font-medium text-gray-200">
                        3. {t('pasteDecisionJson')}
                      </label>
                      <div className="flex gap-2">
                        <button
                          onClick={async () => {
                            try {
                              const text = await navigator.clipboard.readText();
                              setManualAIDecision(text);
                              notify.success(language === 'zh' ? '已从剪贴板粘贴内容' : 'Content pasted from clipboard');
                            } catch (err) {
                              console.error('Failed to read clipboard contents: ', err);
                              notify.error(language === 'zh' ? '无法访问剪贴板，请检查权限' : 'Could not access clipboard');
                            }
                          }}
                          className="text-xs px-2 py-1 text-gray-400 hover:text-white flex items-center gap-1"
                          title={language === 'zh' ? '从剪贴板粘贴' : 'Paste from clipboard'}
                        >
                          <ClipboardPaste className="w-3 h-3" />
                          {t('paste')}
                        </button>
                        <button
                          onClick={() => setManualAIDecision('')}
                          className="text-xs px-2 py-1 text-gray-400 hover:text-white"
                        >
                          {t('clear')}
                        </button>
                      </div>
                    </div>
                    <textarea
                      value={manualAIDecision}
                      onChange={(e) => setManualAIDecision(e.target.value)}
                      placeholder='粘贴 AI 返回的 JSON，例如: { "decisions": [...] }'
                      className="w-full p-3 rounded bg-gray-900 border border-gray-600 text-white font-mono text-sm focus:ring-2 focus:ring-emerald-500 h-[150px]"
                    />
                    <button
                      onClick={async () => {
                        if (!selectedSemiAutoTraderId) {
                          notify.warning(t('pleaseSelectTrader'));
                          return;
                        }
                        if (!manualAIDecision.trim()) {
                          notify.warning(t('pleasePasteDecisionJson'));
                          return;
                        }

                        setSubmitAILoading(true);
                        setSubmitAIResult('');
                                  
                        try {
                          const response = await fetch(`${API_BASE}/api/test/submit-ai-decision`, {
                            method: 'POST',
                            headers: {
                              'Content-Type': 'application/json',
                              'Authorization': `Bearer ${token}`,
                            },
                            body: JSON.stringify({
                              trader_id: selectedSemiAutoTraderId,
                              decision_json: manualAIDecision
                            })
                          });

                          const data = await response.json();
                                                  
                          if (!response.ok) {
                            setSubmitAIResult(`❌ Error (${response.status}): ${data.error || JSON.stringify(data, null, 2)}`);
                            notify.error(language === 'zh' ? '提交失败' : 'Submission failed');
                          } else {
                            setSubmitAIResult(`✅ ${t('executionCompleted')}\n\n${JSON.stringify(data, null, 2)}`);
                            notify.success(t('executionCompleted'));
                          }
                        } catch (error) {
                          setSubmitAIResult(`❌ ${language === 'zh' ? '错误：' : 'Error: '}${error}`);
                          notify.error('Network Error');
                        } finally {
                          setSubmitAILoading(false);
                        }
                      }}
                      disabled={submitAILoading || !selectedSemiAutoTraderId || !manualAIDecision.trim()}
                      className="mt-4 w-full px-4 py-3 rounded-lg bg-gradient-to-r from-emerald-600 to-green-600 text-white font-bold hover:from-emerald-700 hover:to-green-700 disabled:from-gray-700 disabled:to-green-700 disabled:cursor-not-allowed transition-all flex items-center justify-center gap-2"
                    >
                      {submitAILoading ? <Loader2 className="w-5 h-5 animate-spin" /> : <Send className="w-5 h-5" />}
                      {t('submitDecision')}
                    </button>
                  </div>
        
                  {/* 执行结果 */}
                  {submitAIResult && (
                    <div className="bg-black/50 p-4 rounded-lg border border-gray-700">
                      <p className="text-xs font-bold text-gray-400 mb-2 uppercase tracking-wider">{t('executionResult')}:</p>
                      <pre className="text-xs text-emerald-400 whitespace-pre-wrap font-mono overflow-auto max-h-[200px]">
                        {submitAIResult}
                      </pre>
                    </div>
                  )}
                </div>
              </div>
            </div>
          )}
          
          {activeTool === 'auto-login' && (
            <div className="p-4 h-full">
              <AutoLoginTestPage />
            </div>
          )}
        </div>
      </div>
    </DeepVoidBackground>
  )
}

export default ToolsPage