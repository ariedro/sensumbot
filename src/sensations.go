package main

import (
	"log"
	"math/big"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Sensation struct {
	Author  string
	Message string
}

type Receiver struct {
	ChatID int64
}

func setupContract() *Contract {
	ethClientUrl := Configs.EthClientUrl

	client, err := ethclient.Dial(ethClientUrl)
	if err != nil {
		log.Fatal(err)
	}

	contractAddress := common.HexToAddress(Configs.SensumContractAddress)
	contract, err := NewContract(contractAddress, client)
	if err != nil {
		log.Fatal(err)
	}

	return contract
}

func updateIndex(newIndex int) {
	CachedIndex = newIndex
}

func getSensations(contract *Contract) ([]Sensation, error) {
	indexBigInt, err := contract.GetSensationsLength(nil)
	if err != nil {
		log.Fatal(err)
	}

	lastIndex := int(indexBigInt.Int64())

	var sensations []Sensation
	start := CachedIndex

	for i := start; i < lastIndex-1; i += 1 {
		contractSensation, err := contract.Sensations(nil, big.NewInt(int64(i)))
		if err != nil {
			log.Fatal(err)
		}
		// TODO: Get real author avatar
		sensation := Sensation{Author: "fafa", Message: contractSensation.Message}

		sensations = append(sensations, sensation)
	}

	updateIndex(lastIndex)

	return sensations, nil
}

func SensumPoll(bot *tgbotapi.BotAPI) {
	contract := setupContract()
	c := time.Tick(Configs.PollTick * time.Second)
	for range c {
		log.Println("Checking")
		// Get new sensations
		sensations, err := getSensations(contract)
		if err != nil {
			continue
		}

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
