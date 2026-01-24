package guardian

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"nofx/config"
	"time"
)

// DataTransfer 数据传输器
type DataTransfer struct {
	config *config.DataTransferConfig
	client *http.Client
}

// NewDataTransfer 创建数据传输实例
func NewDataTransfer(transferConfig *config.DataTransferConfig) *DataTransfer {
	client := &http.Client{
		Timeout: time.Duration(transferConfig.TransferTimeout) * time.Second,
	}

	return &DataTransfer{
		config: transferConfig,
		client: client,
	}
}

// SendDecisionToNoFx 将AI决策发送到NoFx执行
func (dt *DataTransfer) SendDecisionToNoFx(decision string) error {
	log.Printf("📤 Sending decision to NoFx for execution, decision length: %d", len(decision))

	// 尝试解析决策内容以确定其类型
	var parsedDecision interface{}
	err := json.Unmarshal([]byte(decision), &parsedDecision)
	if err != nil {
		log.Printf("⚠️ Decision is not valid JSON, treating as raw content: %v", err)
		// 如果不是JSON，仍然尝试发送
		return dt.sendToNoFx(decision)
	}

	// 如果是有效的JSON，格式化后发送
	prettyDecision, err := json.MarshalIndent(parsedDecision, "", "  ")
	if err != nil {
		log.Printf("⚠️ Could not prettify decision JSON: %v", err)
		return dt.sendToNoFx(decision)
	}

	return dt.sendToNoFx(string(prettyDecision))
}

// sendToNoFx 向NoFx发送决策的实际方法
func (dt *DataTransfer) sendToNoFx(decision string) error {
	// 准备请求数据
	requestData := map[string]interface{}{
		"decision":  decision,
		"timestamp": time.Now().Unix(),
		"source":    "guardian", // 表示来自守护程序
	}

	// 将请求数据编码为JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %w", err)
	}

	// 尝试多次发送以提高成功率
	var lastErr error
	for attempt := 1; attempt <= dt.config.MaxTransferRetries; attempt++ {
		log.Printf("📡 Attempting to send to NoFx (attempt %d/%d)...", attempt, dt.config.MaxTransferRetries)

		// 创建HTTP请求
		req, err := http.NewRequest("POST", dt.config.NofxAPIEndpoint+"/guardian/execute", bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		// 设置请求头
		req.Header.Set("Content-Type", "application/json")
		if dt.config.APIKey != "" {
			req.Header.Set("Authorization", "Bearer "+dt.config.APIKey)
		}
		if dt.config.AuthToken != "" {
			req.Header.Set("X-Auth-Token", dt.config.AuthToken)
		}

		// 发送请求
		resp, err := dt.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			log.Printf("❌ Request attempt %d failed: %v", attempt, lastErr)

			if attempt < dt.config.MaxTransferRetries {
				delay := time.Duration(dt.config.TransferRetryDelay) * time.Millisecond
				log.Printf("⏳ Waiting %v before retry...", delay)
				time.Sleep(delay)
			}
			continue
		}

		// 读取响应
		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %w", err)
			log.Printf("❌ Response read attempt %d failed: %v", attempt, lastErr)

			if attempt < dt.config.MaxTransferRetries {
				delay := time.Duration(dt.config.TransferRetryDelay) * time.Millisecond
				log.Printf("⏳ Waiting %v before retry...", delay)
				time.Sleep(delay)
			}
			continue
		}

		// 检查响应状态码
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			log.Printf("✅ Decision successfully sent to NoFx, response: %s", string(respBody))
			return nil
		} else {
			lastErr = fmt.Errorf("received error response %d: %s", resp.StatusCode, string(respBody))
			log.Printf("❌ Request attempt %d failed with status %d: %v", attempt, resp.StatusCode, lastErr)

			if attempt < dt.config.MaxTransferRetries {
				delay := time.Duration(dt.config.TransferRetryDelay) * time.Millisecond
				log.Printf("⏳ Waiting %v before retry...", delay)
				time.Sleep(delay)
			}
			continue
		}
	}

	// 所有重试都失败了
	return fmt.Errorf("failed to send decision to NoFx after %d attempts: %w", dt.config.MaxTransferRetries, lastErr)
}

// TestConnection 测试与NoFx的连接
func (dt *DataTransfer) TestConnection() error {
	log.Println("🔍 Testing connection to NoFx...")

	req, err := http.NewRequest("GET", dt.config.NofxAPIEndpoint+"/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	if dt.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+dt.config.APIKey)
	}

	resp, err := dt.client.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("health check returned status %d: %s", resp.StatusCode, string(body))
	}

	log.Println("✅ Connection to NoFx successful")
	return nil
}

// ValidateDecision 验证决策格式是否正确
func (dt *DataTransfer) ValidateDecision(decision string) error {
	var parsed interface{}
	err := json.Unmarshal([]byte(decision), &parsed)
	if err != nil {
		return fmt.Errorf("decision is not valid JSON: %w", err)
	}

	// 这里可以添加更多验证逻辑
	log.Println("✅ Decision format validation passed")
	return nil
}
