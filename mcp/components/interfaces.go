package components

import (
	"context"
	"time"
)

// BrowserEngine defines the interface for browser automation
type BrowserEngine interface {
	NavigateToURL(ctx context.Context, url string) error
	GetContext(parentCtx context.Context) (context.Context, context.CancelFunc, error)
}

// DOMInspectorInterface defines the interface for DOM inspection
type DOMInspectorInterface interface {
	CaptureDOMSnapshot(ctx context.Context) error
	GetDSThemeElements(ctx context.Context) ([]map[string]interface{}, error)
	FindInputElement(ctx context.Context, selectors []string) (string, bool)
	FindSubmitButton(ctx context.Context, selectors []string) (string, bool)
	WaitForElement(ctx context.Context, selector string, timeout time.Duration) error
	GetElementText(ctx context.Context, selector string) (string, error)
	GetAllPageElements(ctx context.Context) ([]map[string]interface{}, error)
	GetDOMTree(ctx context.Context) (map[string]interface{}, error)
	ScrollAndCapture(ctx context.Context) error
}
