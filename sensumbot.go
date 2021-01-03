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
)

type Configuration struct {
	BotToken                string
	StartCommand            string
	SensumUrl               string
	PollTick                time.Duration
	TrackedSensationsLength int
}

var Configs = loadConfigs()

type Sensation struct {
	SensumID  string    `json:"id"`
	Author    string    `json:"author"`
	Message   string    `json:"message"`
	Likes     int       `json:"likes"`
	Dislikes  int       `json:"dislikes"`
	Timestamp time.Time `json:"timestamp"`
}

type TrackedSensation struct {
	SensationID string
	MessageID   int
}

type Receiver struct {
	ChatID            int64
	TrackedSensations []TrackedSensation
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
	requestBody, err := json.Marshal(map[string]interface{}{"offset": 0, "limit": Configs.TrackedSensationsLength})
	req, err := http.NewRequest("POST", Configs.SensumUrl, bytes.NewBuffer(requestBody))
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

func filterUnalteredSensations(sensations []Sensation) ([]Sensation, error) {
	var updatedSensations []Sensation
	var newSensation bool

	if len(CachedSensations) == 0 {
		return sensations, nil
	}
	for _, sensation := range sensations {
		newSensation = true
		for _, cachedSensation := range CachedSensations {
			if cachedSensation.SensumID == sensation.SensumID {
				newSensation = false
				if cachedSensation.Likes != sensation.Likes || cachedSensation.Dislikes != sensation.Dislikes {
					updatedSensations = append(updatedSensations, sensation)
				}
			}
		}
		if newSensation {
			updatedSensations = append(updatedSensations, sensation)
		}
	}

	return updatedSensations, nil
}

func addSensationsToCache(sensations []Sensation) {
	for _, sensation := range sensations {
		UpdateSensationsCache(sensation)
	}
}

func findMessage(trackedSensations []TrackedSensation, sensationId string) int {
	for _, trackedSensation := range trackedSensations {
		if trackedSensation.SensationID == sensationId {
			return trackedSensation.MessageID
		}
	}
	return 0
}

func sensumPoll(bot *tgbotapi.BotAPI) {
	c := time.Tick(Configs.PollTick * time.Second)
	for range c {
		log.Println("Checking")
		// Get new sensations
		sensations, err := getSensations()
		if err != nil {
			continue
		}
		// Filter the ones that didn't change
		sensations, err = filterUnalteredSensations(sensations)
		if err != nil {
			continue
		}
		// Update the cache
		addSensationsToCache(sensations)

		for receiverIndex, receiver := range CachedReceivers {
			for _, sensation := range sensations {
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
					_, err = bot.Send(edit)
				} else {
					msg := tgbotapi.NewMessage(receiver.ChatID, messageText)
					messageSent, err := bot.Send(msg)
					if err != nil {
						continue
					}
					TrackSensation(receiverIndex, sensation, messageSent)
				}
			}
			SaveData()
		}
		log.Println(CachedReceivers)
	}
}

func telegramPoll(bot *tgbotapi.BotAPI) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil || update.Message.Text != Configs.StartCommand {
			continue
		}
		AddNewReceiver(Receiver{ChatID: update.Message.Chat.ID})

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Agrega2")
		bot.Send(msg)
	}
}

func main() {
	LoadData()
	bot, err := tgbotapi.NewBotAPI(Configs.BotToken)
	if err != nil {
		log.Panic(err)
		panic(err)
	}

	log.Println("Bot started")

	go telegramPoll(bot)
	sensumPoll(bot)
}
