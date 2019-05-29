package pistol

import (
	"io"
	"regexp"
)

var internalWritersRegexMap = map[string] func(string, string) (func(w io.Writer) error, error) {
	"text/*": NewChromaWriter,
	"application/(x-(xz|bzip|bzip2|rar|tar)|zip)": NewArchiveLister,
}

var emptyWriter = func(w io.Writer) error {
	return nil
}

func MatchInternalWriter(mimeType, filePath string) (func(w io.Writer) error, error) {
	for regex, writerCreator := range internalWritersRegexMap {
		match, err := regexp.MatchString(regex, mimeType)
		if err != nil {
			return emptyWriter, err
		}
		if match {
			writer, err := writerCreator(mimeType, filePath)
			if err != nil {
				return emptyWriter, err
			}
			return writer, nil
		}
	}
	writer, err := NewFallbackWriter(mimeType, filePath)
	if err != nil {
		return emptyWriter, err
	}
	return writer, err
}
