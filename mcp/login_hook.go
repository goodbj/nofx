package mcp

import (
	"context"
)

// LoginCheckHook 定义登录检测钩子接口
type LoginCheckHook interface {
	// Execute 在浏览器自动化执行前后进行登录状态检查
	Execute(ctx context.Context, gc *GuardianClient, targetURL string) error
}

// DefaultLoginCheckHook 默认的登录检测钩子实现
type DefaultLoginCheckHook struct{}

// NewDefaultLoginCheckHook 创建默认登录检测钩子
func NewDefaultLoginCheckHook() LoginCheckHook {
	return &DefaultLoginCheckHook{}
}

// Execute 实现登录检测钩子逻辑
func (h *DefaultLoginCheckHook) Execute(ctx context.Context, gc *GuardianClient, targetURL string) error {
	gc.logger.Printf("🔐 Executing login check hook for URL: %s", targetURL)

	// 检查登录状态
	isLoggedIn, err := gc.checkLoginStatus(ctx)
	if err != nil {
		gc.logger.Printf("⚠️ Error checking login status in hook: %v", err)
		// 即使检查失败也继续，因为可能是因为页面还没完全加载
		return nil
	}

	if !isLoggedIn {
		gc.logger.Printf("🔒 User is not logged in. Please log in manually to the AI service before continuing.")
		gc.logger.Printf("⏳ Waiting for user to log in... (this is a reminder to ensure you're logged in)")
		// 这里可以添加更复杂的逻辑，比如等待用户确认已登录
	}

	return nil
}

// HookManager 钩子管理器
type HookManager struct {
	loginHooks []LoginCheckHook
}

// NewHookManager 创建新的钩子管理器
func NewHookManager() *HookManager {
	return &HookManager{
		loginHooks: make([]LoginCheckHook, 0),
	}
}

// AddLoginHook 添加登录检测钩子
func (hm *HookManager) AddLoginHook(hook LoginCheckHook) {
	hm.loginHooks = append(hm.loginHooks, hook)
}

// ExecuteLoginHooks 执行所有登录检测钩子
func (hm *HookManager) ExecuteLoginHooks(ctx context.Context, gc *GuardianClient, targetURL string) error {
	for i, hook := range hm.loginHooks {
		gc.logger.Printf("🏃 Executing login hook #%d", i+1)
		if err := hook.Execute(ctx, gc, targetURL); err != nil {
			gc.logger.Printf("⚠️ Login hook #%d failed: %v", i+1, err)
			// 根据需要决定是否继续执行其他钩子或返回错误
			// 这里我们继续执行其他钩子
		}
	}
	return nil
}

// GlobalHookManager 全局钩子管理器
var GlobalHookManager = NewHookManager()

// init 初始化时注册默认钩子
func init() {
	// 清空现有钩子以避免重复注册
	GlobalHookManager = NewHookManager()
	GlobalHookManager.AddLoginHook(NewDefaultLoginCheckHook())
}
