package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sonyarouje/simdb/db"
)

type Configuration struct {
	BotToken     string
	StartCommand string
	SensumUrl    string
	PollTick     time.Duration
}

var lastPostedSensationId = ""

var configs = loadConfigs()

type Sensation struct {
	ID        string
	Author    string
	Message   string
	Likes     int
	Dislikes  int
	Timestamp time.Time
}

type Receiver struct {
	ChatID int64
}

func (r Receiver) ID() (jsonField string, value interface{}) {
	value = r.ChatID
	jsonField = "chatid"
	return
}

func loadConfigs() Configuration {
	file, _ := os.Open("config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Fatalln("error:", err)
	}
	log.Println(configuration.BotToken)
	return configuration
}

func getLastSensation() Sensation {
	requestBody, err := json.Marshal(map[string]interface{}{"offset": 0, "limit": 1})
	req, err := http.NewRequest("POST", configs.SensumUrl, bytes.NewBuffer(requestBody))
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

func sensumPoll(bot *tgbotapi.BotAPI, dbDriver *db.Driver) {
	c := time.Tick(configs.PollTick * time.Second)
	for range c {
		log.Println("Checking")
		lastSensation := getLastSensation()
		if lastSensation.ID != lastPostedSensationId {
			lastPostedSensationId = lastSensation.ID
			log.Println(lastSensation.Message)

			var receivers []Receiver
			dbDriver.Open(Receiver{}).Get().AsEntity(&receivers)

			for _, receiver := range receivers {
				msg := tgbotapi.NewMessage(receiver.ChatID, lastSensation.Message+"\n~ "+lastSensation.Author)
				bot.Send(msg)
			}
		}
	}
}

func telegramPoll(bot *tgbotapi.BotAPI, dbDriver *db.Driver) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil || update.Message.Text != configs.StartCommand {
			continue
		}
		dbDriver.Insert(Receiver{ChatID: update.Message.Chat.ID})

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Agrega2")
		bot.Send(msg)
	}
}

func main() {
	bot, err := tgbotapi.NewBotAPI(configs.BotToken)
	if err != nil {
		log.Panic(err)
	}
	dbDriver, err := db.New("data")
	if err != nil {
		panic(err)
	}

	log.Println("Bot started")

	go telegramPoll(bot, dbDriver)
	sensumPoll(bot, dbDriver)
}
