package telebot

import (
	"fmt"
	"os"
	"strconv"
	"testing"
)

func TestBot(t *testing.T) {
	Temp()
}

func TestRun(t *testing.T) {

	token := os.Getenv("token")
	chatIdTemp := os.Getenv("chatId")
	chatId, _ := strconv.ParseInt(chatIdTemp, 10, 64)

	tele, err := NewTeleBot(&TeleBotConfig{
		token:  token,
		chatId: chatId,
	})
	if err != nil {
		t.Error(err)
	}

	t.Run("chanel_test", func(t *testing.T) {
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

	})

	t.Run("telegram_send", func(t *testing.T) {
		tele.sendOptions("HI", []string{"덱1", "덱2"}, []int{1, 2})
	})

}
