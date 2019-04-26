package main

import (
	"os"
	"log"

	"github.com/gabriel-vasile/mimetype"
)

func handle(configFile string, filePath string, verbosity bool) (error) {
	// get mimetype of given file, we don't care about the extension
	mime, _, err := mimetype.DetectFile(filePath)
	if err != nil {
		log.Fatalf("%s\n", err)
		os.Exit(1)
	}
	if verbosity {
		log.Printf("detected mimetype is %s", mime)
		log.Printf("reading configuration from %s", configFile)
	}
	return nil
}
