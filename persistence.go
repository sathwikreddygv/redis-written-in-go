package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"time"
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

func periodicSave(data interface{}, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// fmt.Print("\nSaving to disk...\n")
			if err := saveToDisk(data); err != nil {
				fmt.Println("Error saving to file:", err)
			}
		}
	}
}
