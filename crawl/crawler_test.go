package crawl

import (
	"fmt"
	"testing"
)

/*
#__next > div > div.css-1x48m3k.eetc6ox0 > div.content > div > section > div.css-s9pipd.e2kj5ne0 > div > div > div > div.css-5x9ld.emls75t2 > div.css-35tzvc.emls75t4 > div
#content-container > section > div.css-s9pipd.e2kj5ne0 > div:nth-child(2) > div > div > div.css-5x9ld.emls75t2 > div.css-35tzvc.emls75t4 > div
*/
func TestCrawl(t *testing.T) {
	url := "https://lolchess.gg/meta"
	css := "#content-container > section > div.css-s9pipd.e2kj5ne0 > div > div > div > div.css-5x9ld.emls75t2 > div.css-35tzvc.emls75t4 > div"

	crwaler := Crawler{}
	rtn, err := crwaler.crawl(url, css)
	if err != nil {
		t.Error(err)
	}
	for _, dec := range rtn {
		fmt.Println(dec)
	}
}

func TestCurrentMeta(t *testing.T) {

	crwaler := New()
	rtn, err := crwaler.CurrentMeta()
	if err != nil {
		t.Error(err)
	}
	for _, dec := range rtn {
		fmt.Println(dec)
	}
}

func TestPbeMeta(t *testing.T) {

	crwaler := New()
	rtn, err := crwaler.PbeMeta()
	if err != nil {
		t.Error(err)
	}
	for _, dec := range rtn {
		fmt.Println(dec)
	}
}
