package crawl

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

// PopupScraper handles the scraping operations
type PopupScraper struct {
	ctx           context.Context
	cancel        context.CancelFunc
	buttonPath    string
	dataSelectors []string
}

// ScrapedData represents the extracted data structure
type ScrapedData struct {
	ButtonPath  string            `json:"button_path"`
	PopupData   map[string]string `json:"popup_data"`
	PopupHTML   string            `json:"popup_html"`
	ExtractedAt string            `json:"extracted_at"`
}

// NewPopupScraper creates a new scraper instance
func NewPopupScraper(buttonPath string, dataSelectors []string) *PopupScraper {
	// Configure Chrome options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false), // Set to true for headless mode
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.WindowSize(1920, 1080),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"),
	)

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx)

	return &PopupScraper{
		ctx:           ctx,
		cancel:        cancel,
		buttonPath:    buttonPath,
		dataSelectors: dataSelectors,
	}
}

// Close cleans up resources
func (ps *PopupScraper) Close() {
	ps.cancel()
}

// ScrapePopupData is the main scraping method
func (ps *PopupScraper) ScrapePopupData(url string) (*ScrapedData, error) {
	var popupHTML string
	var buttonExists bool

	// Step 1: Navigate to the page and wait for it to load
	err := chromedp.Run(ps.ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(3*time.Second), // Wait for dynamic content
	)
	if err != nil {
		return nil, fmt.Errorf("failed to navigate to %s: %v", url, err)
	}

	// Step 2: Check if the button exists
	err = chromedp.Run(ps.ctx,
		chromedp.Evaluate(fmt.Sprintf(`
			document.querySelector('%s') !== null
		`, ps.buttonPath), &buttonExists),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check button existence: %v", err)
	}

	if !buttonExists {
		return nil, fmt.Errorf("button with path '%s' not found", ps.buttonPath)
	}

	fmt.Printf("Button found: %s\n", ps.buttonPath)

	// Step 3: Set up popup monitoring
	err = chromedp.Run(ps.ctx,
		chromedp.Evaluate(`
			// Store original state
			window.popupMonitor = {
				popupDetected: false,
				popupElement: null,
				originalHTML: document.documentElement.innerHTML
			};

			// Monitor for DOM changes (new popups/modals)
			const observer = new MutationObserver(function(mutations) {
				mutations.forEach(function(mutation) {
					mutation.addedNodes.forEach(function(node) {
						if (node.nodeType === 1) { // Element node
							// Check if it's a popup/modal
							const style = getComputedStyle(node);
							if (node.classList.contains('modal') || 
								node.classList.contains('popup') || 
								node.classList.contains('dialog') ||
								node.getAttribute('role') === 'dialog' ||
								style.position === 'fixed' ||
								style.position === 'absolute') {
								
								window.popupMonitor.popupDetected = true;
								window.popupMonitor.popupElement = node;
							}
						}
					});
				});
			});
			
			observer.observe(document.body, { 
				childList: true, 
				subtree: true 
			});
		`, nil),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set up popup monitoring: %v", err)
	}

	// Step 4: Click the button
	fmt.Println("Clicking button...")
	err = chromedp.Run(ps.ctx,
		chromedp.Click(ps.buttonPath, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second), // Wait for popup to appear
	)
	if err != nil {
		return nil, fmt.Errorf("failed to click button: %v", err)
	}

	// Step 5: Check if popup appeared and get its HTML
	var popupDetected bool
	err = chromedp.Run(ps.ctx,
		chromedp.Evaluate(`window.popupMonitor.popupDetected`, &popupDetected),
	)
	if err != nil {
		log.Printf("Warning: Failed to check popup status: %v", err)
	}

	if popupDetected {
		fmt.Println("Popup detected!")
		err = chromedp.Run(ps.ctx,
			chromedp.Evaluate(`
				window.popupMonitor.popupElement ? 
				window.popupMonitor.popupElement.outerHTML : 
				document.documentElement.innerHTML
			`, &popupHTML),
		)
	} else {
		fmt.Println("No popup detected, getting current page HTML...")
		err = chromedp.Run(ps.ctx,
			chromedp.OuterHTML("html", &popupHTML),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get popup HTML: %v", err)
	}

	// Step 6: Parse HTML with goquery and extract data
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(popupHTML))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML with goquery: %v", err)
	}

	// Step 7: Extract data using the provided selectors
	extractedData := make(map[string]string)
	for i, selector := range ps.dataSelectors {
		selectorKey := fmt.Sprintf("selector_%d", i+1)
		if strings.Contains(selector, "=") {
			// Handle custom selector with key (e.g., "title=.popup-title")
			parts := strings.SplitN(selector, "=", 2)
			selectorKey = parts[0]
			selector = parts[1]
		}

		doc.Find(selector).Each(func(j int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if text != "" {
				if j == 0 {
					extractedData[selectorKey] = text
				} else {
					extractedData[fmt.Sprintf("%s_%d", selectorKey, j+1)] = text
				}
			}

			// Also try to get attribute values if text is empty
			if text == "" {
				if val, exists := s.Attr("value"); exists {
					extractedData[selectorKey+"_value"] = val
				}
				if val, exists := s.Attr("data-code"); exists {
					extractedData[selectorKey+"_data_code"] = val
				}
				if val, exists := s.Attr("href"); exists {
					extractedData[selectorKey+"_href"] = val
				}
			}
		})
	}

	// Step 8: Create result
	result := &ScrapedData{
		ButtonPath:  ps.buttonPath,
		PopupData:   extractedData,
		PopupHTML:   popupHTML,
		ExtractedAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	return result, nil
}

// ScrapeMultipleButtons scrapes data from multiple buttons
func (ps *PopupScraper) ScrapeMultipleButtons(url string, buttonPaths []string) ([]*ScrapedData, error) {
	var results []*ScrapedData

	for i, buttonPath := range buttonPaths {
		fmt.Printf("\nProcessing button %d/%d: %s\n", i+1, len(buttonPaths), buttonPath)

		// Update button path for this iteration
		ps.buttonPath = buttonPath

		result, err := ps.ScrapePopupData(url)
		if err != nil {
			log.Printf("Error scraping button %s: %v", buttonPath, err)
			continue
		}

		results = append(results, result)

		// Close popup if it exists before next iteration
		chromedp.Run(ps.ctx,
			chromedp.Evaluate(`
				// Try common popup close methods
				const closeButtons = document.querySelectorAll(
					'.modal-close, .popup-close, .close, [aria-label="Close"], [aria-label="close"], .btn-close'
				);
				for (let btn of closeButtons) {
					if (btn.offsetParent !== null) { // visible
						btn.click();
						break;
					}
				}
				
				// Also try ESC key
				document.dispatchEvent(new KeyboardEvent('keydown', {key: 'Escape'}));
			`, nil),
			chromedp.Sleep(1*time.Second),
		)
	}

	return results, nil
}

// ExtractSpecificData extracts data using custom extraction logic
func (ps *PopupScraper) ExtractSpecificData(doc *goquery.Document, extractors map[string]string) map[string]string {
	result := make(map[string]string)

	for key, selector := range extractors {
		var values []string

		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if text != "" {
				values = append(values, text)
			}
		})

		if len(values) > 0 {
			result[key] = strings.Join(values, " | ")
		}
	}

	return result
}

func main() {
	// Configuration - Replace these with your actual values
	config := struct {
		url           string
		buttonPath    string
		dataSelectors []string
	}{
		url:        "https://lolchess.gg/meta", // Replace with your URL
		buttonPath: "button[data-copy]",        // Replace with your CSS path
		dataSelectors: []string{
			"team_code=.team-code",  // Custom key=selector format
			".popup-title",          // Simple selector
			".modal-body",           // Another selector
			"[data-clipboard-text]", // Attribute selector
			".code-display",         // Code display area
		},
	}

	fmt.Println("Starting popup scraper...")
	fmt.Printf("URL: %s\n", config.url)
	fmt.Printf("Button Path: %s\n", config.buttonPath)
	fmt.Printf("Data Selectors: %v\n", config.dataSelectors)

	// Create scraper
	scraper := NewPopupScraper(config.buttonPath, config.dataSelectors)
	defer scraper.Close()

	// Method 1: Scrape single button
	fmt.Println("\n=== Scraping Single Button ===")
	result, err := scraper.ScrapePopupData(config.url)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Success! Extracted %d data points:\n", len(result.PopupData))
		for key, value := range result.PopupData {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	// Method 2: Scrape multiple buttons (if you have multiple paths)
	fmt.Println("\n=== Scraping Multiple Buttons ===")
	buttonPaths := []string{
		"button[data-copy]:nth-child(1)", // Replace with your paths
		"button[data-copy]:nth-child(2)",
		".copy-btn:first-of-type",
	}

	results, err := scraper.ScrapeMultipleButtons(config.url, buttonPaths)
	if err != nil {
		log.Printf("Error scraping multiple buttons: %v", err)
	} else {
		fmt.Printf("Successfully scraped %d buttons\n", len(results))
		for i, res := range results {
			fmt.Printf("\nButton %d (%s):\n", i+1, res.ButtonPath)
			for key, value := range res.PopupData {
				fmt.Printf("  %s: %s\n", key, value)
			}
		}
	}

	fmt.Println("\nScraping completed!")
}
