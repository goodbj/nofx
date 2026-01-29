package components

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// DOMInspector handles DOM inspection and element detection
type DOMInspector struct {
	logger *log.Logger
}

// NewDOMInspector creates a new DOM inspector
func NewDOMInspector(logger *log.Logger) *DOMInspector {
	return &DOMInspector{
		logger: logger,
	}
}

// CaptureDOMSnapshot captures a snapshot of the current DOM
func (di *DOMInspector) CaptureDOMSnapshot(ctx context.Context) error {
	// 使用新的chromedp动作来获取完整的页面HTML内容并保存到文件
	var pageHTML string
	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			// 获取完整的页面HTML内容
			err := chromedp.OuterHTML("html", &pageHTML).Do(ctx)
			if err != nil {
				return err
			}

			// 将完整的HTML内容保存到文件
			filename := fmt.Sprintf("dom_snapshot_%d.html", time.Now().Unix())
			err = os.WriteFile(filename, []byte(pageHTML), 0644)
			if err == nil {
				di.logger.Printf("Full DOM snapshot saved to %s, size: %d bytes", filename, len(pageHTML))
			} else {
				di.logger.Printf("Error saving DOM snapshot to file: %v", err)
				// 如果保存文件失败，至少输出部分HTML内容到日志
				maxLength := 2000
				if len(pageHTML) < maxLength {
					maxLength = len(pageHTML)
				}
				di.logger.Printf("DOM Snapshot (first %d chars):\n%s", maxLength, pageHTML[:maxLength])
			}
			return nil
		}),
	)
	return err
}

// GetDSThemeElements gets detailed information about ds-theme elements
func (di *DOMInspector) GetDSThemeElements(ctx context.Context) ([]map[string]interface{}, error) {
	var themeElements []map[string]interface{}
	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			err := chromedp.Evaluate(`(() => {
				const elements = document.querySelectorAll('div.ds-theme');
				return Array.from(elements).map(el => ({
					tagName: el.tagName,
					className: el.className,
					style: el.style.cssText,
					innerHTML: el.innerHTML ? el.innerHTML.substring(0, 500) : '', // 增加到前500个字符
					attributes: Array.prototype.reduce.call(el.attributes, function(acc, attr) {
						acc[attr.name] = attr.value;
						return acc;
					}, {})
				}));
			})()`, &themeElements).Do(ctx)
			if err != nil {
				return err
			}

			// 将ds-theme元素详情也保存到文件
			themeFilename := fmt.Sprintf("ds_theme_elements_%d.json", time.Now().Unix())
			themeJSON, jsonErr := json.MarshalIndent(themeElements, "", "  ")
			if jsonErr == nil {
				err = os.WriteFile(themeFilename, themeJSON, 0644)
				if err == nil {
					di.logger.Printf("ds-theme Elements Detail saved to %s", themeFilename)
				} else {
					di.logger.Printf("Error saving ds-theme elements to file: %v", err)
					di.logger.Printf("ds-theme Elements Detail: %+v", themeElements)
				}
			} else {
				di.logger.Printf("Error marshaling ds-theme elements to JSON: %v", jsonErr)
				di.logger.Printf("ds-theme Elements Detail: %+v", themeElements)
			}
			return nil
		}),
	)
	return themeElements, err
}

// FindInputElement finds an input element using multiple selectors
func (di *DOMInspector) FindInputElement(ctx context.Context, selectors []string) (string, bool) {
	for _, selector := range selectors {
		di.logger.Printf("Attempting to find input with selector: %s", selector)

		// 首先等待元素可见
		err := chromedp.WaitVisible(selector).Do(ctx)
		if err == nil {
			di.logger.Printf("Found input element with selector: %s", selector)
			return selector, true
		} else {
			di.logger.Printf("Input selector %s not found or not visible: %v", selector, err)
		}
	}
	return "", false
}

// FindSubmitButton finds a submit button using multiple selectors
func (di *DOMInspector) FindSubmitButton(ctx context.Context, selectors []string) (string, bool) {
	for _, selector := range selectors {
		di.logger.Printf("Attempting to find submit button with selector: %s", selector)

		// 首先等待按钮可见
		err := chromedp.WaitVisible(selector).Do(ctx)
		if err == nil {
			di.logger.Printf("Found submit button: %s", selector)
			return selector, true
		} else {
			di.logger.Printf("Submit button %s not visible: %v", selector, err)
		}
	}
	return "", false
}

// WaitForElement waits for an element to appear
func (di *DOMInspector) WaitForElement(ctx context.Context, selector string, timeout time.Duration) error {
	done := make(chan bool, 1)

	go func() {
		err := chromedp.WaitVisible(selector).Do(ctx)
		if err == nil {
			done <- true
		} else {
			done <- false
		}
	}()

	select {
	case result := <-done:
		if result {
			return nil
		} else {
			return fmt.Errorf("element not found: %s", selector)
		}
	case <-time.After(timeout):
		return fmt.Errorf("timeout waiting for element: %s", selector)
	}
}

// GetElementText retrieves text from an element
func (di *DOMInspector) GetElementText(ctx context.Context, selector string) (string, error) {
	var text string
	err := chromedp.Run(ctx,
		chromedp.Text(selector, &text, chromedp.ByQuery),
	)
	return strings.TrimSpace(text), err
}

// GetAllPageElements gets detailed information about all page elements
func (di *DOMInspector) GetAllPageElements(ctx context.Context) ([]map[string]interface{}, error) {
	var allPageElements []map[string]interface{}
	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			err := chromedp.Evaluate(`(() => {
				// 获取所有可能的按钮相关元素
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
							// 获取元素的rect信息以了解其位置
							const rect = el.getBoundingClientRect();
							allElements.push({
								selector: selector,
								tagName: el.tagName,
								className: el.className,
								style: el.style.cssText,
								innerHTML: el.innerHTML ? el.innerHTML.substring(0, 500) : '',
								attributes: Object.fromEntries(Array.from(el.attributes).map(attr => [attr.name, attr.value])),
								isVisible: !!(rect.width && rect.height),
								isInViewport: rect.top >= 0 && rect.left >= 0 &&
											rect.bottom <= (window.innerHeight || document.documentElement.clientHeight) &&
											rect.right <= (window.innerWidth || document.documentElement.clientWidth),
								rect: { top: rect.top, left: rect.left, bottom: rect.bottom, right: rect.right, width: rect.width, height: rect.height }
							});
						});
					} catch(e) {}
				});
				return allElements;
			})()`, &allPageElements).Do(ctx)
			if err != nil {
				return err
			}

			// 将所有页面元素详情也保存到文件
			allElementsFilename := fmt.Sprintf("all_page_elements_%d.json", time.Now().Unix())
			allElementsJSON, jsonErr := json.MarshalIndent(allPageElements, "", "  ")
			if jsonErr == nil {
				err = os.WriteFile(allElementsFilename, allElementsJSON, 0644)
				if err == nil {
					di.logger.Printf("All Page Elements saved to %s", allElementsFilename)
				} else {
					di.logger.Printf("Error saving all page elements to file: %v", err)
					di.logger.Printf("All Page Elements count: %d", len(allPageElements))
				}
			} else {
				di.logger.Printf("Error marshaling all page elements to JSON: %v", jsonErr)
				di.logger.Printf("All Page Elements count: %d", len(allPageElements))
			}
			return nil
		}),
	)
	return allPageElements, err
}

// GetDOMTree gets the complete DOM tree structure
func (di *DOMInspector) GetDOMTree(ctx context.Context) (map[string]interface{}, error) {
	var domTree map[string]interface{}
	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			err := chromedp.Evaluate(`(function() {
				// 递归获取DOM树结构
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

			// 将DOM树结构保存到文件
			domTreeFilename := fmt.Sprintf("dom_tree_%d.json", time.Now().Unix())
			domTreeJSON, jsonErr := json.MarshalIndent(domTree, "", "  ")
			if jsonErr == nil {
				err = os.WriteFile(domTreeFilename, domTreeJSON, 0644)
				if err == nil {
					di.logger.Printf("Full DOM Tree saved to %s", domTreeFilename)
				} else {
					di.logger.Printf("Error saving DOM tree to file: %v", err)
				}
			} else {
				di.logger.Printf("Error marshaling DOM tree to JSON: %v", jsonErr)
			}
			return nil
		}),
	)
	return domTree, err
}

// ScrollAndCapture scrolls the page and captures snapshots
func (di *DOMInspector) ScrollAndCapture(ctx context.Context) error {
	var result interface{}
	scrollErr := chromedp.Run(ctx,
		chromedp.Evaluate(`(() => {
			window.scrollTo(0, document.body.scrollHeight);
			return 'scrolled';
		})()`, &result),
		chromedp.Sleep(1*time.Second),
		chromedp.Evaluate(`(() => {
			window.scrollTo(0, 0);
			return 'back to top';
		})()`, &result),
		chromedp.Sleep(1*time.Second),
	)
	if scrollErr != nil {
		di.logger.Printf("Error during scroll operations: %v", scrollErr)
		return scrollErr
	}
	return nil
}
