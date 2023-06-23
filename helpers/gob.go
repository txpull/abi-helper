// Package utils provides utility functions for common operations.
package helpers

import (
	"encoding/gob"
	"os"
)

// WriteGob writes the specified object to a .gob file at the given filePath.
// It uses the encoding/gob package to encode the object and write it to the file.
// If any error occurs during the process, it is returned.
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

// ReadGob reads a .gob file at the given filePath and decodes its content into the specified object.
// It uses the encoding/gob package to decode the file content into the object.
// If any error occurs during the process, it is returned.
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
