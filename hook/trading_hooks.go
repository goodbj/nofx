package hook

// Trading related hook results
type DecisionExecutionResult struct {
	Err     error
	Success bool
	Data    map[string]interface{}
}

func (r *DecisionExecutionResult) Error() error {
	return r.Err
}

func (r *DecisionExecutionResult) GetResult() (bool, map[string]interface{}) {
	return r.Success, r.Data
}

type PriceFetchResult struct {
	Err   error
	Price float64
}

func (r *PriceFetchResult) Error() error {
	return r.Err
}

func (r *PriceFetchResult) GetResult() float64 {
	return r.Price
}

type OrderExecutionResult struct {
	Err    error
	Order  interface{} // Could be different types depending on exchange
	Status string
}

func (r *OrderExecutionResult) Error() error {
	return r.Err
}

func (r *OrderExecutionResult) GetResult() (interface{}, string) {
	return r.Order, r.Status
}

type TradeRecordResult struct {
	Err error
	ID  uint
}

func (r *TradeRecordResult) Error() error {
	return r.Err
}

func (r *TradeRecordResult) GetResult() uint {
	return r.ID
}

// Trading related hook constants
const (
	// Decision execution hooks
	BEFORE_DECISION_EXECUTE  = "BEFORE_DECISION_EXECUTE"  // func(decision *kernel.Decision, traderID string) *DecisionExecutionResult
	AFTER_DECISION_EXECUTE   = "AFTER_DECISION_EXECUTE"   // func(decision *kernel.Decision, success bool, result map[string]interface{}, traderID string) *DecisionExecutionResult
	DECISION_EXECUTION_ERROR = "DECISION_EXECUTION_ERROR" // func(decision *kernel.Decision, err error, traderID string) *DecisionExecutionResult

	// Price fetching hooks
	BEFORE_PRICE_FETCH = "BEFORE_PRICE_FETCH" // func(symbol string, traderID string) *PriceFetchResult
	AFTER_PRICE_FETCH  = "AFTER_PRICE_FETCH"  // func(symbol string, price float64, traderID string) *PriceFetchResult

	// Order execution hooks
	BEFORE_ORDER_EXECUTE  = "BEFORE_ORDER_EXECUTE"  // func(orderParams map[string]interface{}, traderID string) *OrderExecutionResult
	AFTER_ORDER_EXECUTE   = "AFTER_ORDER_EXECUTE"   // func(order interface{}, status string, traderID string) *OrderExecutionResult
	ORDER_EXECUTION_ERROR = "ORDER_EXECUTION_ERROR" // func(orderParams map[string]interface{}, err error, traderID string) *OrderExecutionResult

	// Trade recording hooks
	BEFORE_TRADE_RECORD = "BEFORE_TRADE_RECORD" // func(actionRecord *store.DecisionAction, traderID string) *TradeRecordResult
	AFTER_TRADE_RECORD  = "AFTER_TRADE_RECORD"  // func(actionRecord *store.DecisionAction, recordID uint, traderID string) *TradeRecordResult

	// Trader lifecycle hooks
	TRADER_START         = "TRADER_START"         // func(traderID string, config map[string]interface{}) error
	TRADER_STOP          = "TRADER_STOP"          // func(traderID string) error
	TRADER_STATUS_UPDATE = "TRADER_STATUS_UPDATE" // func(traderID string, status map[string]interface{}) error

	// Balance sync hooks
	BEFORE_BALANCE_SYNC = "BEFORE_BALANCE_SYNC" // func(traderID string) error
	AFTER_BALANCE_SYNC  = "AFTER_BALANCE_SYNC"  // func(traderID string, balanceInfo map[string]interface{}) error

	// Position management hooks
	POSITION_UPDATE = "POSITION_UPDATE" // func(traderID string, positions []map[string]interface{}) error
	POSITION_CLOSE  = "POSITION_CLOSE"  // func(symbol string, closeParams map[string]interface{}, traderID string) error
)
