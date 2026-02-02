// Package trader 交易者接口定义和实现
package trader

import (
	"nofx/logger"
	"time"
)

// ProxyTraderWrapper 代理交易者包装器
// 根据DataAccessMethod配置决定是否通过代理服务调用API
type ProxyTraderWrapper struct {
	trader           Trader // 原始交易者实例
	dataAccessMethod string // "native" 或 "proxy"
	proxyURL         string // 代理服务URL
	// 存储认证信息用于代理模式
	apiKey         string
	secretKey      string
	userId         string
	customEndpoint string
}

// NewProxyTraderWrapper 创建代理交易者包装器
func NewProxyTraderWrapper(originalTrader Trader, dataAccessMethod string, proxyURL string) *ProxyTraderWrapper {
	if proxyURL == "" {
		proxyURL = "http://localhost:8082" // 默认代理URL
	}

	logger.Infof("🔄 Creating ProxyTraderWrapper with method: %s, proxy: %s", dataAccessMethod, proxyURL)

	return &ProxyTraderWrapper{
		trader:           originalTrader,
		dataAccessMethod: dataAccessMethod,
		proxyURL:         proxyURL,
	}
}

// NewProxyTraderWrapperWithAuth 创建带认证信息的代理交易者包装器
func NewProxyTraderWrapperWithAuth(originalTrader Trader, dataAccessMethod string, proxyURL, apiKey, secretKey, customAPIURL string) *ProxyTraderWrapper {
	if proxyURL == "" {
		proxyURL = "http://localhost:8082" // 默认代理URL
	}

	logger.Infof("🔄 Creating ProxyTraderWrapper with method: %s, proxy: %s", dataAccessMethod, proxyURL)

	pw := &ProxyTraderWrapper{
		trader:           originalTrader,
		dataAccessMethod: dataAccessMethod,
		proxyURL:         proxyURL,
		apiKey:           apiKey,
		secretKey:        secretKey,
		userId:           "", // userId 通常在创建时确定，可以从原始交易者中获取或设为默认值
		customEndpoint:   customAPIURL,
	}

	return pw
}

// shouldUseProxy 判断是否应该使用代理
func (p *ProxyTraderWrapper) shouldUseProxy() bool {
	return p.dataAccessMethod == "proxy"
}

// GetBalance 获取账户余额
// 这是统一的API调用入口点
func (p *ProxyTraderWrapper) GetBalance() (map[string]interface{}, error) {
	if p.shouldUseProxy() {
		logger.Infof("🏦 [%s] Using proxy for balance query", p.dataAccessMethod)
		// 在代理模式下，创建一个新的FuturesTrader实例，连接到代理服务，同时告知真正的目标端点
		proxyTrader := NewFuturesTraderViaProxy(p.apiKey, p.secretKey, p.userId, p.proxyURL, p.customEndpoint)
		return proxyTrader.GetBalance()
	}

	logger.Infof("🏦 [%s] Using direct connection for balance query", p.dataAccessMethod)
	return p.trader.GetBalance()
}

// GetPositions 获取持仓信息
func (p *ProxyTraderWrapper) GetPositions() ([]map[string]interface{}, error) {
	if p.shouldUseProxy() {
		logger.Infof("📈 [%s] Using proxy for positions query", p.dataAccessMethod)
		// 在代理模式下，创建一个新的FuturesTrader实例，连接到代理服务，同时告知真正的目标端点
		proxyTrader := NewFuturesTraderViaProxy(p.apiKey, p.secretKey, p.userId, p.proxyURL, p.customEndpoint)
		return proxyTrader.GetPositions()
	}

	logger.Infof("📈 [%s] Using direct connection for positions query", p.dataAccessMethod)
	positions, err := p.trader.GetPositions()
	if err != nil {
		return nil, err
	}
	return positions, nil
}

// OpenLong 开多
func (p *ProxyTraderWrapper) OpenLong(symbol string, quantity float64, leverage int) (map[string]interface{}, error) {
	if p.shouldUseProxy() {
		logger.Infof("📈 [%s] Using proxy for opening long position", p.dataAccessMethod)
		// 在代理模式下，创建一个新的FuturesTrader实例，连接到代理服务，同时告知真正的目标端点
		proxyTrader := NewFuturesTraderViaProxy(p.apiKey, p.secretKey, p.userId, p.proxyURL, p.customEndpoint)
		return proxyTrader.OpenLong(symbol, quantity, leverage)
	}

	logger.Infof("📈 [%s] Using direct connection for opening long position", p.dataAccessMethod)
	return p.trader.OpenLong(symbol, quantity, leverage)
}

// OpenShort 开空
func (p *ProxyTraderWrapper) OpenShort(symbol string, quantity float64, leverage int) (map[string]interface{}, error) {
	if p.shouldUseProxy() {
		logger.Infof("📉 [%s] Using proxy for opening short position", p.dataAccessMethod)
		// 在代理模式下，创建一个新的FuturesTrader实例，连接到代理服务，同时告知真正的目标端点
		proxyTrader := NewFuturesTraderViaProxy(p.apiKey, p.secretKey, p.userId, p.proxyURL, p.customEndpoint)
		return proxyTrader.OpenShort(symbol, quantity, leverage)
	}

	logger.Infof("📉 [%s] Using direct connection for opening short position", p.dataAccessMethod)
	return p.trader.OpenShort(symbol, quantity, leverage)
}

// CloseLong 平多
func (p *ProxyTraderWrapper) CloseLong(symbol string, quantity float64) (map[string]interface{}, error) {
	if p.shouldUseProxy() {
		logger.Infof("📈 [%s] Using proxy for closing long position", p.dataAccessMethod)
		// 在代理模式下，创建一个新的FuturesTrader实例，连接到代理服务，同时告知真正的目标端点
		proxyTrader := NewFuturesTraderViaProxy(p.apiKey, p.secretKey, p.userId, p.proxyURL, p.customEndpoint)
		return proxyTrader.CloseLong(symbol, quantity)
	}

	logger.Infof("📈 [%s] Using direct connection for closing long position", p.dataAccessMethod)
	return p.trader.CloseLong(symbol, quantity)
}

// CloseShort 平空
func (p *ProxyTraderWrapper) CloseShort(symbol string, quantity float64) (map[string]interface{}, error) {
	if p.shouldUseProxy() {
		logger.Infof("📉 [%s] Using proxy for closing short position", p.dataAccessMethod)
		// 在代理模式下，创建一个新的FuturesTrader实例，连接到代理服务，同时告知真正的目标端点
		proxyTrader := NewFuturesTraderViaProxy(p.apiKey, p.secretKey, p.userId, p.proxyURL, p.customEndpoint)
		return proxyTrader.CloseShort(symbol, quantity)
	}

	logger.Infof("📉 [%s] Using direct connection for closing short position", p.dataAccessMethod)
	return p.trader.CloseShort(symbol, quantity)
}

// PartialClose 部分平仓
func (p *ProxyTraderWrapper) PartialClose(symbol string, side string, percentage float64) (map[string]interface{}, error) {
	if p.shouldUseProxy() {
		logger.Infof("🔄 [%s] Using proxy for partial close", p.dataAccessMethod)
		// 在代理模式下，创建一个新的FuturesTrader实例，连接到代理服务，同时告知真正的目标端点
		proxyTrader := NewFuturesTraderViaProxy(p.apiKey, p.secretKey, p.userId, p.proxyURL, p.customEndpoint)
		return proxyTrader.PartialClose(symbol, side, percentage)
	}

	logger.Infof("🔄 [%s] Using direct connection for partial close", p.dataAccessMethod)
	return p.trader.PartialClose(symbol, side, percentage)
}

// UpdateStopLoss 更新止损单
func (p *ProxyTraderWrapper) UpdateStopLoss(symbol string, positionSide string, newStopPrice float64) error {
	if p.shouldUseProxy() {
		logger.Infof("🛑 [%s] Using proxy for updating stop loss", p.dataAccessMethod)
		// 在代理模式下，创建一个新的FuturesTrader实例，连接到代理服务，同时告知真正的目标端点
		proxyTrader := NewFuturesTraderViaProxy(p.apiKey, p.secretKey, p.userId, p.proxyURL, p.customEndpoint)
		return proxyTrader.UpdateStopLoss(symbol, positionSide, newStopPrice)
	}

	logger.Infof("🛑 [%s] Using direct connection for updating stop loss", p.dataAccessMethod)
	return p.trader.UpdateStopLoss(symbol, positionSide, newStopPrice)
}

// UpdateTakeProfit 更新止盈单
func (p *ProxyTraderWrapper) UpdateTakeProfit(symbol string, positionSide string, newTakeProfitPrice float64) error {
	if p.shouldUseProxy() {
		logger.Infof("🎯 [%s] Using proxy for updating take profit", p.dataAccessMethod)
		// 在代理模式下，创建一个新的FuturesTrader实例，连接到代理服务，同时告知真正的目标端点
		proxyTrader := NewFuturesTraderViaProxy(p.apiKey, p.secretKey, p.userId, p.proxyURL, p.customEndpoint)
		return proxyTrader.UpdateTakeProfit(symbol, positionSide, newTakeProfitPrice)
	}

	logger.Infof("🎯 [%s] Using direct connection for updating take profit", p.dataAccessMethod)
	return p.trader.UpdateTakeProfit(symbol, positionSide, newTakeProfitPrice)
}

// SetLeverage 设置杠杆
func (p *ProxyTraderWrapper) SetLeverage(symbol string, leverage int) error {
	if p.shouldUseProxy() {
		logger.Infof("⚙️ [%s] Using proxy for setting leverage", p.dataAccessMethod)
		// 在代理模式下，创建一个新的FuturesTrader实例，连接到代理服务，同时告知真正的目标端点
		proxyTrader := NewFuturesTraderViaProxy(p.apiKey, p.secretKey, p.userId, p.proxyURL, p.customEndpoint)
		return proxyTrader.SetLeverage(symbol, leverage)
	}

	logger.Infof("⚙️ [%s] Using direct connection for setting leverage", p.dataAccessMethod)
	return p.trader.SetLeverage(symbol, leverage)
}

// SetMarginMode 设置保证金模式
func (p *ProxyTraderWrapper) SetMarginMode(symbol string, isCrossMargin bool) error {
	if p.shouldUseProxy() {
		logger.Infof("⚖️ [%s] Using proxy for setting margin mode", p.dataAccessMethod)
		// 在代理模式下，创建一个新的FuturesTrader实例，连接到代理服务，同时告知真正的目标端点
		proxyTrader := NewFuturesTraderViaProxy(p.apiKey, p.secretKey, p.userId, p.proxyURL, p.customEndpoint)
		return proxyTrader.SetMarginMode(symbol, isCrossMargin)
	}

	logger.Infof("⚖️ [%s] Using direct connection for setting margin mode", p.dataAccessMethod)
	return p.trader.SetMarginMode(symbol, isCrossMargin)
}

// GetMarketPrice 获取市场价格
func (p *ProxyTraderWrapper) GetMarketPrice(symbol string) (float64, error) {
	if p.shouldUseProxy() {
		logger.Infof("💰 [%s] Using proxy for getting market price", p.dataAccessMethod)
		// 在代理模式下，创建一个新的FuturesTrader实例，连接到代理服务，同时告知真正的目标端点
		proxyTrader := NewFuturesTraderViaProxy(p.apiKey, p.secretKey, p.userId, p.proxyURL, p.customEndpoint)
		return proxyTrader.GetMarketPrice(symbol)
	}

	logger.Infof("💰 [%s] Using direct connection for getting market price", p.dataAccessMethod)
	return p.trader.GetMarketPrice(symbol)
}

// SetStopLoss 设置止损单
func (p *ProxyTraderWrapper) SetStopLoss(symbol string, positionSide string, quantity, stopPrice float64) error {
	if p.shouldUseProxy() {
		logger.Infof("🛑 [%s] Using proxy for setting stop loss", p.dataAccessMethod)
		// 在代理模式下，创建一个新的FuturesTrader实例，连接到代理服务，同时告知真正的目标端点
		proxyTrader := NewFuturesTraderViaProxy(p.apiKey, p.secretKey, p.userId, p.proxyURL, p.customEndpoint)
		return proxyTrader.SetStopLoss(symbol, positionSide, quantity, stopPrice)
	}

	logger.Infof("🛑 [%s] Using direct connection for setting stop loss", p.dataAccessMethod)
	return p.trader.SetStopLoss(symbol, positionSide, quantity, stopPrice)
}

// SetTakeProfit 设置止盈单
func (p *ProxyTraderWrapper) SetTakeProfit(symbol string, positionSide string, quantity, takeProfitPrice float64) error {
	if p.shouldUseProxy() {
		logger.Infof("🎯 [%s] Using proxy for setting take profit", p.dataAccessMethod)
		// 在代理模式下，创建一个新的FuturesTrader实例，连接到代理服务，同时告知真正的目标端点
		proxyTrader := NewFuturesTraderViaProxy(p.apiKey, p.secretKey, p.userId, p.proxyURL, p.customEndpoint)
		return proxyTrader.SetTakeProfit(symbol, positionSide, quantity, takeProfitPrice)
	}

	logger.Infof("🎯 [%s] Using direct connection for setting take profit", p.dataAccessMethod)
	return p.trader.SetTakeProfit(symbol, positionSide, quantity, takeProfitPrice)
}

// SetTrailingStop 设置追踪止损单
func (p *ProxyTraderWrapper) SetTrailingStop(symbol string, positionSide string, quantity, callbackRate, activationPrice float64) error {
	if p.shouldUseProxy() {
		logger.Infof("➰ [%s] Using proxy for setting trailing stop", p.dataAccessMethod)
		// 在代理模式下，创建一个新的FuturesTrader实例，连接到代理服务，同时告知真正的目标端点
		proxyTrader := NewFuturesTraderViaProxy(p.apiKey, p.secretKey, p.userId, p.proxyURL, p.customEndpoint)
		return proxyTrader.SetTrailingStop(symbol, positionSide, quantity, callbackRate, activationPrice)
	}

	logger.Infof("➰ [%s] Using direct connection for setting trailing stop", p.dataAccessMethod)
	return p.trader.SetTrailingStop(symbol, positionSide, quantity, callbackRate, activationPrice)
}

// CancelStopLossOrders 取消止损单
func (p *ProxyTraderWrapper) CancelStopLossOrders(symbol string) error {
	if p.shouldUseProxy() {
		logger.Infof("🚫 [%s] Using proxy for cancelling stop loss orders", p.dataAccessMethod)
		// 在代理模式下，创建一个新的FuturesTrader实例，连接到代理服务，同时告知真正的目标端点
		proxyTrader := NewFuturesTraderViaProxy(p.apiKey, p.secretKey, p.userId, p.proxyURL, p.customEndpoint)
		return proxyTrader.CancelStopLossOrders(symbol)
	}

	logger.Infof("🚫 [%s] Using direct connection for cancelling stop loss orders", p.dataAccessMethod)
	return p.trader.CancelStopLossOrders(symbol)
}

// CancelTakeProfitOrders 取消止盈单
func (p *ProxyTraderWrapper) CancelTakeProfitOrders(symbol string) error {
	if p.shouldUseProxy() {
		logger.Infof("🚫 [%s] Using proxy for cancelling take profit orders", p.dataAccessMethod)
		// 在代理模式下，创建一个新的FuturesTrader实例，连接到代理服务，同时告知真正的目标端点
		proxyTrader := NewFuturesTraderViaProxy(p.apiKey, p.secretKey, p.userId, p.proxyURL, p.customEndpoint)
		return proxyTrader.CancelTakeProfitOrders(symbol)
	}

	logger.Infof("🚫 [%s] Using direct connection for cancelling take profit orders", p.dataAccessMethod)
	return p.trader.CancelTakeProfitOrders(symbol)
}

// CancelAllOrders 取消所有订单
func (p *ProxyTraderWrapper) CancelAllOrders(symbol string) error {
	if p.shouldUseProxy() {
		logger.Infof("🚫 [%s] Using proxy for cancelling all orders", p.dataAccessMethod)
		// 在代理模式下，创建一个新的FuturesTrader实例，连接到代理服务，同时告知真正的目标端点
		proxyTrader := NewFuturesTraderViaProxy(p.apiKey, p.secretKey, p.userId, p.proxyURL, p.customEndpoint)
		return proxyTrader.CancelAllOrders(symbol)
	}

	logger.Infof("🚫 [%s] Using direct connection for cancelling all orders", p.dataAccessMethod)
	return p.trader.CancelAllOrders(symbol)
}

// CancelStopOrders 取消止盈止损单
func (p *ProxyTraderWrapper) CancelStopOrders(symbol string) error {
	if p.shouldUseProxy() {
		logger.Infof("🚫 [%s] Using proxy for cancelling stop orders", p.dataAccessMethod)
		// 在代理模式下，创建一个新的FuturesTrader实例，连接到代理服务，同时告知真正的目标端点
		proxyTrader := NewFuturesTraderViaProxy(p.apiKey, p.secretKey, p.userId, p.proxyURL, p.customEndpoint)
		return proxyTrader.CancelStopOrders(symbol)
	}

	logger.Infof("🚫 [%s] Using direct connection for cancelling stop orders", p.dataAccessMethod)
	return p.trader.CancelStopOrders(symbol)
}

// FormatQuantity 格式化数量
func (p *ProxyTraderWrapper) FormatQuantity(symbol string, quantity float64) (string, error) {
	// Formatting is usually done locally, so use direct connection
	logger.Infof("🔢 [%s] Using direct connection for formatting quantity", p.dataAccessMethod)
	return p.trader.FormatQuantity(symbol, quantity)
}

// GetOrderStatus 获取订单状态
func (p *ProxyTraderWrapper) GetOrderStatus(symbol string, orderID string) (map[string]interface{}, error) {
	if p.shouldUseProxy() {
		logger.Infof("📋 [%s] Using proxy for getting order status", p.dataAccessMethod)
		// 在代理模式下，创建一个新的FuturesTrader实例，连接到代理服务，同时告知真正的目标端点
		proxyTrader := NewFuturesTraderViaProxy(p.apiKey, p.secretKey, p.userId, p.proxyURL, p.customEndpoint)
		return proxyTrader.GetOrderStatus(symbol, orderID)
	}

	logger.Infof("📋 [%s] Using direct connection for getting order status", p.dataAccessMethod)
	return p.trader.GetOrderStatus(symbol, orderID)
}

// GetClosedPnL 获取已平仓盈亏记录
func (p *ProxyTraderWrapper) GetClosedPnL(startTime time.Time, limit int) ([]ClosedPnLRecord, error) {
	if p.shouldUseProxy() {
		logger.Infof("📊 [%s] Using proxy for getting closed PnL", p.dataAccessMethod)
		// 在代理模式下，创建一个新的FuturesTrader实例，连接到代理服务，同时告知真正的目标端点
		proxyTrader := NewFuturesTraderViaProxy(p.apiKey, p.secretKey, p.userId, p.proxyURL, p.customEndpoint)
		return proxyTrader.GetClosedPnL(startTime, limit)
	}

	logger.Infof("📊 [%s] Using direct connection for getting closed PnL", p.dataAccessMethod)
	return p.trader.GetClosedPnL(startTime, limit)
}

// GetOpenOrders 获取挂单
func (p *ProxyTraderWrapper) GetOpenOrders(symbol string) ([]OpenOrder, error) {
	if p.shouldUseProxy() {
		logger.Infof("🛒 [%s] Using proxy for getting open orders", p.dataAccessMethod)
		// 在代理模式下，创建一个新的FuturesTrader实例，连接到代理服务，同时告知真正的目标端点
		proxyTrader := NewFuturesTraderViaProxy(p.apiKey, p.secretKey, p.userId, p.proxyURL, p.customEndpoint)
		return proxyTrader.GetOpenOrders(symbol)
	}

	logger.Infof("🛒 [%s] Using direct connection for getting open orders", p.dataAccessMethod)
	return p.trader.GetOpenOrders(symbol)
}
