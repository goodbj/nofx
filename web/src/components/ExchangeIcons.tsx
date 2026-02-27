import React from 'react'

interface IconProps {
  width?: number
  height?: number
  className?: string
}

// 本地图标路径映射
const ICON_PATHS: Record<string, string> = {
  binance: '/exchange-icons/binance.jpg',
  binance_demo: '/exchange-icons/binance.jpg', // 虚拟盘使用相同图标
  bybit: '/exchange-icons/bybit.png',
  okx: '/exchange-icons/okx.svg',
  bitget: '/exchange-icons/bitget.svg',
  hyperliquid: '/exchange-icons/hyperliquid.png',
  aster: '/exchange-icons/aster.svg',
  lighter: '/exchange-icons/lighter.png',
}

// 通用图标组件 - 支持点击导航
const ExchangeImage: React.FC<IconProps & { src: string; alt: string; exchangeType?: string; onClick?: () => void }> = ({
  width = 24,
  height = 24,
  className,
  src,
  alt,
  exchangeType,
  onClick
}) => {
  const handleClick = () => {
    if (onClick) {
      onClick();
    } else if (exchangeType) {
      // 根据交易所类型打开对应的链接
      let exchangeUrl = '';
      const lowerType = exchangeType.toLowerCase();
      
      if (lowerType.includes('binance_demo')) {
        //币安虚拟盘
        exchangeUrl = 'https://demo.binance.com';
      } else if (lowerType.includes('binance')) {
        //币安实盘
        exchangeUrl = 'https://www.binance.com';
      } else if (lowerType.includes('bybit')) {
        // Bybit
        exchangeUrl = 'https://www.bybit.com';
      } else if (lowerType.includes('okx')) {
        // OKX
        exchangeUrl = 'https://www.okx.com';
      } else if (lowerType.includes('bitget')) {
        // Bitget
        exchangeUrl = 'https://www.bitget.com';
      } else if (lowerType.includes('hyperliquid')) {
        // Hyperliquid
        exchangeUrl = 'https://app.hyperliquid.xyz';
      } else if (lowerType.includes('aster')) {
        // Aster
        exchangeUrl = 'https://www.asterdex.com';
      } else if (lowerType.includes('lighter')) {
        // Lighter
        exchangeUrl = 'https://app.lighter.xyz';
      }
      
      if (exchangeUrl) {
        window.open(exchangeUrl, '_blank');
      }
    }
  };

  return (
    <div
      className={`${className} cursor-pointer hover:opacity-80 transition-opacity`}
      style={{
        width,
        height,
        borderRadius: 6,
        overflow: 'hidden',
        flexShrink: 0,
        background: '#2B3139',
      }}
      onClick={handleClick}
      title={`访问${alt}交易所`}
    >
      <img
        src={src}
        alt={alt}
        style={{
          width: '100%',
          height: '100%',
          objectFit: 'cover',
        }}
      />
    </div>
  );
}

// Fallback 图标
const FallbackIcon: React.FC<IconProps & { label: string }> = ({
  width = 24,
  height = 24,
  className,
  label,
}) => (
  <div
    className={className}
    style={{
      width,
      height,
      borderRadius: 6,
      background: '#2B3139',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      fontSize: Math.max(10, (width || 24) * 0.4),
      fontWeight: 'bold',
      color: '#EAECEF',
      flexShrink: 0,
    }}
  >
    {label[0]?.toUpperCase() || '?'}
  </div>
)

// 获取交易所图标的函数
export const getExchangeIcon = (
  exchangeType: string,
  props: IconProps = {}
) => {
  const lowerType = exchangeType.toLowerCase()
  
  const type = lowerType.includes('binance_demo')
    ? 'binance_demo'
    : lowerType.includes('binance')
      ? 'binance'
      : lowerType.includes('bybit')
        ? 'bybit'
        : lowerType.includes('okx')
          ? 'okx'
          : lowerType.includes('bitget')
            ? 'bitget'
            : lowerType.includes('hyperliquid')
              ? 'hyperliquid'
              : lowerType.includes('aster')
                ? 'aster'
                : lowerType.includes('lighter')
                  ? 'lighter'
                  : lowerType
  
  const iconProps = {
    width: props.width || 24,
    height: props.height || 24,
    className: props.className,
  }

  const path = ICON_PATHS[type]
  if (path) {
    //为binance_demo添加特殊标识
    const isDemo = type === 'binance_demo';
    const additionalStyle = isDemo ? {
      border: '2px solid #F0B90B',
      boxShadow: '0 0 10px rgba(240, 185, 11, 0.5)'
    } : {};
    
    return (
      <div style={{ display: 'inline-block', ...additionalStyle }}>
        <ExchangeImage {...iconProps} src={path} alt={type} exchangeType={exchangeType} />
        {isDemo && (
          <div style={{
            position: 'absolute',
            top: '-5px',
            right: '-5px',
            background: '#F0B90B',
            color: 'black',
            fontSize: '10px',
            padding: '1px 4px',
            borderRadius: '3px',
            fontWeight: 'bold'
          }}>
            DEMO
          </div>
        )}
      </div>
    );
  }

  return <FallbackIcon {...iconProps} label={type} />
}
