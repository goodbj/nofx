package trader

import (
	"time"
)

// SymbolPrecisionInfo 交易对精度信息
type SymbolPrecisionInfo struct {
	Symbol     string
	StepSize   float64
	TickSize   float64
	MinQty     float64
	MaxQty     float64
	Precision  int
	LastUpdate time.Time
}
