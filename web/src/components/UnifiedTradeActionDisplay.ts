// 统一的交易动作展示规范
// 适用于所有交易动作类型：open_long, open_short, close_long, close_short, 
// hold, wait, partial_close, trailing_stop, update_stop_loss, update_take_profit等

export interface TradeActionDisplayConfig {
  // 基础配置
  action: string;
  symbol: string;
  price: number;
  confidence: number;
  reasoning: string;
  
  // 可选参数
  leverage?: number;
  quantity?: number;
  stopLoss?: number;
  takeProfit?: number;
  closePercentage?: number;
  trailPercentage?: number;
  callbackRate?: number;
  newStopLoss?: number;
  newTakeProfit?: number;
  timestamp?: string;
  success?: boolean;
}

// 统一的动作类型配置
export const UNIFIED_ACTION_CONFIG: Record<string, { 
  color: string; 
  bg: string; 
  icon: string; 
  label: string;
  category: 'open' | 'close' | 'modify' | 'wait' | 'other';
}> = {
  open_long: { 
    color: '#0ECB81', 
    bg: 'rgba(14, 203, 129, 0.15)', 
    icon: '📈', 
    label: 'Open Long',
    category: 'open'
  },
  open_short: { 
    color: '#F6465D', 
    bg: 'rgba(246, 70, 93, 0.15)', 
    icon: '📉', 
    label: 'Open Short',
    category: 'open'
  },
  close_long: { 
    color: '#F0B90B', 
    bg: 'rgba(240, 185, 11, 0.15)', 
    icon: '💰', 
    label: 'Close Long',
    category: 'close'
  },
  close_short: { 
    color: '#F0B90B', 
    bg: 'rgba(240, 185, 11, 0.15)', 
    icon: '💰', 
    label: 'Close Short',
    category: 'close'
  },
  hold: { 
    color: '#848E9C', 
    bg: 'rgba(132, 142, 156, 0.15)', 
    icon: '⏸️', 
    label: 'Hold',
    category: 'wait'
  },
  wait: { 
    color: '#848E9C', 
    bg: 'rgba(132, 142, 156, 0.15)', 
    icon: '⏳', 
    label: 'Wait',
    category: 'wait'
  },
  partial_close: { 
    color: '#9C27B0', 
    bg: 'rgba(156, 39, 176, 0.15)', 
    icon: '📉', 
    label: 'Partial Close',
    category: 'close'
  },
  update_stop_loss: { 
    color: '#FF9800', 
    bg: 'rgba(255, 152, 0, 0.15)', 
    icon: '🛡️', 
    label: 'Update Stop Loss',
    category: 'modify'
  },
  update_take_profit: { 
    color: '#4CAF50', 
    bg: 'rgba(76, 175, 80, 0.15)', 
    icon: '🎯', 
    label: 'Update Take Profit',
    category: 'modify'
  },
  trailing_stop: { 
    color: '#2196F3', 
    bg: 'rgba(33, 150, 243, 0.15)', 
    icon: '👣', 
    label: 'Trailing Stop',
    category: 'modify'
  },
  dynamic_take_profit: { 
    color: '#00BCD4', 
    bg: 'rgba(0, 188, 212, 0.15)', 
    icon: '🚀', 
    label: 'Dynamic Take Profit',
    category: 'modify'
  },
  add_position: { 
    color: '#795548', 
    bg: 'rgba(121, 85, 72, 0.15)', 
    icon: '➕', 
    label: 'Add Position',
    category: 'open'
  },
  full_close: { 
    color: '#607D8B', 
    bg: 'rgba(96, 125, 139, 0.15)', 
    icon: '🔚', 
    label: 'Full Close',
    category: 'close'
  }
};

// 统一的4列网格展示配置
export const UNIFIED_GRID_CONFIG = {
  columns: [
    {
      key: 'price',
      label: { en: 'Price', zh: '价格' },
      getValue: (action: TradeActionDisplayConfig) => action.price,
      format: 'price'
    },
    {
      key: 'secondary',
      label: { en: 'Secondary', zh: '次要信息' },
      getValue: (action: TradeActionDisplayConfig) => {
        // 根据动作类型返回不同的次要信息
        switch(action.action) {
          case 'partial_close':
            return action.closePercentage ? `${action.closePercentage}%` : '-';
          case 'trailing_stop':
            return action.trailPercentage ? `${action.trailPercentage}%` : '-';
          case 'hold':
            return `${action.confidence || 0}%`;
          case 'wait':
            return action.reasoning?.substring(0, 15) || '-';
          default:
            return action.stopLoss || '-';
        }
      },
      format: 'secondary'
    },
    {
      key: 'tertiary',
      label: { en: 'Tertiary', zh: '第三信息' },
      getValue: (action: TradeActionDisplayConfig) => {
        switch(action.action) {
          case 'trailing_stop':
            return action.callbackRate ? `${action.callbackRate}%` : '-';
          case 'hold':
            return action.reasoning?.substring(0, 15) || '-';
          case 'wait':
            return action.timestamp ? 
              `${Math.floor((new Date().getTime() - new Date(action.timestamp).getTime()) / 60000)}分钟` : 
              '-';
          default:
            return action.takeProfit || '-';
        }
      },
      format: 'tertiary'
    },
    {
      key: 'metrics',
      label: { en: 'Metrics', zh: '指标' },
      getValue: (action: TradeActionDisplayConfig) => {
        switch(action.action) {
          case 'hold':
          case 'wait':
            return `${action.confidence || 0}%`;
          case 'open_long':
          case 'open_short':
            return action.leverage ? `${action.leverage}x` : '-';
          default:
            return action.quantity || '-';
        }
      },
      format: 'metrics'
    }
  ]
};

// 格式化函数
export function formatPrice(price: number | undefined): string {
  if (price === undefined || Number.isNaN(price)) return '-';
  if (price === 0) return '0';
  if (price >= 1000) return price.toFixed(2);
  if (price >= 1) return price.toFixed(4);
  return price.toFixed(6);
}

export function getConfidenceColor(confidence: number | undefined): string {
  if (!confidence) return '#848E9C';
  if (confidence >= 80) return '#0ECB81';
  if (confidence >= 60) return '#F0B90B';
  return '#F6465D';
}

export function formatNumber(num: number | undefined): string {
  if (num === undefined || Number.isNaN(num)) return '-';
  return num.toString();
}