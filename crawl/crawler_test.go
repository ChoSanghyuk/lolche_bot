package crawl

import (
	"fmt"
	"io"
	"lolcheBot"
	"net/http"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

/*
	#content-container > section > div.css-s9pipd.e2kj5ne0 > div > div > div > div.css-5x9ld.emls75t2 > div.css-35tzvc.emls75t4 > div

div#content-container > section.css-1v8my8o.esg9lhj0 > div.css-s9pipd.e2kj5ne0 > div > div.css-1iudmso.emls75t0 > div.css-1r1x0j5.emls75t1 > div.css-5x9ld.emls75t2 > div.css-35tzvc.emls75t4 > div
*/
func TestCrawl(t *testing.T) {
	url := "https://lolchess.gg/meta"
	css := "div#content-container > section.css-1v8my8o.esg9lhj0 > div.css-s9pipd.e2kj5ne0 > div > div.css-1iudmso.emls75t0 > div.css-1r1x0j5.emls75t1 > div.css-5x9ld.emls75t2 > div.css-35tzvc.emls75t4 > div"

	rtn, err := crawlTexts(url, css)
	if err != nil {
		t.Error(err)
	}
	for _, dec := range rtn {
		fmt.Println(dec)
	}
}

func TestMeta(t *testing.T) {

	crwaler := New()
	t.Run("Main Mode", func(t *testing.T) {
		rtn, err := crwaler.Meta(lolcheBot.MainMode)
		if err != nil {
			t.Error(err)
		}
		for _, dec := range rtn {
			fmt.Println(dec)
		}
	})

	t.Run("Pbe Mode", func(t *testing.T) {
		rtn, err := crwaler.Meta(lolcheBot.PbeMode)
		if err != nil {
			t.Error(err)
		}
		for _, dec := range rtn {
			fmt.Println(dec)
		}
	})
}

func TestDeckUrl(t *testing.T) {

	crwaler := New()
	rtn, err := crwaler.DeckUrl(lolcheBot.MainMode, "1")
	if err != nil {
		t.Error(err)
	}

	fmt.Println(rtn)
}

func TestCssPath(t *testing.T) {
	path, err := cssPath("https://lolchess.gg/meta", "초반 빌드업 요약")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(path)
}

///builder/guide/32354bc9cb5d579cfaad177dfd9e1eac363fedc9?type=guide
///builder/guide/1645b1a4dd615b928c293e4647b61c0e6323cded?type=guide

func TestPrintCssTarget(t *testing.T) {
	url := "https://lolchess.gg/meta"
	// css := "#content-container > section > div.css-s9pipd.e2kj5ne0 > div:nth-child(1) > div > div > div.css-1vo3wqf.emls75t3 > div.css-zg1vud.emls75t5 > button"
	css := "#content-container > section > div.css-s9pipd.e2kj5ne0 > div:nth-child(5) > div > div > div.css-1vo3wqf.emls75t3 > div.css-cchicn.emls75t7 > div.link-wrapper > a"

	// Send the request
	res, err := http.Get(url)
	if err != nil {
		t.Errorf("error making request\n%s", err)
	}

	defer res.Body.Close()

	// Check the response status
	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		t.Errorf("status code error: %d %s %s", res.StatusCode, res.Status, body)
	}

	// Create a goquery document from the response body
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		t.Errorf("error creating document\n%s", err)
	}

	doc.Find(css).Each(func(i int, s *goquery.Selection) {
		// Extract and print the data
		href, exists := s.Attr("href")
		if exists {
			fmt.Printf("  Class match: %s\n", href)
		}
	})

	// Find the elements by the CSS selector
	// fmt.Printf("%s", doc.Find(css).Text())
}
func TestPrintCssTarget2(t *testing.T) {
	url := "https://lolchess.gg/meta"
	// Send the request
	res, err := http.Get(url)
	if err != nil {
		t.Errorf("error making request\n%s", err)
	}

	defer res.Body.Close()

	// Check the response status
	body, _ := io.ReadAll(res.Body)
	fmt.Println(string(body))

	// Find the elements by the CSS selector
}

func TestPopupCrawl(t *testing.T) {
	pop := NewPopupScraper("#content-container > section > div.css-s9pipd.e2kj5ne0 > div:nth-child(1) > div > div > div.css-1vo3wqf.emls75t3 > div.css-zg1vud.emls75t5 > button",
		[]string{"#__next > div.Modal.isOpen.center.css-a59nmb.e1d1zcva0 > div.css-1hglyba.e1d1zcva1 > div > div.css-6r2fzw.e1dus1pt2 > div.css-o4ip4g.e1dus1pt4 "})
	rtn, err := pop.ScrapePopupData("https://lolchess.gg/meta")
	if err != nil {
		t.Error(err)
	}
	for key, value := range rtn.PopupData {
		fmt.Printf("  %s: %s\n", key, value)
	}
}

// 0201816617e16e16419a19915e16b000TFTSet15
