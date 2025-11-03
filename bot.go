package lolcheBot

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TeleBot struct {
	bot    *tgbotapi.BotAPI
	chatId int64
	stg    Stoage
	dc     DeckCrawler
}

func NewTeleBot(conf *TeleBotConfig, stg Stoage, dc DeckCrawler) (*TeleBot, error) {

	bot, err := tgbotapi.NewBotAPI(conf.token)
	if err != nil {
		return nil, err
	}
	// bot.Debug = true

	return &TeleBot{
		bot:    bot,
		chatId: conf.chatId,
		stg:    stg,
		dc:     dc,
	}, nil
}

type TeleBotConfig struct {
	token  string
	chatId int64
}

func NewTeleBotConfig(token string, chatId int64) *TeleBotConfig {

	return &TeleBotConfig{
		token:  token,
		chatId: chatId,
	}
}

var candidateDeckMap map[string]string = map[string]string{}

// todo deck index +1
func (t TeleBot) Run() { // channel 받아
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := t.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			switch Command(update.Message.Text) {
			case help:
				t.helpJob()
			case mode:
				t.modeJob()
			case switching:
				t.switchJob()
			case updating:
				t.updateJob()
			case reset:
				t.resetJob()
			case done:
				t.doneJob()
			case fix:
				t.fixJob()
			default:
				t.SendMessage("미등록 작업")
			}

		}

		if update.CallbackQuery != nil {
			switch update.CallbackQuery.Message.Text {
			case titleNormalDeck, titleSpecDeck:
				t.selectJob(&update)
			case titleWhetherCompleted:
				t.completeJob(&update)
			case titleCompletionList:
				t.restoreJob(&update)
			default:
				t.SendMessage("세션 완료. /update로 덱 갱신 필요")
			}

		}
	}

}

func (t TeleBot) helpJob() {
	cmds := ""
	for i, c := range allCommands() {
		cmds += string(c)
		if i != len(allCommands())-1 {
			cmds += "\n"
		}
	}
	t.SendMessage(cmds)
}

func (t TeleBot) modeJob() {
	mode := t.stg.Mode()
	t.SendMessage("현재 모드: " + mode.Str())
}

func (t TeleBot) switchJob() {
	mode := t.stg.Mode()
	mode = !mode
	t.stg.SaveMode(mode)

	t.SendMessage("모드 변환 완료. 현재 모드: " + mode.Str())
}

func (t TeleBot) updateJob() {
	mode := t.stg.Mode()

	decLi, err := t.dc.Meta(mode)
	doneLi, _ := t.stg.All(mode)

	if err != nil {
		t.SendMessage(fmt.Sprintf("오류 발생 %s", err.Error()))
	} else {
		decs := makeDecRcmd(decLi, doneLi)
		if len(decs) > 0 {
			for i := 0; i < len(decs); i++ {
				t.sendOptions(&decs[i])
				for j := 0; j < len(decs[i].Rcmds); j++ {
					candidateDeckMap[strconv.Itoa(decs[i].Ids[j])] = decs[i].Rcmds[j]
				}
			}

		} else {
			t.SendMessage("Congratulation! All Completed")
		}
	}
}

func (t TeleBot) resetJob() { // todo. 지우기전에 한번 물어봐
	mode := t.stg.Mode()
	err := t.stg.DeleteAll(mode)
	if err != nil {
		t.SendMessage(fmt.Sprintf("%s 기록 삭제 오류 발생. %s", mode.Str(), err.Error()))
	}
	t.SendMessage(fmt.Sprintf("%s 기록 삭제 완료", mode.Str()))
}

func (t TeleBot) doneJob() { // todo. 지우기전에 한번 물어봐
	mode := t.stg.Mode()
	doneLi, err := t.stg.All(mode)
	if err != nil {
		t.SendMessage(fmt.Sprintf("오류 발생 %s", err.Error()))
		return
	}

	dec := makeDecDone(doneLi)
	t.sendOptions(&dec)

}

func (t TeleBot) fixJob() {
	err := t.dc.UpdateCssPath("")
	if err != nil {
		t.SendMessage(fmt.Sprintf("fix 실패. %s", err.Error()))
	}
}

func (t TeleBot) restoreJob(update *tgbotapi.Update) {

	// Update the inline keyboard with the new checkbox state
	newKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("RESTORE", update.CallbackQuery.Data),
		),
	)
	// Edit the message with the updated keyboard
	editMsg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, newKeyboard)
	if _, err := t.bot.Send(editMsg); err != nil {
		t.SendMessage("Callback 오류. " + err.Error())
		return
	}

	restoreTarget := update.CallbackQuery.Message.Text // todo.이거 Text 부분 가져오는거 맞는지 확인
	mode := t.stg.Mode()
	t.stg.DeleteByName(mode, restoreTarget)
}

func (t TeleBot) selectJob(update *tgbotapi.Update) {

	// 여기서는 url 정보 한번 보내고
	idx := update.CallbackQuery.Data
	mode := t.stg.Mode()
	url, err := t.dc.DeckUrl(mode, idx)
	if err != nil {
		t.SendMessage("Deck url 가져오기 오류. " + err.Error())
	}
	t.SendMessage(url)

	i, _ := strconv.Atoi(idx)

	// 완료버튼에 data 부터 덱명 담아서 보내야함.
	t.sendOptions(&DecOptMsg{
		Title: titleWhetherCompleted,
		Rcmds: []string{candidateDeckMap[idx]},
		Ids:   []int{i},
	})
}

func (t TeleBot) completeJob(update *tgbotapi.Update) {

	// Update the inline keyboard with the new checkbox state
	newKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("SUCCESSFULLY COMPLETED", update.CallbackQuery.Data),
		),
	)
	// Edit the message with the updated keyboard
	editMsg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, newKeyboard)
	if _, err := t.bot.Send(editMsg); err != nil {
		t.SendMessage("Callback 오류. " + err.Error())
		return
	}

	doneNum := update.CallbackQuery.Data

	mode := t.stg.Mode()
	t.stg.Save(mode, candidateDeckMap[doneNum])
}

func (t TeleBot) SendMessage(msg string) { // todo. private.
	t.bot.Send(tgbotapi.NewMessage(t.chatId, msg))
}

/*
덱이 현재 목록에서 몇 번째인지 data로 보내고, 그걸 활용해서 덱의 url 가져오자
*/
func (t TeleBot) sendOptions(optMsg *DecOptMsg) {

	msg := tgbotapi.NewMessage(t.chatId, optMsg.Title)

	buttons := make([][]tgbotapi.InlineKeyboardButton, len(optMsg.Rcmds))
	for i := 0; i < len(optMsg.Rcmds); i++ {
		buttons[i] = tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(optMsg.Rcmds[i], strconv.Itoa(optMsg.Ids[i])),
		)
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		buttons...,
	)

	msg.ReplyMarkup = keyboard

	if _, err := t.bot.Send(msg); err != nil {
		log.Panic(err)
	}
}

func makeDecRcmd(decLi []string, doneLi []string) []DecOptMsg {

	rtn := []DecOptMsg{}

	m := make(map[string]bool)
	for _, d := range doneLi {
		m[d] = true
	}

	selected := false
	specialDec := make([]string, 0)
	specailIdx := make([]int, 0)

	for i := len(decLi) - 1; i >= 0; i-- {
		if strings.HasPrefix(decLi[i], "[") && !m[decLi[i]] { // && !strings.Contains(decLi[i], "[상징]")
			specialDec = append(specialDec, decLi[i])
			specailIdx = append(specailIdx, i+1) // index 보정. 이 숫자로 몇 번째 덱 link 가져오는지 정하기 때문에 css nth-child의 숫자에 맞게 보정
		} else if !selected && !m[decLi[i]] {
			rtn = append(rtn, DecOptMsg{
				Title: titleNormalDeck,
				Rcmds: []string{decLi[i]},
				Ids:   []int{i + 1}, // index 보정
			})
			selected = true
		}
	}

	if len(specialDec) > 0 {
		rtn = append(rtn, DecOptMsg{
			Title: titleSpecDeck,
			Rcmds: specialDec,
			Ids:   specailIdx,
		})
	}

	return rtn

}
func makeDecDone(doneLi []string) DecOptMsg {

	ids := make([]int, len(doneLi))
	for i := range ids {
		ids[i] = i
	}

	return DecOptMsg{
		Title: titleCompletionList,
		Rcmds: doneLi,
		Ids:   ids,
	}
}

/***************************************************************** DELETE *******************************************************************************************/

/*
ch1 - 단순 메시지 교환
ch2 - 추천 덱
ch3 - 완료 덱
*/
// func (t TeleBot) Run(offset int, chReq chan<- string, chMsg <-chan string, chDec <-chan DecMsg, chId chan<- DecResp) { // channel 받아

// 	// 텔레그램 updates 지속 수행
// 	u := tgbotapi.NewUpdate(offset)
// 	u.Timeout = 60
// 	updates := t.bot.GetUpdatesChan(u)

// 	go func() {
// 		for true {
// 			select {
// 			case msg := <-chMsg:
// 				t.SendMessage(msg)
// 			case resp := <-chDec:
// 				for i := 0; i < len(resp.Title); i++ {
// 					t.sendOptions(resp.Title[i], resp.Rcmds[i], resp.Ids[i])
// 				}
// 			}
// 		}
// 	}()

// 	for update := range updates {
// 		if update.Message != nil {

// 			switch update.Message.Text {
// 			case "/help":
// 				chReq <- "help"
// 			case "/mode": // 현재 모드
// 				chReq <- "mode"
// 			case "/switch": // 모드 전환
// 				chReq <- "switch"
// 			case "/update":
// 				chReq <- "update"
// 			case "/reset":
// 				chReq <- "reset"
// 			case "/done":
// 				chReq <- "done"
// 			case "/fix":
// 				chReq <- "fix"
// 			}
// 		}

// 		if update.CallbackQuery != nil {

// 			resp := DecResp{}
// 			var buttonText string
// 			if update.CallbackQuery.Message.Text == "완료 목록" {
// 				resp.Type = "restore"
// 				buttonText = "RESTORE"
// 			} else {
// 				resp.Type = "add"
// 				buttonText = "COMPLETE"
// 			}

// 			// Update the inline keyboard with the new checkbox state
// 			newKeyboard := tgbotapi.NewInlineKeyboardMarkup(
// 				tgbotapi.NewInlineKeyboardRow(
// 					tgbotapi.NewInlineKeyboardButtonData(buttonText, update.CallbackQuery.Data),
// 				),
// 			)
// 			// Edit the message with the updated keyboard
// 			editMsg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, newKeyboard)
// 			if _, err := t.bot.Send(editMsg); err != nil {
// 				t.SendMessage("Callback 오류. " + err.Error())
// 				continue
// 			}

// 			decID, _ := strconv.Atoi(update.CallbackQuery.Data)
// 			resp.Id = decID

// 			chId <- resp
// 		}
// 	}
// }

func Temp() {
	bot, err := tgbotapi.NewBotAPI("")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true // Enable debug logging

	// Set up the bot to handle updates
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Send a message with "checkbox" style inline keyboard buttons
	msg := tgbotapi.NewMessage(0, "Choose your options:") // Replace 123456789 with the chat ID

	// Create inline keyboard buttons (initially unchecked)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("sjlfe", "1"), //☑️Done
		),
	)

	// Attach the inline keyboard to the message
	msg.ReplyMarkup = keyboard

	// Send the message
	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}

	// Handle updates
	for update := range updates {
		if update.CallbackQuery != nil {
			// Handle button clicks

			// Get callback data (which button was clicked)
			// optionID, _ := strconv.Atoi(update.CallbackQuery.Data)

			// Toggle button state (checkbox logic)
			// buttonText := "☑️ Option " + update.CallbackQuery.Data
			// if update.CallbackQuery.Data == strconv.Itoa(optionID) {
			// 	buttonText = "✅ Option " + update.CallbackQuery.Data // Change ☑️ to ✅ (checked)
			// }
			buttonText := "COMPLETE"

			// Update the inline keyboard with the new checkbox state
			newKeyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(buttonText, update.CallbackQuery.Data),
				),
			)

			// Edit the message with the updated keyboard
			editMsg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, newKeyboard)
			if _, err := bot.Send(editMsg); err != nil {
				log.Panic(err)
			}
		}
	}
}
