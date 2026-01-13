package decision

import (
	"fmt"
	"testing"
)

// TestDynamicStopLossFunctionality 测试动态止损功能的基本逻辑
func TestDynamicStopLossFunctionality(t *testing.T) {
	fmt.Println("Testing Dynamic Stop Loss functionality...")
	
	// 测试决策结构体是否包含新字段
	decision := Decision{
		Symbol:          "BTCUSDT",
		Action:          "update_stop_loss",
		NewStopLoss:     45000.0,
		NewTakeProfit:   55000.0,
		ClosePercentage: 50.0,
		Reasoning:       "测试动态止损更新功能",
	}
	
	fmt.Printf("Decision created: %+v\n", decision)
	
	// 验证新操作类型是否被接受
	accountEquity := 10000.0
	btcEthLeverage := 10
	altcoinLeverage := 5
	
	err := validateDecision(&decision, accountEquity, btcEthLeverage, altcoinLeverage)
	if err != nil {
		t.Errorf("Validation failed for update_stop_loss: %v", err)
	} else {
		fmt.Println("✓ update_stop_loss validation passed")
	}
	
	// 测试partial_close操作
	decision.Action = "partial_close"
	decision.ClosePercentage = 25.0
	err = validateDecision(&decision, accountEquity, btcEthLeverage, altcoinLeverage)
	if err != nil {
		t.Errorf("Validation failed for partial_close: %v", err)
	} else {
		fmt.Println("✓ partial_close validation passed")
	}
	
	// 测试update_take_profit操作
	decision.Action = "update_take_profit"
	decision.NewTakeProfit = 52000.0
	err = validateDecision(&decision, accountEquity, btcEthLeverage, altcoinLeverage)
	if err != nil {
		t.Errorf("Validation failed for update_take_profit: %v", err)
	} else {
		fmt.Println("✓ update_take_profit validation passed")
	}
	
	// 测试无效的百分比
	decision.Action = "partial_close"
	decision.ClosePercentage = 150.0 // 无效值
	err = validateDecision(&decision, accountEquity, btcEthLeverage, altcoinLeverage)
	if err == nil {
		t.Errorf("Expected validation error for invalid percentage, but got none")
	} else {
		fmt.Printf("✓ Correctly rejected invalid percentage: %v\n", err)
	}
	
	fmt.Println("Dynamic Stop Loss functionality tests completed.")
}