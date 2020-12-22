package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sonyarouje/simdb/db"
)

type Configuration struct {
	BotToken                string
	StartCommand            string
	SensumUrl               string
	PollTick                time.Duration
	TrackedSensationsLength int
}

var lastPostedSensationId = ""

var configs = loadConfigs()

type Sensation struct {
	SensumID  string    `json:"id"`
	Author    string    `json:"author"`
	Message   string    `json:"message"`
	Likes     int       `json:"likes"`
	Dislikes  int       `json:"dislikes"`
	Timestamp time.Time `json:"timestamp"`
}

func (s Sensation) ID() (jsonField string, value interface{}) {
	value = s.SensumID
	jsonField = "id"
	return
}

type TrackedSensation struct {
	SensationID string
	MessageID   int
}

type Receiver struct {
	ChatID            int64              `json:"chatId"`
	TrackedSensations []TrackedSensation `json:"trackedSensations"`
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
		log.Fatalln(err)
	}
	return configuration
}

func getSensations() ([]Sensation, error) {
	requestBody, err := json.Marshal(map[string]interface{}{"offset": 0, "limit": configs.TrackedSensationsLength})
	req, err := http.NewRequest("POST", configs.SensumUrl, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var sensations []Sensation
	json.Unmarshal(body, &sensations)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	return sensations, nil
}

func isDBEmpty(dbDriver *db.Driver) bool {
	rawSensations := dbDriver.Open(Sensation{}).Raw().([]interface{})
	return len(rawSensations) == 0
}

func getUpdatedSensations(newSensations []Sensation, dbDriver *db.Driver) ([]Sensation, error) {
	var updatedSensations []Sensation
	var currentSensation Sensation
	var err error

	if isDBEmpty(dbDriver) {
		return newSensations, nil
	}
	for _, newSensation := range newSensations {
		err = dbDriver.Open(Sensation{}).Where("id", "=", newSensation.SensumID).First().AsEntity(&currentSensation)
		if err != nil {
			log.Panicln(err)
			return nil, err
		}
		if currentSensation.Likes != newSensation.Likes || currentSensation.Dislikes != newSensation.Dislikes {
			updatedSensations = append(updatedSensations, newSensation)
		}
	}

	return updatedSensations, nil
}

func insertNewSensations(newSensations []Sensation, dbDriver *db.Driver) error {
	for _, sensation := range newSensations {
		if err := dbDriver.Insert(sensation); err != nil {
			log.Panicln(err)
			return err
		}
	}
	return nil
}

func findMessage(trackedSensations []TrackedSensation, sensationId string) int {
	for _, trackedSensation := range trackedSensations {
		if trackedSensation.SensationID == sensationId {
			return trackedSensation.MessageID
		}
	}
	return 0
}

func sensumPoll(bot *tgbotapi.BotAPI, dbDriver *db.Driver) {
	c := time.Tick(configs.PollTick * time.Second)
	for range c {
		log.Println("Checking")
		newSensations, err := getSensations()
		if err != nil {
			continue
		}
		updatedSensations, err := getUpdatedSensations(newSensations, dbDriver)
		if err != nil {
			continue
		}
		if insertNewSensations(updatedSensations, dbDriver) != nil {
			continue
		}
		var receivers []Receiver
		dbDriver.Open(Receiver{}).Get().AsEntity(&receivers)

		for _, receiver := range receivers {
			for _, sensation := range updatedSensations {
				messageText := sensation.Message + "\n~ " + sensation.Author + "\n -" + strconv.Itoa(sensation.Dislikes) + " +" + strconv.Itoa(sensation.Likes)
				messageID := findMessage(receiver.TrackedSensations, sensation.SensumID)
				if messageID > 0 {
					edit := tgbotapi.EditMessageTextConfig{
						BaseEdit: tgbotapi.BaseEdit{
							ChatID:    receiver.ChatID,
							MessageID: messageID,
						},
						Text: messageText,
					}
					_, err = bot.Send(edit)
				} else {
					msg := tgbotapi.NewMessage(receiver.ChatID, messageText)
					messageSent, err := bot.Send(msg)
					if err != nil {
						continue
					}
					receiver.TrackedSensations = append(receiver.TrackedSensations, TrackedSensation{SensationID: sensation.SensumID, MessageID: messageSent.MessageID})
				}
			}
			// FIXME: This is broken
			dbDriver.Insert(receiver)
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
