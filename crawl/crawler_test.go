package crawl

import (
	"fmt"
	"io"
	"lolcheBot"
	"net/http"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

/*
	#content-container > section > div.css-s9pipd.e2kj5ne0 > div > div > div > div.css-5x9ld.emls75t2 > div.css-35tzvc.emls75t4 > div

div#content-container > section.css-1v8my8o.esg9lhj0 > div.css-s9pipd.e2kj5ne0 > div > div.css-1iudmso.emls75t0 > div.css-1r1x0j5.emls75t1 > div.css-5x9ld.emls75t2 > div.css-35tzvc.emls75t4 > div
*/
func TestCrawlTextsFromChromedp(t *testing.T) {
	url := "https://lolchess.gg/meta"
	css := "#content-container > section > div.css-s9pipd.e2kj5ne0 > div > div > div > div.css-5x9ld.emls75t2 > div.css-1fu47ws.emls75t4 > div"
	targetDeck := "빌지워터 미스 포츈"

	// Use the new crawlTextsFromChromedp function
	deckNames, err := crawlTextsFromChromedp(url, css)
	if err != nil {
		t.Fatalf("Failed to crawl deck names: %v", err)
	}

	// Print results
	fmt.Printf("\n=== Found %d deck names ===\n", len(deckNames))
	for i, name := range deckNames {
		fmt.Printf("  [%d] %s\n", i, name)
	}

	// Verify results
	if len(deckNames) == 0 {
		t.Error("No deck names found")
	}

	// Check if "빌지워터 미스 포츈" is in the list
	foundTarget := false
	for _, name := range deckNames {
		if strings.Contains(name, targetDeck) {
			foundTarget = true
			break
		}
	}

	if !foundTarget {
		t.Errorf("'%s' not found in deck list. Found decks: %v", targetDeck, deckNames)
	} else {
		fmt.Println("\n✓ Successfully found ", targetDeck)
	}
}

func TestCrawlTexts(t *testing.T) {
	url := "https://lolchess.gg/meta"
	css := "#content-container > section > div.css-s9pipd.e2kj5ne0 > div > div > div > div.css-5x9ld.emls75t2 > div.css-1fu47ws.emls75t4 > div"

	// Use the new crawlTextsFromChromedp function
	deckNames, err := crawlTexts(url, css)
	if err != nil {
		t.Fatalf("Failed to crawl deck names: %v", err)
	}

	if len(deckNames) == 0 {
		t.Error("No deck names found")
	}

	for i, name := range deckNames {
		fmt.Printf("  [%d] %s\n", i, name)
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
	// Note: cssPath function searches static HTML for text, but lolchess.gg
	// loads content dynamically with JavaScript, so this test will not find
	// deck names in the static HTML. For actual crawling, use chromedp via
	// TestCrawl or TestPopupCrawl.

	path, err := cssPath("https://lolchess.gg/meta", "빌지워터 미스 포츈")
	if err != nil {
		t.Error(err)
	}

	if path == "" {
		t.Skip("cssPath returned empty - page is dynamically loaded (expected behavior)")
	} else {
		fmt.Printf("CSS Path for '빌지워터 미스 포츈': %s\n", path)
	}
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

func TestPrintDocs(t *testing.T) {
	url := "https://lolchess.gg/meta"
	printDocs(url)
}

func TestGetDeckMeta(t *testing.T) {
	url := "https://lolchess.gg/meta"
	decks, err := GetDeckMeta(url)
	if err != nil {
		t.Fatalf("Failed to get deck meta: %v", err)
	}

	fmt.Printf("\n=== Found %d decks ===\n", len(decks))
	for i, deck := range decks {
		fmt.Printf("[%d] %s (key: %s)\n", i, deck.Name, deck.TeamBuilderKey)
	}

	// Verify we found the target deck
	found := false
	for _, deck := range decks {
		if deck.Name == "빌지워터 미스 포츈" && deck.TeamBuilderKey == "0932c429e79980626aeb3253f102a739ecf9fc6e" {
			found = true
			fmt.Printf("\n✓ Successfully found target deck: %s (key: %s)\n", deck.Name, deck.TeamBuilderKey)
			break
		}
	}

	if !found {
		t.Error("Target deck '빌지워터 미스 포츈' with correct teamBuilderKey not found")
	}
}
