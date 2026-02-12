package mcp

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// AutoLoginProcess 鑷姩寮瑰嚭鐧诲綍绐楀彛骞剁瓑寰呯敤鎴峰畬鎴愮櫥褰?
func (gc *GuardianClient) AutoLoginProcess() bool {
	gc.logger.Println("馃殌 Starting automatic login process...")

	// 浣跨敤闀跨敓鍛藉懆鏈熸祻瑙堝櫒绐楀彛杩涜鐧诲綍
	targetURL := gc.ProviderConfig.BaseURL
	if targetURL == "" {
		targetURL = DefaultGuardianBaseURL
	}

	// 鎵撳紑闀挎椂闂翠繚鎸佺殑娴忚鍣ㄧ獥鍙ｇ敤浜庣櫥褰?
	loginTimeout := 300 // 5鍒嗛挓鐧诲綍瓒呮椂
	result, err := gc.performLoginBrowserAutomation(targetURL, loginTimeout)
	if err != nil {
		gc.logger.Printf("鉂?Login browser automation failed: %v", err)
		return false
	}

	gc.logger.Printf("鉁?Login browser session completed: %s", result)

	// 楠岃瘉鐧诲綍鐘舵€?
	isLoggedIn, err := gc.CheckLoginStatus(targetURL)
	if err != nil {
		gc.logger.Printf("鈿狅笍 Error verifying login status after login attempt: %v", err)
		return false
	}

	if isLoggedIn {
		gc.logger.Println("鉁?Login verification successful")
		return true
	} else {
		gc.logger.Println("鉂?Login verification failed")
		return false
	}
}

// performLoginBrowserAutomation 涓撻棬鐢ㄤ簬鐧诲綍鐨勬祻瑙堝櫒鑷姩鍖栨祦绋?
func (gc *GuardianClient) performLoginBrowserAutomation(targetURL string, timeoutSeconds int) (string, error) {
	gc.logger.Printf("馃攼 Opening login browser window for: %s", targetURL)

	// 璁剧疆Chrome閫夐」
	uniqueID := gc.TraderID
	if uniqueID == "" {
		timestamp := time.Now().UnixNano()
		uniqueID = fmt.Sprintf("%d_%p", timestamp, gc)
	}

	userDir := fmt.Sprintf("%s_%s", GuardianBrowserDataDir, uniqueID)
	gc.logger.Printf("馃搧 Using user data directory: %s", userDir)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-web-security", false),
		chromedp.Flag("disable-features", "VizDisplayCompositor"),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("blink-settings", "imagesEnabled=true"),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("exclude-switches", "enable-automation"),
		chromedp.Flag("disable-extensions", false),
		chromedp.Flag("disable-plugins-discovery", false),
		chromedp.Flag("incognito", false),
		chromedp.Flag("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("renderer-process-limit", "1"),
		chromedp.Flag("max_old_space_size", "4096"),
		chromedp.Flag("no-first-run", "true"),
		chromedp.Flag("no-default-browser-check", "true"),
		chromedp.Flag("disable-backgrounding-occluded-windows", "false"),
		chromedp.Flag("disable-renderer-backgrounding", "true"),
		chromedp.Flag("disable-background-timer-throttling", "true"),
		chromedp.Flag("disable-background-networking", "false"),
		chromedp.Flag("window-size", "1000,850"),
		chromedp.Flag("user-data-dir", userDir),
		chromedp.Flag("profile-directory", "Default"),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	ctx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	// 璁剧疆鐧诲綍瓒呮椂
	timeout := time.Duration(timeoutSeconds) * time.Second
	ctx, timeoutCancel := context.WithTimeout(ctx, timeout)
	defer timeoutCancel()

	gc.logger.Printf("鈴?Login browser started with timeout: %v", timeout)

	// 瀵艰埅鍒扮櫥褰曢〉闈?
	if !strings.HasPrefix(targetURL, "http") {
		targetURL = "https://" + targetURL
	}

	if err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(targetURL),
		chromedp.Sleep(3*time.Second), // 缁欓〉闈㈡洿澶氬姞杞芥椂闂?
	); err != nil {
		return "", fmt.Errorf("failed to navigate to login page: %w", err)
	}

	gc.logger.Printf("鉁?Successfully navigated to login page: %s", targetURL)

	// 绛夊緟鐢ㄦ埛瀹屾垚鐧诲綍
	gc.logger.Println("鈴?Waiting for user to complete login... (browser will close automatically after login or timeout)")

	// 瀹氭湡妫€鏌ョ櫥褰曠姸鎬?
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			isLoggedIn, err := gc.checkLoginStatus(ctx)
			if err != nil {
				gc.logger.Printf("鈿狅笍 Error checking login status: %v", err)
				continue
			}

			if isLoggedIn {
				gc.logger.Println("鉁?Login detected, closing browser...")
				return "Login successful", nil
			}

			gc.logger.Println("鈴?Still waiting for login completion...")

		case <-ctx.Done():
			gc.logger.Println("鈴?Login timeout reached")
			return "Login timeout", nil
		}
	}
}

// call implements the actual AI call using browser automation
func (gc *GuardianClient) call(systemPrompt, userPrompt string) (string, error) {
	gc.logger.Println("馃 Guardian Browser Automation: Starting browser automation for AI processing")

	// 鍚堝苟绯荤粺鎻愮ず鍜岀敤鎴锋彁绀?
	var combinedPrompt string
	if systemPrompt != "" && userPrompt != "" {
		combinedPrompt = fmt.Sprintf("System Prompt:\n%s\n\nUser Prompt:\n%s", systemPrompt, userPrompt)
	} else if userPrompt != "" {
		combinedPrompt = userPrompt // 濡傛灉鍙湁鐢ㄦ埛鎻愮ず锛屽垯鐩存帴浣跨敤鐢ㄦ埛鎻愮ず
	} else {
		combinedPrompt = systemPrompt // 鍚﹀垯浣跨敤绯荤粺鎻愮ず
	}

	// 娉ㄦ剰锛氫笉鍐嶈嚜鍔ㄦ鏌ョ櫥褰曠姸鎬侊紝鍥犱负宸叉湁涓撻棬鐨勯娆＄櫥褰曞伐鍏?
	// 濡傞渶鐧诲綍锛岃浣跨敤鍓嶇鐨勪氦鏄撳憳棣栨鐧诲綍宸ュ叿椤甸潰杩涜璁剧疆

	// 鍚姩娴忚鍣ㄨ嚜鍔ㄥ寲娴佺▼
	result, err := gc.performBrowserAutomation(combinedPrompt)
	if err != nil {
		gc.logger.Printf("鉂?Guardian Browser Automation failed: %v", err)
		return "", fmt.Errorf("guardian browser automation failed: %w", err)
	}

	gc.logger.Println("鉁?Guardian Browser Automation completed successfully")
	return result, nil
}

// performBrowserAutomation handles the browser automation process
func (gc *GuardianClient) performBrowserAutomation(prompt string) (string, error) {
	// 鍒濆鍖栧搷搴斿彉閲?
	var response string

	// 璁剧疆Chrome閫夐」 - 姣忎釜浜ゆ槗鍛樹娇鐢ㄧ嫭绔嬬殑鐢ㄦ埛鏁版嵁鐩綍锛岀‘淇濇瘡涓氦鏄撳憳鏈夌嫭绔嬬殑鐧诲綍鐘舵€?
	uniqueID := gc.TraderID
	if uniqueID == "" {
		// 濡傛灉娌℃湁浜ゆ槗鍛業D锛屽垯浣跨敤鏃堕棿鎴冲拰瀹炰緥鍦板潃浣滀负鍚庡
		timestamp := time.Now().UnixNano()
		uniqueID = fmt.Sprintf("%d_%p", timestamp, gc)
	}

	// 璁板綍璋冭瘯淇℃伅
	gc.logger.Printf("馃毃 [GUARDIAN BROWSER DEBUG] Creating browser with TraderID: '%s', uniqueID: '%s'", gc.TraderID, uniqueID)
	userDir := fmt.Sprintf("%s_%s", GuardianBrowserDataDir, uniqueID)
	gc.logger.Printf("馃毃 [GUARDIAN BROWSER DEBUG] User data directory: %s", userDir)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),             // 闈炴棤澶存ā寮忎互渚胯瀵?
		chromedp.Flag("disable-web-security", false), // 鍚敤缃戠粶瀹夊叏浠ユ敮鎸佹甯哥綉绔欏姛鑳?
		chromedp.Flag("disable-features", "VizDisplayCompositor,TranslateUI"),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-gpu", false), // 鍚敤GPU鍔犻€?
		chromedp.Flag("disable-software-rasterizer", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", false), // 纭繚绐楀彛鍙
		chromedp.Flag("disable-background-timer-throttling", true),     // 闃叉瀹氭椂鍣ㄨ妭娴?
		chromedp.Flag("disable-background-networking", false),          // 鍏佽鍚庡彴缃戠粶娲诲姩
		chromedp.Flag("disable-renderer-backgrounding", true),          // 闃叉鍚庡彴娓叉煋
		chromedp.Flag("blink-settings", "imagesEnabled=true"),          // 鍚敤鍥剧墖鍔犺浇浠ユ敮鎸侀獙璇佺爜绛夊姛鑳?
		chromedp.Flag("enable-automation", false),                      // 闃叉琚綉绔欐娴嬩负鑷姩鍖?
		chromedp.Flag("exclude-switches", "enable-automation"),         // 鎺掗櫎鑷姩鍖栧紑鍏?
		chromedp.Flag("disable-extensions", false),                     // 鍚敤鎵╁睍
		chromedp.Flag("disable-plugins-discovery", false),              // 鍚敤鎻掍欢鍙戠幇
		chromedp.Flag("disable-plugins", true),
		chromedp.Flag("disable-image-animation-resampling", true),
		chromedp.Flag("disable-session-crashed-bubble", true),
		chromedp.Flag("disable-breakpad", true),
		chromedp.Flag("disable-field-trial-config", true),
		chromedp.Flag("disable-ipc-flooding-protection", false),
		chromedp.Flag("incognito", false), // 涓嶄娇鐢ㄩ殣韬ā寮?
		chromedp.Flag("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"), // 璁剧疆姝ｅ父鐢ㄦ埛浠ｇ悊
		chromedp.Flag("disable-blink-features", "AutomationControlled"),                                                                                // 绂佺敤鑷姩鍖栨帶鍒剁壒寰?
		chromedp.Flag("renderer-process-limit", "1"),
		chromedp.Flag("max-old-space-size", "4096"),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("window-size", "1000,850"),       // 璁剧疆娴忚鍣ㄧ獥鍙ｅ昂瀵镐负1000x850
		chromedp.Flag("user-data-dir", userDir),        // 姣忎釜瀹炰緥浣跨敤鐙珛鐨勭敤鎴锋暟鎹洰褰?
		chromedp.Flag("profile-directory", "Default"),  // 浣跨敤榛樿閰嶇疆鏂囦欢
		chromedp.Flag("remote-debugging-port", "9222"), // 鍚敤杩滅▼璋冭瘯绔彛
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel() // 鍙栨秷娴忚鍣ㄥ垎閰嶅櫒

	// 鍒涘缓chrome瀹炰緥涓婁笅鏂?
	ctx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel() // 鍙栨秷娴忚鍣ㄥ疄渚?

	// 璁剧疆瓒呮椂 - 浣跨敤寮哄埗鍏抽棴鏃堕棿浣滀负涓昏秴鏃讹紝纭繚娴忚鍣ㄤ笉浼氳繍琛岃秴杩?鍒?0绉?
	forceCloseTimeoutSeconds := getEnvInt("GUARDIAN_FORCE_CLOSE_TIMEOUT_SECONDS", 290)
	timeout := time.Duration(forceCloseTimeoutSeconds) * time.Second
	ctx, timeoutCancel := context.WithTimeout(ctx, timeout)
	defer timeoutCancel() // 鍙栨秷甯﹁秴鏃剁殑涓婁笅鏂?

	// 璁板綍瀹為檯浣跨敤鐨勮秴鏃舵椂闂?
	gc.logger.Printf("鈴?Browser automation started with force close timeout: %v", timeout)

	// 璁板綍寮€濮嬫椂闂?
	startTime := time.Now()
	gc.logger.Printf("鈴?Browser automation started, timeout: %v", timeout)

	// 鍚姩寮哄埗鍏抽棴瀹氭椂鍣細4鍒?0绉掑悗寮哄埗鍏抽棴娴忚鍣紝闃叉褰卞搷涓嬩竴杞煡璇?
	forceCloseTimer := time.AfterFunc(time.Duration(forceCloseTimeoutSeconds)*time.Second, func() {
		gc.logger.Printf("馃毃 Force closing browser after %d seconds (safety timeout)", forceCloseTimeoutSeconds)
		// 娉ㄦ剰锛氳繖閲屾垜浠笉鑳界洿鎺ヨ皟鐢╟ancel鍑芥暟锛屽洜涓烘垜浠渶瑕侀€氳繃鍏朵粬鏂瑰紡閫氱煡鍑芥暟缁堟
		// 鎴戜滑灏嗗湪鍚庣画妫€鏌ヤ腑妫€娴嬭繖涓秴鏃?
	})
	defer forceCloseTimer.Stop() // 纭繚鍦ㄥ嚱鏁版甯哥粨鏉熸椂鍋滄瀹氭椂鍣?

	// 璁块棶鐩爣URL
	targetURL := gc.ProviderConfig.BaseURL
	if targetURL == "" {
		// 濡傛灉BaseURL涓虹┖锛屼娇鐢ㄩ粯璁RL
		targetURL = DefaultGuardianBaseURL
		gc.logger.Printf("鈿狅笍 BaseURL is empty, using default: %s", targetURL)
	}
	if !strings.HasPrefix(targetURL, "http") {
		targetURL = "https://" + targetURL
	}

	// 楠岃瘉URL鏍煎紡
	if _, err := url.Parse(targetURL); err != nil {
		return "", fmt.Errorf("invalid target URL '%s': %w", targetURL, err)
	}

	gc.logger.Printf("馃寪 Navigating to URL: %s", targetURL)

	// 璁剧疆璇锋眰鎷︽埅浠ュ鐞嗘綔鍦ㄧ殑CORS鎴栧叾浠栫綉缁滈棶棰?
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			gc.logger.Printf("馃摛 Request: %s", ev.Request.URL)
		case *network.EventResponseReceived:
			gc.logger.Printf("馃摜 Response: %d", ev.Response.Status)
		}
	})

	// 瀵艰埅鍒扮洰鏍囬〉闈?
	if err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(targetURL),
		chromedp.Sleep(2*time.Second), // 绛夊緟椤甸潰鍔犺浇
	); err != nil {
		return "", fmt.Errorf("failed to navigate to target URL: %w", err)
	}

	gc.logger.Printf("鉁?Successfully navigated to target URL")

	// 鏍规嵁涓嶅悓鐨凙I鏈嶅姟浣跨敤鐩稿簲鐨勯€夋嫨鍣?
	var inputSelectors []string
	var submitSelectors []string
	var responseSelectors []string

	serviceType := strings.ToLower(gc.ProviderConfig.Provider)

	switch {
	case strings.Contains(serviceType, "deepseek"):
	case strings.Contains(serviceType, "chat"):
	case targetURL == "https://chat.deepseek.com" || strings.Contains(targetURL, "deepseek"):
		// 濡傛灉鏄疍eepSeek鎴栧寘鍚玞hat鐨勬湇鍔★紝鎴栫洰鏍嘦RL鍖呭惈deepseek
		inputSelectors = []string{
			"textarea[placeholder='缁?DeepSeek 鍙戦€佹秷鎭?']",
			"textarea._27c9245.ds-scroll-area.d96f2d2a",
			"textarea._27c9245.ds-scroll-area.d96f2d2a[placeholder='缁?DeepSeek 鍙戦€佹秷鎭?']",
			"textarea[placeholder='Send a message']",
			"div.public-DraftEditor-content",
			"textarea[aria-label='Chat text input']",
			"#chat-input",
			"[data-testid='chat-input']",
		}
		submitSelectors = []string{
			"div._7436101.ds-icon-button.ds-icon-button--l.ds-icon-button--sizing-container[role='button'][aria-disabled='false']", // DeepSeek鐗规湁鎸夐挳鏍峰紡 - 鍚敤鐘舵€?
			"div._7436101.ds-icon-button[role='button']:not([aria-disabled='true'])",                                               // DeepSeek鐗规湁鎸夐挳鏍峰紡 - 鍚敤鐘舵€?
			"div.ds-icon-button[role='button']:not([aria-disabled='true'])",                                                        // 鍚敤鐘舵€佺殑鎸夐挳
			"button._3quh._30yy._2t_",
			"button[type='submit']",
			"button[data-testid='send-button']",
			"button.send-button",
			"button[aria-label='Send']",
			".send-btn",
			"button._2t_",
			"button[type='button'][aria-label='Send Message']",
			"button:enabled:not([disabled])",
		}
		responseSelectors = []string{
			"div.ds-flex._0a3d93b", // 涓昏鐨凙I杈撳嚭瀹屾垚鏍囪瘑瀹瑰櫒
			"div.text-message span",
			"div[data-testid='response-container']",
			"div.message-response",
			".chat-message-content",
			"div.markdown-body",
			"pre",
			"code",
		}
	default:
		// 榛樿浣跨敤閫氱敤閫夋嫨鍣?
		inputSelectors = []string{
			"textarea[placeholder*='message'], textarea[placeholder*='Message']",
			"textarea[placeholder*='input'], textarea[placeholder*='Input']",
			"textarea[aria-label*='input'], textarea[aria-label*='text']",
			"textarea[role='textbox']",
			"input[type='text']",
			"div[contenteditable='true']",
		}
		submitSelectors = []string{
			"button[type='submit']",
			"button[aria-label*='send'], button[title*='send']",
			"button[data-testid*='send']",
			".send-button, #send-button",
		}
		responseSelectors = []string{
			"[data-testid*='response'], [data-testid*='answer']",
			".response, .answer, .result",
			"div[class*='message']",
		}
	case strings.Contains(serviceType, "chatgpt"):
		inputSelectors = []string{
			"textarea[placeholder='Send a message']",
			"textarea[id='prompt-textarea']",
			"textarea[aria-label='Chat text input']",
			"#chat-input",
			"[data-testid='chat-input']",
		}
		submitSelectors = []string{
			"button[data-testid='send-button']",
			"button.send-button",
			"button[aria-label='Send']",
			".send-btn",
		}
		responseSelectors = []string{
			"div.ds-flex._0a3d93b", // 涓昏鐨凙I杈撳嚭瀹屾垚鏍囪瘑瀹瑰櫒
			"div.text-message span",
			"div[data-testid='response-container']",
			"div.message-response",
			".chat-message-content",
			"div.markdown-body",
		}
	case strings.Contains(serviceType, "qwen") || strings.Contains(serviceType, "閫氫箟鍗冮棶"):
		inputSelectors = []string{
			"textarea[placeholder*='璇疯緭鍏?], textarea[placeholder*='杈撳叆']",
			"textarea[placeholder*='message'], textarea[placeholder*='Message']",
			"textarea[aria-label*='input'], textarea[aria-label*='text']",
			"div[contenteditable='true']",
			"textarea[data-testid='chat-input']",
			"textarea#text-input",
		}
		submitSelectors = []string{
			"button[data-testid='send-button'], button.send-btn",
			"button[aria-label*='鍙戦€?], button[aria-label*='Send']",
			"button[type='submit']",
			".submit-btn, #submit-btn",
			"button:has(svg[class*='send']), button:has(i[class*='send'])",
		}
		responseSelectors = []string{
			"div[data-testid='assistant-response'], div[data-role='assistant']",
			"div.qwen-response, div.chat-message.assistant",
			"div.markdown-body, div.response-content",
			"pre, code",
			"div[class*='message']:not([data-role='user'])",
		}
	case strings.Contains(serviceType, "other"):
		// 棰勭暀鍏朵粬鏈嶅姟绫诲瀷
		inputSelectors = []string{
			"textarea[placeholder*='message'], textarea[placeholder*='Message']",
			"textarea[placeholder*='input'], textarea[placeholder*='Input']",
			"textarea[aria-label*='input'], textarea[aria-label*='text']",
			"textarea[role='textbox']",
			"input[type='text']",
			"div[contenteditable='true']",
		}
		submitSelectors = []string{
			"button[type='submit']",
			"button[aria-label*='send'], button[title*='send']",
			"button[data-testid*='send']",
			".send-button, #send-button",
		}
		responseSelectors = []string{
			"[data-testid*='response'], [data-testid*='answer']",
			".response, .answer, .result",
			"div[class*='message']",
		}
	}

	gc.logger.Printf("馃攳 Attempting to find and fill input field with selectors: %v", inputSelectors)

	// 灏濊瘯濉厖杈撳叆妗?
	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			gc.logger.Println("馃摑 Starting to find and fill input field...")
			gc.logger.Printf("馃攳 Trying input selectors: %v", inputSelectors)

			inputFound := false // 鏍囧織鍙橀噺锛岃窡韪槸鍚︽垚鍔熸壘鍒板苟杈撳叆

			for _, selector := range inputSelectors {
				gc.logger.Printf("馃攳 Attempting to find input with selector: %s", selector)
				// 棣栧厛绛夊緟鍏冪礌鍙
				err := chromedp.WaitVisible(selector).Do(ctx)
				if err == nil {
					gc.logger.Printf("鉁?Found input element with selector: %s", selector)

					// 娓呯┖杈撳叆妗嗗苟杈撳叆鏂板唴瀹?
					gc.logger.Printf("馃棏锔?Clearing input field: %s", selector)
					err = chromedp.Clear(selector).Do(ctx)
					if err != nil {
						gc.logger.Printf("鈿狅笍 Could not clear input field %s, proceeding anyway: %v", selector, err)
					}

					// 浣跨敤chromedp.Focus鍜孲etValue鏂规硶蹇€熷～鍏呮枃鏈苟婵€娲绘帶浠?
					// 1. 閫変腑鎺т欢骞惰仛鐒?
					// 2. 浣跨敤SetValue璁剧疆鍊?
					// 3. 瑙﹀彂蹇呰鐨勪簨浠朵互婵€娲绘彁浜ゆ寜閽?
					textLength := len(strings.ReplaceAll(prompt, "\n", " ")) // 浣跨敤鍘熷闀垮害璁＄畻鐢ㄤ簬鏃ュ織
					gc.logger.Printf("馃殌 Fast input mode: Setting %d characters using chromedp.Focus + SetValue...", textLength)

					// 棣栧厛浣跨敤chromedp.Focus鏂规硶鑱氱劍鍏冪礌
					err = chromedp.Focus(selector).Do(ctx)
					if err != nil {
						gc.logger.Printf("鉂?Focus activation failed with chromedp.Focus: %v", err)
						// 鐩存帴璺宠繃姝ら€夋嫨鍣?
						gc.logger.Printf("鈴笍 Skipping this selector due to focus activation failure")
						continue // 缁х画灏濊瘯涓嬩竴涓€夋嫨鍣?
					} else {
						gc.logger.Printf("鉁?Control focused successfully with chromedp.Focus, now using optimized chunked input approach...")

						// 棣栧厛娓呯┖杈撳叆妗?
						err = chromedp.Clear(selector).Do(ctx)
						if err != nil {
							gc.logger.Printf("鈿狅笍 Could not clear input field, proceeding anyway: %v", err)
						}

						// 棣栧厛鐐瑰嚮杈撳叆妗嗕互鑾峰緱鐒︾偣
						err = chromedp.Click(selector).Do(ctx)
						// 鐭殏绛夊緟纭繚鐒︾偣璁剧疆瀹屾垚
						time.Sleep(100 * time.Millisecond)

						// 棣栧厛鑱氱劍杈撳叆妗?
						err = chromedp.Focus(selector).Do(ctx)
						if err != nil {
							gc.logger.Printf("鈿狅笍 Focus failed: %v", err)
						}

						// 鍏堣缃畬鏁村唴瀹?
						err = chromedp.SetValue(selector, prompt).Do(ctx)
						if err != nil {
							gc.logger.Printf("鉂?SetValue failed: %v", err)
							// 鐩存帴璺宠繃姝ら€夋嫨鍣?
							gc.logger.Printf("鈴笍 Skipping this selector due to SetValue failure")
							continue // 缁х画灏濊瘯涓嬩竴涓€夋嫨鍣?
						}

						gc.logger.Printf("鉁?Full content set with %d characters using SetValue approach", len(prompt))

						// 鏍规嵁鎮ㄧ殑鍙戠幇锛岀幇鍦ㄦā鎷熻緭鍏ヤ竴涓瓧绗︽潵婵€娲绘寜閽?
						// 杩欎釜鍏抽敭鎿嶄綔浼氳Е鍙戝墠绔鏋剁殑鐘舵€佹洿鏂?
						err = chromedp.SendKeys(selector, " ").Do(ctx) // 鍙戦€佷竴涓┖鏍煎瓧绗?
						if err != nil {
							gc.logger.Printf("鈿狅笍 SendKeys activation character failed: %v", err)
							// 鍗充娇鍙戦€佹縺娲诲瓧绗﹀け璐ワ紝涔熻缁х画锛屽洜涓轰富瑕佸唴瀹瑰凡缁忚缃簡
						} else {
							gc.logger.Printf("鉁?Activation character sent to trigger button state")
						}

						// 鐒跺悗鍒犻櫎杩欎釜棰濆鐨勭┖鏍煎瓧绗︼紝鎭㈠鍘熷鍐呭
						// 閫氳繃JavaScript妯℃嫙Backspace閿?
						backspaceScript := fmt.Sprintf(`
								(function() {
									var element = document.querySelector('%s');
									if (element) {
										// 鑾峰彇褰撳墠鍊煎苟鍘绘帀鏈€鍚庝竴涓瓧绗︼紙绌烘牸锛?
										var currentValue = element.value;
										if (currentValue.length > 0) {
											element.value = currentValue.substring(0, currentValue.length - 1);
				
											// 瑙﹀彂Backspace鐩稿叧鐨勪簨浠?
											var keydownEvent = new KeyboardEvent('keydown', {
												bubbles: true,
												cancelable: true,
												key: 'Backspace',
												code: 'Backspace',
												keyCode: 8,
												which: 8
											});
											element.dispatchEvent(keydownEvent);
				
											var inputEvent = new InputEvent('input', {
												bubbles: true,
												cancelable: true,
												inputType: 'deleteContentBackward',
												data: null
											});
											element.dispatchEvent(inputEvent);
				
											var keyupEvent = new KeyboardEvent('keyup', {
												bubbles: true,
												cancelable: true,
												key: 'Backspace',
												code: 'Backspace',
												keyCode: 8,
												which: 8
											});
											element.dispatchEvent(keyupEvent);
				
											return true;
										}
									}
									return false;
								})();
							`, selector)

						var backspaceResult bool
						err = chromedp.Evaluate(backspaceScript, &backspaceResult).Do(ctx)
						if err != nil || !backspaceResult {
							gc.logger.Printf("鈿狅笍 JavaScript Backspace simulation failed: %v, success: %v", err, backspaceResult)
						} else {
							gc.logger.Printf("鉁?Removed activation character via JavaScript")
						}

						// 鍐嶆瑙﹀彂浜嬩欢浠ョ‘淇濈姸鎬佹洿鏂?
						ensureStateUpdateScript := fmt.Sprintf(`
								(function() {
									try {
										var element = document.querySelector('%s');
										if (element) {
											var inputEvent = new Event('input', { bubbles: true, cancelable: true });
											element.dispatchEvent(inputEvent);
											var changeEvent = new Event('change', { bubbles: true, cancelable: true });
											element.dispatchEvent(changeEvent);
											return true;
									}
										return false;
									} catch(e) {
											console.error('Error in state update script:', e);
											return false;
									}
									})();
							`, selector)

						var ensureResult bool
						err = chromedp.Evaluate(ensureStateUpdateScript, &ensureResult).Do(ctx)
						if err != nil || !ensureResult {
							gc.logger.Printf("鈿狅笍 State update script failed: %v, success: %v", err, ensureResult)
						} else {
							gc.logger.Printf("鉁?State update script executed successfully")
						}

						// 鐭殏寤惰繜锛岀‘淇濋〉闈㈠搷搴?
						time.Sleep(200 * time.Millisecond)

						// 鏍囪杈撳叆鎴愬姛
						inputFound = true

						// 杈撳叆瀹屾垚锛岃烦鍑洪€夋嫨鍣ㄥ惊鐜紝缁х画鎵ц鎻愪氦鎸夐挳閫昏緫
						break // 璺冲嚭閫夋嫨鍣ㄥ惊鐜紝缁х画鎵ц鎻愪氦鎸夐挳閫昏緫
					}

					// 鐭殏寤惰繜锛岀‘淇濋〉闈㈠搷搴?
					time.Sleep(200 * time.Millisecond)

					// 鏍囪杈撳叆鎴愬姛
					inputFound = true

					// 杈撳叆瀹屾垚锛岃烦鍑洪€夋嫨鍣ㄥ惊鐜紝缁х画鎵ц鎻愪氦鎸夐挳閫昏緫
					break // 璺冲嚭閫夋嫨鍣ㄥ惊鐜紝缁х画鎵ц鎻愪氦鎸夐挳閫昏緫
				} else {
					gc.logger.Printf("鉂?Input selector %s not found or not visible: %v", selector, err)
				}
			}

			// 鍙湁鍦ㄦ墍鏈夐€夋嫨鍣ㄩ兘澶辫触鐨勬儏鍐典笅鎵嶈繑鍥為敊璇?
			if !inputFound {
				gc.logger.Printf("鉂?No input element found with any of the attempted selectors: %v", inputSelectors)
				return fmt.Errorf("no input element found with any of the attempted selectors: %v", inputSelectors)
			}

			// 濡傛灉鎵惧埌浜嗚緭鍏ュ厓绱犲苟鎴愬姛杈撳叆锛屽垯涓嶈繑鍥為敊璇?
			return nil
		}),
	)
	if err != nil {
		return "", fmt.Errorf("failed to fill input field: %w", err)
	}

	gc.logger.Printf("鉁?Successfully filled input field")

	// 鐐瑰嚮鎻愪氦鎸夐挳
	gc.logger.Printf("馃憜 Attempting to click submit button with selectors: %v", submitSelectors)

	// 澶勭悊鎹㈣绗︼紝纭繚闀垮害姣旇緝浣跨敤鐨勬槸瀹為檯杈撳叆鍒版枃鏈鐨勫唴瀹归暱搴?
	safePrompt := strings.ReplaceAll(prompt, "\n", " ")

	// 绛夊緟杈撳叆妗嗗唴瀹归暱搴︿笌澶勭悊鍚庢彁绀鸿瘝闀垮害涓€鑷村悗鐐瑰嚮鎻愪氦鎸夐挳
	gc.logger.Printf("鈴?Waiting for input content to match processed prompt length (%d characters)", len(safePrompt))

	// 浣跨敤鏂扮殑chromedp鎿嶄綔鏉ョ瓑寰呰緭鍏ュ畬鎴愬苟鐐瑰嚮鎻愪氦鎸夐挳
	waitErr := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			maxWait := time.Now().Add(10 * time.Second) // 鏈€澶х瓑寰?0绉?
			for time.Now().Before(maxWait) {
				// 灏濊瘯鎵惧埌杈撳叆妗嗗苟妫€鏌ュ叾鍐呭闀垮害
				var inputLength int
				foundInput := false

				for _, inputSelector := range inputSelectors {
					var inputValue string
					err := chromedp.Value(inputSelector, &inputValue).Do(ctx)
					if err == nil && len(inputValue) > 0 {
						inputLength = len(inputValue)
						foundInput = true
						gc.logger.Printf("馃搳 Input box length check: %d/%d characters", inputLength, len(safePrompt))

						// 濡傛灉杈撳叆妗嗗唴瀹归暱搴︿笌澶勭悊鍚庣殑鎻愮ず璇嶉暱搴︿竴鑷达紝鍙互鐐瑰嚮鎻愪氦鎸夐挳
						if inputLength == len(safePrompt) {
							gc.logger.Printf("鉁?Input length matches processed prompt length (%d), attempting to click submit button", len(safePrompt))

							// 灏濊瘯鐐瑰嚮鎻愪氦鎸夐挳
							for _, submitSelector := range submitSelectors {
								gc.logger.Printf("馃憜 Clicking submit button with selector: %s", submitSelector)
								err := chromedp.Click(submitSelector).Do(ctx)
								if err == nil {
									gc.logger.Printf("鉁?Successfully clicked submit button: %s", submitSelector)
									return nil // 鎴愬姛鐐瑰嚮锛岄€€鍑篈ctionFunc
								} else {
									gc.logger.Printf("鈿狅笍 Failed to click submit button %s: %v, trying next selector", submitSelector, err)
								}
							}
							// 濡傛灉鎵€鏈夋彁浜ゆ寜閽兘鐐瑰嚮澶辫触锛岀户缁瓑寰?
						}
						break
					}
				}

				if !foundInput {
					gc.logger.Println("鈿狅笍 Could not find input box, continuing to wait...")
				}

				time.Sleep(1 * time.Second) // 绛夊緟1绉掑悗鍐嶆妫€鏌?
			}
			return fmt.Errorf("timeout waiting for input length to match processed prompt length")
		}),
	)

	if waitErr != nil {
		gc.logger.Printf("鈿狅笍 Wait for input completion failed or timeout: %v", waitErr)
		// 鍗充娇绛夊緟澶辫触锛屾垜浠篃灏濊瘯鐐瑰嚮鎻愪氦鎸夐挳锛屼互闃蹭竾涓€
		gc.logger.Println("馃攧 Proceeding to try clicking submit button anyway...")
		// 灏濊瘯鐐瑰嚮鎻愪氦鎸夐挳
		clickErr := chromedp.Run(ctx,
			chromedp.ActionFunc(func(ctx context.Context) error {
				for _, submitSelector := range submitSelectors {
					gc.logger.Printf("馃憜 Clicking submit button with selector: %s", submitSelector)
					err := chromedp.Click(submitSelector).Do(ctx)
					if err == nil {
						gc.logger.Printf("鉁?Successfully clicked submit button: %s", submitSelector)
						return nil
					} else {
						gc.logger.Printf("鈿狅笍 Failed to click submit button %s: %v, trying next selector", submitSelector, err)
					}
				}
				return fmt.Errorf("failed to click any submit button")
			}),
		)
		if clickErr != nil {
			gc.logger.Printf("鈿狅笍 All submit button attempts failed: %v", clickErr)
		}
	}

	// 鎻愪氦鎸夐挳鐐瑰嚮宸插湪涓婇潰鐨勯€昏緫涓鐞?
	gc.logger.Println("鉁?Submit button processing completed")

	// 鍦ㄧ瓑寰匒I鍝嶅簲涔嬪墠锛屽厛杈撳嚭褰撳墠椤甸潰鐨凞OM缁撴瀯鐢ㄤ簬璋冭瘯
	gc.logger.Println("馃攳 Setting up DOM snapshot for debugging...")
	// 绛夊緟1绉掕椤甸潰鏇存柊鍚庡啀杈撳嚭DOM
	time.Sleep(1 * time.Second)

	// 浣跨敤鏂扮殑绛夊緟鍜岄獙璇佹満鍒舵娴婣I杈撳嚭瀹屾垚
	gc.logger.Println("鈴?Waiting for AI to complete output using optimized detection mechanism...")
	// 浣跨敤杈冪煭鐨勮秴鏃舵椂闂达紝纭繚涓嶄細瓒呰繃寮哄埗鍏抽棴鏃堕棿
	validationCtx, validationCancel := context.WithTimeout(ctx, time.Duration(forceCloseTimeoutSeconds)*time.Second)
	extractedContent, err := gc.waitForAndValidateOutput(validationCtx)
	validationCancel() // 纭繚鍙栨秷涓婁笅鏂囦互閲婃斁璧勬簮
	if err != nil {
		gc.logger.Printf("鈿狅笍 Error waiting for AI completion: %v", err)
		// 缁х画鎵ц锛屽嵆浣跨瓑寰呭嚭鐜伴棶棰?
	} else {
		gc.logger.Println("鉁?AI output completed successfully detected")
		// 濡傛灉鎴愬姛鎻愬彇浜嗗唴瀹癸紝浣跨敤鎻愬彇鐨勫唴瀹逛綔涓哄搷搴?
		if extractedContent != "" {
			response = extractedContent
			gc.logger.Printf("鉁?Using extracted content as response, length: %d", len(response))
		}

		// 鍦ˋI杈撳嚭瀹屾垚鍚庡叧闂祻瑙堝櫒
		gc.logger.Println("鉁?Closing browser after AI output completion")

		// 鏄惧紡鍙栨秷娴忚鍣ㄤ笂涓嬫枃浠ョ‘淇濇祻瑙堝櫒琚叧闂?
		// browserCancel搴旇閫氳繃defer璇彞鑷姩璋冪敤锛屼絾涓轰簡纭繚锛岃繖閲岃褰?
		gc.logger.Println("馃挕 Browser context will be cancelled via defer statement when function exits")
	}

	// 鑾峰彇瀹屾暣鐨勯〉闈TML鍐呭浣嗕笉淇濆瓨鍒版枃浠?
	var pageHTML string
	err = chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			// 鑾峰彇瀹屾暣鐨勯〉闈TML鍐呭
			err := chromedp.OuterHTML("html", &pageHTML).Do(ctx)
			if err != nil {
				return err
			}
			// 浠呭湪鍐呭瓨涓娇鐢℉TML鍐呭锛屼笉淇濆瓨鍒版枃浠?
			gc.logger.Printf("馃搫 Full DOM snapshot acquired, size: %d bytes", len(pageHTML))
			return nil
		}),
	)
	if err != nil {
		gc.logger.Printf("鈿狅笍 Error getting page HTML: %v", err)
	}

	// 鍚敤锛氬寮虹殑DOM鎶撳彇鍜岃皟璇曞姛鑳?
	// 鑾峰彇椤甸潰涓婃墍鏈塪s-theme鍏冪礌鐨勮缁嗕俊鎭紙涓嶄繚瀛樺埌鏂囦欢锛?
	var themeElements []map[string]interface{}
	err = chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			err := chromedp.Evaluate(`(() => {
				const elements = document.querySelectorAll('div.ds-theme');
				return Array.from(elements).map(el => ({
					tagName: el.tagName,
					className: el.className,
					style: el.style.cssText,
					innerHTML: el.innerHTML ? el.innerHTML.substring(0, 500) : '',
					attributes: Array.prototype.reduce.call(el.attributes, function(acc, attr) {
						acc[attr.name] = attr.value;
						return acc;
					}, {})
				}));
			})()`, &themeElements).Do(ctx)
			if err != nil {
				return err
			}

			// 浠呭湪鍐呭瓨涓娇鐢ㄤ富棰樺厓绱犱俊鎭紝涓嶄繚瀛樺埌鏂囦欢
			gc.logger.Printf("馃攳 ds-theme Elements Detail acquired, count: %d", len(themeElements))
			return nil
		}),
	)
	if err != nil {
		gc.logger.Printf("鈿狅笍 Error getting ds-theme elements detail: %v", err)
	}

	// 涔熻幏鍙栭〉闈笂鎵€鏈夊彲鑳界殑AI瀹屾垚鏍囧織鍏冪礌锛堜笉淇濆瓨鍒版枃浠讹級
	var aiCompletionElements []map[string]interface{}
	err = chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			err := chromedp.Evaluate(`(() => {
				const selectors = ['div.ds-flex._0a3d93b', '.ds-floating-position-wrapper', '[class*="_0a3d93b"]', '[class*="ds-flex"]'];
				let allElements = [];
				selectors.forEach(selector => {
					try {
						const elements = document.querySelectorAll(selector);
						Array.from(elements).forEach(el => {
							allElements.push({
								selector: selector,
								tagName: el.tagName,
								className: el.className,
								style: el.style.cssText,
								innerHTML: el.innerHTML ? el.innerHTML.substring(0, 500) : '',
								attributes: Array.prototype.reduce.call(el.attributes, function(acc, attr) {
									acc[attr.name] = attr.value;
									return acc;
								}, {})
							});
						});
					} catch(e) {
						// 蹇界暐閫夋嫨鍣ㄩ敊璇?
					}
				});
				return allElements;
			})()`, &aiCompletionElements).Do(ctx)
			if err != nil {
				return err
			}

			// 浠呭湪鍐呭瓨涓娇鐢ˋI瀹屾垚鏍囧織鍏冪礌淇℃伅锛屼笉淇濆瓨鍒版枃浠?
			gc.logger.Printf("馃攳 Potential AI Completion Elements acquired, count: %d", len(aiCompletionElements))
			return nil
		}),
	)
	if err != nil {
		gc.logger.Printf("鈿狅笍 Error getting potential AI completion elements: %v", err)
	}

	// 鍦ㄧ幇鏈堿I瀹屾垚鏍囧織鍏冪礌鎶撳彇鍚庢坊鍔犲寮虹殑DOM鎶撳彇鍔熻兘
	// 澧炲姞鏇村DOM鎶撳彇鏂规硶锛岀壒鍒槸閽堝鍙兘鍔ㄦ€佸姞杞界殑鍏冪礌锛堜笉淇濆瓨鍒版枃浠讹級
	var allPageElements []map[string]interface{}
	err = chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			err := chromedp.Evaluate(`(() => {
				// 鑾峰彇鎵€鏈夊彲鑳界殑鎸夐挳鐩稿叧鍏冪礌
				const buttonSelectors = [
					'div.ds-icon-button',
					'button',
					'[role="button"]',
					'.ds-flex',
					'[class*="button"]',
					'[class*="icon"]',
					'[class*="action"]',
					'div[tabindex]',
					'[class*="position-wrapper"]',
					'[class*="floating"]'
				];
				let allElements = [];
				buttonSelectors.forEach(selector => {
					try {
						const elements = document.querySelectorAll(selector);
						Array.from(elements).forEach(el => {
							// 鑾峰彇鍏冪礌鐨剅ect淇℃伅浠ヤ簡瑙ｅ叾浣嶇疆
							const rect = el.getBoundingClientRect();
							allElements.push({
								selector: selector,
								tagName: el.tagName,
								className: el.className,
								style: el.style.cssText,
								innerHTML: el.innerHTML ? el.innerHTML.substring(0, 500) : '',
								attributes: Object.fromEntries(Array.from(el.attributes).map(attr => [attr.name, attr.value])),
								isVisible: !!(rect.width && rect.height),
								isInViewport: rect.top >= 0 && rect.left >= 0 && rect.bottom <= window.innerHeight && rect.right <= window.innerWidth,
								rect: { top: rect.top, left: rect.left, bottom: rect.bottom, right: rect.right, width: rect.width, height: rect.height }
							});
						});
					} catch(e) {
						// 蹇界暐閫夋嫨鍣ㄩ敊璇?
					}
				});
				return allElements;
			})()`, &allPageElements).Do(ctx)
			if err != nil {
				return err
			}

			// 浠呭湪鍐呭瓨涓娇鐢ㄦ墍鏈夐〉闈㈠厓绱犱俊鎭紝涓嶄繚瀛樺埌鏂囦欢
			gc.logger.Printf("馃攳 All Page Elements acquired, count: %d", len(allPageElements))
			return nil
		}),
	)
	if err != nil {
		gc.logger.Printf("鈿狅笍 Error getting all page elements: %v", err)
	}

	// 鍐嶆鑾峰彇瀹屾暣鐨凞OM鏍戠粨鏋?
	var domTree map[string]interface{}
	err = chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			err := chromedp.Evaluate(`(function() {
				// 閫掑綊鑾峰彇DOM鏍戠粨鏋?
				function getDOMTree(node) {
					var result = {
						tagName: node.tagName || '#text',
						className: node.className || '',
						attributes: node.nodeType === 1 ? Array.prototype.reduce.call(node.attributes || [], function(acc, attr) {
							acc[attr.name] = attr.value;
							return acc;
						}, {}) : {},
						textContent: node.nodeType === 3 ? (node.textContent || '').trim().substring(0, 100) : '',
						children: []
					};

					if (node.childNodes) {
						for (var i = 0; i < node.childNodes.length; i++) {
							var child = node.childNodes[i];
							if (child.nodeType === 1 || child.nodeType === 3) { // Element or Text node
								result.children.push(getDOMTree(child));
							}
						}
					}
					return result;
				}

				return getDOMTree(document.documentElement);
			})()`, &domTree).Do(ctx)
			if err != nil {
				return err
			}

			// 浠呭湪鍐呭瓨涓娇鐢―OM鏍戜俊鎭紝涓嶄繚瀛樺埌鏂囦欢
			gc.logger.Printf("馃攳 Full DOM Tree acquired")
			return nil
		}),
	)
	if err != nil {
		gc.logger.Printf("鈿狅笍 Error getting DOM tree: %v", err)
	}

	// 妫€鏌ユ槸鍚﹀凡缁忔娴嬪埌AI杈撳嚭瀹屾垚锛屽鏋滄槸锛屽垯璺宠繃涓荤瓑寰呭惊鐜?
	if extractedContent != "" {
		gc.logger.Println("鈴笍 Skipping main wait loop as AI output has already been detected and captured")
		// 鐩存帴璺冲埌鍑芥暟鏈熬锛岄伩鍏嶈繘鍏ヤ富绛夊緟寰幆
		// 涓嶅啀绛夊緟涓诲惊鐜紝鐩存帴鎵ц娓呯悊骞惰繑鍥?
		gc.logger.Println("鉁?AI processing completed, closing browser immediately")
		return response, nil
	}

	// 寤堕暱绐楀彛瀛樻椿鏃堕棿锛岀‘淇滱I鏈夎冻澶熸椂闂村畬鎴愯緭鍑?
	gc.logger.Printf("鈴?Keeping browser window alive for configured timeout: %d seconds", forceCloseTimeoutSeconds)
	// 涓嶉渶瑕侀澶栫殑sleep锛屽洜涓烘暣浣撹秴鏃跺凡缁忓湪閰嶇疆涓缃?

	// 浣跨敤閰嶇疆鐨勮秴鏃舵椂闂寸瓑寰匒I澶勭悊骞惰幏鍙栧搷搴?
	gc.logger.Printf("鈴?Waiting for AI response with selectors: %v, timeout: %d seconds", responseSelectors, forceCloseTimeoutSeconds)

	// 绛夊緟鍝嶅簲鍑虹幇
	waitCtx, cancel := context.WithTimeout(ctx, time.Duration(forceCloseTimeoutSeconds)*time.Second) // 浣跨敤寮哄埗鍏抽棴瓒呮椂鏃堕棿
	defer cancel()

Loop:
	for {
		select {
		case <-waitCtx.Done():
			gc.logger.Println("鈴?Timeout waiting for AI response")
			break Loop
		default:
			// 棣栧厛灏濊瘯鑾峰彇鍝嶅簲
			err = chromedp.Run(ctx,
				chromedp.ActionFunc(func(ctx context.Context) error {
					for _, selector := range responseSelectors {
						// 妫€鏌ラ€夋嫨鍣ㄥ厓绱犳槸鍚﹀瓨鍦?
						var nodes []*cdp.Node
						if err := chromedp.Nodes(selector, &nodes).Do(ctx); err == nil && len(nodes) > 0 {
							// 鑾峰彇鏈€鏂扮殑鍝嶅簲鍐呭
							var respText string
							err = chromedp.Text(selector, &respText).Do(ctx)
							if err == nil && strings.TrimSpace(respText) != "" {
								// 馃毃 TRANSPARENCY PIPELINE NOTICE 馃毃
								// 閲嶈閫氱煡锛氬姭鍖鍒掗€忔槑绠￠亾鍘熷垯 - 淇濈暀鍘熷鍐呭鏍煎紡
								// 涓ョ鍦ㄦ澶勬坊鍔犱换浣曞舰寮忕殑鍐呭杩囨护銆佹牸寮忓寲鎴栨竻鐞嗘搷浣?
								response = respText
								gc.logger.Printf("鉁?Got response from selector: %s, length: %d", selector, len(response))
								return nil
							}
						}
					}
					return fmt.Errorf("no response found yet")
				}),
			)

			if err == nil {
				// 鎴愬姛鑾峰彇鍒板搷搴旓紝鐜板湪绛夊緟AI澶勭悊瀹屾垚鐨勬爣蹇楋細鎻愪氦鎸夐挳鍙樹负鍙敤鐘舵€?鎴?澶嶅埗鎸夐挳鍑虹幇
				gc.logger.Println("鉁?Response received, waiting for AI processing to complete (checking submit button or copy button)...")

				// 妫€鏌ョ壒瀹氬叧閿瘝鏄惁鍑虹幇锛圓I瀹屾垚鐨勬爣蹇楋級
				// 妫€娴嬫槸鍚﹀嚭鐜?div class="ds-flex _0a3d93b" style="align-items: center; gap: 10px;"><div class="ds-flex
				//杩欐槸涓€涓尮閰嶆槸鍚﹀畬鎴愯緭鍑虹殑鍒ゆ柇
				/*
					completionCtx, cancel := context.WithTimeout(ctx, time.Duration(GUARDIAN_BROWSER_TIMEOUT_SECONDS)*time.Second)
					defer cancel()

						for {
							select {
							case <-completionCtx.Done():
								gc.logger.Println("鈴?Timeout waiting for AI processing to complete")
								break Loop // 鍗充娇娌℃湁鏄庣‘瀹屾垚鏍囧織锛屾垜浠篃宸叉湁鍝嶅簲锛屾墍浠ラ€€鍑轰富寰幆
							default:
								// 妫€鏌ョ壒瀹氬叧閿瘝鏄惁鍑虹幇鍦ㄩ〉闈腑
								keywordFound := false
								err = chromedp.EvaluateAsDevTools(
									`(function() {
											var html = document.documentElement.outerHTML;
											return html.indexOf('div class=\"ds-flex _0a3d93b\" style=\"align-items: center; gap: 10px;\"') !== -1 &&
												html.indexOf('<div class=\"ds-flex') !== -1;
										})();`, &keywordFound).Do(ctx)

								if err == nil && keywordFound {
									gc.logger.Println("鉁?Specific keyword found, indicating AI processing completed")
									break Loop
								}

								time.Sleep(1 * time.Second) // 绛夊緟涓€绉掑悗鍐嶆妫€鏌?
							}
						}
				*/
				// 浣跨敤澶氱绛栫暐妫€娴婣I鏄惁瀹屾垚杈撳嚭
				completionCtx, cancel := context.WithTimeout(ctx, time.Duration(forceCloseTimeoutSeconds)*time.Second)
				defer cancel()

				for {
					select {
					case <-completionCtx.Done():
						gc.logger.Println("鈴?Timeout waiting for AI processing to complete")
						break Loop // 鍗充娇娌℃湁鏄庣‘瀹屾垚鏍囧織锛屾垜浠篃宸叉湁鍝嶅簲锛屾墍浠ラ€€鍑轰富寰幆
					default:
						// 妫€娴婣I鏄惁瀹屾垚杈撳嚭鐨勫绉嶆柟娉?- 浣跨敤鏂扮殑妫€娴嬫満鍒?
						// 浣跨敤鏂板鐨凙I瀹屾垚妫€娴嬫柟娉?
						err := gc.ProcessAIOutputCompletion(ctx)
						if err == nil {
							gc.logger.Println("鉁?AI processing completed using enhanced detection")
							break Loop
						} else {
							gc.logger.Printf("鈴?AI processing still in progress: %v", err)
						}

						time.Sleep(1 * time.Second) // 绛夊緟涓€绉掑悗鍐嶆妫€鏌?
					}
				}

				// 鏆傛椂娉ㄩ噴鎺夋墍鏈堿I瀹屾垚妫€娴嬫潯浠讹紝浠ヤ究閲嶆柊瀵绘壘鏈夋晥鐨勬娴嬫潯浠?
				/*
					for {
						select {
						case <-completionCtx.Done():
							gc.logger.Println("鈴?Timeout waiting for AI processing to complete")
							break Loop // 鍗充娇娌℃湁鏄庣‘瀹屾垚鏍囧織锛屾垜浠篃宸叉湁鍝嶅簲锛屾墍浠ラ€€鍑轰富寰幆
						default:
							// 妫€鏌ユ彁浜ゆ寜閽槸鍚﹀彉涓哄悜涓婄澶翠笖绂佺敤鐘舵€侊紙AI瀹屾垚鐨勬爣蹇椾箣涓€锛?
							// 鏍规嵁鎮ㄦ彁渚涚殑淇℃伅锛孉I姝ｅ湪杈撳嚭鏃舵寜閽槸鏂瑰舰鍥炬爣+鍙偣鍑伙紙aria-disabled="false"锛?
							// AI杈撳嚭瀹屾瘯鍚庢寜閽彉鎴愬悜涓婄澶?绂佺敤锛坅ria-disabled="true")
							buttonChanged := false
							for _, submitSel := range submitSelectors {
								err = chromedp.EvaluateAsDevTools(
									fmt.Sprintf(
										`(function() {
											var element = document.querySelector('%s');
											if (element) {
												// 妫€鏌ユ寜閽槸鍚﹀彉涓虹鐢ㄧ姸鎬侊紙aria-disabled="true"锛夛紝杩欒〃鏄嶢I宸插畬鎴?
												var isAriaDisabled = element.hasAttribute('aria-disabled') && element.getAttribute('aria-disabled') === 'true';
												return isAriaDisabled;
											}
											return false; // 鍏冪礌涓嶅瓨鍦ㄨ涓烘寜閽笉鍙敤
										})();`, submitSel), &buttonChanged).Do(ctx)

								if err == nil && buttonChanged {
									gc.logger.Println("鈿狅笍鈿狅笍 WARNING: 宸茬粡妫€娴嬪埌杈撳嚭瀹屾垚锛屽噯澶囧鍒?鈿狅笍鈿狅笍")
									break Loop
								}
							}

							// 鍚屾椂妫€鏌ュ鍒舵寜閽槸鍚﹀嚭鐜帮紙AI瀹屾垚鐨勫彟涓€涓爣蹇楋級
							copyButtonExists := false
							copySelectors := []string{
								"div.ds-flex._0a3d93b div.db183363.ds-icon-button.ds-icon-button--m.ds-icon-button--sizing-container[role='button'][aria-disabled='false']:first-child", // 绗竴涓寜閽?- 澶嶅埗
								"div.db183363.ds-icon-button.ds-icon-button--m.ds-icon-button--sizing-container[role='button'][aria-disabled='false']",                                  // DeepSeek澶嶅埗鎸夐挳 - 鍚敤鐘舵€?
								"div.db183363.ds-icon-button[role='button'][aria-disabled='false']",                                                                                     // DeepSeek澶嶅埗鎸夐挳 - 鍚敤鐘舵€侊紙澶囬€夛級
							}

							for _, copySelector := range copySelectors {
								err = chromedp.EvaluateAsDevTools(
									fmt.Sprintf(
										`(function() {
											var element = document.querySelector('%s');
											if (element !== null && element.hasAttribute('aria-disabled') && element.getAttribute('aria-disabled') === 'false') {
												// 妫€鏌ユ寜閽槸鍚﹀彲瑙佷笖鍦ㄨ鍙ｄ腑
												var rect = element.getBoundingClientRect();
												var isVisible = rect.top >= 0 && rect.left >= 0 &&
																rect.bottom <= (window.innerHeight || document.documentElement.clientHeight) &&
																rect.right <= (window.innerWidth || document.documentElement.clientWidth);

												// 妫€鏌ユ寜閽槸鍚︽湁灏哄锛堜笉涓?锛?
												var hasDimensions = rect.width > 0 && rect.height > 0;

												return isVisible && hasDimensions;
											}
											return false;
										})();`, copySelector), &copyButtonExists).Do(ctx)

								if err == nil && copyButtonExists {
									gc.logger.Println("鉁?Copy button detected as enabled and visible, indicating AI processing is complete")

									// 绔嬪嵆灏濊瘯鐐瑰嚮澶嶅埗鎸夐挳 - 浼樺厛鐐瑰嚮绗竴涓紙澶嶅埗鎸夐挳锛?
									gc.logger.Println("馃搵 Immediately attempting to click copy button...")

									// 浣跨敤鏇寸簿纭殑閫夋嫨鍣ㄦ潵鐐瑰嚮绗竴涓鍒舵寜閽?
									firstCopySelector := "div.ds-flex._0a3d93b div.db183363.ds-icon-button.ds-icon-button--m.ds-icon-button--sizing-container[role='button'][aria-disabled='false']:first-child"
									clickErr := chromedp.Click(firstCopySelector).Do(ctx)
									if clickErr == nil {
										gc.logger.Printf("鉁?Successfully clicked first copy button: %s", firstCopySelector)
									} else {
										gc.logger.Printf("鈿狅笍 Failed to click first copy button %s: %v, trying original selector", firstCopySelector, clickErr)
										// 濡傛灉绗竴涓€夋嫨鍣ㄥけ璐ワ紝灏濊瘯鍘熸潵鐨勯€夋嫨鍣?
										clickErr2 := chromedp.Click(copySelector).Do(ctx)
										if clickErr2 == nil {
											gc.logger.Printf("鉁?Successfully clicked copy button with original selector: %s", copySelector)
										} else {
											gc.logger.Printf("鈿狅笍 Failed to click copy button %s: %v", copySelector, clickErr2)
										}
									}

									break Loop
								}
							}

							// 妫€鏌ョ壒瀹氬畬鎴愭枃鏈槸鍚﹀嚭鐜帮紙鏈€楂樹紭鍏堢骇锛?
							completionTextExists := false
							completionTextSelectors := []string{
								"div.dbe8cf4a", // AI鐢熸垚瀹屾垚鏍囪鏂囨湰
							}

							for _, textSelector := range completionTextSelectors {
								err = chromedp.EvaluateAsDevTools(
									fmt.Sprintf(
										`(function() {
											var element = document.querySelector('%s');
											if (element) {
												var text = element.textContent || element.innerText;
												return text && text.includes('鏈洖绛旂敱 AI 鐢熸垚');
											}
											return false;
										})();`, textSelector), &completionTextExists).Do(ctx)

								if err == nil && completionTextExists {
									gc.logger.Println("鉁?AI completion text detected: '鏈洖绛旂敱 AI 鐢熸垚'")

									// 闅忔満绛夊緟1-2绉?
									randomWait := time.Duration(1000+rand.Intn(1000)) * time.Millisecond
									gc.logger.Printf("鈴?Waiting %.1f seconds before copying...", randomWait.Seconds())
									time.Sleep(randomWait)

									// 绔嬪嵆灏濊瘯鐐瑰嚮澶嶅埗鎸夐挳
									gc.logger.Println("馃搵 Immediately attempting to click copy button after detecting completion text...")

									// 鏌ユ壘骞剁偣鍑诲彲鐢ㄧ殑澶嶅埗鎸夐挳 - 浣跨敤鎮ㄦ彁渚涚殑瀹為檯鎸夐挳瀹氫綅鐐癸紝浼樺厛鐐瑰嚮绗竴涓紙澶嶅埗鎸夐挳锛?
									copySelectors := []string{
										"div.ds-flex._0a3d93b div.db183363.ds-icon-button.ds-icon-button--m.ds-icon-button--sizing-container[role='button'][aria-disabled='false']:first-child",  // 绗竴涓寜閽?- 澶嶅埗
										"div.ds-flex._0a3d93b div.db183363.ds-icon-button.ds-icon-button--m.ds-icon-button--sizing-container[role='button'][aria-disabled='false']:nth-child(1)", // 绗竴涓寜閽?- 澶嶅埗
										"div.ds-icon-button__hover-bg:first-child", // 绗竴涓寜閽殑瀹為檯浣嶇疆
										"div.ds-icon-button__hover-bg",             // 澶嶅埗鎸夐挳鐨勫疄闄呬綅缃?
										"div.db183363.ds-icon-button.ds-icon-button--m.ds-icon-button--sizing-container[role='button'][aria-disabled='false']",
										"div.db183363.ds-icon-button[role='button'][aria-disabled='false']",
									}

									clicked := false
									for _, copySelector := range copySelectors {
										// 棣栧厛绛夊緟鎸夐挳鍙
										err := chromedp.WaitVisible(copySelector).Do(ctx)
										if err == nil {
											clickErr := chromedp.Click(copySelector).Do(ctx)
											if clickErr == nil {
												gc.logger.Printf("鉁?Successfully clicked copy button after detecting completion text: %s", copySelector)
												clicked = true
												break
											} else {
												gc.logger.Printf("鈿狅笍 Failed to click copy button %s: %v", copySelector, clickErr)
											}
										} else {
											gc.logger.Printf("鈿狅笍 Copy button %s not visible: %v", copySelector, err)
										}
									}

									if !clicked {
										gc.logger.Println("鈿狅笍 Failed to click any copy button after detecting completion text")
									}

									break Loop
								}
							}

							time.Sleep(1 * time.Second) // 绛夊緟涓€绉掑悗鍐嶆妫€鏌?
						}
					}
				*/

			}
			// 缁х画寰幆绛夊緟
			time.Sleep(1 * time.Second)
		}
	}

	if response == "" {
		// 濡傛灉浠嶆病鏈夊搷搴旓紝灏濊瘯鑾峰彇椤甸潰鐨勬墍鏈夋枃鏈唴瀹?
		gc.logger.Println("馃摙 No specific response found, trying to get all page content...")
		err = chromedp.Run(ctx,
			chromedp.Evaluate("document.body.innerText", &response),
		)
		if err != nil {
			gc.logger.Printf("鉂?Failed to get page content: %v", err)
			return "", fmt.Errorf("failed to retrieve AI response: %w", err)
		}
		response = strings.TrimSpace(response)
	}

	gc.logger.Printf("鉁?Successfully retrieved AI response, length: %d", len(response))

	// 璁＄畻鎬昏€楁椂
	duration := time.Since(startTime)
	gc.logger.Printf("鈴憋笍 Browser automation completed in %v", duration)

	// 馃毃 TRANSPARENCY PIPELINE NOTICE 馃毃
	// ===================================
	// 閲嶈閫氱煡锛氬姭鍖鍒掗€忔槑绠￠亾鍘熷垯
	// 浠庢祻瑙堝櫒鑾峰彇鐨凙I杈撳嚭鍐呭蹇呴』淇濇寔鍘熷鏍煎紡锛屼笉鍋氫换浣曞鐞?
	// 浠讳綍瀵瑰唴瀹圭殑淇敼閮戒細褰卞搷AI鎬濈淮閾剧殑鍙鎬?
	// 涓ョ鍦ㄦ澶勬坊鍔犱换浣曞舰寮忕殑鍐呭杩囨护銆佹牸寮忓寲鎴栨竻鐞嗘搷浣?
	// ===================================
	gc.logger.Printf("馃Ч Final response content, length: %d", len(response))

	// 鏍规嵁閰嶇疆鍐冲畾鏄惁淇濇寔娴忚鍣ㄦ墦寮€
	// 鐢变簬Config缁撴瀯浣撲腑娌℃湁KeepAlive瀛楁锛屾殏鏃剁Щ闄よ鏉′欢
	// gc.logger.Printf("馃槾 Keeping browser alive for %d seconds as configured", GUARDIAN_AUTO_KEEP_OPEN_SECONDS)
	// time.Sleep(time.Duration(GUARDIAN_AUTO_KEEP_OPEN_SECONDS) * time.Second)

	// 鍦ˋI澶勭悊瀹屾垚鍚庣珛鍗冲叧闂祻瑙堝櫒锛屼笉鍐嶇瓑寰呴暱鏃堕棿寤惰繜
	gc.logger.Println("鉁?AI processing completed, closing browser immediately")

	// 涓嶅啀绛夊緟閰嶇疆鐨勫欢杩熸椂闂达紝鐩存帴杩斿洖鍘熷鍐呭
	// 馃敟 FUNDAMENTAL PRINCIPLE: RETURN RAW CONTENT WITHOUT ANY PROCESSING 馃敟
	return response, nil
}

// OpenLongLivedBrowser 鎵撳紑闀跨敓鍛藉懆鏈熺殑娴忚鍣ㄧ獥鍙ｏ紝鐢ㄤ簬棣栨鐧诲綍璁剧疆
func (gc *GuardianClient) OpenLongLivedBrowser(targetURL string) error {
	// 鍚姩娴忚鍣ㄨ嚜鍔ㄥ寲娴佺▼锛屼絾淇濇寔娴忚鍣ㄩ暱鏃堕棿鎵撳紑
	_, err := gc.performBrowserAutomationWithKeepAlive(targetURL)
	return err
}

// performBrowserAutomationWithKeepAlive 绫讳技浜巔erformBrowserAutomation锛屼絾淇濇寔娴忚鍣ㄩ暱鏃堕棿鎵撳紑
func (gc *GuardianClient) performBrowserAutomationWithKeepAlive(targetURL string) (string, error) {
	// 璁剧疆Chrome閫夐」 - 姣忎釜浜ゆ槗鍛樹娇鐢ㄧ嫭绔嬬殑鐢ㄦ埛鏁版嵁鐩綍锛岀‘淇濇瘡涓氦鏄撳憳鏈夌嫭绔嬬殑鐧诲綍鐘舵€?
	uniqueID := gc.TraderID
	if uniqueID == "" {
		// 濡傛灉娌℃湁浜ゆ槗鍛業D锛屽垯浣跨敤鏃堕棿鎴冲拰瀹炰緥鍦板潃浣滀负鍚庡
		timestamp := time.Now().UnixNano()
		uniqueID = fmt.Sprintf("%d_%p", timestamp, gc)
	}
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),             // 闈炴棤澶存ā寮忎互渚胯瀵?
		chromedp.Flag("disable-web-security", false), // 鍚敤缃戠粶瀹夊叏浠ユ敮鎸佹甯哥綉绔欏姛鑳?
		chromedp.Flag("disable-features", "VizDisplayCompositor"),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-gpu", false),                    // 鍚敤GPU鍔犻€?
		chromedp.Flag("blink-settings", "imagesEnabled=true"),  // 鍚敤鍥剧墖鍔犺浇浠ユ敮鎸侀獙璇佺爜绛夊姛鑳?
		chromedp.Flag("enable-automation", false),              // 闃叉琚綉绔欐娴嬩负鑷姩鍖?
		chromedp.Flag("exclude-switches", "enable-automation"), // 鎺掗櫎鑷姩鍖栧紑鍏?
		chromedp.Flag("disable-extensions", false),             // 鍚敤鎵╁睍
		chromedp.Flag("disable-plugins-discovery", false),      // 鍚敤鎻掍欢鍙戠幇
		chromedp.Flag("incognito", false),                      // 涓嶄娇鐢ㄩ殣韬ā寮?
		chromedp.Flag("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"), // 璁剧疆姝ｅ父鐢ㄦ埛浠ｇ悊
		chromedp.Flag("disable-blink-features", "AutomationControlled"),                                                                                // 绂佺敤鑷姩鍖栨帶鍒剁壒寰?
		chromedp.Flag("renderer-process-limit", "1"),
		chromedp.Flag("max_old_space_size", "4096"),
		chromedp.Flag("no-first-run", "true"),
		chromedp.Flag("no-default-browser-check", "true"),
		chromedp.Flag("disable-backgrounding-occluded-windows", "false"),                       // 纭繚绐楀彛鍙
		chromedp.Flag("disable-renderer-backgrounding", "true"),                                // 闃叉鍚庡彴娓叉煋
		chromedp.Flag("disable-background-timer-throttling", "true"),                           // 闃叉瀹氭椂鍣ㄨ妭娴?
		chromedp.Flag("disable-background-networking", "false"),                                // 鍏佽鍚庡彴缃戠粶娲诲姩
		chromedp.Flag("window-size", "1000,850"),                                               // 璁剧疆娴忚鍣ㄧ獥鍙ｅ昂瀵镐负1000x850
		chromedp.Flag("user-data-dir", fmt.Sprintf("%s_%s", GuardianBrowserDataDir, uniqueID)), // 姣忎釜瀹炰緥浣跨敤鐙珛鐨勭敤鎴锋暟鎹洰褰?
		chromedp.Flag("profile-directory", "Default"),                                          // 浣跨敤榛樿閰嶇疆鏂囦欢
	)

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	// defer cancel() // 涓存椂娉ㄩ噴鎺夛紝淇濇寔娴忚鍣ㄧ獥鍙ｆ墦寮€

	// 鍒涘缓chrome瀹炰緥涓婁笅鏂?
	ctx, _ := chromedp.NewContext(allocCtx)
	// defer cancel() // 涓存椂娉ㄩ噴鎺夛紝淇濇寔娴忚鍣ㄧ獥鍙ｆ墦寮€

	// 璁剧疆杈冮暱鏃堕棿鐨勮秴鏃?
	timeout := time.Duration(GUARDIAN_LONG_BROWSER_TIMEOUT_SECONDS) * time.Second
	ctx, _ = context.WithTimeout(ctx, timeout)
	// defer cancel() // 涓存椂娉ㄩ噴鎺夛紝淇濇寔娴忚鍣ㄧ獥鍙ｆ墦寮€

	gc.logger.Printf("鈴?Long-lived browser automation started, timeout: %v", timeout)

	// 璁块棶鐩爣URL
	if !strings.HasPrefix(targetURL, "http") {
		targetURL = "https://" + targetURL
	}

	gc.logger.Printf("馃寪 Opening long-lived browser window for URL: %s", targetURL)

	// 瀵艰埅鍒扮洰鏍囬〉闈?
	if err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(targetURL),
		chromedp.Sleep(2*time.Second), // 绛夊緟椤甸潰鍔犺浇
	); err != nil {
		return "", fmt.Errorf("failed to navigate to target URL: %w", err)
	}

	gc.logger.Printf("鉁?Successfully opened long-lived browser window for URL: %s", targetURL)

	// 淇濇寔娴忚鍣ㄦ墦寮€鎸囧畾鏃堕棿
	time.Sleep(time.Duration(GUARDIAN_LONG_BROWSER_TIMEOUT_SECONDS) * time.Second)

	return "Browser window opened successfully", nil
}

// CallWithMessages 璋冪敤AI鏈嶅姟鐨勪富瑕佹柟娉?
func (gc *GuardianClient) CallWithMessages(systemPrompt, userPrompt string) (string, error) {
	return gc.call(systemPrompt, userPrompt)
}

// CallWithRequest 浣跨敤Request瀵硅薄璋冪敤AI鏈嶅姟
func (gc *GuardianClient) CallWithRequest(req *Request) (string, error) {
	// 瀵逛簬娴忚鍣ㄨ嚜鍔ㄥ寲瀹㈡埛绔紝鎴戜滑绠€鍖栧鐞?
	// 鍒嗙绯荤粺鎻愮ず鍜屽叾浠栨秷鎭?
	var systemPrompt string
	var userPrompt string

	for _, msg := range req.Messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
		} else {
			userPrompt += msg.Content + "\n"
		}
	}

	return gc.call(systemPrompt, userPrompt)
}

func (gc *GuardianClient) buildMCPRequestBody(systemPrompt, userPrompt string) map[string]any {
	// 瀵逛簬娴忚鍣ㄨ嚜鍔ㄥ寲锛屼笉闇€瑕佹瀯寤篐TTP璇锋眰浣?
	return nil
}

func (gc *GuardianClient) buildRequestBodyFromRequest(req *Request) map[string]any {
	// 瀵逛簬娴忚鍣ㄨ嚜鍔ㄥ寲锛屼笉闇€瑕佹瀯寤篐TTP璇锋眰浣?
	return nil
}

func (gc *GuardianClient) buildUrl() string {
	// 杩斿洖閰嶇疆鐨勫熀鏈琔RL
	return gc.ProviderConfig.BaseURL
}

func (gc *GuardianClient) buildRequest(url string, jsonData []byte) (*http.Request, error) {
	// 瀵逛簬娴忚鍣ㄨ嚜鍔ㄥ寲锛屾垜浠笉浣跨敤鏍囧噯HTTP璇锋眰
	return nil, fmt.Errorf("GuardianClient does not use standard HTTP requests")
}

func (gc *GuardianClient) setAuthHeader(reqHeaders http.Header) {
	// 瀵逛簬娴忚鍣ㄨ嚜鍔ㄥ寲锛岃璇侀€氳繃娴忚鍣ㄤ細璇濆鐞?
}

func (gc *GuardianClient) marshalRequestBody(requestBody map[string]any) ([]byte, error) {
	// 瀵逛簬娴忚鍣ㄨ嚜鍔ㄥ寲锛屼笉闇€瑕佸簭鍒楀寲璇锋眰浣?
	return nil, nil
}

func (gc *GuardianClient) parseMCPResponse(body []byte) (string, error) {
	// 瀵逛簬娴忚鍣ㄨ嚜鍔ㄥ寲锛屽搷搴旂敱娴忚鍣ㄨ嚜鍔ㄥ寲娴佺▼澶勭悊
	return "", nil
}

func (gc *GuardianClient) isRetryableError(err error) bool {
	// 瀵逛簬娴忚鍣ㄨ嚜鍔ㄥ寲锛屾墍鏈夐敊璇兘鍙互閲嶈瘯
	return true
}

