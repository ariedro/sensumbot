package main

import (
	"log"
	"time"
)

var UpTime = time.Now()

func main() {
	LoadData()
	bot := CreateBot(Configs.BotToken)
	log.Println("Bot started")

	go telegramPoll(bot)
	SensumPoll(bot)
}
