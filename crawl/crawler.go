package crawl

import (
	"fmt"
	"io"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

type Crawler struct {
	mainUrl string
	pbeUrl  string
	cssPath string
}

func New() Crawler {

	main := "https://lolchess.gg/meta"
	pbe := "https://lolchess.gg/meta?pbe=true"
	css := "#content-container > section > div.css-s9pipd.e2kj5ne0 > div > div > div > div.css-5x9ld.emls75t2 > div.css-35tzvc.emls75t4 > div"

	return Crawler{
		mainUrl: main,
		pbeUrl:  pbe,
		cssPath: css,
	}
}

func (c Crawler) CurrentMeta() ([]string, error) {
	return c.crawl(c.mainUrl, c.cssPath)
}

func (c Crawler) PbeMeta() ([]string, error) {
	return c.crawl(c.pbeUrl, c.cssPath)
}

func (s Crawler) crawl(url string, cssPath string) ([]string, error) {

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

	var v []string
	// Find the elements by the CSS selector
	doc.Find(cssPath).Each(func(i int, s *goquery.Selection) {
		// Extract and print the data
		v = append(v, s.Text())
	})

	return v, nil
}
