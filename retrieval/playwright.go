package retrieval

// playwrightFetcher fetches pages using a real Firefox browser managed by
// Playwright. It satisfies the same (body, mimeType, error) contract as the
// plain HTTP fetch path, so the rest of Client.SaveDocument is unchanged.

import (
	"context"
	"fmt"

	"github.com/playwright-community/playwright-go"
)

// firefoxUserAgent matches a current desktop Firefox on macOS, used for both
// the HTTP client and the Playwright browser context.
const firefoxUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:138.0) Gecko/20100101 Firefox/138.0"

type playwrightFetcher struct {
	pw             *playwright.Playwright
	browser        playwright.Browser
	timeoutSeconds int
}

// newPlaywrightFetcher starts Playwright and launches a Firefox instance.
//
// Key anti-detection measures applied:
//   - Non-headless (visible window): eliminates dozens of headless-only
//     fingerprint signals that Akamai and similar systems check.
//   - dom.webdriver.enabled = false: stops Firefox from advertising
//     navigator.webdriver = true, which bot-detection JS reads directly.
//   - Realistic user agent and 1920×1080 viewport set on every context.
//
// The caller must call close() when done.
func newPlaywrightFetcher(timeoutSeconds int) (*playwrightFetcher, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("playwright: start: %w", err)
	}
	browser, err := pw.Firefox.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
		FirefoxUserPrefs: map[string]interface{}{
			// Prevents navigator.webdriver from being set to true.
			"dom.webdriver.enabled": false,
		},
	})
	if err != nil {
		pw.Stop() //nolint:errcheck
		return nil, fmt.Errorf("playwright: launch firefox: %w", err)
	}
	return &playwrightFetcher{pw: pw, browser: browser, timeoutSeconds: timeoutSeconds}, nil
}

// fetch navigates to url in a fresh browser context, waits for network
// activity to settle (giving JS challenges time to complete), then returns
// the fully-rendered HTML.  A new context is created per call so cookies and
// storage do not leak between requests.
func (p *playwrightFetcher) fetch(_ context.Context, url string) ([]byte, string, error) {
	ctx, err := p.browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String(firefoxUserAgent),
		Viewport: &playwright.Size{
			Width:  1920,
			Height: 1080,
		},
	})
	if err != nil {
		return nil, "", fmt.Errorf("playwright: new context: %w", err)
	}
	defer ctx.Close()

	page, err := ctx.NewPage()
	if err != nil {
		return nil, "", fmt.Errorf("playwright: new page: %w", err)
	}

	totalMS := playwright.Float(float64(p.timeoutSeconds) * 1000)

	// Step 1: navigate and wait for the initial HTML + subresources to load.
	if _, err := page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateLoad,
		Timeout:   totalMS,
	}); err != nil {
		return nil, "", fmt.Errorf("playwright: goto %s: %w", url, err)
	}

	// Step 2: wait for network activity to go quiet.  This lets the SPA
	// framework finish its initial XHR/fetch calls (data loading, bot-challenge
	// token exchange, etc.) before we start watching the DOM.
	if err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: totalMS,
	}); err != nil {
		return nil, "", fmt.Errorf("playwright: wait for network idle: %w", err)
	}

	// Step 3: wait until the DOM has been stable for 2 seconds.  We observe
	// document.documentElement (the <html> element) rather than document.body
	// because body can be null during mid-navigation / redirect states.
	// The debounce timer resets on every mutation; once 2 full seconds pass
	// with no DOM changes the page has finished rendering.
	if _, err := page.WaitForFunction(`() => new Promise(resolve => {
		let timer;
		const reset = () => {
			clearTimeout(timer);
			timer = setTimeout(() => { observer.disconnect(); resolve(true); }, 5000);
		};
		const observer = new MutationObserver(reset);
		observer.observe(document.documentElement, { childList: true, subtree: true, characterData: true });
		reset(); // start the timer even if the DOM is already stable
	})`, nil, playwright.PageWaitForFunctionOptions{Timeout: totalMS}); err != nil {
		return nil, "", fmt.Errorf("playwright: wait for DOM stability: %w", err)
	}

	content, err := page.Content()
	if err != nil {
		return nil, "", fmt.Errorf("playwright: content: %w", err)
	}

	return []byte(content), "text/html; charset=utf-8", nil
}

// close shuts down the browser and the Playwright subprocess.
func (p *playwrightFetcher) close() error {
	if err := p.browser.Close(); err != nil {
		return fmt.Errorf("playwright: close browser: %w", err)
	}
	if err := p.pw.Stop(); err != nil {
		return fmt.Errorf("playwright: stop: %w", err)
	}
	return nil
}
