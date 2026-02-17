package trader

// PrecisionManagerInterface 精度管理器接口
type PrecisionManagerInterface interface {
	GetPrecisionInfo(symbol string) (*SymbolPrecisionInfo, error)
	FormatQuantityWithValidation(symbol string, quantity float64) (string, error)
	FormatPriceWithValidation(symbol string, price float64) (string, error)
	ClearCache()
	GetCacheStats() map[string]interface{}
	CheckAndRefreshFileIfNeeded() // 新增：手动触发文件更新检查
}
