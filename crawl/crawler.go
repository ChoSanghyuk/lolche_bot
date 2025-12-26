package crawl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"lolcheBot"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

type Crawler struct {
	mainUrl     string
	pbeUrl      string
	cssPath     string
	deckCache   map[lolcheBot.Mode][]DeckMeta
	refreshTime map[lolcheBot.Mode]time.Time
}

func New() *Crawler {
	crawler := Crawler{
		mainUrl: "https://lolchess.gg/meta",
		pbeUrl:  "https://lolchess.gg/meta?pbe=true",
	}

	// err := crawler.UpdateCssPath("")
	// if err != nil {
	// 	crawler.cssPath = "#content-container > section > div.css-s9pipd.e2kj5ne0 > div > div > div > div.css-5x9ld.emls75t2 > div.css-1fu47ws.emls75t4 > div"
	// }

	go crawler.cleanCache()
	return &crawler
}

func (c *Crawler) cleanCache() {
	for key := range c.deckCache {
		if len(c.deckCache[key]) != 0 && c.refreshTime[key].Before(time.Now().Add(time.Minute*-5)) {
			c.deckCache[key] = nil
		}
		time.Sleep(10 * time.Minute)
	}
}

func (c *Crawler) Meta(mode lolcheBot.Mode) ([]string, error) {
	var deckMeta []DeckMeta
	var err error
	if mode == lolcheBot.MainMode {
		deckMeta, err = GetDeckMeta(c.mainUrl)
	} else {
		deckMeta, err = GetDeckMeta(c.pbeUrl)
	}
	if err != nil {
		return nil, fmt.Errorf("크롤링 실패. %w", err)
	}
	if len(deckMeta) == 0 {
		return nil, fmt.Errorf("크롤링 조회 결과 없음")
	}
	dec := make([]string, len(c.deckCache))
	for i, dm := range deckMeta {
		dec[i] = dm.Name
	}
	c.deckCache[mode] = deckMeta
	return dec, nil
}

func (c *Crawler) DeckBuilderUrl(mode lolcheBot.Mode, id int) (string, error) {
	var deckMeta []DeckMeta
	var err error
	if len(c.deckCache[mode]) == 0 {
		if mode == lolcheBot.MainMode {
			deckMeta, err = GetDeckMeta(c.mainUrl)
		} else {
			deckMeta, err = GetDeckMeta(c.pbeUrl)
		}
		if err != nil {
			return "", fmt.Errorf("크롤링 실패. %w", err)
		}
		if len(deckMeta) == 0 {
			return "", fmt.Errorf("크롤링 조회 결과 없음")
		}
	} else {
		deckMeta = c.deckCache[mode]
	}

	builderKey := deckMeta[id].TeamBuilderKey

	return "https://lolchess.gg/builder/guide/" + builderKey, nil
}

// DeckMeta represents a deck with its key and name
type DeckMeta struct {
	TeamBuilderKey string `json:"teamBuilderKey"`
	Name           string `json:"name"`
}

// GetDeckMeta fetches deck metadata (teamBuilderKey and name) from the lolchess.gg meta page
func GetDeckMeta(url string) ([]DeckMeta, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	return extractDecksFromJSON(string(body))
}

// deprecated. web page rendering 방식 변화로 첫 조회 시 html 형식으로 오지 않음
func (c *Crawler) UpdateCssPath(target string) error {

	if target == "" {
		target = "초반 빌드업 요약"
	}
	path, err := cssPath(c.mainUrl, target)
	if err != nil {
		return fmt.Errorf("css 경로 갱신 실패. %w", err)
	}

	sl, err := crawlTexts(c.mainUrl, path)
	if err != nil {
		return fmt.Errorf("css 경로로 조회 실패. path : %s, error: %w", path, err)
	}

	if len(sl) == 0 {
		return fmt.Errorf("잘못된 css 경로(조회된 덱 없음). path : %s, error: %w", path, err)
	}

	c.cssPath = path

	return nil
}

// deprecated. web page rendering 방식 변화로 첫 조회 시 html 형식으로 오지 않음
func crawlTexts(url string, cssPath string) ([]string, error) {

	// Send the request
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error making request\n%w", err)
	}

	defer res.Body.Close()

	// Check the response status
	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("status code error: %d %s %s", res.StatusCode, res.Status, body)
	}

	// Create a goquery document from the response body
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error creating document\n%w", err)
	}

	// fmt.Println(doc.Text())
	// fmt.Println(strings.Repeat("*", 100))

	var v []string
	// Find the elements by the CSS selector
	doc.Find(cssPath).Each(func(i int, s *goquery.Selection) {
		// Extract and print the data
		v = append(v, s.Text())
	})

	return v, nil
}

// deprecated. web page rendering 방식 변화로 첫 조회 시 html 형식으로 오지 않음
func (c Crawler) DeckUrl(mode lolcheBot.Mode, id string) (string, error) {

	var target string
	if mode == lolcheBot.MainMode {
		target = c.mainUrl
	} else {
		target = c.pbeUrl
	}
	cssPath := fmt.Sprintf("#content-container > section > div.css-s9pipd.e2kj5ne0 > div:nth-child(%s) > div > div > div.css-1vo3wqf.emls75t3 > div.css-cchicn.emls75t7 > div.link-wrapper > a", id)

	urlPath, err := crawlUrl(target, cssPath)
	if err != nil {
		return "", fmt.Errorf("크롤링 실패. %w", err)
	}
	url := "https://lolchess.gg" + urlPath

	return url, nil
}

// chrome browser가 조회하는 것처럼 조회하여 html 형식으로 받아옴. 단, 대기 시간이 존재하여 미사용.
func crawlTextsFromChromedp(url string, cssPath string) ([]string, error) {
	// Configure Chrome options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.WindowSize(1920, 1080),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Add timeout
	ctx, timeoutCancel := context.WithTimeout(ctx, 60*time.Second)
	defer timeoutCancel()

	// Navigate to page and get rendered HTML
	var htmlContent string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(5*time.Second), // Wait for dynamic content to load
		chromedp.OuterHTML("html", &htmlContent),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load page with chromedp: %w", err)
	}

	// Parse the rendered HTML with goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract text from elements matching the CSS selector
	var results []string
	doc.Find(cssPath).Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if len(text) > 0 && len(text) < 100 && !strings.Contains(text, "\n") {
			results = append(results, text)
		}
	})

	return results, nil
}

func crawlUrl(url string, cssPath string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error making request\n%w", err)
	}

	defer res.Body.Close()

	// Check the response status
	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("status code error: %d %s %s", res.StatusCode, res.Status, body)
	}

	// Create a goquery document from the response body
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", fmt.Errorf("error creating document\n%w", err)
	}

	// fmt.Println(doc.Text())
	// fmt.Println(strings.Repeat("*", 100))

	// Find the elements by the CSS selector
	var rtn string
	doc.Find(cssPath).Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			rtn = href
		}
	})

	if rtn == "" {
		return "", fmt.Errorf("크롤링 조회 결과 없음")
	}

	return rtn, nil
}

func cssPath(url string, target string) (string, error) {

	// Send the request
	res, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error making request\n%w", err)
	}

	defer res.Body.Close()

	// Check the response status
	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("status code error: %d %s %s", res.StatusCode, res.Status, body)
	}

	// Create a goquery document from the response body
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", fmt.Errorf("error creating document\n%w", err)
	}

	var path string
	doc.Find("div:not(:has(*))").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), target) {
			// Construct the CSS path
			path = constructCSSPath(s)
			// fmt.Printf("Found text: %s\nCSS Path: %s\n", target, path)
		}
	})
	return path, nil

}

func constructCSSPath(selection *goquery.Selection) string {
	path := ""

	for len(selection.Nodes) > 0 {
		node := selection.Nodes[0]
		if node == nil {
			break
		}

		tag := goquery.NodeName(selection)
		if tag == "" {
			break
		}

		id, _ := selection.Attr("id")
		if id != "" {
			path = fmt.Sprintf("%s#%s", tag, id) + " > " + path
			break
		}

		classes, _ := selection.Attr("class")
		classSelector := ""
		if classes != "" {
			classSelector = "." + strings.ReplaceAll(classes, " ", ".")
		}

		path = fmt.Sprintf("%s%s", tag, classSelector) + " > " + path

		selection = selection.Parent()
	}

	// Trim the trailing " > "
	return strings.TrimSuffix(path, " > ")
}

func printDocs(url string) error {

	// Send the request
	res, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error making request\n%w", err)
	}

	defer res.Body.Close()

	// Check the response status
	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("status code error: %d %s %s", res.StatusCode, res.Status, body)
	}

	// Create a goquery document from the response body
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return fmt.Errorf("error creating document\n%w", err)
	}

	fmt.Println(doc.Text())
	return nil
}

// extractDecksFromJSON parses the HTML and extracts deck metadata from embedded JSON
func extractDecksFromJSON(htmlContent string) ([]DeckMeta, error) {
	// Find the JSON data embedded in a script tag
	// Look for the pattern: <script id="__NEXT_DATA__" type="application/json">...JSON...</script>
	// Or just find {"props":{...}} with balanced braces

	// Find the start of the JSON
	startIdx := strings.Index(htmlContent, `{"props":`)
	if startIdx == -1 {
		return nil, fmt.Errorf("no JSON data found in HTML")
	}

	// Find the matching closing brace by counting braces
	braceCount := 0
	endIdx := startIdx
	for i := startIdx; i < len(htmlContent); i++ {
		if htmlContent[i] == '{' {
			braceCount++
		} else if htmlContent[i] == '}' {
			braceCount--
			if braceCount == 0 {
				endIdx = i + 1
				break
			}
		}
	}

	if braceCount != 0 {
		return nil, fmt.Errorf("unbalanced braces in JSON data")
	}

	jsonData := htmlContent[startIdx:endIdx]

	// Parse the JSON
	var data struct {
		Props struct {
			PageProps struct {
				DehydratedState struct {
					Queries []struct {
						State struct {
							Data struct {
								GuideDecks []DeckMeta `json:"guideDecks"`
							} `json:"data"`
						} `json:"state"`
					} `json:"queries"`
				} `json:"dehydratedState"`
			} `json:"pageProps"`
		} `json:"props"`
	}

	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Find the query that contains guideDecks
	for _, query := range data.Props.PageProps.DehydratedState.Queries {
		if len(query.State.Data.GuideDecks) > 0 {
			return query.State.Data.GuideDecks, nil
		}
	}

	return nil, fmt.Errorf("no guideDecks found in JSON data")
}
