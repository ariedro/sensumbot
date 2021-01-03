package main

import (
	"log"
)

func main() {
	LoadData()
	bot := CreateBot(Configs.BotToken)
	log.Println("Bot started")

	go telegramPoll(bot)
	SensumPoll(bot)
}
