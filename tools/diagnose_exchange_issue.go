package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Exchange struct {
	ID           string `json:"id"`
	ExchangeType string `json:"exchange_type"`
	AccountName  string `json:"account_name"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Enabled      bool   `json:"enabled"`
}

type TraderInfo struct {
	TraderID     string `json:"trader_id"`
	TraderName   string `json:"trader_name"`
	ExchangeID   string `json:"exchange_id"`
	ExchangeName string `json:"exchange_name"`
	ExchangeType string `json:"exchange_type"`
	Status       string `json:"status"`
}

func main() {
	fmt.Println("🔍 交易所模块初始化问题诊断")
	fmt.Println("==============================")

	// 1. 检查基础服务连接
	fmt.Println("\n1. 基础服务连接检查:")
	testAPI("http://localhost:8888/api/health", "后端健康检查")
	testAPI("http://localhost:8888/api/supported-exchanges", "交易所支持列表")

	// 2. 检查交易所配置API
	fmt.Println("\n2. 交易所配置检查:")
	resp, err := http.Get("http://localhost:8888/api/supported-exchanges")
	if err != nil {
		fmt.Printf("❌ 无法获取交易所支持列表: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var exchanges []Exchange
		if err := json.NewDecoder(resp.Body).Decode(&exchanges); err != nil {
			fmt.Printf("❌ 解析交易所数据失败: %v\n", err)
			return
		}

		fmt.Printf("✅ 支持的交易所类型 (%d个):\n", len(exchanges))
		for i, ex := range exchanges {
			fmt.Printf("   %d. %s (%s) - %s\n", i+1, ex.Name, ex.ExchangeType, ex.Type)
		}
	} else {
		fmt.Printf("❌ 交易所API返回错误状态: %d\n", resp.StatusCode)
	}

	// 3. 检查交易员API
	fmt.Println("\n3. 交易员状态检查:")
	testAPI("http://localhost:8888/api/my-traders", "我的交易员列表")

	// 4. 模拟前端初始化流程
	fmt.Println("\n4. 模拟前端初始化流程:")
	simulateFrontendInit()
}

func testAPI(url, name string) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("❌ %s: 连接失败 - %v\n", name, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Printf("✅ %s: 连接正常 (状态码: %d)\n", name, resp.StatusCode)
	} else {
		fmt.Printf("⚠️  %s: 状态码异常 - %d\n", name, resp.StatusCode)
	}
}

func simulateFrontendInit() {
	fmt.Println("   模拟前端启动流程...")

	// 步骤1: 获取系统配置
	fmt.Println("   1. 获取系统配置...")
	testAPI("http://localhost:8888/api/config", "系统配置")

	// 步骤2: 获取支持的模型
	fmt.Println("   2. 获取AI模型列表...")
	testAPI("http://localhost:8888/api/supported-models", "支持的AI模型")

	// 步骤3: 获取交易所配置（这可能就是问题所在）
	fmt.Println("   3. 获取交易所配置...")
	testAPI("http://localhost:8888/api/supported-exchanges", "交易所配置")

	// 步骤4: 尝试获取交易员列表（需要认证）
	fmt.Println("   4. 尝试获取交易员数据...")
	resp, err := http.Get("http://localhost:8888/api/my-traders")
	if err != nil {
		fmt.Printf("   ❌ 获取交易员列表失败: %v\n", err)
	} else {
		defer resp.Body.Close()
		if resp.StatusCode == 401 {
			fmt.Println("   ⚠️  需要认证才能获取交易员数据")
		} else if resp.StatusCode == 200 {
			fmt.Println("   ✅ 交易员数据获取成功")
		} else {
			fmt.Printf("   ⚠️  交易员API状态码: %d\n", resp.StatusCode)
		}
	}

	fmt.Println("   前端初始化流程模拟完成")
}
