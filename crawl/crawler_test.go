package crawl

import (
	"fmt"
	"testing"
)

/*
	#content-container > section > div.css-s9pipd.e2kj5ne0 > div > div > div > div.css-5x9ld.emls75t2 > div.css-35tzvc.emls75t4 > div

div#content-container > section.css-1v8my8o.esg9lhj0 > div.css-s9pipd.e2kj5ne0 > div > div.css-1iudmso.emls75t0 > div.css-1r1x0j5.emls75t1 > div.css-5x9ld.emls75t2 > div.css-35tzvc.emls75t4 > div
*/
func TestCrawl(t *testing.T) {
	url := "https://lolchess.gg/meta"
	css := "div#content-container > section.css-1v8my8o.esg9lhj0 > div.css-s9pipd.e2kj5ne0 > div > div.css-1iudmso.emls75t0 > div.css-1r1x0j5.emls75t1 > div.css-5x9ld.emls75t2 > div.css-35tzvc.emls75t4 > div"

	rtn, err := crawl(url, css)
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

func TestCssPath(t *testing.T) {
	path, err := cssPath("https://lolchess.gg/meta", "초반 빌드업 요약")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(path)
}
