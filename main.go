package main

import (
	"fmt"
	"lolcheBot/config"
	"lolcheBot/crawl"
	"lolcheBot/db"
	"lolcheBot/telebot"
	"strings"
)

func main() {
	conf, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	bot, err := telebot.NewTeleBot(conf.Telebot())
	if err != nil {
		panic(err)
	}
	crawler := crawl.Crawler{}
	db, err := db.NewStorage()
	if err != nil {
		panic(err)
	}

	var isMain bool = db.Mode()
	var decLi []string
	var doneLi []string

	chReq := make(chan string)
	chMsg := make(chan string)
	chDec := make(chan telebot.DecMsg)
	chResp := make(chan telebot.DecResp)
	go func() {
		bot.Run(0, chReq, chMsg, chDec, chResp)
	}()

	for true {
		select {
		case msg := <-chReq:
			switch msg {
			case "mode":
				if isMain {
					chMsg <- "정규 모드"
				} else {
					chMsg <- "PBE 모드"
				}
			case "switch":
				isMain = !isMain
				db.SaveMode(isMain)
				rtn := "모드 변환 완료. 현재 모드: "
				if isMain {
					rtn += "정규 모드"
				} else {
					rtn += "PBE 모드"
				}
				chMsg <- rtn
			case "update":

				if isMain {
					decLi, err = crawler.CurrentMeta() // todo 여기서의 error 처리
					doneLi, err = db.AllMain()
				} else {
					decLi, err = crawler.PbeMeta() // todo 여기서의 error 처리
					doneLi, err = db.AllPbe()
				}
				if err != nil {
					chMsg <- fmt.Sprintf("오류 발생 %s", err.Error())
				} else {
					dec := makeDecRcmd(decLi, doneLi)
					if len(dec.Title) > 0 {
						chDec <- dec
					} else {
						chMsg <- "Congratulation! All Completed"
					}
				}
			case "done":
				if isMain {
					doneLi, err = db.AllMain()
				} else {
					doneLi, err = db.AllPbe()
				}
				if err != nil {
					chMsg <- fmt.Sprintf("오류 발생 %s", err.Error())
				} else {
					dec := makeDecDone(doneLi)
					chDec <- dec
				}
			case "reset":
				// todo. DB 삭제
				if isMain {
					err = db.DeleteAllMain()
					if err != nil {
						chMsg <- fmt.Sprintf("정규 모드 기록 삭제 오류 발생. %s", err.Error())
					}
					chMsg <- "정규 모드 기록 삭제 완료"
				} else {
					err = db.DeleteAllPbe()
					if err != nil {
						chMsg <- fmt.Sprintf("PBE 모드 기록 삭제 오류 발생. %s", err.Error())
					}
					chMsg <- "PBE 모드 기록 삭제 완료"
				}
			}

		case resp := <-chResp:
			if resp.Type == "restore" {
				if len(doneLi) == 0 {
					chMsg <- "만료된 캐시 정보. /done 수행 필요"
				}
				restore := doneLi[resp.Id]
				if isMain {
					db.DeleteMain(restore)
				} else {
					db.DeletePbe(restore)
				}

			} else {
				if len(decLi) == 0 {
					chMsg <- "만료된 캐시 정보. /update 수행 필요"
				}
				done := decLi[resp.Id]
				if isMain {
					db.SaveMain(done)
				} else {
					db.SavePbe(done)
				}
			}
		}
	}
}

func makeDecRcmd(decLi []string, doneLi []string) telebot.DecMsg {

	rtn := telebot.DecMsg{
		// Title: []string{},
		// Rcmds: [][]string{{}},
		// Ids:   [][]int{{}},
	}

	m := make(map[string]bool)
	for _, d := range doneLi {
		m[d] = true
	}

	selected := false
	specialDec := make([]string, 0)
	specailIdx := make([]int, 0)

	for i := len(decLi) - 1; i >= 0; i-- {
		if strings.HasPrefix(decLi[i], "[") && !strings.Contains(decLi[i], "[상징]") && !m[decLi[i]] {
			specialDec = append(specialDec, decLi[i])
			specailIdx = append(specailIdx, i)
		} else if !selected && !m[decLi[i]] {
			rtn.Title = append(rtn.Title, "일반 덱")
			rtn.Rcmds = append(rtn.Rcmds, []string{decLi[i]})
			rtn.Ids = append(rtn.Ids, []int{i})
			selected = true
		}
	}

	if len(specialDec) > 0 {
		rtn.Title = append(rtn.Title, "증강 덱")
		rtn.Rcmds = append(rtn.Rcmds, specialDec)
		rtn.Ids = append(rtn.Ids, specailIdx)
	}
	return rtn

}

func makeDecDone(doneLi []string) telebot.DecMsg {

	ids := make([]int, len(doneLi))
	for i := range ids {
		ids[i] = i
	}

	rtn := telebot.DecMsg{
		Title: []string{"완료 목록"},
		Rcmds: [][]string{doneLi},
		Ids:   [][]int{ids},
	}

	return rtn
}
