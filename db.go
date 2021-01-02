package main

import (
	"encoding/gob"
	"log"
	"os"
)

const FILE_PATH = "data.bin"

var CachedSensations []Sensation

func LoadData() {
	var data []Sensation

	dataFile, err := os.Open(FILE_PATH)

	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	dataDecoder := gob.NewDecoder(dataFile)
	err = dataDecoder.Decode(&data)

	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	dataFile.Close()

	CachedSensations = data
}

func PushNew(sensation Sensation) {
	CachedSensations = append(CachedSensations, sensation)
}

func PopOldest() {
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

	dataFile.Close()
}
