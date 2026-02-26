package main

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Decision 修复后的决策结构体
type Decision struct {
	Symbol            string      `json:"symbol"`
	Action            string      `json:"action"`
	Leverage          int         `json:"leverage,omitempty"`
	PositionSizeUSD   float64     `json:"position_size_usd,omitempty"`
	StopLoss          float64     `json:"stop_loss,omitempty"`
	TakeProfit        float64     `json:"take_profit,omitempty"`
	NewStopLoss       interface{} `json:"new_stop_loss,omitempty"` //支持字符串和数字
	NewTakeProfit     float64     `json:"new_take_profit,omitempty"`
	ClosePercentage   float64     `json:"close_percentage,omitempty"`
	PositionDirection string      `json:"position_direction,omitempty"`
	Confidence        int         `json:"confidence,omitempty"`
	RiskUSD           float64     `json:"risk_usd,omitempty"`
	Reasoning         string      `json:"reasoning"`
}

// GetNewStopLoss 获取new_stop_loss的float64值，支持字符串和数字类型
func (d *Decision) GetNewStopLoss() (float64, error) {
	switch v := d.NewStopLoss.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case string:
		//尝解析字符串为浮点数
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, nil
		}
		return 0, fmt.Errorf("invalid new_stop_loss string value: %s", v)
	case nil:
		return 0, fmt.Errorf("new_stop_loss is nil")
	default:
		return 0, fmt.Errorf("unsupported new_stop_loss type: %T", v)
	}
}

// GetNewTakeProfit 获取new_take_profit的float64值
func (d *Decision) GetNewTakeProfit() (float64, error) {
	return d.NewTakeProfit, nil
}

func main() {
	// 测试有问题的JSON
	testJSON := `[{"action":"close_long","confidence":100,"reasoning":"执行清仓指令","symbol":"SIRENUSDT"},{"action":"open_long","confidence":75,"leverage":5,"position_size_usd":4000,"reasoning":"放量突破后强势震荡看涨","risk_usd":200,"stop_loss":0.78,"symbol":"PIPPINUSDT","take_profit":0.866},{"action":"update_stop_loss","confidence":50,"new_stop_loss":"1.2000","position_direction":"long","reasoning":"随机决策无持仓","symbol":"POWERUSDT"}]`

	var decisions []Decision
	err := json.Unmarshal([]byte(testJSON), &decisions)
	if err != nil {
		fmt.Printf("❌ JSON解析失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 成功解析 %d 个决策\n", len(decisions))

	// 验证每个决策
	for i, decision := range decisions {
		fmt.Printf("\n决策 #%d:\n", i+1)
		fmt.Printf("  Action: %s\n", decision.Action)
		fmt.Printf("  Symbol: %s\n", decision.Symbol)
		fmt.Printf("  Confidence: %d\n", decision.Confidence)
		fmt.Printf("  Reasoning: %s\n", decision.Reasoning)

		if decision.Action == "update_stop_loss" {
			newStopLoss, err := decision.GetNewStopLoss()
			if err != nil {
				fmt.Printf("  ❌ new_stop_loss解析失败: %v\n", err)
			} else {
				fmt.Printf("  ✅ new_stop_loss解析成功: %.4f\n", newStopLoss)
			}
		}
	}

	fmt.Println("\n✅所有测试完成！")
}
