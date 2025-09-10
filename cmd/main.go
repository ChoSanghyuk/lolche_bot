package main

import (
	"lolcheBot"
	"lolcheBot/config"
	"lolcheBot/crawl"
	"lolcheBot/db"
)

func main() {
	conf, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	crawler := crawl.New()
	db, err := db.NewStorage(conf.StorageConfig())
	if err != nil {
		panic(err)
	}

	bot, err := lolcheBot.NewTeleBot(conf.Telebot(), db, crawler)
	if err != nil {
		panic(err)
	}

	bot.Run()
}
