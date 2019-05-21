package pistol

import (
	"io"
	"regexp"
)

var internalWritersRegexMap = map[string] func(string, string, bool) (func(w io.Writer) error, error) {
	"text/*": NewChromaWriter,
	"inode/directory": NewDirectoryLister,
	"application/x-xz": NewArchiveLister,
}

var emptyWriter = func(w io.Writer) error {
	return nil
}

func MatchInternalWriter(mimeType, filePath string, verbose bool) (func(w io.Writer) error, error) {
	for regex, writerCreator := range internalWritersRegexMap {
		match, err := regexp.MatchString(regex, mimeType)
		if err != nil {
			return emptyWriter, err
		}
		if match {
			writer, err := writerCreator(mimeType, filePath, verbose)
			if err != nil {
				return emptyWriter, err
			}
			return writer, nil
		}
	}
	writer, err := NewFallbackWriter(mimeType, filePath, verbose)
	if err != nil {
		return emptyWriter, err
	}
	return writer, err
}
