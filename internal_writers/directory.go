package pistol

import (
	"io"
	"log"
)

func NewDirectoryLister(mimeType, filePath string, verbose bool) (func(w io.Writer) error, error) {
	if verbose {
		log.Printf("listing directory\n")
	}
	return emptyWriter, nil
}
