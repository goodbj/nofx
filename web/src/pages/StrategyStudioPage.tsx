import { useState, useEffect, useCallback, useRef } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { useLanguage } from '../contexts/LanguageContext'
import {
  Plus,
  Copy,
  Trash2,
  Check,
  ChevronDown,
  ChevronRight,
  Settings,
  BarChart3,
  Target,
  Shield,
  Zap,
  Activity,
  Save,
  Sparkles,
  Eye,
  Play,
  FileText,
  Loader2,
  RefreshCw,
  Clock,
  Bot,
  Terminal,
  Code,
  Send,
  Download,
  Upload,
  Globe,
  Clipboard,
  ClipboardCheck,
} from 'lucide-react'
import type { Strategy, StrategyConfig, AIModel, SystemConfig } from '../types'
import { confirmToast, notify } from '../lib/notify'
import { CoinSourceEditor } from '../components/strategy/CoinSourceEditor'
import { IndicatorEditor } from '../components/strategy/IndicatorEditor'
import { RiskControlEditor } from '../components/strategy/RiskControlEditor'
import { PromptSectionsEditor } from '../components/strategy/PromptSectionsEditor'
import { PublishSettingsEditor } from '../components/strategy/PublishSettingsEditor'
import { DeepVoidBackground } from '../components/DeepVoidBackground'

const API_BASE = import.meta.env.VITE_API_BASE || ''

export function StrategyStudioPage() {
  const { token } = useAuth()
  const { language } = useLanguage()

  const [strategies, setStrategies] = useState<Strategy[]>([])
  const [selectedStrategy, setSelectedStrategy] = useState<Strategy | null>(null)
  const [editingConfig, setEditingConfig] = useState<StrategyConfig | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isSaving, setIsSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [hasChanges, setHasChanges] = useState(false)

  // AI Models for test run
  const [aiModels, setAiModels] = useState<AIModel[]>([])
  const [selectedModelId, setSelectedModelId] = useState<string>('')
  
  // System configuration for model token limits
  const [systemConfig, setSystemConfig] = useState<SystemConfig | null>(null)

  // Accordion states for left panel
  const [expandedSections, setExpandedSections] = useState({
    coinSource: true,
    indicators: false,
    riskControl: false,
    promptSections: false,
    customPrompt: false,
    publishSettings: false,
  })

  // Right panel states
  const [activeRightTab, setActiveRightTab] = useState<'prompt' | 'test'>('prompt')
  const [promptPreview, setPromptPreview] = useState<{
    system_prompt: string
    user_prompt?: string
    prompt_variant: string
    config_summary: Record<string, unknown>
  } | null>(null)
  const [isLoadingPrompt, setIsLoadingPrompt] = useState(false)
  const [selectedVariant, setSelectedVariant] = useState('balanced')

  // AI Test Run states
  const [aiTestResult, setAiTestResult] = useState<{
    system_prompt?: string
    user_prompt?: string
    ai_response?: string
    reasoning?: string
    decisions?: unknown[]
    error?: string
    duration_ms?: number
  } | null>(null)
  const [isRunningAiTest, setIsRunningAiTest] = useState(false)
  const [copiedStates, setCopiedStates] = useState<Record<string, boolean>>({})

  // 获取AI模型的最大token限制
  const getMaxTokensForModel = (modelId: string): number => {
    console.log('getMaxTokensForModel called with modelId:', modelId);
    console.log('Current systemConfig:', systemConfig);
    console.log('Current aiModels:', aiModels);
    
    const model = aiModels.find(m => m.id === modelId)
    if (!model) {
      console.log('Model not found, returning default');
      return systemConfig?.model_max_tokens?.default || 4096 // 默认值
    }
    
    console.log('Found model:', model);
    
    // 根据系统配置中的模型token限制
    if (systemConfig?.model_max_tokens) {
      // 检查模型名称中是否包含特定关键词
      const modelNameLower = model.name.toLowerCase()
      
      if (modelNameLower.includes('deepseek')) {
        const result = systemConfig.model_max_tokens['deepseek'] || 32768;
        console.log('DeepSeek model, returning:', result);
        return result;
      } else if (modelNameLower.includes('gpt-4')) {
        const result = systemConfig.model_max_tokens['gpt-4'] || 128000;
        console.log('GPT-4 model, returning:', result);
        return result;
      } else if (modelNameLower.includes('gpt-3.5')) {
        const result = systemConfig.model_max_tokens['gpt-3.5'] || 16384;
        console.log('GPT-3.5 model, returning:', result);
        return result;
      } else if (modelNameLower.includes('claude')) {
        const result = systemConfig.model_max_tokens['claude'] || 200000;
        console.log('Claude model, returning:', result);
        return result;
      } else if (modelNameLower.includes('qwen')) {
        const result = systemConfig.model_max_tokens['qwen'] || 32768;
        console.log('Qwen model, returning:', result);
        return result;
      }
    }
    
    // 如果没有找到特定配置，使用默认值
    const defaultValue = systemConfig?.model_max_tokens?.['default'] || 4096;
    console.log('Returning default value:', defaultValue);
    return defaultValue;
  }

  // 估算AI模型调用成本
  const estimateCost = (modelId: string, inputLength: number, outputLength: number = 0): number => {
    console.log('estimateCost called with modelId:', modelId, 'inputLength:', inputLength, 'outputLength:', outputLength);
    const model = aiModels.find(m => m.id === modelId)
    if (!model) {
      console.log('Model not found for cost estimation');
      return 0
    }
    
    console.log('Found model for cost estimation:', model);
    console.log('Current systemConfig:', systemConfig);
    
    // 使用从后端获取的模型定价信息
    const modelPricing = systemConfig?.model_pricing;
    if (!modelPricing) {
      console.log('Model pricing not available, using fallback');
      // 如果定价信息不可用，使用默认定价
      return calculateFallbackCost(model.name, inputLength, outputLength);
    }
    
    // 根据模型名称匹配定价
    const lowerModelName = model.name.toLowerCase();
    let matchedPricing = null;
    
    // 精确匹配
    if (modelPricing[lowerModelName]) {
      matchedPricing = modelPricing[lowerModelName];
    } else {
      // 模糊匹配
      const availableKeys = Object.keys(modelPricing);
      for (const key of availableKeys) {
        if (lowerModelName.includes(key) || key.includes(lowerModelName)) {
          matchedPricing = modelPricing[key];
          break;
        }
      }
    }
    
    console.log('Matched pricing for model:', matchedPricing);
    
    if (matchedPricing) {
      // 定价是每百万tokens的价格，需要转换为每千tokens的价格
      const pricePerThousandInput = matchedPricing.input_price / 1000;
      const pricePerThousandOutput = matchedPricing.output_price / 1000;
      
      console.log('Using matched pricing - input:', pricePerThousandInput, 'output:', pricePerThousandOutput);
      return calculateCost({ input: pricePerThousandInput, output: pricePerThousandOutput }, inputLength, outputLength);
    } else {
      console.log('No pricing found, using fallback');
      return calculateFallbackCost(model.name, inputLength, outputLength);
    }
  }
  
  // 辅助函数：计算成本
  const calculateCost = (price: { input: number; output: number }, inputLength: number, outputLength: number): number => {
    // 将字符数粗略转换为token数（通常1 token ≈ 4 characters）
    const inputTokens = Math.ceil(inputLength / 4)
    const outputTokens = Math.ceil(outputLength / 4)
    console.log('Calculated tokens - input:', inputTokens, 'output:', outputTokens);
    
    // 计算成本
    const inputCost = (inputTokens / 1000) * price.input;
    const outputCost = (outputTokens / 1000) * price.output;
    const totalCost = inputCost + outputCost;
    console.log('Calculated cost - input:', inputCost, 'output:', outputCost, 'total:', totalCost);
    
    return totalCost;
  };
  
  // 备用成本计算函数
  const calculateFallbackCost = (modelName: string, inputLength: number, outputLength: number): number => {
    // 这里使用一些典型的模型定价（每1000 tokens的价格）
    const pricing: Record<string, { input: number; output: number }> = {
      'gpt-3.5-turbo': { input: 0.0005, output: 0.0015 }, // $0.5/1M tokens 输入, $1.5/1M tokens 输出
      'gpt-4': { input: 0.03, output: 0.06 }, // $30/1M tokens 输入, $60/1M tokens 输出
      'gpt-4-turbo': { input: 0.01, output: 0.03 }, // $10/1M tokens 输入, $30/1M tokens 输出
      'gpt-4o': { input: 0.01, output: 0.03 }, // $10/1M tokens 输入, $30/1M tokens 输出
      'gpt-4omini': { input: 0.0006, output: 0.0018 }, // $0.6/1M tokens 输入, $1.8/1M tokens 输出
      'claude-3-haiku': { input: 0.00025, output: 0.00125 }, // $0.25/1M tokens 输入, $1.25/1M tokens 输出
      'claude-3-sonnet': { input: 0.003, output: 0.015 }, // $3/1M tokens 输入, $15/1M tokens 输出
      'claude-3-opus': { input: 0.015, output: 0.075 }, // $15/1M tokens 输入, $75/1M tokens 输出
      'gemini-pro': { input: 0.000125, output: 0.000375 }, // $0.125/1M tokens 输入, $0.375/1M tokens 输出
      'gemini-flash': { input: 0.00005, output: 0.00015 }, // $0.05/1M tokens 输入, $0.15/1M tokens 输出
      'deepseek': { input: 0.0001, output: 0.0002 }, // $0.1/1M tokens 输入, $0.2/1M tokens 输出
      'deepseek-coder': { input: 0.0001, output: 0.0002 }, // $0.1/1M tokens 输入, $0.2/1M tokens 输出
      'qwen-max': { input: 0.0005, output: 0.002 }, // $0.5/1M tokens 输入, $2/1M tokens 输出
      'qwen-plus': { input: 0.0004, output: 0.0016 }, // $0.4/1M tokens 输入, $1.6/1M tokens 输出
      'qwen-turbo': { input: 0.0001, output: 0.0002 }, // $0.1/1M tokens 输入, $0.2/1M tokens 输出
    };
    
    // 根据模型名称匹配定价
    const modelKey = Object.keys(pricing).find(key => 
      modelName.toLowerCase().includes(key.toLowerCase()) || 
      key.toLowerCase().includes(modelName.toLowerCase())
    );
    
    if (!modelKey) {
      // 尝试使用通用匹配
      const lowerModelName = modelName.toLowerCase();
      if (lowerModelName.includes('gpt-4')) {
        return calculateCost(pricing['gpt-4'], inputLength, outputLength);
      } else if (lowerModelName.includes('gpt-3.5')) {
        return calculateCost(pricing['gpt-3.5-turbo'], inputLength, outputLength);
      } else if (lowerModelName.includes('claude')) {
        if (lowerModelName.includes('opus')) {
          return calculateCost(pricing['claude-3-opus'], inputLength, outputLength);
        } else if (lowerModelName.includes('sonnet')) {
          return calculateCost(pricing['claude-3-sonnet'], inputLength, outputLength);
        } else {
          return calculateCost(pricing['claude-3-haiku'], inputLength, outputLength);
        }
      } else if (lowerModelName.includes('gemini')) {
        if (lowerModelName.includes('flash')) {
          return calculateCost(pricing['gemini-flash'], inputLength, outputLength);
        } else {
          return calculateCost(pricing['gemini-pro'], inputLength, outputLength);
        }
      } else if (lowerModelName.includes('deepseek')) {
        return calculateCost(pricing['deepseek'], inputLength, outputLength);
      } else if (lowerModelName.includes('qwen')) {
        if (lowerModelName.includes('max')) {
          return calculateCost(pricing['qwen-max'], inputLength, outputLength);
        } else if (lowerModelName.includes('plus')) {
          return calculateCost(pricing['qwen-plus'], inputLength, outputLength);
        } else {
          return calculateCost(pricing['qwen-turbo'], inputLength, outputLength);
        }
      }
      
      // 如果还是找不到，返回0
      return 0;
    }
    
    const price = pricing[modelKey];
    return calculateCost(price, inputLength, outputLength);
  };
  
  // 格式化成本显示，使其更易读
  const formatCostForDisplay = (cost: number): string => {
    if (cost >= 1) {
      // 如果成本大于等于1美元，显示两位小数
      return cost.toFixed(2);
    } else if (cost >= 0.01) {
      // 如果成本大于等于0.01美元，显示四位小数
      return cost.toFixed(4);
    } else {
      // 如果成本小于0.01美元，显示六位小数
      return cost.toFixed(6);
    }
  };
  
  // 获取成本估算的显示文本（根据配置的显示单位）
  const getCostDisplayText = (modelId: string, inputLength: number, outputLength: number = 0): string => {
    const costUSD = estimateCost(modelId, inputLength, outputLength);
    if (costUSD <= 0) return '';
    
    // 从系统配置获取显示单位和汇率
    const displayUnit = systemConfig?.cost_display_unit || 'BOTH';
    const exchangeRate = systemConfig?.usd_to_cny_rate || 7.2;
    
    const costCNY = costUSD * exchangeRate;
    
    // 格式化成本显示，使其更易读
    const formattedUSD = formatCostForDisplay(costUSD);
    const formattedCNY = formatCostForDisplay(costCNY);
    
    switch(displayUnit) {
      case 'USD':
        return `${t('estimatedCost')}: $${formattedUSD}`;
      case 'CNY':
        return `${t('estimatedCost')}: ¥${formattedCNY}`;
      case 'BOTH':
      default:
        return `${t('estimatedCost')}: $${formattedUSD} (¥${formattedCNY})`;
    }
  };
  
  // 生成系统提示词用于成本估算
  const generateSystemPrompt = (): string => {
    try {
      if (!editingConfig) return '';
      
      // 构建一个简化版的系统提示词用于成本估算
      const { coin_source, indicators, risk_control, prompt_sections, custom_prompt } = editingConfig;
      
      let prompt = "你是一个专业的量化交易AI助手。\n";
      
      // 添加币种来源信息
      if (coin_source) {
        prompt += `币种来源：${JSON.stringify(coin_source)}\n`;
      }
      
      // 添加技术指标信息
      if (indicators) {
        prompt += `技术指标：${JSON.stringify(indicators)}\n`;
      }
      
      // 添加风控参数
      if (risk_control) {
        prompt += `风控参数：${JSON.stringify(risk_control)}\n`;
      }
      
      // 添加提示词部分
      if (prompt_sections) {
        prompt += `提示词部分：${JSON.stringify(prompt_sections)}\n`;
      }
      
      // 添加自定义提示
      if (custom_prompt) {
        prompt += `自定义提示：${custom_prompt}\n`;
      }
      
      prompt += "\n请根据以上配置生成交易决策。";
      
      return prompt;
    } catch (error) {
      console.error('Error generating system prompt:', error);
      return '';
    }
  };
  
  // 优化AI提示词内容，去除冗余信息
  const optimizePromptContent = (content: string): string => {
    try {
      if (!content || typeof content !== 'string') return content;
      
      // 移除多余的空白字符和换行
      let optimized = content.trim();
      
      // 替换多个连续的空格为单个空格
      optimized = optimized.replace(/\s+/g, ' ');
      
      // 替换多个连续的换行为单个换行
      optimized = optimized.replace(/\n\s*\n/g, '\n');
      
      // 移除重复的标点符号
      optimized = optimized.replace(/[.!?]{2,}/g, '.');
      
      // 移除不必要的注释标记（如"注意:"、"重要:"等）
      optimized = optimized.replace(/(注意|重要|提醒|提示)[:：]\s*/g, '');
      
      // 移除过于冗长的重复表述
      const sentences = optimized.split(/[.!?]/);
      const uniqueSentences: string[] = [];
      const seenSentences = new Set<string>();
      
      for (const sentence of sentences) {
        const cleanSentence = sentence.trim().toLowerCase();
        // 忽略太短的句子（可能是标点符号）
        if (cleanSentence.length < 5) continue;
        
        // 如果句子没有重复，则添加
        if (!seenSentences.has(cleanSentence)) {
          seenSentences.add(cleanSentence);
          uniqueSentences.push(sentence.trim());
        }
      }
      
      optimized = uniqueSentences.join('. ') + '.';
      
      return optimized;
    } catch (error) {
      console.error('Error optimizing prompt content:', error);
      // 如果优化过程中出现错误，返回原始内容
      return content;
    }
  };
  
  // 估算优化后的成本
  const estimateOptimizedCost = (modelId: string, originalContent: string): { originalCost: number; optimizedCost: number; savings: number; reductionPercentage: number } => {
    try {
      const originalLength = originalContent.length;
      const optimizedContent = optimizePromptContent(originalContent);
      const optimizedLength = optimizedContent.length;
      
      const originalCost = estimateCost(modelId, originalLength);
      const optimizedCost = estimateCost(modelId, optimizedLength);
      const savings = originalCost > optimizedCost ? originalCost - optimizedCost : 0;
      const reductionPercentage = originalLength > 0 ? ((originalLength - optimizedLength) / originalLength) * 100 : 0;
      
      return {
        originalCost,
        optimizedCost,
        savings,
        reductionPercentage
      };
    } catch (error) {
      console.error('Error estimating optimized cost:', error);
      // 如果计算过程中出现错误，返回零值
      return {
        originalCost: 0,
        optimizedCost: 0,
        savings: 0,
        reductionPercentage: 0
      };
    }
  };
  
  
  


  // 复制到剪贴板的函数
  const copyToClipboard = async (text: string, type: string) => {
    try {
      await navigator.clipboard.writeText(text)
      setCopiedStates(prev => ({ ...prev, [type]: true }))
      setTimeout(() => {
        setCopiedStates(prev => ({ ...prev, [type]: false }))
      }, 2000) // 2秒后重置图标
    } catch (err) {
      console.error('Failed to copy text: ', err)
    }
  }

  const toggleSection = (section: keyof typeof expandedSections) => {
    setExpandedSections((prev) => ({
      ...prev,
      [section]: !prev[section],
    }))
  }

  // Fetch system configuration
  const fetchSystemConfig = useCallback(async () => {
    try {
      console.log('Fetching system configuration...')
      const response = await fetch(`${API_BASE}/api/system-config`)
      if (response.ok) {
        const data = await response.json()
        console.log('Received system configuration:', data)
        setSystemConfig(data)
      } else {
        console.error('Failed to fetch system config:', response.status, response.statusText)
        // 设置默认配置以防获取失败
        setSystemConfig({
          registration_enabled: true,
          btc_eth_leverage: 10,
          altcoin_leverage: 5,
          model_max_tokens: {
            'deepseek': 32768,
            'gpt-4': 128000,
            'gpt-3.5': 16384,
            'claude': 200000,
            'qwen': 32768,
            'default': 4096,
          },
          model_pricing: {
            'qwen-max': { input_price: 2.40, output_price: 9.60 },
            'qwen-plus': { input_price: 1.00, output_price: 4.00 },
            'gpt-4o': { input_price: 2.50, output_price: 10.00 },
            'gpt-4omini': { input_price: 0.15, output_price: 0.60 },
            'claude-sonnet': { input_price: 3.00, output_price: 15.00 },
            'deepseek': { input_price: 0.20, output_price: 0.80 },
            'ernie': { input_price: 0.08, output_price: 0.20 },
            'default': { input_price: 0.50, output_price: 2.00 },
          }
        })
      }
    } catch (err) {
      console.error('Error fetching system configuration:', err)
      // 设置默认配置以防获取失败
      setSystemConfig({
        registration_enabled: true,
        btc_eth_leverage: 10,
        altcoin_leverage: 5,
        model_max_tokens: {
          'deepseek': 32768,
          'gpt-4': 128000,
          'gpt-3.5': 16384,
          'claude': 200000,
          'qwen': 32768,
          'default': 4096,
        },
        model_pricing: {
          'qwen-max': { input_price: 2.40, output_price: 9.60 },
          'qwen-plus': { input_price: 1.00, output_price: 4.00 },
          'gpt-4o': { input_price: 2.50, output_price: 10.00 },
          'gpt-4omini': { input_price: 0.15, output_price: 0.60 },
          'claude-sonnet': { input_price: 3.00, output_price: 15.00 },
          'deepseek': { input_price: 0.20, output_price: 0.80 },
          'ernie': { input_price: 0.08, output_price: 0.20 },
          'default': { input_price: 0.50, output_price: 2.00 },
        }
      })
    }
  }, [])

  // Fetch AI Models
  const fetchAiModels = useCallback(async () => {
    if (!token) return
    try {
      const response = await fetch(`${API_BASE}/api/models`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (response.ok) {
        const data = await response.json()
        // 后端返回的是数组，不是 { models: [] }
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
  }, [token, selectedModelId])

  // Fetch strategies
  const fetchStrategies = useCallback(async () => {
    if (!token) return
    try {
      const response = await fetch(`${API_BASE}/api/strategies`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (!response.ok) throw new Error('Failed to fetch strategies')
      const data = await response.json()
      setStrategies(data.strategies || [])

      // Select active or first strategy
      const active = data.strategies?.find((s: Strategy) => s.is_active)
      if (active) {
        setSelectedStrategy(active)
        setEditingConfig(active.config)
      } else if (data.strategies?.length > 0) {
        setSelectedStrategy(data.strategies[0])
        setEditingConfig(data.strategies[0].config)
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setIsLoading(false)
    }
  }, [token])

  useEffect(() => {
    fetchSystemConfig()
    fetchStrategies()
    fetchAiModels()
  }, [fetchSystemConfig, fetchStrategies, fetchAiModels])

  // Track previous language to detect actual changes
  const prevLanguageRef = useRef(language)

  // When language changes, update prompt sections to match the new language
  useEffect(() => {
    const updatePromptSectionsForLanguage = async () => {
      // Only update if language actually changed (not on initial mount)
      if (prevLanguageRef.current === language) return
      prevLanguageRef.current = language

      if (!token) return

      try {
        // Fetch default config for the new language
        const response = await fetch(
          `${API_BASE}/api/strategies/default-config?lang=${language}`,
          { headers: { Authorization: `Bearer ${token}` } }
        )
        if (!response.ok) return
        const defaultConfig = await response.json()

        // Update only the prompt sections and language field
        setEditingConfig(prev => {
          if (!prev) return prev
          return {
            ...prev,
            language: language as 'zh' | 'en',
            prompt_sections: defaultConfig.prompt_sections,
          }
        })
        setHasChanges(true)
      } catch (err) {
        console.error('Failed to update prompt sections for language:', err)
      }
    }

    updatePromptSectionsForLanguage()
  }, [language, token]) // Only trigger when language changes

  // Create new strategy
  const handleCreateStrategy = async () => {
    if (!token) return
    try {
      const configResponse = await fetch(
        `${API_BASE}/api/strategies/default-config?lang=${language}`,
        { headers: { Authorization: `Bearer ${token}` } }
      )
      const defaultConfig = await configResponse.json()

      const response = await fetch(`${API_BASE}/api/strategies`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          name: language === 'zh' ? '新策略' : 'New Strategy',
          description: '',
          config: defaultConfig,
        }),
      })
      if (!response.ok) throw new Error('Failed to create strategy')
      const result = await response.json()
      await fetchStrategies()
      // Auto-select the newly created strategy
      if (result.id) {
        const now = new Date().toISOString()
        const newStrategy = {
          id: result.id,
          name: language === 'zh' ? '新策略' : 'New Strategy',
          description: '',
          is_active: false,
          is_default: false,
          is_public: false,
          config_visible: true,
          config: defaultConfig,
          created_at: now,
          updated_at: now,
        }
        setSelectedStrategy(newStrategy)
        setEditingConfig(defaultConfig)
        setHasChanges(false)
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    }
  }

  // Delete strategy
  const handleDeleteStrategy = async (id: string) => {
    if (!token) return

    const confirmed = await confirmToast(
      language === 'zh' ? '确定删除此策略？' : 'Delete this strategy?',
      {
        title: language === 'zh' ? '确认删除' : 'Confirm Delete',
        okText: language === 'zh' ? '删除' : 'Delete',
        cancelText: language === 'zh' ? '取消' : 'Cancel',
      }
    )
    if (!confirmed) return

    try {
      const response = await fetch(`${API_BASE}/api/strategies/${id}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` },
      })
      if (!response.ok) throw new Error('Failed to delete strategy')
      notify.success(language === 'zh' ? '策略已删除' : 'Strategy deleted')
      // Clear selection if deleted strategy was selected
      if (selectedStrategy?.id === id) {
        setSelectedStrategy(null)
        setEditingConfig(null)
        setHasChanges(false)
      }
      await fetchStrategies()
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Unknown error'
      setError(errorMsg)
      
      // 在开发模式下显示更详细的错误信息
      if (process.env.NODE_ENV === 'development') {
        console.error('Strategy deletion error details:', {
          message: err instanceof Error ? err.message : 'Unknown error',
          stack: err instanceof Error ? err.stack : undefined,
          error: err,
        });
        
        notify.error(`${errorMsg}\nURL: /api/strategies/${id}\nMethod: DELETE`)
      } else {
        notify.error(errorMsg)
      }
    }
  }

  // Duplicate strategy
  const handleDuplicateStrategy = async (id: string) => {
    if (!token) return
    try {
      const response = await fetch(`${API_BASE}/api/strategies/${id}/duplicate`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          name: language === 'zh' ? '策略副本' : 'Strategy Copy',
        }),
      })
      if (!response.ok) throw new Error('Failed to duplicate strategy')
      await fetchStrategies()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    }
  }

  // Activate strategy
  const handleActivateStrategy = async (id: string) => {
    if (!token) return
    try {
      const response = await fetch(`${API_BASE}/api/strategies/${id}/activate`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
      })
      if (!response.ok) throw new Error('Failed to activate strategy')
      await fetchStrategies()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    }
  }

  // Export strategy as JSON file
  const handleExportStrategy = (strategy: Strategy) => {
    const exportData = {
      name: strategy.name,
      description: strategy.description,
      config: strategy.config,
      exported_at: new Date().toISOString(),
      version: '1.0',
    }
    const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `strategy_${strategy.name.replace(/\s+/g, '_')}_${new Date().toISOString().split('T')[0]}.json`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    notify.success(language === 'zh' ? '策略已导出' : 'Strategy exported')
  }

  // Import strategy from JSON file
  const handleImportStrategy = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (!file || !token) return

    try {
      const text = await file.text()
      const importData = JSON.parse(text)

      // Validate imported data
      if (!importData.config || !importData.name) {
        throw new Error(language === 'zh' ? '无效的策略文件' : 'Invalid strategy file')
      }

      // Create new strategy with imported config
      const response = await fetch(`${API_BASE}/api/strategies`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          name: `${importData.name} (${language === 'zh' ? '导入' : 'Imported'})`,
          description: importData.description || '',
          config: importData.config,
        }),
      })
      if (!response.ok) throw new Error('Failed to import strategy')

      notify.success(language === 'zh' ? '策略已导入' : 'Strategy imported')
      await fetchStrategies()
    } catch (err) {
      // 在开发模式下显示更详细的错误信息
      if (process.env.NODE_ENV === 'development') {
        console.error('Strategy import error details:', {
          message: err instanceof Error ? err.message : 'Unknown error',
          stack: err instanceof Error ? err.stack : undefined,
          error: err,
        });
        
        notify.error(`${err instanceof Error ? err.message : 'Unknown error'}\nURL: /api/strategies\nMethod: POST (import)`)
      } else {
        notify.error(err instanceof Error ? err.message : 'Unknown error')
      }
    } finally {
      // Reset file input
      event.target.value = ''
    }
  }

  // Download annotated strategy config template
  const handleDownloadAnnotatedTemplate = async () => {
    if (!token) return
    
    try {
      const response = await fetch(`${API_BASE}/api/strategies/annotated-config?lang=${language}`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      
      if (!response.ok) {
        throw new Error('Failed to fetch annotated template')
      }
      
      const annotatedConfig = await response.json()
      
      // Create a downloadable file with annotations
      const blob = new Blob([JSON.stringify(annotatedConfig, null, 2)], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `strategy_config_annotated_template_${new Date().toISOString().split('T')[0]}.json`
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
      
      notify.success(language === 'zh' ? '带注释的配置模板已下载' : 'Annotated config template downloaded')
    } catch (err) {
      // 在开发模式下显示更详细的错误信息
      if (process.env.NODE_ENV === 'development') {
        console.error('Annotated template download error details:', {
          message: err instanceof Error ? err.message : 'Unknown error',
          stack: err instanceof Error ? err.stack : undefined,
          error: err,
        });
        
        notify.error(`${err instanceof Error ? err.message : 'Unknown error'}\nURL: /api/strategies/annotated-config\nMethod: GET`)
      } else {
        notify.error(err instanceof Error ? err.message : 'Unknown error')
      }
      console.error('Failed to download annotated template:', err)
    }
  }

  // Save strategy
  const handleSaveStrategy = async () => {
    if (!token || !selectedStrategy || !editingConfig) return
    setIsSaving(true)
    try {
      // Always sync the config language with the current interface language
      const configWithLanguage = {
        ...editingConfig,
        language: language as 'zh' | 'en',
      }
      const response = await fetch(
        `${API_BASE}/api/strategies/${selectedStrategy.id}`,
        {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
          },
          body: JSON.stringify({
            name: selectedStrategy.name,
            description: selectedStrategy.description,
            config: configWithLanguage,
            is_public: selectedStrategy.is_public,
            config_visible: selectedStrategy.config_visible,
          }),
        }
      )
      if (!response.ok) throw new Error('Failed to save strategy')
      setHasChanges(false)
      notify.success(language === 'zh' ? '策略已保存' : 'Strategy saved')
      await fetchStrategies()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
      
      // 在开发模式下显示更详细的错误信息
      if (process.env.NODE_ENV === 'development') {
        console.error('Strategy save error details:', {
          message: err instanceof Error ? err.message : 'Unknown error',
          stack: err instanceof Error ? err.stack : undefined,
          error: err,
          url: `/api/strategies/${selectedStrategy?.id}`,
          method: 'PUT',
        });
        
        notify.error(`${err instanceof Error ? err.message : 'Unknown error'}\nURL: /api/strategies/${selectedStrategy?.id}\nMethod: PUT (save)`)
      }
    } finally {
      setIsSaving(false)
    }
  }

  // Update config section
  const updateConfig = <K extends keyof StrategyConfig>(
    section: K,
    value: StrategyConfig[K]
  ) => {
    if (!editingConfig) return
    setEditingConfig({
      ...editingConfig,
      [section]: value,
    })
    setHasChanges(true)
  }

  // Fetch prompt preview
  const fetchPromptPreview = async () => {
    if (!token || !editingConfig) return
    setIsLoadingPrompt(true)
    try {
      const response = await fetch(`${API_BASE}/api/strategies/preview-prompt`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          config: editingConfig,
          account_equity: 1000,
          prompt_variant: selectedVariant,
        }),
      })
      if (!response.ok) throw new Error('Failed to fetch prompt preview')
      const data = await response.json()
      setPromptPreview(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
      
      // 在开发模式下显示更详细的错误信息
      if (process.env.NODE_ENV === 'development') {
        console.error('Prompt preview fetch error details:', {
          message: err instanceof Error ? err.message : 'Unknown error',
          stack: err instanceof Error ? err.stack : undefined,
          error: err,
          url: `/api/strategies/preview-prompt`,
          method: 'POST',
        });
        
        notify.error(`${err instanceof Error ? err.message : 'Unknown error'}\nURL: /api/strategies/preview-prompt\nMethod: POST (preview)`)
      }
    } finally {
      setIsLoadingPrompt(false)
    }
  }

  // Run AI test with real AI model
  const runAiTest = async () => {
    if (!token || !editingConfig || !selectedModelId) return
    setIsRunningAiTest(true)
    setAiTestResult(null)
    try {
      const response = await fetch(`${API_BASE}/api/strategies/test-run`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          config: editingConfig,
          prompt_variant: selectedVariant,
          ai_model_id: selectedModelId,
          run_real_ai: true,
        }),
      })
      if (!response.ok) throw new Error('Failed to run AI test')
      const data = await response.json()
      setAiTestResult(data)
    } catch (err) {
      // 在开发模式下显示更详细的错误信息
      if (process.env.NODE_ENV === 'development') {
        console.error('AI test run error details:', {
          message: err instanceof Error ? err.message : 'Unknown error',
          stack: err instanceof Error ? err.stack : undefined,
          error: err,
          url: `/api/strategies/test-run`,
          method: 'POST',
        });
        
        notify.error(`${err instanceof Error ? err.message : 'Unknown error'}\nURL: /api/strategies/test-run\nMethod: POST (AI test)`)
      }
      
      setAiTestResult({
        error: err instanceof Error ? err.message : 'Unknown error',
      })
    } finally {
      setIsRunningAiTest(false)
    }
  }

  const t = (key: string) => {
    const translations: Record<string, Record<string, string>> = {
      strategyStudio: { zh: '策略工作室', en: 'Strategy Studio' },
      subtitle: { zh: '可视化配置和测试交易策略', en: 'Configure and test trading strategies' },
      strategies: { zh: '策略', en: 'Strategies' },
      newStrategy: { zh: '新建', en: 'New' },
      coinSource: { zh: '币种来源', en: 'Coin Source' },
      indicators: { zh: '技术指标', en: 'Indicators' },
      riskControl: { zh: '风控参数', en: 'Risk Control' },
      promptSections: { zh: 'Prompt 编辑', en: 'Prompt Editor' },
      customPrompt: { zh: '附加提示', en: 'Extra Prompt' },
      save: { zh: '保存', en: 'Save' },
      saving: { zh: '保存中...', en: 'Saving...' },
      activate: { zh: '激活', en: 'Activate' },
      active: { zh: '激活中', en: 'Active' },
      default: { zh: '默认', en: 'Default' },
      promptPreview: { zh: 'Prompt 预览', en: 'Prompt Preview' },
      aiTestRun: { zh: 'AI 测试', en: 'AI Test' },
      systemPrompt: { zh: 'System Prompt', en: 'System Prompt' },
      userPrompt: { zh: 'User Prompt', en: 'User Prompt' },
      loadPrompt: { zh: '生成 Prompt', en: 'Generate Prompt' },
      refreshPrompt: { zh: '刷新', en: 'Refresh' },
      promptVariant: { zh: '风格', en: 'Style' },
      balanced: { zh: '平衡', en: 'Balanced' },
      aggressive: { zh: '激进', en: 'Aggressive' },
      conservative: { zh: '保守', en: 'Conservative' },
      selectModel: { zh: '选择 AI 模型', en: 'Select AI Model' },
      runTest: { zh: '运行 AI 测试', en: 'Run AI Test' },
      running: { zh: '运行中...', en: 'Running...' },
      aiOutput: { zh: 'AI 输出', en: 'AI Output' },
      reasoning: { zh: '思维链', en: 'Reasoning' },
      decisions: { zh: '决策', en: 'Decisions' },
      duration: { zh: '耗时', en: 'Duration' },
      noModel: { zh: '请先配置 AI 模型', en: 'Please configure AI model first' },
      testNote: { zh: '使用真实 AI 模型测试，不执行交易', en: 'Test with real AI, no trading' },
      publishSettings: { zh: '发布设置', en: 'Publish' },
    }
    return translations[key]?.[language] || key
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[70vh]">
        <div className="text-center">
          <div className="relative">
            <div className="w-16 h-16 rounded-full border-4 border-yellow-500/20 border-t-yellow-500 animate-spin" />
            <Zap className="w-6 h-6 text-yellow-500 absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2" />
          </div>
        </div>
      </div>
    )
  }

  const configSections = [
    {
      key: 'coinSource' as const,
      icon: Target,
      color: '#F0B90B',
      title: t('coinSource'),
      content: editingConfig && (
        <CoinSourceEditor
          config={editingConfig.coin_source}
          onChange={(coinSource) => updateConfig('coin_source', coinSource)}
          disabled={selectedStrategy?.is_default}
          language={language}
        />
      ),
    },
    {
      key: 'indicators' as const,
      icon: BarChart3,
      color: '#0ECB81',
      title: t('indicators'),
      content: editingConfig && (
        <IndicatorEditor
          config={editingConfig.indicators}
          onChange={(indicators) => updateConfig('indicators', indicators)}
          disabled={selectedStrategy?.is_default}
          language={language}
        />
      ),
    },
    {
      key: 'riskControl' as const,
      icon: Shield,
      color: '#F6465D',
      title: t('riskControl'),
      content: editingConfig && (
        <RiskControlEditor
          config={editingConfig.risk_control}
          onChange={(riskControl) => updateConfig('risk_control', riskControl)}
          disabled={selectedStrategy?.is_default}
          language={language}
        />
      ),
    },
    {
      key: 'promptSections' as const,
      icon: FileText,
      color: '#a855f7',
      title: t('promptSections'),
      content: editingConfig && (
        <PromptSectionsEditor
          config={editingConfig.prompt_sections}
          onChange={(promptSections) => updateConfig('prompt_sections', promptSections)}
          disabled={selectedStrategy?.is_default}
          language={language}
        />
      ),
    },
    {
      key: 'customPrompt' as const,
      icon: Settings,
      color: '#60a5fa',
      title: t('customPrompt'),
      content: editingConfig && (
        <div>
          <p className="text-xs mb-2" style={{ color: '#848E9C' }}>
            {language === 'zh' ? '附加在 System Prompt 末尾的额外提示，用于补充个性化交易风格' : 'Extra prompt appended to System Prompt for personalized trading style'}
          </p>
          <textarea
            value={editingConfig.custom_prompt || ''}
            onChange={(e) => updateConfig('custom_prompt', e.target.value)}
            disabled={selectedStrategy?.is_default}
            placeholder={language === 'zh' ? '输入自定义提示词...' : 'Enter custom prompt...'}
            className="w-full h-32 px-3 py-2 rounded-lg resize-none font-mono text-xs"
            style={{ background: '#0B0E11', border: '1px solid #2B3139', color: '#EAECEF' }}
          />
        </div>
      ),
    },
    {
      key: 'publishSettings' as const,
      icon: Globe,
      color: '#0ECB81',
      title: t('publishSettings'),
      content: selectedStrategy && (
        <PublishSettingsEditor
          isPublic={selectedStrategy.is_public ?? false}
          configVisible={selectedStrategy.config_visible ?? true}
          onIsPublicChange={(value) => {
            setSelectedStrategy({ ...selectedStrategy, is_public: value })
            setHasChanges(true)
          }}
          onConfigVisibleChange={(value) => {
            setSelectedStrategy({ ...selectedStrategy, config_visible: value })
            setHasChanges(true)
          }}
          disabled={selectedStrategy?.is_default}
          language={language}
        />
      ),
    },
  ]

  return (
    <DeepVoidBackground className="h-[calc(100vh-64px)] flex flex-col bg-nofx-bg relative overflow-hidden">

      {/* Header */}
      {/* Header */}
      <div className="flex-shrink-0 px-4 py-3 border-b border-nofx-gold/20 bg-nofx-bg/60 backdrop-blur-md z-10">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-gradient-to-br from-nofx-gold to-yellow-500">
              <Sparkles className="w-5 h-5 text-black" />
            </div>
            <div>
              <h1 className="text-lg font-bold text-nofx-text">{t('strategyStudio')}</h1>
              <p className="text-xs text-nofx-text-muted">{t('subtitle')}</p>
            </div>
          </div>
          {error && (
            <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs bg-nofx-danger/10 text-nofx-danger">
              {error}
              <button onClick={() => setError(null)} className="hover:underline">×</button>
            </div>
          )}
        </div>
      </div>

      {/* Main Content - Three Columns */}
      <div className="flex-1 flex overflow-hidden">
        {/* Left Column - Strategy List */}
        <div className="w-48 flex-shrink-0 border-r border-nofx-gold/20 overflow-y-auto bg-nofx-bg/30 backdrop-blur-sm z-10">
          <div className="p-2">
            <div className="flex items-center justify-between mb-2 px-2">
              <span className="text-xs font-medium text-nofx-text-muted">{t('strategies')}</span>
              <div className="flex items-center gap-1">
                {/* Download annotated template button */}
                <button
                  onClick={handleDownloadAnnotatedTemplate}
                  className="p-1 rounded hover:bg-blue-500/20 transition-colors text-blue-400"
                  title={language === 'zh' ? '下载带注释的配置模板' : 'Download annotated config template'}
                >
                  <Download className="w-4 h-4" />
                </button>
                {/* Import button with hidden file input */}
                <label className="p-1 rounded hover:bg-white/10 transition-colors cursor-pointer text-nofx-text-muted hover:text-white" title={language === 'zh' ? '导入策略' : 'Import Strategy'}>
                  <Upload className="w-4 h-4" />
                  <input
                    type="file"
                    accept=".json"
                    onChange={handleImportStrategy}
                    className="hidden"
                  />
                </label>
                <button
                  onClick={handleCreateStrategy}
                  className="p-1 rounded hover:bg-white/10 transition-colors text-nofx-gold"
                  title={language === 'zh' ? '新建策略' : 'New Strategy'}
                >
                  <Plus className="w-4 h-4" />
                </button>
              </div>
            </div>
            <div className="space-y-1">
              {strategies.map((strategy) => (
                <div
                  key={strategy.id}
                  onClick={() => {
                    setSelectedStrategy(strategy)
                    setEditingConfig(strategy.config)
                    setHasChanges(false)
                    setPromptPreview(null)
                    setAiTestResult(null)
                  }}
                  className={`group px-2 py-2 rounded-lg cursor-pointer transition-all ${selectedStrategy?.id === strategy.id
                    ? 'ring-1 ring-nofx-gold/50 bg-nofx-gold/10 shadow-[0_0_15px_rgba(240,185,11,0.1)]'
                    : 'hover:bg-nofx-bg-lighter/60 hover:ring-1 hover:ring-nofx-gold/20 bg-transparent'
                    }`}
                >
                  <div className="flex items-center justify-between">
                    <span className="text-sm truncate text-nofx-text">{strategy.name}</span>
                    <div className="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
                      <button
                        onClick={(e) => { e.stopPropagation(); handleExportStrategy(strategy) }}
                        className="p-1 rounded hover:bg-white/10 text-nofx-text-muted hover:text-white"
                        title={language === 'zh' ? '导出' : 'Export'}
                      >
                        <Download className="w-3 h-3" />
                      </button>
                      {!strategy.is_default && (
                        <>
                          <button
                            onClick={(e) => { e.stopPropagation(); handleDuplicateStrategy(strategy.id) }}
                            className="p-1 rounded hover:bg-white/10 text-nofx-text-muted hover:text-white"
                            title={language === 'zh' ? '复制' : 'Duplicate'}
                          >
                            <Copy className="w-3 h-3" />
                          </button>
                          <button
                            onClick={(e) => { e.stopPropagation(); handleDeleteStrategy(strategy.id) }}
                            className="p-1 rounded hover:bg-nofx-danger/20 text-nofx-danger"
                            title={language === 'zh' ? '删除' : 'Delete'}
                          >
                            <Trash2 className="w-3 h-3" />
                          </button>
                        </>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center gap-1 mt-1 flex-wrap">
                    {strategy.is_active && (
                      <span className="px-1.5 py-0.5 text-[10px] rounded bg-nofx-success/15 text-nofx-success">
                        {t('active')}
                      </span>
                    )}
                    {strategy.is_default && (
                      <span className="px-1.5 py-0.5 text-[10px] rounded bg-nofx-gold/15 text-nofx-gold">
                        {t('default')}
                      </span>
                    )}
                    {strategy.is_public && (
                      <span className="px-1.5 py-0.5 text-[10px] rounded flex items-center gap-0.5 bg-blue-400/15 text-blue-400">
                        <Globe className="w-2.5 h-2.5" />
                        {language === 'zh' ? '公开' : 'Public'}
                      </span>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Middle Column - Config Editor */}
        <div className="flex-1 min-w-0 overflow-y-auto border-r border-nofx-gold/20">
          {selectedStrategy && editingConfig ? (
            <div className="p-4">
              {/* Strategy Name & Actions */}
              <div className="flex items-center justify-between mb-4">
                <div className="flex-1 min-w-0">
                  <input
                    type="text"
                    value={selectedStrategy.name}
                    onChange={(e) => {
                      setSelectedStrategy({ ...selectedStrategy, name: e.target.value })
                      setHasChanges(true)
                    }}
                    disabled={selectedStrategy.is_default}
                    className="text-lg font-bold bg-transparent border-none outline-none w-full text-nofx-text placeholder-nofx-text-muted"
                  />
                  <input
                    type="text"
                    value={selectedStrategy.description || ''}
                    onChange={(e) => {
                      setSelectedStrategy({ ...selectedStrategy, description: e.target.value })
                      setHasChanges(true)
                    }}
                    disabled={selectedStrategy.is_default}
                    placeholder={language === 'zh' ? '添加策略简介...' : 'Add strategy description...'}
                    className="text-xs bg-transparent border-none outline-none w-full text-nofx-text-muted placeholder-nofx-text-muted/50 mt-1"
                  />
                  {hasChanges && (
                    <span className="text-xs text-nofx-gold">● {language === 'zh' ? '未保存' : 'Unsaved'}</span>
                  )}
                </div>
                <div className="flex items-center gap-2 flex-shrink-0">
                  {!selectedStrategy.is_active && (
                    <button
                      onClick={() => handleActivateStrategy(selectedStrategy.id)}
                      className="flex items-center gap-1 px-3 py-1.5 rounded-lg text-xs transition-colors bg-nofx-success/10 border border-nofx-success/30 text-nofx-success hover:bg-nofx-success/20"
                    >
                      <Check className="w-3 h-3" />
                      {t('activate')}
                    </button>
                  )}
                  {!selectedStrategy.is_default && (
                    <button
                      onClick={handleSaveStrategy}
                      disabled={isSaving || !hasChanges}
                      className={`flex items-center gap-1 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors disabled:opacity-50
                        ${hasChanges ? 'bg-nofx-gold text-black hover:bg-yellow-500' : 'bg-nofx-bg-lighter text-nofx-text-muted cursor-not-allowed'}`}
                    >
                      <Save className="w-3 h-3" />
                      {isSaving ? t('saving') : t('save')}
                    </button>
                  )}
                </div>
              </div>

              {/* Config Sections */}
              <div className="space-y-2">
                {configSections.map(({ key, icon: Icon, color, title, content }) => (
                  <div
                    key={key}
                    className="rounded-lg overflow-hidden bg-nofx-bg-lighter border border-nofx-gold/20"
                  >
                    <button
                      onClick={() => toggleSection(key)}
                      className="w-full flex items-center justify-between px-3 py-2.5 hover:bg-white/5 transition-colors"
                    >
                      <div className="flex items-center gap-2">
                        <Icon className="w-4 h-4" style={{ color }} />
                        <span className="text-sm font-medium text-nofx-text">{title}</span>
                      </div>
                      {expandedSections[key] ? (
                        <ChevronDown className="w-4 h-4 text-nofx-text-muted" />
                      ) : (
                        <ChevronRight className="w-4 h-4 text-nofx-text-muted" />
                      )}
                    </button>
                    {expandedSections[key] && (
                      <div className="px-3 pb-3">
                        {content}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </div>
          ) : (
            <div className="flex items-center justify-center h-full">
              <div className="text-center">
                <Activity className="w-12 h-12 mx-auto mb-2 opacity-30 text-nofx-text-muted" />
                <p className="text-sm text-nofx-text-muted">
                  {language === 'zh' ? '选择或创建策略' : 'Select or create a strategy'}
                </p>
              </div>
            </div>
          )}
        </div>

        {/* Right Column - Prompt Preview & AI Test */}
        <div className="w-[420px] flex-shrink-0 flex flex-col overflow-hidden">
          {/* Tabs */}
          <div className="flex-shrink-0 flex border-b border-nofx-gold/20">
            <button
              onClick={() => setActiveRightTab('prompt')}
              className={`flex-1 flex items-center justify-center gap-2 px-4 py-2.5 text-sm font-medium transition-colors ${activeRightTab === 'prompt' ? 'border-b-2 border-purple-500 text-purple-500' : 'opacity-60 hover:opacity-100 text-nofx-text-muted'
                }`}
            >
              <Eye className="w-4 h-4" />
              {t('promptPreview')}
            </button>
            <button
              onClick={() => setActiveRightTab('test')}
              className={`flex-1 flex items-center justify-center gap-2 px-4 py-2.5 text-sm font-medium transition-colors ${activeRightTab === 'test' ? 'border-b-2 border-green-500 text-green-500' : 'opacity-60 hover:opacity-100 text-nofx-text-muted'
                }`}
            >
              <Play className="w-4 h-4" />
              {t('aiTestRun')}
            </button>
          </div>

          {/* Tab Content */}
          <div className="flex-1 overflow-y-auto">
            {activeRightTab === 'prompt' ? (
              /* Prompt Preview Tab */
              <div className="p-3 space-y-3">
                {/* Controls */}
                <div className="flex items-center gap-2 flex-wrap">
                  <select
                    value={selectedVariant}
                    onChange={(e) => setSelectedVariant(e.target.value)}
                    className="px-2 py-1.5 rounded text-xs bg-nofx-bg border border-nofx-gold/20 text-nofx-text outline-none focus:border-nofx-gold"
                  >
                    <option value="balanced">{t('balanced')}</option>
                    <option value="aggressive">{t('aggressive')}</option>
                    <option value="conservative">{t('conservative')}</option>
                  </select>
                  <button
                    onClick={fetchPromptPreview}
                    disabled={isLoadingPrompt || !editingConfig}
                    className="flex items-center gap-1.5 px-3 py-1.5 rounded text-xs font-medium transition-colors disabled:opacity-50 bg-purple-600 hover:bg-purple-700 text-white"
                  >
                    {isLoadingPrompt ? <Loader2 className="w-3 h-3 animate-spin" /> : <RefreshCw className="w-3 h-3" />}
                    {promptPreview ? t('refreshPrompt') : t('loadPrompt')}
                  </button>
                </div>

                {promptPreview ? (
                  <>
                    {/* Config Summary */}
                    <div className="p-2 rounded-lg bg-nofx-bg border border-nofx-gold/20">
                      <div className="flex items-center gap-1.5 mb-2">
                        <Code className="w-3 h-3 text-purple-500" />
                        <span className="text-xs font-medium text-purple-500">Config</span>
                      </div>
                      <div className="grid grid-cols-3 gap-2 text-xs">
                        {Object.entries(promptPreview.config_summary || {}).map(([key, value]) => (
                          <div key={key}>
                            <div className="text-nofx-text-muted">{key.replace(/_/g, ' ')}</div>
                            <div className="text-nofx-text">{String(value)}</div>
                          </div>
                        ))}
                      </div>
                    </div>

                    {/* System Prompt */}
                    <div>
                      <div className="flex items-center justify-between mb-1.5">
                        <div className="flex items-center gap-1.5">
                          <FileText className="w-3 h-3 text-purple-500" />
                          <span className="text-xs font-medium text-nofx-text">{t('systemPrompt')}</span>
                        </div>
                        <div className="flex items-center gap-1.5">
                          <span className="text-[10px] px-1.5 py-0.5 rounded bg-nofx-bg-lighter text-nofx-text-muted">
                            {promptPreview.system_prompt.length.toLocaleString()} chars
                          </span>
                          <button
                            onClick={() => copyToClipboard(promptPreview.system_prompt, 'systemPrompt')}
                            className="p-1 rounded text-xs hover:bg-nofx-bg-lighter transition-colors"
                            title="Copy to clipboard"
                          >
                            {copiedStates['systemPrompt'] ? (
                              <ClipboardCheck className="w-3 h-3 text-green-500" />
                            ) : (
                              <Clipboard className="w-3 h-3 text-nofx-text-muted hover:text-nofx-text" />
                            )}
                          </button>
                        </div>
                      </div>
                      <pre
                        className="p-2 rounded-lg text-[11px] font-mono overflow-auto bg-nofx-bg border border-nofx-gold/20 text-nofx-text"
                        style={{ maxHeight: '400px' }}
                      >
                        {promptPreview.system_prompt}
                      </pre>
                    </div>
                  </>
                ) : (
                  <div className="flex flex-col items-center justify-center py-12 text-nofx-text-muted">
                    <Eye className="w-10 h-10 mb-2 opacity-30" />
                    <p className="text-sm">{language === 'zh' ? '点击生成 Prompt 预览' : 'Click to generate prompt preview'}</p>
                  </div>
                )}
              </div>
            ) : (
              /* AI Test Tab */
              <div className="p-3 space-y-3">
                {/* Controls */}
                <div className="space-y-2">
                  <div className="flex items-center gap-2">
                    <Bot className="w-4 h-4 text-green-500" />
                    <span className="text-xs font-medium text-nofx-text">{t('selectModel')}</span>
                  </div>
                  {aiModels.length > 0 ? (
                    <select
                      value={selectedModelId}
                      onChange={(e) => setSelectedModelId(e.target.value)}
                      className="w-full px-3 py-2 rounded-lg text-sm bg-nofx-bg border border-nofx-gold/20 text-nofx-text"
                    >
                      {aiModels.map((model) => (
                        <option key={model.id} value={model.id}>
                          {model.name} ({model.provider})
                        </option>
                      ))}
                    </select>
                  ) : (
                    <div className="px-3 py-2 rounded-lg text-sm bg-nofx-danger/10 text-nofx-danger">
                      {t('noModel')}
                    </div>
                  )}

                  <div className="flex items-center gap-2">
                    <select
                      value={selectedVariant}
                      onChange={(e) => setSelectedVariant(e.target.value)}
                      className="px-2 py-1.5 rounded text-xs bg-nofx-bg border border-nofx-gold/20 text-nofx-text"
                    >
                      <option value="balanced">{t('balanced')}</option>
                      <option value="aggressive">{t('aggressive')}</option>
                      <option value="conservative">{t('conservative')}</option>
                    </select>
                    <div className="flex-1 space-y-2">
                      <button
                        onClick={runAiTest}
                        disabled={isRunningAiTest || !editingConfig || !selectedModelId}
                        className="w-full flex items-center justify-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all disabled:opacity-50 text-white shadow-lg shadow-green-500/20 bg-gradient-to-br from-green-500 to-green-600"
                      >
                        {isRunningAiTest ? (
                          <>
                            <Loader2 className="w-4 h-4 animate-spin" />
                            {t('running')}
                          </>
                        ) : (
                          <>
                            <Send className="w-4 h-4" />
                            {t('runTest')}
                          </>
                        )}
                      </button>
                                    
                      {/* 成本优化提示 */}
                      {selectedModelId && editingConfig && (
                        <div className="text-xs text-gray-500 dark:text-gray-400">
                          {(() => {
                            try {
                              const samplePrompt = generateSystemPrompt(); // 使用当前策略生成一个样本提示词
                              if (samplePrompt) {
                                const costEstimate = estimateOptimizedCost(selectedModelId, samplePrompt);
                                if (costEstimate.savings > 0) {
                                  const savingsFormatted = formatCostForDisplay(costEstimate.savings);
                                  return `${t('potentialSavings')}: ${getCostDisplayText(selectedModelId, 0, Math.round(costEstimate.savings * 1000000))} (${costEstimate.reductionPercentage.toFixed(1)}% reduction)`;
                                }
                              }
                              return `${t('optimizePromptForCost')}`;
                            } catch (error) {
                              console.error('Error in cost estimation display:', error);
                              return `${t('optimizePromptForCost')}`;
                            }
                          })()}
                        </div>
                      )}
                    </div>
                  </div>
                  <p className="text-[10px] text-nofx-text-muted">{t('testNote')}</p>
                </div>

                {/* Test Results */}
                {aiTestResult ? (
                  <div className="space-y-3">
                    {aiTestResult.error ? (
                      <div className="p-3 rounded-lg bg-nofx-danger/10 border border-nofx-danger/30">
                        <p className="text-sm text-nofx-danger">{aiTestResult.error}</p>
                      </div>
                    ) : (
                      <>
                        {aiTestResult.duration_ms && (
                          <div className="flex items-center gap-3 p-2 rounded-lg bg-nofx-bg/50 border border-nofx-gold/30">
                            <div className="flex items-center gap-1">
                              <Clock className="w-3 h-3 text-nofx-gold" />
                              <span className="text-xs text-nofx-text">
                                {t('duration')}: {(aiTestResult.duration_ms / 1000).toFixed(2)}s
                              </span>
                            </div>
                            <div className="h-4 w-px bg-nofx-border"></div>
                            <div className="flex items-center gap-1">
                              <Zap className="w-3 h-3 text-yellow-500" />
                              <span className="text-xs text-nofx-text">
                                {(aiTestResult.user_prompt?.length || 0) + (aiTestResult.reasoning?.length || 0) + (JSON.stringify(aiTestResult.decisions || []).length) + (aiTestResult.ai_response?.length || 0)} {t('chars')}
                              </span>
                            </div>
                          </div>
                        )}

                        {/* User Prompt Input */}
                        {aiTestResult.user_prompt && (
                          <div>
                            <div className="flex items-center justify-between gap-1.5 mb-1.5">
                              <div className="flex flex-col">
                                <div className="flex items-center gap-1.5">
                                  <Terminal className="w-3 h-3 text-blue-400" />
                                  <span className="text-xs font-medium text-nofx-text">{t('userPrompt')} ({t('input')})</span>
                                </div>
                                <div className="flex items-center gap-1.5 ml-4">
                                  <span className="text-[10px] text-nofx-text-muted">
                                    ({aiTestResult.user_prompt.length} chars)
                                  </span>
                                  {selectedModelId && systemConfig && (
                                    <span className="text-[10px] text-nofx-text-muted ml-1">
                                      / {getMaxTokensForModel(selectedModelId)} {t('charLimit')}
                                    </span>
                                  )}
                                  {selectedModelId && systemConfig && (() => {
                                    const costEstimate = estimateCost(selectedModelId, aiTestResult.user_prompt.length)
                                    return costEstimate > 0 ? (
                                      <span className="text-[8px] text-nofx-text-muted ml-1">
                                        ({t('estimatedCost')}: ${costEstimate.toFixed(6)})
                                      </span>
                                    ) : null
                                  })()}
                                </div>
                              </div>
                              <button
                                onClick={() => copyToClipboard(aiTestResult.user_prompt || '', 'userPrompt')}
                                className="p-1 rounded text-xs hover:bg-nofx-bg-lighter transition-colors"
                                title="Copy to clipboard"
                              >
                                {copiedStates['userPrompt'] ? (
                                  <ClipboardCheck className="w-3 h-3 text-green-500" />
                                ) : (
                                  <Clipboard className="w-3 h-3 text-nofx-text-muted hover:text-nofx-text" />
                                )}
                              </button>
                            </div>
                            <div
                              className="p-2 rounded-lg text-[10px] font-mono overflow-auto bg-nofx-bg border border-nofx-gold/20 text-nofx-text select-text"
                              style={{ maxHeight: '200px' }}
                              onClick={(e) => {
                                if (e.detail === 3) { // triple click for full select
                                  const selection = window.getSelection()
                                  if (selection) {
                                    const range = document.createRange()
                                    range.selectNodeContents(e.currentTarget)
                                    selection.removeAllRanges()
                                    selection.addRange(range)
                                  }
                                }
                              }}
                            >
                              {aiTestResult.user_prompt}
                            </div>
                          </div>
                        )}

                        {/* AI Reasoning */}
                        {aiTestResult.reasoning && (
                          <div>
                            <div className="flex items-center justify-between gap-1.5 mb-1.5">
                              <div className="flex flex-col">
                                <div className="flex items-center gap-1.5">
                                  <Sparkles className="w-3 h-3 text-nofx-gold" />
                                  <span className="text-xs font-medium text-nofx-text">{t('reasoning')}</span>
                                </div>
                                <div className="flex items-center gap-1.5 ml-4">
                                  <span className="text-[10px] text-nofx-text-muted">
                                    ({aiTestResult.reasoning.length} chars)
                                  </span>
                                  {selectedModelId && aiTestResult.user_prompt && systemConfig && (
                                    <span className="text-[8px] text-nofx-text-muted ml-1">
                                      {getCostDisplayText(selectedModelId, aiTestResult.user_prompt.length, aiTestResult.reasoning.length)}
                                    </span>
                                  )}
                                </div>
                              </div>
                              <button
                                onClick={() => copyToClipboard(aiTestResult.reasoning || '', 'reasoning')}
                                className="p-1 rounded text-xs hover:bg-nofx-bg-lighter transition-colors"
                                title="Copy to clipboard"
                              >
                                {copiedStates['reasoning'] ? (
                                  <ClipboardCheck className="w-3 h-3 text-green-500" />
                                ) : (
                                  <Clipboard className="w-3 h-3 text-nofx-text-muted hover:text-nofx-text" />
                                )}
                              </button>
                            </div>
                            <div
                              className="p-2 rounded-lg text-[10px] font-mono overflow-auto whitespace-pre-wrap bg-nofx-bg border border-nofx-gold/30 text-nofx-text select-text"
                              style={{ maxHeight: '200px' }}
                              onClick={(e) => {
                                if (e.detail === 3) { // triple click for full select
                                  const selection = window.getSelection()
                                  if (selection) {
                                    const range = document.createRange()
                                    range.selectNodeContents(e.currentTarget)
                                    selection.removeAllRanges()
                                    selection.addRange(range)
                                  }
                                }
                              }}
                            >
                              {aiTestResult.reasoning}
                            </div>
                          </div>
                        )}

                        {/* AI Decisions */}
                        {aiTestResult.decisions && aiTestResult.decisions.length > 0 && (
                          <div>
                            <div className="flex items-center justify-between gap-1.5 mb-1.5">
                              <div className="flex flex-col">
                                <div className="flex items-center gap-1.5">
                                  <Activity className="w-3 h-3 text-green-500" />
                                  <span className="text-xs font-medium text-nofx-text">{t('decisions')}</span>
                                </div>
                                <div className="flex items-center gap-1.5 ml-4">
                                  <span className="text-[10px] text-nofx-text-muted">
                                    ({JSON.stringify(aiTestResult.decisions).length} chars)
                                  </span>
                                  {selectedModelId && aiTestResult.user_prompt && systemConfig && (
                                    <span className="text-[8px] text-nofx-text-muted ml-1">
                                      {getCostDisplayText(selectedModelId, aiTestResult.user_prompt.length, JSON.stringify(aiTestResult.decisions).length)}
                                    </span>
                                  )}
                                </div>
                              </div>
                              <button
                                onClick={() => copyToClipboard(JSON.stringify(aiTestResult.decisions, null, 2), 'decisions')}
                                className="p-1 rounded text-xs hover:bg-nofx-bg-lighter transition-colors"
                                title="Copy to clipboard"
                              >
                                {copiedStates['decisions'] ? (
                                  <ClipboardCheck className="w-3 h-3 text-green-500" />
                                ) : (
                                  <Clipboard className="w-3 h-3 text-nofx-text-muted hover:text-nofx-text" />
                                )}
                              </button>
                            </div>
                            <div
                              className="p-2 rounded-lg text-[10px] font-mono overflow-auto bg-nofx-bg border border-green-500/30 text-nofx-text select-text"
                              style={{ maxHeight: '200px' }}
                              onClick={(e) => {
                                if (e.detail === 3) { // triple click for full select
                                  const selection = window.getSelection()
                                  if (selection) {
                                    const range = document.createRange()
                                    range.selectNodeContents(e.currentTarget)
                                    selection.removeAllRanges()
                                    selection.addRange(range)
                                  }
                                }
                              }}
                            >
                              {JSON.stringify(aiTestResult.decisions, null, 2)}
                            </div>
                          </div>
                        )}

                        {/* Raw AI Response */}
                        {aiTestResult.ai_response && (
                          <div>
                            <div className="flex items-center justify-between gap-1.5 mb-1.5">
                              <div className="flex flex-col">
                                <div className="flex items-center gap-1.5">
                                  <FileText className="w-3 h-3 text-nofx-text-muted" />
                                  <span className="text-xs font-medium text-nofx-text">{t('aiOutput')} ({t('raw')})</span>
                                </div>
                                <div className="flex items-center gap-1.5 ml-4">
                                  <span className="text-[10px] text-nofx-text-muted">
                                    ({aiTestResult.ai_response.length} chars)
                                  </span>
                                  {selectedModelId && aiTestResult.user_prompt && systemConfig && (
                                    <span className="text-[8px] text-nofx-text-muted ml-1">
                                      {getCostDisplayText(selectedModelId, aiTestResult.user_prompt.length, aiTestResult.ai_response.length)}
                                    </span>
                                  )}
                                </div>
                              </div>
                              <button
                                onClick={() => copyToClipboard(aiTestResult.ai_response || '', 'aiResponse')}
                                className="p-1 rounded text-xs hover:bg-nofx-bg-lighter transition-colors"
                                title="Copy to clipboard"
                              >
                                {copiedStates['aiResponse'] ? (
                                  <ClipboardCheck className="w-3 h-3 text-green-500" />
                                ) : (
                                  <Clipboard className="w-3 h-3 text-nofx-text-muted hover:text-nofx-text" />
                                )}
                              </button>
                            </div>
                            <div
                              className="p-2 rounded-lg text-[10px] font-mono overflow-auto whitespace-pre-wrap bg-nofx-bg border border-nofx-gold/20 text-nofx-text select-text"
                              style={{ maxHeight: '300px' }}
                              onClick={(e) => {
                                if (e.detail === 3) { // triple click for full select
                                  const selection = window.getSelection()
                                  if (selection) {
                                    const range = document.createRange()
                                    range.selectNodeContents(e.currentTarget)
                                    selection.removeAllRanges()
                                    selection.addRange(range)
                                  }
                                }
                              }}
                            >
                              {aiTestResult.ai_response}
                            </div>
                          </div>
                        )}
                      </>
                    )}
                  </div>
                ) : (
                  <div className="flex flex-col items-center justify-center py-12 text-nofx-text-muted">
                    <Play className="w-10 h-10 mb-2 opacity-30" />
                    <p className="text-sm">{language === 'zh' ? '点击运行 AI 测试' : 'Click to run AI test'}</p>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    </DeepVoidBackground>
  )
}

export default StrategyStudioPage
