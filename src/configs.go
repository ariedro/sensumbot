package main

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type Configuration struct {
	BotToken     string
	StartCommand string
	Commands     struct {
		Start   string
		Version string
	}
	SensumContractAddress string
	EthClientUrl          string
	PollTick              time.Duration
}

var Configs = loadConfigs()

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
