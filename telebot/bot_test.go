package telebot

import (
	"fmt"
	"lolcheBot/config"
	"testing"
)

func TestBot(t *testing.T) {
	Temp()
}

func TestRun(t *testing.T) {
	conf, err := config.NewConfig()
	if err != nil {
		panic(err)
	}
	tele, err := NewTeleBot(conf.Telebot())
	if err != nil {
		t.Error(err)
	}

	chReq := make(chan string)
	chMsg := make(chan string)
	chDec := make(chan DecMsg)
	chId := make(chan DecResp)
	go func() {
		tele.Run(0, chReq, chMsg, chDec, chId)
	}()
	chDec <- DecMsg{
		Title: []string{"TITLE1", "TITLE2"},
		Rcmds: [][]string{{"dec1", "dec2"}, {"dec3", "dec4"}},
		Ids:   [][]int{{1, 2}, {3, 4}},
	}
	fmt.Println(<-chId)
}

func TestSendOptions(t *testing.T) {
	tele, err := NewTeleBot("", 0)
	if err != nil {
		t.Error(err)
	}
	tele.sendOptions("HI", []string{"덱1", "덱2"}, []int{1, 2})
}
