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

func formatMessage(sensation Sensation) string {
	message := sensation.Message

	if IsTrending(sensation) {
		message = "<b>" + message + "</b>"
	}
	if ShouldBeDenied(sensation) {
		message = "<s>" + message + "</s>"
	}

	return message + "\n~ " +
		sensation.Author + "\n -" +
		strconv.Itoa(sensation.Dislikes) + " +" +
		strconv.Itoa(sensation.Likes)
}

func SendMessage(sensation Sensation, receiver Receiver, receiverIndex int, bot *tgbotapi.BotAPI) error {
	messageText := formatMessage(sensation)
	messageID := findMessage(receiver.TrackedSensations, sensation.SensumID)
	// Post or edit?
	if messageID > 0 {
		edit := tgbotapi.EditMessageTextConfig{
			BaseEdit: tgbotapi.BaseEdit{
				ChatID:    receiver.ChatID,
				MessageID: messageID,
			},
			Text:      messageText,
			ParseMode: "HTML",
		}
		_, err := bot.Send(edit)
		if err != nil {
			return err
		}
	} else {
		msg := tgbotapi.NewMessage(receiver.ChatID, messageText)
		msg.ParseMode = "HTML"
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
