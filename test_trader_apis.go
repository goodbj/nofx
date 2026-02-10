package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func main() {
	fmt.Println("🔑 获取JWT Token并测试交易员API")
	fmt.Println("=================================")

	// 先进行登录获取token
	token, err := getJWTToken()
	if err != nil {
		fmt.Printf("❌ 获取Token失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 获取到JWT Token: %s\n", token[:20]+"...")

	// 测试两个交易员
	traderIDs := []string{
		"83faf7b3_deepseek_1770656396",    // 实盘交易员
		"75103af7_guardian-ai_1770569967", // 虚拟盘交易员
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	for _, traderID := range traderIDs {
		fmt.Printf("\n🔍 测试交易员: %s\n", traderID)
		fmt.Println("------------------------")

		// 测试账户信息API
		apiURL := fmt.Sprintf("http://localhost:8888/api/account?trader_id=%s", traderID)
		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			fmt.Printf("❌ 创建请求失败: %v\n", err)
			continue
		}

		// 添加认证头
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("❌ 请求失败: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		// 读取响应内容
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("❌ 读取响应失败: %v\n", err)
			continue
		}

		fmt.Printf("状态码: %d\n", resp.StatusCode)
		fmt.Printf("响应内容: %s\n", string(body))

		if resp.StatusCode == 200 {
			fmt.Printf("✅ 交易员 %s 余额获取成功\n", traderID)
			// 解析余额信息
			var accountInfo map[string]interface{}
			if err := json.Unmarshal(body, &accountInfo); err == nil {
				if balance, ok := accountInfo["totalWalletBalance"].(float64); ok {
					fmt.Printf("   总余额: %.2f USDT\n", balance)
				}
				if available, ok := accountInfo["availableBalance"].(float64); ok {
					fmt.Printf("   可用余额: %.2f USDT\n", available)
				}
			}
		} else if resp.StatusCode == 404 {
			fmt.Printf("❌ 交易员 %s 未找到 (404) - 可能未正确加载到内存\n", traderID)
		} else if resp.StatusCode == 500 {
			fmt.Printf("❌ 交易员 %s 服务器内部错误 (500) - 网络或API连接问题\n", traderID)
		} else {
			fmt.Printf("⚠️  交易员 %s 返回状态: %d\n", traderID, resp.StatusCode)
		}
	}

	fmt.Println("\n📋 诊断总结:")
	fmt.Println("============")
	fmt.Println("如果看到404错误，说明交易员未正确加载到内存")
	fmt.Println("如果看到500错误，说明交易员存在但获取余额时出现问题")
	fmt.Println("如果看到200成功，说明余额获取正常")
}

func getJWTToken() (string, error) {
	// 尝试使用默认的测试用户登录
	loginURL := "http://localhost:8888/api/auth/login"

	// 准备登录数据
	loginData := url.Values{}
	loginData.Set("email", "test@example.com")
	loginData.Set("password", "test123")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 如果默认登录失败，尝试注册
	req, err := http.NewRequest("POST", loginURL, strings.NewReader(loginData.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		// 登录成功，解析token
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			return "", err
		}

		if token, ok := result["token"].(string); ok {
			return token, nil
		}
	}

	// 如果登录失败，尝试注册
	return registerAndGetToken()
}

func registerAndGetToken() (string, error) {
	registerURL := "http://localhost:8888/api/auth/register"

	registerData := map[string]interface{}{
		"email":    "test@example.com",
		"password": "test123",
		"name":     "Test User",
	}

	jsonData, err := json.Marshal(registerData)
	if err != nil {
		return "", err
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("POST", registerURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("注册失败: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if token, ok := result["token"].(string); ok {
		return token, nil
	}

	return "", fmt.Errorf("无法获取token")
}
