package main

import (
	"fmt"
	"nofx/store"
)

func main() {
	// 创建数据库连接
	db, err := store.New("data/data.db")
	if err != nil {
		fmt.Printf("无法连接数据库: %v\n", err)
		return
	}
	defer db.Close()

	// 查询所有交易所配置
	var exchanges []*store.Exchange
	err = db.GormDB().Find(&exchanges).Error
	if err != nil {
		fmt.Printf("查询交易所配置失败: %v\n", err)
		return
	}

	fmt.Println("=== 当前交易所配置 ===")
	for i, exchange := range exchanges {
		fmt.Printf("\n[%d] 交易所配置:\n", i+1)
		fmt.Printf("  ID: %s\n", exchange.ID)
		fmt.Printf("  类型: %s\n", exchange.ExchangeType)
		fmt.Printf("  账户名: %s\n", exchange.AccountName)
		fmt.Printf("  名称: %s\n", exchange.Name)
		fmt.Printf("  启用: %t\n", exchange.Enabled)
		fmt.Printf("  测试网: %t\n", exchange.Testnet)
		fmt.Printf("  自定义API URL: %s\n", exchange.CustomAPIURL)
		fmt.Printf("  用户ID: %s\n", exchange.UserID)

		// 如果是币安且配置为测试网，则提示需要修改
		if exchange.ExchangeType == "binance" {
			if exchange.Testnet || exchange.CustomAPIURL == "https://testnet.binancefuture.com" {
				fmt.Printf("  ⚠️  此为测试网配置，需要修改为实盘\n")

				// 提供修改选项
				fmt.Printf("  尝试更新为实盘配置...\n")

				// 更新为实盘配置
				err := db.Exchange().Update(
					exchange.UserID,
					exchange.ID,
					exchange.Enabled,
					string(exchange.APIKey),
					string(exchange.SecretKey),
					string(exchange.Passphrase),
					false, // 改为实盘
					"",    // 清空自定义API URL，使用默认主网
					exchange.HyperliquidWalletAddr,
					exchange.AsterUser,
					exchange.AsterSigner,
					string(exchange.AsterPrivateKey),
					exchange.LighterWalletAddr,
					string(exchange.LighterPrivateKey),
					string(exchange.LighterAPIKeyPrivateKey),
					exchange.LighterAPIKeyIndex,
				)

				if err != nil {
					fmt.Printf("  ❌ 更新失败: %v\n", err)
				} else {
					fmt.Printf("  ✅ 更新成功！\n")

					// 验证更新结果
					updatedExchange, err := db.Exchange().GetByID(exchange.UserID, exchange.ID)
					if err != nil {
						fmt.Printf("  ❌ 验证失败: %v\n", err)
					} else {
						fmt.Printf("  验证结果 - 测试网: %t, 自定义API URL: %s\n", updatedExchange.Testnet, updatedExchange.CustomAPIURL)
					}
				}
			} else {
				fmt.Printf("  ✅ 此为实盘配置\n")
			}
		}
	}

	if len(exchanges) == 0 {
		fmt.Println("未找到任何交易所配置")
	}
}
