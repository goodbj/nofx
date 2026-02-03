package main

import (
	"fmt"
	"nofx/mcp"
)

func main() {
	fmt.Println("Testing Guardian AI Integration...")

	// Test listing available providers
	providers := mcp.ListBrowserAIProviders()
	fmt.Printf("Available Browser AI Providers: %v\n", providers)

	// Test creating a DeepSeek browser provider
	provider, err := mcp.CreateBrowserAIProvider("deepseek-browser")
	if err != nil {
		fmt.Printf("Error creating DeepSeek provider: %v\n", err)
	} else {
		fmt.Printf("Successfully created DeepSeek provider: %T\n", provider)
		fmt.Printf("Service name: %s\n", provider.GetServiceName())
		fmt.Printf("Default URL: %s\n", provider.GetDefaultURL())
	}

	// Test creating a Guardian AI provider
	guardianProvider, err := mcp.CreateBrowserAIProvider("guardian-ai")
	if err != nil {
		fmt.Printf("Error creating Guardian AI provider: %v\n", err)
	} else {
		fmt.Printf("Successfully created Guardian AI provider: %T\n", guardianProvider)
		fmt.Printf("Service name: %s\n", guardianProvider.GetServiceName())
		fmt.Printf("Default URL: %s\n", guardianProvider.GetDefaultURL())
	}

	// Test creating a ChatGPT browser provider
	chatgptProvider, err := mcp.CreateBrowserAIProvider("chatgpt-browser")
	if err != nil {
		fmt.Printf("Error creating ChatGPT provider: %v\n", err)
	} else {
		fmt.Printf("Successfully created ChatGPT provider: %T\n", chatgptProvider)
		fmt.Printf("Service name: %s\n", chatgptProvider.GetServiceName())
		fmt.Printf("Default URL: %s\n", chatgptProvider.GetDefaultURL())
	}

	// Test creating a Claude browser provider
	claudeProvider, err := mcp.CreateBrowserAIProvider("claude-browser")
	if err != nil {
		fmt.Printf("Error creating Claude provider: %v\n", err)
	} else {
		fmt.Printf("Successfully created Claude provider: %T\n", claudeProvider)
		fmt.Printf("Service name: %s\n", claudeProvider.GetServiceName())
		fmt.Printf("Default URL: %s\n", claudeProvider.GetDefaultURL())
	}

	fmt.Println("Guardian AI Integration test completed.")
}
