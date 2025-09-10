package crawl

import (
	"fmt"
	"io"
	"lolcheBot"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Crawler struct {
	mainUrl string
	pbeUrl  string
	cssPath string
}

func New() *Crawler {
	crawler := Crawler{
		mainUrl: "https://lolchess.gg/meta",
		pbeUrl:  "https://lolchess.gg/meta?pbe=true",
	}

	err := crawler.UpdateCssPath("")
	if err != nil {
		crawler.cssPath = "#content-container > section > div.css-s9pipd.e2kj5ne0 > div > div > div > div.css-5x9ld.emls75t2 > div.css-35tzvc.emls75t4 > div"
	}

	return &crawler
}

func (c Crawler) Meta(mode lolcheBot.Mode) (dec []string, err error) {
	if mode == lolcheBot.MainMode {
		dec, err = crawlTexts(c.mainUrl, c.cssPath)
	} else {
		dec, err = crawlTexts(c.pbeUrl, c.cssPath)
	}
	if err != nil {
		return nil, fmt.Errorf("크롤링 실패. %w", err)
	}
	if len(dec) == 0 {
		return nil, fmt.Errorf("크롤링 조회 결과 없음")
	}
	return dec, nil
}

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
