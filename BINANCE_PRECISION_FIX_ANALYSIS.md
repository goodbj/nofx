# 币安精度问题诊断和修复方案

## 🔍 问题诊断

根据查询结果，AKEUSDT的LOT_SIZE过滤器信息：
- `stepSize`: 0.001
- `minQty`: 0.001  
- `maxQty`: 1000

根据`calculatePrecision`函数的逻辑：
- `stepSize` = "0.001"
- 去掉尾随零后仍为"0.001"
- 小数点后有3位数字
- 因此精度应该是3

## 📋 当前系统状态检查

让我检查系统中精度处理的相关代码：

### 1. 精度计算函数
```go
func calculatePrecision(stepSize string) int {
    // Remove trailing zeros
    stepSize = trimTrailingZeros(stepSize)
    
    // Find decimal point
    dotIndex := -1
    for i := 0; i < len(stepSize); i++ {
        if stepSize[i] == '.' {
            dotIndex = i
            break
        }
    }
    
    // If no decimal point or decimal point is at the end, precision is 0
    if dotIndex == -1 || dotIndex == len(stepSize)-1 {
        return 0
    }
    
    // Return number of digits after decimal point
    return len(stepSize) - dotIndex - 1
}
```

### 2. 格式化函数
```go
func (t *FuturesTrader) FormatQuantity(symbol string, quantity float64) (string, error) {
    precision, err := t.GetSymbolPrecision(symbol)
    if err != nil {
        // If retrieval fails, use default format
        formatted := fmt.Sprintf("%.3f", quantity)
        logger.Warnf("⚠️ Using default precision for %s, formatted quantity: %s", symbol, formatted)
        return formatted, nil
    }
    
    format := fmt.Sprintf("%%.%df", precision)
    formatted := fmt.Sprintf(format, quantity)
    logger.Debugf("Formatted quantity for %s: %s (precision: %d)", symbol, formatted, precision)
    return formatted, nil
}
```

## 🛠️ 修复方案

根据您的建议，参照币安原生代码的过滤和判断系统，我建议：

### 方案1：增强精度处理逻辑
```go
// Enhanced precision handling with step size alignment
func (t *FuturesTrader) FormatQuantityWithStepSize(symbol string, quantity float64) (string, error) {
    // Get exchange info to get step size
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    exchangeInfo, err := t.client.NewExchangeInfoService().Do(ctx)
    if err != nil {
        // Fallback to default precision
        return fmt.Sprintf("%.3f", quantity), nil
    }
    
    // Find symbol and get step size
    var stepSize float64 = 0.001 // Default fallback
    for _, s := range exchangeInfo.Symbols {
        if s.Symbol == symbol {
            for _, filter := range s.Filters {
                if filter["filterType"] == "LOT_SIZE" {
                    if stepSizeStr, ok := filter["stepSize"].(string); ok {
                        stepSize, _ = strconv.ParseFloat(stepSizeStr, 64)
                        break
                    }
                }
            }
            break
        }
    }
    
    // Align quantity to step size (round down to nearest step)
    if stepSize > 0 {
        alignedQty := math.Floor(quantity/stepSize) * stepSize
        // Format to appropriate precision based on step size
        decimals := 0
        if stepSize < 1 {
            stepStr := strconv.FormatFloat(stepSize, 'f', -1, 64)
            if idx := strings.Index(stepStr, "."); idx >= 0 {
                decimals = len(stepStr) - idx - 1
            }
        }
        format := fmt.Sprintf("%%.%df", decimals)
        return fmt.Sprintf(format, alignedQty), nil
    }
    
    // Fallback to default formatting
    return fmt.Sprintf("%.3f", quantity), nil
}
```

### 方案2：改进现有GetSymbolPrecision函数
```go
func (t *FuturesTrader) GetSymbolPrecision(symbol string) (int, error) {
    // Try to get exchange info with retry logic
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    exchangeInfo, err := t.client.NewExchangeInfoService().Do(ctx)
    if err != nil {
        logger.Warnf("⚠️ Failed to get exchange info for %s, using default precision 3: %v", symbol, err)
        return 3, nil
    }
    
    for _, s := range exchangeInfo.Symbols {
        if s.Symbol == symbol {
            // Get precision from LOT_SIZE filter with enhanced parsing
            for _, filter := range s.Filters {
                if filter["filterType"] == "LOT_SIZE" {
                    stepSize := filter["stepSize"].(string)
                    precision := calculatePrecision(stepSize)
                    // Additional validation
                    if precision < 0 || precision > 10 {
                        logger.Warnf("⚠️ Invalid precision %d for %s, using default 3", precision, symbol)
                        return 3, nil
                    }
                    logger.Infof("  %s quantity precision: %d (stepSize: %s)", symbol, precision, stepSize)
                    return precision, nil
                }
            }
        }
    }
    
    logger.Infof("  ⚠ %s precision information not found, using default precision 3", symbol)
    return 3, nil
}
```

## ✅ 推荐实施步骤

1. **立即修复**：在`OpenLong`和`OpenShort`函数中使用更严格的精度处理
2. **长期优化**：实现缓存机制，避免重复查询exchangeInfo
3. **验证测试**：针对AKEUSDT进行专门测试

## 📊 预期效果

修复后应该能够：
- 正确处理AKEUSDT的0.001步长要求
- 避免精度超出限制的错误
- 保持与其他交易对的兼容性