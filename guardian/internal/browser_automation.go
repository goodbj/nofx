package guardian

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"nofx/config"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// BrowserAutomation 娴忚鍣ㄨ嚜鍔ㄥ寲鎺у埗鍣?
type BrowserAutomation struct {
	config  *config.BrowserConfig
	ctx     context.Context
	cancel  context.CancelFunc
	browser context.Context // chromedp娴忚鍣ㄤ笂涓嬫枃
	isReady bool
}

// NewBrowserAutomation 鍒涘缓娴忚鍣ㄨ嚜鍔ㄥ寲瀹炰緥
func NewBrowserAutomation(browserConfig *config.BrowserConfig) (*BrowserAutomation, error) {
	ctx, cancel := context.WithCancel(context.Background())

	ba := &BrowserAutomation{
		config:  browserConfig,
		ctx:     ctx,
		cancel:  cancel,
		isReady: false,
	}

	// 鍒濆鍖栨祻瑙堝櫒
	err := ba.InitBrowser()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize browser: %w", err)
	}

	return ba, nil
}

// InitBrowser 鍒濆鍖栨祻瑙堝櫒
func (ba *BrowserAutomation) InitBrowser() error {
	log.Println("馃寪 Initializing browser automation...")

	// 璁剧疆chromedp閫夐」
	options := []chromedp.ExecAllocatorOption{}

	if ba.config.ExecutablePath != "" {
		options = append(options, chromedp.ExecPath(ba.config.ExecutablePath))
	}

	// 娣诲姞瑙勯伩妫€娴嬬殑閫夐」
	options = append(options,
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("exclude-switches", "enable-automation"),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-plugins-discovery", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("allow-running-insecure-content", true),
		chromedp.Flag("use-fake-ui-for-media-stream", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-ipc-flooding-protection", true),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-features", "TranslateUI,VizDisplayCompositor"),
		chromedp.Flag("enable-features", "NetworkService,NetworkServiceInProcess"),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("disable-software-rasterizer", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-plugins", true),
		chromedp.Flag("disable-image-animation-resampling", true),
		chromedp.Flag("disable-session-crashed-bubble", true),
		chromedp.Flag("disable-breakpad", true),
		chromedp.Flag("disable-field-trial-config", true),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-ipc-flooding-protection", false),
		chromedp.Flag("remote-debugging-port", "9222"),
	)

	if ba.config.Headless {
		options = append(options, chromedp.Headless)
	}

	if ba.config.UserDataDir != "" {
		options = append(options, chromedp.UserDataDir(ba.config.UserDataDir))
	}

	if ba.config.DisableImages {
		options = append(options, chromedp.Flag("disable-images", true))
	}

	if ba.config.DisableJS {
		options = append(options, chromedp.Flag("disable-javascript", true))
	}

	// 娣诲姞棰濆鍙傛暟
	for _, arg := range ba.config.AdditionalArgs {
		options = append(options, chromedp.Flag(strings.Split(arg, "=")[0], strings.Split(arg, "=")[1]))
	}

	// 鍒涘缓鍒嗛厤鍣ㄤ笂涓嬫枃
	allocatorCtx, browserCancel := chromedp.NewExecAllocator(context.Background(), options...)
	ba.browser, ba.cancel = chromedp.NewContext(allocatorCtx)
	ba.cancel = browserCancel // Override with allocator cancel to close browser properly

	// 鍚姩娴忚鍣?
	if err := chromedp.Run(ba.browser); err != nil {
		return fmt.Errorf("failed to start browser: %w", err)
	}

	ba.isReady = true
	log.Println("鉁?Browser automation ready")
	// 娉ㄥ叆JavaScript鏉ラ殣钘弚ebdriver灞炴€?
	go func() {
		chromedp.Run(ba.browser,
			chromedp.Evaluate(
				`(function(){
					Object.defineProperty(navigator, 'webdriver', {
						get: () => undefined,
				});
				window.chrome = {
					runtime: {}
				};
				Object.defineProperty(navigator, 'plugins', {
					get: () => [1, 2, 3, 4, 5],
				});
				Object.defineProperty(navigator, 'languages', {
					get: () => ['zh-CN', 'zh', 'en'],
				});
			})())`, nil,
			),
		)
	}()
	return nil
}

// ProcessPrompt 澶勭悊鎻愮ず璇嶏紝鍙戦€佸埌娴忚鍣ˋI鏈嶅姟骞惰幏鍙栧搷搴?
func (ba *BrowserAutomation) ProcessPrompt(systemPrompt, userPrompt string) (string, error) {
	if !ba.isReady {
		return "", fmt.Errorf("browser automation not ready")
	}

	log.Println("馃挰 Processing prompt in browser AI service...")

	// 瀵艰埅鍒癆I鏈嶅姟椤甸潰
	err := ba.navigateToPage()
	if err != nil {
		return "", fmt.Errorf("failed to navigate to AI page: %w", err)
	}

	// 娓呯┖杈撳叆妗嗭紙濡傛灉瀛樺湪涔嬪墠鐨勫璇濓級
	err = ba.clearChat()
	if err != nil {
		log.Printf("鈿狅笍 Could not clear chat: %v", err)
		// 缁х画鎵ц锛屼笉娓呯┖鍙兘涔熸棤濡?
	}

	// 鍙戦€佹彁绀鸿瘝鍒癆I鏈嶅姟
	err = ba.sendPrompt(systemPrompt, userPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to send prompt: %w", err)
	}

	// 绛夊緟AI鍝嶅簲
	response, err := ba.waitForResponse()
	if err != nil {
		return "", fmt.Errorf("failed to get AI response: %w", err)
	}

	log.Printf("鉁?AI response received, length: %d", len(response))
	return response, nil
}

// navigateToPage 瀵艰埅鍒癆I鏈嶅姟椤甸潰
func (ba *BrowserAutomation) navigateToPage() error {
	log.Printf("馃Л Navigating to AI service page...")

	// 瀵艰埅鍒版寚瀹氱殑AI鏈嶅姟椤甸潰
	err := chromedp.Run(ba.browser,
		chromedp.Navigate("https://chat.deepseek.com"),                               // 榛樿浣跨敤DeepSeek锛屽彲浠ラ€氳繃閰嶇疆鏇存敼
		chromedp.WaitVisible("textarea[data-testid='chat-input']", chromedp.ByQuery), // 绛夊緟杈撳叆妗嗗彲瑙?
	)
	if err != nil {
		return fmt.Errorf("failed to navigate to page: %w", err)
	}

	log.Println("鉁?Navigation completed")
	return nil
}

// clearChat 娓呯┖鑱婂ぉ璁板綍
func (ba *BrowserAutomation) clearChat() error {
	log.Println("馃Ч Clearing chat history...")

	// 灏濊瘯鎵惧埌骞剁偣鍑绘竻闄ゆ寜閽?
	// 娉ㄦ剰锛氳繖鍙栧喅浜庡叿浣撶綉绔欑殑DOM缁撴瀯锛岄渶瑕佹牴鎹疄闄呮儏鍐佃皟鏁?
	err := chromedp.Run(ba.browser,
		chromedp.Click("button[aria-label='Clear conversation']", chromedp.ByQuery),
	)

	// 濡傛灉娓呴櫎鎸夐挳涓嶅瓨鍦紝蹇界暐閿欒
	if err != nil {
		log.Printf("鈿狅笍 Could not find clear button, skipping: %v", err)
		// 涓嶈繑鍥為敊璇紝鍥犱负杩欓€氬父鏄彲閫夋搷浣?
	}

	log.Println("鉁?Chat cleared or clear button not found")
	return nil
}

// sendPrompt 鍙戦€佹彁绀鸿瘝鍒癆I鏈嶅姟
func (ba *BrowserAutomation) sendPrompt(systemPrompt, userPrompt string) error {
	log.Printf("馃摛 Sending prompt to AI service, user prompt length: %d", len(userPrompt))

	// 灏嗘彁绀鸿瘝鍙戦€佸埌杈撳叆妗嗗苟鐐瑰嚮鍙戦€?
	err := chromedp.Run(ba.browser,
		chromedp.Clear("textarea[data-testid='chat-input']", chromedp.ByQuery),
		chromedp.SetValue("textarea[data-testid='chat-input']", userPrompt, chromedp.ByQuery),
		chromedp.Click("button[data-testid='chat-send-button']", chromedp.ByQuery),
	)

	if err != nil {
		return fmt.Errorf("failed to send prompt: %w", err)
	}

	log.Println("鉁?Prompt sent to AI service")
	return nil
}

// waitForResponse 绛夊緟AI鏈嶅姟鍝嶅簲
func (ba *BrowserAutomation) waitForResponse() (string, error) {
	log.Println("鈴?Waiting for AI response...")

	var responseText string
	timeout := time.After(60 * time.Second) // 60绉掕秴鏃?
	tick := time.Tick(2 * time.Second)      // 姣?绉掓鏌ヤ竴娆?

	// 绛夊緟AI鍝嶅簲锛岀洿鍒拌秴鏃?
	for {
		select {
		case <-timeout:
			return "", fmt.Errorf("timeout waiting for AI response")
		case <-tick:
			// 灏濊瘯鑾峰彇AI鍝嶅簲
			err := chromedp.Run(ba.browser,
				chromedp.Text("div[data-testid='chat-response']", &responseText, chromedp.ByQuery),
			)

			if err == nil && responseText != "" {
				log.Println("鉁?Response received from AI service")
				// 灏濊瘯瑙ｆ瀽鍝嶅簲涓篔SON鏍煎紡锛屽鏋滀笉鏄疛SON鍒欒繑鍥炲師濮嬫枃鏈?
				var parsed interface{}
				err = json.Unmarshal([]byte(responseText), &parsed)
				if err != nil {
					// 濡傛灉涓嶆槸鏈夋晥JSON锛屽寘瑁呮垚閫傚綋鐨勫搷搴旀牸寮?
					wrappedResponse := map[string]interface{}{
						"raw_response": responseText,
						"decisions":    []interface{}{},
					}
					wrappedBytes, _ := json.Marshal(wrappedResponse)
					return string(wrappedBytes), nil
				}
				return responseText, nil
			}
		}
	}
}

// Close 鍏抽棴娴忚鍣ㄨ嚜鍔ㄥ寲
func (ba *BrowserAutomation) Close() {
	log.Println("馃洃 Closing browser automation...")

	if ba.cancel != nil {
		ba.cancel()
	}

	// 鍦ㄥ疄闄呭疄鐜颁腑锛岃繖閲屼細鍏抽棴娴忚鍣ㄥ疄渚?
	log.Println("鉁?Browser automation closed")
}

