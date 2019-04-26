package main

import (
	"log"

	"github.com/gabriel-vasile/mimetype"
)

func handle(configFile string, filePath string, verbose bool) (error) {
	// get mimetype of given file, we don't care about the extension
	mime, _, err := mimetype.DetectFile(filePath)
	if err != nil {
		return err
	}
	if verbose {
		log.Printf("detected mimetype is %s", mime)
		log.Printf("reading configuration from %s", configFile)
	}
	return nil
}
