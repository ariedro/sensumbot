package main

import (
	"encoding/gob"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const FILE_PATH = "data.bin"

var CachedSensations []Sensation
var CachedReceivers []Receiver

func LoadData() {
	var sensationsData []Sensation
	var receiversData []Receiver

	dataFile, err := os.Open(FILE_PATH)

	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	dataDecoder := gob.NewDecoder(dataFile)
	dataDecoder.Decode(&sensationsData)
	dataDecoder.Decode(&receiversData)

	dataFile.Close()

	CachedSensations = sensationsData
	CachedReceivers = receiversData
}

func AddNewReceiver(receiver Receiver) {
	CachedReceivers = append(CachedReceivers, receiver)
}

func TrackSensation(receiverIndex int, sensation Sensation, message tgbotapi.Message) {
	CachedReceivers[receiverIndex].TrackedSensations = append(CachedReceivers[receiverIndex].TrackedSensations, TrackedSensation{SensationID: sensation.SensumID, MessageID: message.MessageID})
}

func UpdateSensationsCache(sensation Sensation) {
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

func popOldestSensation() {
	if len(CachedSensations) > 0 {
		CachedSensations = CachedSensations[1:]
	}
}

func SaveData() {
	dataFile, err := os.Create(FILE_PATH)

	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	dataEncoder := gob.NewEncoder(dataFile)
	dataEncoder.Encode(CachedSensations)
	dataEncoder.Encode(CachedReceivers)

	dataFile.Close()
}
