package pistol

import (
	"io"
	"regexp"
)

var internalWritersRegexMap = map[string] func(string, string) (func(w io.Writer) error, error) {
	"text/*": NewChromaWriter,
	// https://github.com/doronbehar/pistol/issues/34
	"application/json": NewChromaWriter,
	"application/zip": NewArchiveLister,
	"application/x-rar-compressed": NewArchiveLister,
	"application/x-tar": NewArchiveLister,
	"application/x-xz": NewArchiveLister,
	"application/x-bzip2": NewArchiveLister,
	"application/gzip": NewArchiveLister,
	"application/x-lz4": NewArchiveLister,
	"application/x-snappy-framed": NewArchiveLister,
	"application/x-zstd": NewArchiveLister,
	// TODO: Match brotli when libmagic supports it
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
