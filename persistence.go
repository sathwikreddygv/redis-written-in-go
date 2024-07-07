package main

import (
	"encoding/gob"
	"os"
)

func saveToDisk(data interface{}) error {
	file, err := os.Create("aof.gob")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(data)
	if err != nil {
		return err
	}

	return nil
}

func loadFromDisk(filename string, data interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(data)
	if err != nil {
		return err
	}

	return nil
}

// func backgroundSavetoDisk
