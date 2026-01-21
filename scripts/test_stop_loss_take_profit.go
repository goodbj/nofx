package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Position represents a trading position
type Position struct {
	Symbol     string  `json:"symbol"`
	Side       string  `json:"side"`
	Quantity   float64 `json:"positionAmt"`
	EntryPrice float64 `json:"entryPrice"`
}

// Order represents a trading order
type Order struct {
	ExchangeOrderID string  `json:"exchange_order_id"`
	Type            string  `json:"type"`
	Status          string  `json:"status"`
	Price           float64 `json:"price"`
}

// ClosePositionRequest represents the request to close a position
type ClosePositionRequest struct {
	Symbol string `json:"symbol"`
	Side   string `json:"side"`
}

func main() {
	fmt.Println("🧪 Starting Stop Loss and Take Profit Test...")

	// Base API URL
	baseURL := "http://localhost:8888/api"

	// Test parameters
	testSymbol := "BTCUSDT"
	testTraderID := "test_trader_1"

	fmt.Printf("📍 Testing with symbol: %s, trader ID: %s\n", testSymbol, testTraderID)

	// 1. Test Get Positions
	fmt.Println("\n📋 1. Testing Get Positions...")
	positions, err := getPositions(baseURL, testTraderID)
	if err != nil {
		fmt.Printf("❌ Error getting positions: %v\n", err)
	} else {
		fmt.Printf("✅ Got %d positions\n", len(positions))
		for _, pos := range positions {
			fmt.Printf("   - Symbol: %s, Side: %s, Quantity: %.4f, Entry Price: %.4f\n",
				pos.Symbol, pos.Side, pos.Quantity, pos.EntryPrice)
		}
	}

	// 2. Test Get Orders
	fmt.Println("\n📋 2. Testing Get Orders...")
	orders, err := getOrders(baseURL, testTraderID, testSymbol)
	if err != nil {
		fmt.Printf("❌ Error getting orders: %v\n", err)
	} else {
		fmt.Printf("✅ Got %d orders for %s\n", len(orders), testSymbol)
		for _, order := range orders {
			fmt.Printf("   - Order ID: %s, Type: %s, Status: %s, Price: %.4f\n",
				order.ExchangeOrderID, order.Type, order.Status, order.Price)
		}
	}

	// 3. Test Manual Position Close (if position exists)
	fmt.Println("\nCloseOperation 3. Testing Manual Position Close...")
	if len(positions) > 0 {
		for _, pos := range positions {
			fmt.Printf("   Attempting to close position: %s %s\n", pos.Symbol, pos.Side)
			err := closePosition(baseURL, testTraderID, pos.Symbol, pos.Side)
			if err != nil {
				fmt.Printf("   ❌ Error closing position: %v\n", err)
			} else {
				fmt.Printf("   ✅ Position close initiated for %s %s\n", pos.Symbol, pos.Side)
			}
		}
	} else {
		fmt.Println("   ⚠️ No positions to close")
	}

	// 4. Test Partial Close (simulated via API call)
	fmt.Println("\nCloseOperation 4. Testing Partial Close...")
	err = simulatePartialClose(testTraderID, testSymbol, "LONG", 50.0) // Close 50%
	if err != nil {
		fmt.Printf("   ❌ Error simulating partial close: %v\n", err)
	} else {
		fmt.Printf("   ✅ Partial close simulated for %s\n", testSymbol)
	}

	// 5. Test Dynamic Stop Loss/Take Profit Updates
	fmt.Println("\n📊 5. Testing Dynamic Stop Loss/Take Profit Updates...")

	// Simulate updating stop loss
	err = simulateUpdateStopLoss(testTraderID, testSymbol, "LONG", 30000.0) // Example price
	if err != nil {
		fmt.Printf("   ❌ Error updating stop loss: %v\n", err)
	} else {
		fmt.Printf("   ✅ Stop loss update simulated for %s\n", testSymbol)
	}

	// Simulate updating take profit
	err = simulateUpdateTakeProfit(testTraderID, testSymbol, "LONG", 60000.0) // Example price
	if err != nil {
		fmt.Printf("   ❌ Error updating take profit: %v\n", err)
	} else {
		fmt.Printf("   ✅ Take profit update simulated for %s\n", testSymbol)
	}

	fmt.Println("\n✅ Stop Loss and Take Profit Test Completed!")
}

// getPositions fetches positions for a trader
func getPositions(baseURL, traderID string) ([]Position, error) {
	url := fmt.Sprintf("%s/positions?trader_id=%s", baseURL, traderID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var positions []Position
	err = json.Unmarshal(body, &positions)
	if err != nil {
		return nil, fmt.Errorf("failed to parse positions response: %v", err)
	}

	return positions, nil
}

// getOrders fetches orders for a trader and symbol
func getOrders(baseURL, traderID, symbol string) ([]Order, error) {
	url := fmt.Sprintf("%s/orders?trader_id=%s&symbol=%s", baseURL, traderID, symbol)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var orders []Order
	err = json.Unmarshal(body, &orders)
	if err != nil {
		return nil, fmt.Errorf("failed to parse orders response: %v", err)
	}

	return orders, nil
}

// closePosition closes a position via API
func closePosition(baseURL, traderID, symbol, side string) error {
	url := fmt.Sprintf("%s/traders/%s/close-position", baseURL, traderID)

	payload := ClosePositionRequest{
		Symbol: symbol,
		Side:   side,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("   Response: %s\n", string(body))
	return nil
}

// simulatePartialClose simulates a partial close operation
func simulatePartialClose(traderID, symbol, side string, percentage float64) error {
	fmt.Printf("   Would send partial close request: trader=%s, symbol=%s, side=%s, percentage=%.2f%%\n",
		traderID, symbol, side, percentage)

	// In a real scenario, this would be an API call to execute partial close
	// For now, we're just simulating the logic

	return nil
}

// simulateUpdateStopLoss simulates updating stop loss
func simulateUpdateStopLoss(traderID, symbol, side string, newStopPrice float64) error {
	fmt.Printf("   Would send stop loss update: trader=%s, symbol=%s, side=%s, new_stop_price=%.2f\n",
		traderID, symbol, side, newStopPrice)

	// In a real scenario, this would be an API call to update stop loss
	// For now, we're just simulating the logic

	return nil
}

// simulateUpdateTakeProfit simulates updating take profit
func simulateUpdateTakeProfit(traderID, symbol, side string, newTakeProfitPrice float64) error {
	fmt.Printf("   Would send take profit update: trader=%s, symbol=%s, side=%s, new_take_profit=%.2f\n",
		traderID, symbol, side, newTakeProfitPrice)

	// In a real scenario, this would be an API call to update take profit
	// For now, we're just simulating the logic

	return nil
}
