package main

import (
	"log"
	"strconv"

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

func SendMessage(sensation Sensation, receiver Receiver, receiverIndex int, bot *tgbotapi.BotAPI) error {
	messageText := sensation.Message + "\n~ " + sensation.Author + "\n -" + strconv.Itoa(sensation.Dislikes) + " +" + strconv.Itoa(sensation.Likes)
	messageID := findMessage(receiver.TrackedSensations, sensation.SensumID)
	// Post or edit?
	if messageID > 0 {
		edit := tgbotapi.EditMessageTextConfig{
			BaseEdit: tgbotapi.BaseEdit{
				ChatID:    receiver.ChatID,
				MessageID: messageID,
			},
			Text: messageText,
		}
		_, err := bot.Send(edit)
		if err != nil {
			return err
		}
	} else {
		msg := tgbotapi.NewMessage(receiver.ChatID, messageText)
		messageSent, err := bot.Send(msg)
		if err != nil {
			return err
		}
		// TODO: This shouldn't be here
		TrackSensation(receiverIndex, sensation, messageSent)
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
