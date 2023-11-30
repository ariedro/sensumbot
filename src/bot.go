package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func addNewReceiver(receiver Receiver) {
	CachedReceivers = append(CachedReceivers, receiver)
}

func telegramPoll(bot *tgbotapi.BotAPI) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		switch update.Message.Text {
		case Configs.Commands.Start:
			addNewReceiver(Receiver{ChatID: update.Message.Chat.ID})
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Agrega2"))
		case Configs.Commands.Version:
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, GetVersion()))
		default:
			continue
		}

	}
}

func formatMessage(sensation Sensation) string {
	return sensation.Message + "\n" + sensation.Author
}

func SendMessage(sensation Sensation, receiver Receiver, receiverIndex int, bot *tgbotapi.BotAPI) error {
	messageText := formatMessage(sensation)
	msg := tgbotapi.NewMessage(receiver.ChatID, messageText)
	msg.ParseMode = "HTML"
	_, err := bot.Send(msg)
	if err != nil {
		return err
	}

	return nil
}

func CreateBot(token string) *tgbotapi.BotAPI {
	bot, err := tgbotapi.NewBotAPI(Configs.BotToken)
	if err != nil {
		log.Panic(err)
		panic(err)
	}

	return bot
}
