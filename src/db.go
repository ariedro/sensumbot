package main

import (
	"encoding/gob"
	"log"
	"os"
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
