package utils

import (
	"encoding/gob"
	"os"
)

func WriteGob(filePath string, object interface{}) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(object)
	if err != nil {
		return err
	}

	return nil
}

func ReadGob(filePath string, object interface{}) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(object)
	if err != nil {
		return err
	}

	return nil
}
