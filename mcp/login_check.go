package mcp

import (
	"context"
	"strings"

	"github.com/chromedp/chromedp"
)

// checkLoginStatus 检查当前页面的登录状态
func (gc *GuardianClient) checkLoginStatus(ctx context.Context) (bool, error) {
	gc.logger.Printf("🔐 Checking login status...")

	// 获取页面标题和URL
	var title, currentURL string
	err := chromedp.Run(ctx,
		chromedp.Title(&title),
		chromedp.Location(&currentURL),
	)
	if err != nil {
		gc.logger.Printf("⚠️ Failed to get page info: %v", err)
		return false, err
	}

	gc.logger.Printf("📋 Page title: %s", title)
	gc.logger.Printf("🔗 Current URL: %s", currentURL)

	// 检查标题或URL中是否包含登录相关关键词
	lowerTitle := strings.ToLower(title)
	lowerURL := strings.ToLower(currentURL)

	loginIndicators := []string{"login", "sign in", "signin", "log in", "auth", "authentication", "account", "access", "welcome", "sign_in", "sign-in"}
	for _, indicator := range loginIndicators {
		if strings.Contains(lowerTitle, indicator) || strings.Contains(lowerURL, indicator) {
			gc.logger.Printf("🔒 Login-related page detected based on title/URL: %s", indicator)
			return false, nil // 表示未登录
		}
	}

	// 特别针对DeepSeek的登录页面进行检测
	if strings.Contains(lowerURL, "deepseek.com/sign_in") || strings.Contains(lowerURL, "deepseek.com/login") {
		gc.logger.Printf("🔒 DeepSeek login page detected: %s", currentURL)
		return false, nil // 表示未登录
	}

	// 检查页面是否包含特定的登录元素
	var hasLoginElements bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(
			`(function() {
				// 检查是否存在登录相关元素
				const loginSelectors = [
					'input[type="email"]',
					'input[type="password"]',
					'input[name="email"]',
					'input[name="password"]',
					'input[id*="email" i]',
					'input[id*="password" i]',
					'form[action*="login" i]',
					'form[action*="signin" i]',
					'div.login',
					'div.signin',
					'.login-form',
					'.signin-form',
					'[data-testid*="login" i]',
					'[data-testid*="signin" i]',
					'button.login',
					'button.signin',
					'#login',
					'#signin',
					'a[href*="login" i]',
					'a[href*="signin" i]',
					'div[aria-label*="login" i]',
					'div[aria-label*="signin" i]',
					// DeepSeek特定的登录元素
					'input[placeholder*="Email" i]',
					'input[placeholder*="Password" i]',
					'button[type="submit"][data-testid*="sign" i]',
					'div[data-testid*="sign-in" i]',
					'div[data-testid*="login" i]',
					'form[data-testid*="sign-in" i]',
					'form[data-testid*="login" i]',
				];

				for (const selector of loginSelectors) {
					const element = document.querySelector(selector);
					if (element) {
						console.log('Found login element:', selector);
						return true; // 找到登录相关元素
					}
				}

				// 检查页面上是否有明显的登录提示文本
				const pageText = document.body.innerText.toLowerCase();
				const loginTextIndicators = ['please sign in', 'please login', 'log in to continue', 'sign in to continue', 'sign in to continue', 'to continue, please sign in'];
				for (const text of loginTextIndicators) {
					if (pageText.includes(text)) {
						console.log('Found login text:', text);
						return true; // 找到登录提示文本
					}
				}

				// 检查是否存在登录页面的特定文本
				const lowerText = document.body.innerText.toLowerCase();
				if (lowerText.includes('sign in') && (lowerText.includes('continue') || lowerText.includes('access') || lowerText.includes('proceed'))) {
					return true; // 表示需要登录才能继续
				}

				return false; // 没有找到登录相关元素
			})()`, &hasLoginElements),
	)

	if err != nil {
		gc.logger.Printf("⚠️ Error checking for login elements: %v", err)
		return false, err
	}

	if hasLoginElements {
		gc.logger.Printf("🔑 Login elements detected on page")
		return false, nil // 存在登录元素，表示未登录
	}

	// 检查是否包含AI聊天界面的典型元素（表示已登录）
	var hasChatElements bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(
			`(function() {
				// 检查是否存在AI聊天界面的典型元素
				const chatSelectors = [
					'textarea[data-testid*="chat" i]',
					'textarea[placeholder*="message" i]',
					'textarea[placeholder*="prompt" i]',
					'div.chat',
					'div.message',
					'.ds-message', // DeepSeek特有
					'.chat-interface',
					'.conversation',
					'.thread',
					'[data-testid*="input" i]',
					'button.send',
					'button[data-testid*="send" i]',
					// DeepSeek特定的已登录元素
					'[data-testid*="user" i]',
					'[data-testid*="avatar" i]',
					'[aria-label*="user" i]',
					'div[aria-label*="account" i]',
					'button[aria-label*="menu" i]',
					'div[class*="user-profile" i]',
					'img[src*="/avatar" i]',
					'img[alt*="user" i]',
				];

				for (const selector of chatSelectors) {
					const element = document.querySelector(selector);
					if (element) {
						console.log('Found chat element:', selector);
						return true; // 找到聊天相关元素
					}
				}

				// 检查是否存在用户菜单或头像（表示已登录）
				const hasUserMenu = document.querySelector('[data-testid*="user" i], [aria-label*="user" i], [data-testid*="avatar" i]');
				if (hasUserMenu) {
					return true; // 找到用户相关元素，表示已登录
				}

				return false; // 没有找到聊天相关元素
			})()`, &hasChatElements),
	)

	if err != nil {
		gc.logger.Printf("⚠️ Error checking for chat elements: %v", err)
		return false, err
	}

	if hasChatElements {
		gc.logger.Printf("💬 Chat elements detected, assuming logged in")
		return true, nil // 存在聊天元素，表示已登录
	}

	// 如果既没有登录元素也没有聊天元素，返回不确定状态
	gc.logger.Printf("❓ Unable to determine login status, assuming not logged in")
	return false, nil
}
