package pistol

import (
	"io"
	"regexp"
)

var internalWritersRegexMap = map[string] func(string, string, string) (func(w io.Writer) error, error) {
	"text/*": NewChromaWriter,
	// https://github.com/doronbehar/pistol/issues/34
	"application/json": NewJsonWriter,
	// See:
	// - https://github.com/doronbehar/pistol/issues/106
	// - https://stackoverflow.com/a/21098951/4935114
	"application/javascript": NewChromaWriter,
	"application/zip": NewArchiveLister,
	"application/x-rar-compressed": NewArchiveLister,
	"application/x-tar": NewArchiveLister,
	"application/x-xz": NewArchiveLister,
	"application/x-bzip2": NewArchiveLister,
	"application/gzip": NewArchiveLister,
	"application/x-lz4": NewArchiveLister,
	"application/x-7z-compressed": NewArchiveLister,
	"application/x-snappy-framed": NewArchiveLister,
	"application/x-zstd": NewArchiveLister,
	// TODO: Match brotli when libmagic supports it
}

var emptyWriter = func(w io.Writer) error {
	return nil
}

func MatchInternalWriter(magic_db, mimeType, filePath string) (func(w io.Writer) error, error) {
	for regex, writerCreator := range internalWritersRegexMap {
		match, err := regexp.MatchString(regex, mimeType)
		if err != nil {
			return emptyWriter, err
		}
		if match {
			writer, err := writerCreator(magic_db, mimeType, filePath)
			if err != nil {
				return emptyWriter, err
			}
			return writer, nil
		}
	}
	writer, err := NewFallbackWriter(magic_db, mimeType, filePath)
	if err != nil {
		return emptyWriter, err
	}
	return writer, err
}
