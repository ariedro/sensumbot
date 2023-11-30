package main

import (
	"fmt"
	"log"
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

	fmt.Println(ethClientUrl)

	client, err := ethclient.Dial(ethClientUrl)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(Configs.SensumContractAddress)
	contractAddress := common.HexToAddress(Configs.SensumContractAddress)
	contract, err := NewContract(contractAddress, client)
	if err != nil {
		log.Fatal(err)
	}

	return contract
}

func getSensations(contract *Contract) ([]Sensation, error) {
	amount, err := contract.GetSensationsLength(nil) // TODO: Cache the last index sent
	if err != nil {
		log.Fatal(err)
	}

	contractSensation, err := contract.Sensations(nil, amount.Sub(amount, common.Big1)) // TODO: Bring all of them
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	var sensation Sensation = Sensation{Author: "fafa", Message: contractSensation.Message} // TODO: Get real author avatar

	return []Sensation{sensation}, nil
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
