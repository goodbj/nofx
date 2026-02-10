package main

import (
	"fmt"
	"log"
	"nofx/config"
	"nofx/crypto"
	"nofx/store"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("🔍 API密钥调试工具")
	fmt.Println("==================")

	// 加载环境变量
	_ = godotenv.Load()
	config.Init()

	// 初始化加密服务
	cryptoService, err := crypto.NewCryptoService()
	if err != nil {
		log.Fatalf("加密服务初始化失败: %v", err)
	}

	// 初始化数据库
	s, err := store.NewWithConfig(store.DBConfig{
		Type: store.DBTypeSQLite,
		Path: "data/data.db",
	})
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer s.Close()

	// 获取所有交易所配置
	var exchanges []store.Exchange
	err = s.GormDB().Find(&exchanges).Error
	if err != nil {
		log.Fatalf("获取交易所配置失败: %v", err)
	}

	fmt.Printf("📊 找到 %d 个交易所配置\n\n", len(exchanges))

	for i, ex := range exchanges {
		fmt.Printf("[%d] 交易所: %s (ID: %s)\n", i+1, ex.AccountName, ex.ID)
		fmt.Printf("   类型: %s\n", ex.ExchangeType)
		fmt.Printf("   启用: %t\n", ex.Enabled)
		fmt.Printf("   测试网: %t\n", ex.Testnet)
		fmt.Printf("   自定义API URL: %s\n", ex.CustomAPIURL)

		// 解密并打印API密钥
		apiKeyStr := string(ex.APIKey)
		secretKeyStr := string(ex.SecretKey)

		fmt.Printf("   API密钥状态: ")
		if len(apiKeyStr) > 0 {
			if len(apiKeyStr) > 20 && apiKeyStr[:7] == "ENC:v1:" {
				// 解密API密钥
				decryptedAPIKey, err := cryptoService.DecryptFromStorage(apiKeyStr)
				if err != nil {
					fmt.Printf("❌ 解密失败: %v\n", err)
				} else {
					fmt.Printf("✅ 已解密 (长度: %d)\n", len(decryptedAPIKey))
					fmt.Printf("   明文API密钥: %s\n", decryptedAPIKey)
				}
			} else {
				fmt.Printf("⚠️  未加密 (长度: %d)\n", len(apiKeyStr))
				fmt.Printf("   明文API密钥: %s\n", apiKeyStr)
			}
		} else {
			fmt.Printf("❌ 空值\n")
		}

		fmt.Printf("   密钥状态: ")
		if len(secretKeyStr) > 0 {
			if len(secretKeyStr) > 20 && secretKeyStr[:7] == "ENC:v1:" {
				// 解密密钥
				decryptedSecretKey, err := cryptoService.DecryptFromStorage(secretKeyStr)
				if err != nil {
					fmt.Printf("❌ 解密失败: %v\n", err)
				} else {
					fmt.Printf("✅ 已解密 (长度: %d)\n", len(decryptedSecretKey))
					fmt.Printf("   明文密钥: %s\n", decryptedSecretKey)
				}
			} else {
				fmt.Printf("⚠️  未加密 (长度: %d)\n", len(secretKeyStr))
				fmt.Printf("   明文密钥: %s\n", secretKeyStr)
			}
		} else {
			fmt.Printf("❌ 空值\n")
		}

		fmt.Println()
	}

	// 测试Binance API连接
	fmt.Println("=== Binance API连接测试 ===")
	testBinanceConnection(cryptoService, s)
}

func testBinanceConnection(cryptoService *crypto.CryptoService, s *store.Store) {
	// 获取第一个启用的Binance交易所配置
	var exchange store.Exchange
	err := s.GormDB().Where("exchange_type = ? AND enabled = ?", "binance", true).First(&exchange).Error
	if err != nil {
		fmt.Printf("❌ 未找到启用的Binance交易所配置: %v\n", err)
		return
	}

	// 解密API密钥
	apiKey, err := cryptoService.DecryptFromStorage(string(exchange.APIKey))
	if err != nil {
		fmt.Printf("❌ API密钥解密失败: %v\n", err)
		return
	}

	secretKey, err := cryptoService.DecryptFromStorage(string(exchange.SecretKey))
	if err != nil {
		fmt.Printf("❌ 密钥解密失败: %v\n", err)
		return
	}

	fmt.Printf("📋 测试交易所: %s\n", exchange.AccountName)
	fmt.Printf("🔑 API密钥: %s\n", apiKey)
	fmt.Printf("🔑 密钥: %s\n", secretKey)
	fmt.Printf("🌐 端点: %s\n", exchange.CustomAPIURL)
	fmt.Printf("🧪 测试网: %t\n", exchange.Testnet)

	// 验证密钥长度（Binance要求至少64字符）
	if len(apiKey) < 64 {
		fmt.Printf("⚠️  API密钥长度不足: %d 字符 (要求至少64字符)\n", len(apiKey))
	} else {
		fmt.Printf("✅ API密钥长度符合要求: %d 字符\n", len(apiKey))
	}

	if len(secretKey) < 64 {
		fmt.Printf("⚠️  密钥长度不足: %d 字符 (要求至少64字符)\n", len(secretKey))
	} else {
		fmt.Printf("✅ 密钥长度符合要求: %d 字符\n", len(secretKey))
	}
}
