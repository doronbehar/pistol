package pistol

import (
	"log"
	"io/ioutil"
	"io"
)

func NewArchiveLister(mimeType, filePath string, verbose bool) (func(w io.Writer) error, error) {
	if verbose {
		log.Printf("listing files in archive %s\n", filePath)
	}
	return emptyWriter, nil
}
