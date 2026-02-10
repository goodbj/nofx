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
	fmt.Println("🔧 交易所配置修复工具")
	fmt.Println("====================")

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

	// 测试当前加密密钥
	fmt.Println("=== 当前加密环境测试 ===")
	testCurrentEncryption(cryptoService)

	// 修复解密失败的配置
	fmt.Println("\n=== 修复解密失败的配置 ===")
	fixFailedDecryptions(s, cryptoService)
}

func testCurrentEncryption(cs *crypto.CryptoService) {
	// 测试加密/解密功能
	testData := "test_api_key_1234567890"

	encrypted, err := cs.EncryptForStorage(testData)
	if err != nil {
		fmt.Printf("❌ 加密测试失败: %v\n", err)
		return
	}

	decrypted, err := cs.DecryptFromStorage(encrypted)
	if err != nil {
		fmt.Printf("❌ 解密测试失败: %v\n", err)
		return
	}

	if decrypted == testData {
		fmt.Printf("✅ 加密/解密功能正常\n")
		fmt.Printf("   原文: %s\n", testData)
		fmt.Printf("   密文: %s\n", encrypted)
		fmt.Printf("   解密: %s\n", decrypted)
	} else {
		fmt.Printf("❌ 加密/解密结果不匹配\n")
	}
}

func fixFailedDecryptions(s *store.Store, cs *crypto.CryptoService) {
	var exchanges []store.Exchange
	err := s.GormDB().Find(&exchanges).Error
	if err != nil {
		log.Fatalf("获取交易所配置失败: %v", err)
	}

	fixedCount := 0
	for i, ex := range exchanges {
		apiKeyStr := string(ex.APIKey)
		secretKeyStr := string(ex.SecretKey)

		// 检查是否能成功解密
		_, apiKeyErr := cs.DecryptFromStorage(apiKeyStr)
		_, secretKeyErr := cs.DecryptFromStorage(secretKeyStr)

		if apiKeyErr != nil || secretKeyErr != nil {
			fmt.Printf("[%d] ❌ 需要修复: %s (ID: %s)\n", i+1, ex.AccountName, ex.ID)

			// 这里可以选择重新加密配置
			// 但更安全的做法是让用户手动重新输入密钥
			fmt.Printf("   建议: 请在前端界面重新配置API密钥\n")
		} else {
			fmt.Printf("[%d] ✅ 正常: %s (ID: %s)\n", i+1, ex.AccountName, ex.ID)
			fixedCount++
		}
	}

	fmt.Printf("\n📊 统计: %d/%d 配置正常\n", fixedCount, len(exchanges))

	if fixedCount < len(exchanges) {
		fmt.Println("\n💡 修复建议:")
		fmt.Println("1. 在前端界面重新配置所有API密钥")
		fmt.Println("2. 或者联系系统管理员获取原始加密密钥")
	}
}
