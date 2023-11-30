package main

import (
	"encoding/gob"
	"log"
	"os"
)

const FILE_PATH = "data.bin"

var CachedReceivers []Receiver

func LoadData() {
	var receiversData []Receiver

	dataFile, err := os.OpenFile(FILE_PATH, os.O_RDONLY|os.O_CREATE, 0666)

	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	dataDecoder := gob.NewDecoder(dataFile)
	dataDecoder.Decode(&receiversData)

	dataFile.Close()

	CachedReceivers = receiversData
}

func SaveData() {
	dataFile, err := os.Create(FILE_PATH)

	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	dataEncoder := gob.NewEncoder(dataFile)
	dataEncoder.Encode(CachedReceivers)

	dataFile.Close()
}
