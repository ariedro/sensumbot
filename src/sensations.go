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

var AVATARS = []string{
	"◕‿◕", "◪_◪", "°▽°", "◔_◔", "・ω・", "◣_◢", "⌐■_■", "◉_◉", "◔̯◔", "⌒_⌒",
	"ʘ益ʘ", "ಥ_ಥ", "ಠ▃ಠ", "◡w◡", "▼ｪ▼", "ಠ_๏", "⚆_⚆", "ↁ_ↁ", "°□°", "Ф.Ф", "♥‿♥",
	"╭ರ_⊙", "◡ᴥ◡", "￣Д￣", "●_●", "Ò‸Ó", "︶︿︶", "ಠ﹏ಠ", "◔︿◔", ".益.", "*‿*", "👽",
	"⌒ω⌒", "ʘ ͜ʖ ʘ", "ಠ_ಠ", "ಠิ﹏ಠิ", "ಠ⌣ಠ", "( ͡° ͜ʖ ͡°)", "´༎ຶٹ༎ຶ`", "￢_￢",
	" ͠° ͟ʖ ͡°", " ͠Ò ‸ Ó", "◣_◢ /", "◣_◢ —", "◣∩◢", "◣∀◢", "￢з￢", "ᔑ•ﺪ͟͠•ᔐ",
	"ლ,ᔑ•ﺪ͟͠•ᔐ.ლ", "(｡◝‿◜｡)", "( ▀ ͜͞ʖ▀) =ε/̵͇̿/’̿’̿ ̿ ̿̿ ̿̿ ̿̿",
	"̿̿ ̿̿ ̿̿ ̿'̿'̵͇̿з= ( ▀ ͜͞ʖ▀)", "(ง︡'-'︠)ง", "( T____T)", "(╯°□°）╯︵ ┻━┻",
	"(ﾉಥ益ಥ）ﾉ ┻━┻", "ヽ༼ຈل͜ຈ༽ﾉ", "ヽ(´▽`)/", "(☞ﾟヮﾟ)☞", "m9(・∀・)", "♪┏(・o･)┛♪",
	"\\[T]/",
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

func getAvatar(avatarIndex int) string {
	if avatarIndex > len(AVATARS)-1 || avatarIndex < 0 {
		return AVATARS[0]
	}
	return AVATARS[avatarIndex]
}

func getSensations(contract *Contract) ([]Sensation, error) {
	indexBigInt, err := contract.GetSensationsLength(nil)
	if err != nil {
		log.Fatal(err)
	}

	lastIndex := int(indexBigInt.Int64())

	var sensations []Sensation
	start := CachedIndex

	for i := start; i < lastIndex; i += 1 {
		contractSensation, err := contract.Sensations(nil, big.NewInt(int64(i)))
		if err != nil {
			log.Fatal(err)
		}
		avatarIndexBigInt := contractSensation.Avatar
		avatarIndex := int(avatarIndexBigInt.Int64())

		sensation := Sensation{Author: getAvatar(avatarIndex), Message: contractSensation.Message}

		sensations = append(sensations, sensation)
	}

	updateIndex(lastIndex)

	return sensations, nil
}

func SensumPoll(bot *tgbotapi.BotAPI) {
	contract := setupContract()
	c := time.Tick(Configs.PollTick * time.Second)
	for range c {
		if len(CachedReceivers) == 0 {
			log.Println("Skipping, no receivers")
			continue
		}
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
