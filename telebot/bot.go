package telebot

import (
	"fmt"
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TeleBot struct {
	bot    *tgbotapi.BotAPI
	chatId int64
}

func NewTeleBot(token string, chatId int64) (*TeleBot, error) {

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	// bot.Debug = true

	return &TeleBot{
		bot:    bot,
		chatId: chatId,
	}, nil
}

/*
ch1 - 단순 메시지 교환
ch2 - 추천 덱
ch3 - 완료 덱
*/
func (t TeleBot) Run(offset int, chReq chan<- string, chMsg <-chan string, chDec <-chan DecMsg, chId chan<- DecResp) { // channel 받아

	// 텔레그램 updates 지속 수행
	u := tgbotapi.NewUpdate(offset)
	u.Timeout = 60
	updates := t.bot.GetUpdatesChan(u)

	go func() {
		for true {
			select {
			case msg := <-chMsg:
				t.SendMessage(msg)
			case resp := <-chDec:
				for i := 0; i < len(resp.Title); i++ {
					t.sendOptions(resp.Title[i], resp.Rcmds[i], resp.Ids[i])
				}
			}
		}
	}()

	for update := range updates {
		if update.Message != nil {

			switch update.Message.Text {
			case "/mode": // 현재 모드
				chReq <- "mode"
			case "/switch": // 모드 전환
				chReq <- "switch"
			case "/update":
				chReq <- "update"
			case "/reset":
				chReq <- "reset"
			case "/done":
				chReq <- "done"
			}
		}

		if update.CallbackQuery != nil {

			resp := DecResp{}
			var buttonText string
			if update.CallbackQuery.Message.Text == "완료 목록" {
				resp.Type = "restore"
				buttonText = "RESTORE"
			} else {
				resp.Type = "add"
				buttonText = "COMPLETE"
			}

			// Update the inline keyboard with the new checkbox state
			newKeyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(buttonText, update.CallbackQuery.Data),
				),
			)
			// Edit the message with the updated keyboard
			editMsg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, newKeyboard)
			if _, err := t.bot.Send(editMsg); err != nil {
				t.SendMessage("Callback 오류. " + err.Error())
				continue
			}

			decID, _ := strconv.Atoi(update.CallbackQuery.Data)
			resp.Id = decID

			chId <- resp
		}
	}
}

func (t TeleBot) sendOptions(title string, msgs []string, ids []int) {

	msg := tgbotapi.NewMessage(t.chatId, title)

	buttons := make([][]tgbotapi.InlineKeyboardButton, len(msgs))
	for i := 0; i < len(msgs); i++ {
		buttons[i] = tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("☑️"+msgs[i], fmt.Sprintf("%d", ids[i])),
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

func (t TeleBot) SendMessage(msg string) {
	t.bot.Send(tgbotapi.NewMessage(t.chatId, msg))
}
