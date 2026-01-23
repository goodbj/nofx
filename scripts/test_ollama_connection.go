package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	fmt.Println("=== Testing Ollama Connection ===")
	fmt.Println()

	// Test 1: Check if Ollama server is running
	fmt.Println("1. Testing Ollama server status...")
	resp, err := http.Get("http://localhost:11434/api/tags")
	if err != nil {
		fmt.Printf("❌ Failed to connect to Ollama: %v\n", err)
		fmt.Println("   Please make sure Ollama is running on port 11434")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ Ollama returned status %d\n", resp.StatusCode)
		return
	}

	// Parse models list
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if models, ok := result["models"].([]interface{}); ok {
		fmt.Printf("✅ Ollama is running on port 11434\n")
		fmt.Printf("   Available models: %d\n", len(models))
		for i, model := range models {
			if m, ok := model.(map[string]interface{}); ok {
				fmt.Printf("   - %s\n", m["name"])
				if i >= 4 { // Show max 5 models
					fmt.Printf("   ... and %d more\n", len(models)-5)
					break
				}
			}
		}
	}
	fmt.Println()

	// Test 2: Test a simple inference
	fmt.Println("2. Testing AI inference...")

	reqBody := map[string]interface{}{
		"model":  "qwen2.5:14b",
		"prompt": "Hello, respond with 'OK' only",
		"stream": false,
	}

	jsonData, _ := json.Marshal(reqBody)

	client := &http.Client{Timeout: 30 * time.Second}
	resp2, err := client.Post(
		"http://localhost:11434/api/generate",
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		fmt.Printf("❌ Failed to call Ollama API: %v\n", err)
		return
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		body2, _ := io.ReadAll(resp2.Body)
		fmt.Printf("❌ Ollama API returned status %d: %s\n", resp2.StatusCode, string(body2))
		return
	}

	body2, _ := io.ReadAll(resp2.Body)
	var result2 map[string]interface{}
	json.Unmarshal(body2, &result2)

	if response, ok := result2["response"].(string); ok {
		fmt.Printf("✅ Ollama AI inference successful\n")
		fmt.Printf("   Response: %s\n", response[:minInt(len(response), 50)])
	}
	fmt.Println()

	fmt.Println("=== Summary ===")
	fmt.Println("✅ Ollama is working correctly on http://localhost:11434")
	fmt.Println("   You can use this URL in NOFX AI model configuration")
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
