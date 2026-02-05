package mcp

import (
	"fmt"
	"sync"
	"time"
)

// BrowserAIProvider defines the interface for browser-based AI services
type BrowserAIProvider interface {
	AIClient
	GetServiceName() string
	GetDefaultURL() string
	GetInputSelectors() []string
	GetSubmitSelectors() []string
	GetResponseSelectors() []string
}

// browserAIRegistry holds all registered browser AI providers
type browserAIRegistry struct {
	providers map[string]func() BrowserAIProvider
	mutex     sync.RWMutex
}

// global registry instance
var globalRegistry = &browserAIRegistry{
	providers: make(map[string]func() BrowserAIProvider),
}

// RegisterBrowserAIProvider registers a new browser AI provider
func RegisterBrowserAIProvider(name string, factory func() BrowserAIProvider) error {
	globalRegistry.mutex.Lock()
	defer globalRegistry.mutex.Unlock()

	if _, exists := globalRegistry.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}
	globalRegistry.providers[name] = factory
	return nil
}

// CreateBrowserAIProvider creates an instance of the specified browser AI provider
func CreateBrowserAIProvider(name string) (BrowserAIProvider, error) {
	globalRegistry.mutex.RLock()
	factory, exists := globalRegistry.providers[name]
	globalRegistry.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	return factory(), nil
}

// ListBrowserAIProviders returns a list of all registered browser AI provider names
func ListBrowserAIProviders() []string {
	globalRegistry.mutex.RLock()
	defer globalRegistry.mutex.RUnlock()

	names := make([]string, 0, len(globalRegistry.providers))
	for name := range globalRegistry.providers {
		names = append(names, name)
	}
	return names
}

// BaseBrowserAIProvider provides common functionality for browser AI providers
type BaseBrowserAIProvider struct {
	*GuardianClient
	ServiceName       string
	DefaultURL        string
	InputSelectors    []string
	SubmitSelectors   []string
	ResponseSelectors []string
}

// NewBaseBrowserAIProvider creates a new base browser AI provider
func NewBaseBrowserAIProvider(name, defaultURL string) *BaseBrowserAIProvider {
	return NewBaseBrowserAIProviderWithTraderID(name, defaultURL, "") // Default no trader ID
}

// NewBaseBrowserAIProviderWithTraderID creates a new base browser AI provider with trader ID
func NewBaseBrowserAIProviderWithTraderID(name, defaultURL, traderID string) *BaseBrowserAIProvider {
	fmt.Printf("🚨 [BASE PROVIDER DEBUG] Creating BaseBrowserAIProvider with traderID: '%s'\n", traderID) // Debug print
	client := NewGuardianClientWithOptions(
		WithTraderID(traderID), // Pass trader ID for browser data isolation
	).(*GuardianClient)
	fmt.Printf("🚨 [BASE PROVIDER DEBUG] Created GuardianClient with traderID: '%s'\n", client.GetTraderID()) // Debug print

	return &BaseBrowserAIProvider{
		GuardianClient:    client,
		ServiceName:       name,
		DefaultURL:        defaultURL,
		InputSelectors:    []string{},
		SubmitSelectors:   []string{},
		ResponseSelectors: []string{},
	}
}

// GetServiceName returns the service name
func (b *BaseBrowserAIProvider) GetServiceName() string {
	return b.ServiceName
}

// GetDefaultURL returns the default URL for the service
func (b *BaseBrowserAIProvider) GetDefaultURL() string {
	return b.DefaultURL
}

// GetInputSelectors returns the input selectors for the service
func (b *BaseBrowserAIProvider) GetInputSelectors() []string {
	return b.InputSelectors
}

// GetSubmitSelectors returns the submit selectors for the service
func (b *BaseBrowserAIProvider) GetSubmitSelectors() []string {
	return b.SubmitSelectors
}

// GetResponseSelectors returns the response selectors for the service
func (b *BaseBrowserAIProvider) GetResponseSelectors() []string {
	return b.ResponseSelectors
}

// Implementation of AIClient interface methods would delegate to GuardianClient
func (b *BaseBrowserAIProvider) SetAPIKey(apiKey string, customURL string, customModel string) {
	b.GuardianClient.SetAPIKey(apiKey, customURL, customModel)
}

func (b *BaseBrowserAIProvider) SetTimeout(timeout time.Duration) {
	b.GuardianClient.SetTimeout(timeout)
}

func (b *BaseBrowserAIProvider) CallWithMessages(systemPrompt, userPrompt string) (string, error) {
	return b.GuardianClient.CallWithMessages(systemPrompt, userPrompt)
}

func (b *BaseBrowserAIProvider) CallWithRequest(req *Request) (string, error) {
	return b.GuardianClient.CallWithRequest(req)
}

func (b *BaseBrowserAIProvider) OpenLongLivedBrowser(targetURL string) error {
	return b.GuardianClient.OpenLongLivedBrowser(targetURL)
}
