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

	// Test creating a Guardian AI provider
	guardianProvider, err := mcp.CreateBrowserAIProvider("guardian-ai")
	if err != nil {
		fmt.Printf("Error creating Guardian AI provider: %v\n", err)
	} else {
		fmt.Printf("Successfully created Guardian AI provider: %T\n", guardianProvider)
		fmt.Printf("Service name: %s\n", guardianProvider.GetServiceName())
		fmt.Printf("Default URL: %s\n", guardianProvider.GetDefaultURL())
	}

	fmt.Println("Guardian AI Integration test completed.")
}
