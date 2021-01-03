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

	for _, sensation := range sensations {
		// Don't care about sensations that happened before bot init
		if sensation.Timestamp.Before(UpTime) {
			continue
		}
		newSensation = true
		for _, cachedSensation := range CachedSensations {
			if cachedSensation.SensumID == sensation.SensumID {
				newSensation = false
				// If sensation is cahed, only care if the like or dislike numbers changed
				if cachedSensation.Likes != sensation.Likes || cachedSensation.Dislikes != sensation.Dislikes {
					updatedSensations = append(updatedSensations, sensation)
				}
			}
		}
		// Always include new sensations
		if newSensation {
			updatedSensations = append(updatedSensations, sensation)
		}
	}

	return updatedSensations, nil
}

func findMessage(trackedSensations []TrackedSensation, sensationId string) int {
	for _, trackedSensation := range trackedSensations {
		if trackedSensation.SensationID == sensationId {
			return trackedSensation.MessageID
		}
	}
	return 0
}

func TrackSensation(receiverIndex int, sensation Sensation, message tgbotapi.Message) {
	CachedReceivers[receiverIndex].TrackedSensations = append(CachedReceivers[receiverIndex].TrackedSensations, TrackedSensation{SensationID: sensation.SensumID, MessageID: message.MessageID})
}

func updateSensationsCache(sensations []Sensation) {
	for _, sensation := range sensations {
		for i := range CachedSensations {
			if CachedSensations[i].SensumID == sensation.SensumID {
				CachedSensations[i] = sensation
				return
			}
		}
		CachedSensations = append(CachedSensations, sensation)
		if len(CachedSensations) > Configs.TrackedSensationsLength {
			popOldestSensation()
		}
	}
}

func popOldestSensation() {
	if len(CachedSensations) > 0 {
		CachedSensations = CachedSensations[1:]
	}
}

func SensumPoll(bot *tgbotapi.BotAPI) {
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
		updateSensationsCache(sensations)

		for receiverIndex, receiver := range CachedReceivers {
			for _, sensation := range sensations {
				err := SendMessage(sensation, receiver, receiverIndex, bot)
				if err != nil {
					continue
				}
			}
		}
		SaveData()
	}
}
