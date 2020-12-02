package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const TOKEN = "SECRET"
const BOT_ID = "SECRET"
const START_COMMAND = "/start@sensum_bot"
const CHAT_IDS_FILE = "./sensumbot_chat_ids"
const SENSUM_URL = "https://sensum-server.herokuapp.com/api/sensations/letThemFlow"

var last_id = "0"
var last_post_id = ""

type Sensation struct {
	ID        string
	Author    string
	Message   string
	Likes     int
	Dislikes  int
	Timestamp time.Time
}

func getLastSensation() Sensation {
	requestBody, err := json.Marshal(map[string]interface{}{"offset": 0, "limit": 1})
	req, err := http.NewRequest("POST", SENSUM_URL, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	var sensations []Sensation
	json.Unmarshal(body, &sensations)
	if err != nil {
		log.Fatalln(err)
	}
	return sensations[0]
}

func main() {
	bot, err := tgbotapi.NewBotAPI(TOKEN)
	if err != nil {
		log.Panic(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	getLastSensation()

	for update := range updates {
		if update.Message == nil || update.Message.Text != START_COMMAND {
			continue
		}
		log.Println(update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, getLastSensation().Message)

		bot.Send(msg)
	}
}
