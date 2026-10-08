package pistol

import (
	"io"
	"os"
	"fmt"
	"encoding/json"

	log "github.com/sirupsen/logrus"
	jc "github.com/nwidger/jsoncolor"
	"github.com/dustin/go-humanize"
)

func jsonPrint(w io.Writer, contents []byte) error {
	var jsonObject any
	err := json.Unmarshal(contents, &jsonObject)
	if err != nil {
		return err
	}
	output, err := jc.MarshalIndent(jsonObject, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(w, string(output))
	return nil
}

func NewJsonWriter(magic_db, mimeType, filePath string) (func(w io.Writer) error, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		log.Fatalf("Encountered error getting the size of file %s: %v", filePath, err)
	}
	// JSON files can't be parsed when truncated, so we don't even read larger
	// ones.
	sizeLimit := chromaSize()
	if info.Size() > sizeLimit {
		return func (w io.Writer) error {
			fmt.Fprintf(
				w,
				"JSON file larger then %s\n",
				humanize.Bytes(uint64(sizeLimit)),
			)
			return nil
		}, nil
	}
	contents, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Encountered error reading file %s: %v", filePath, err)
	}
	return func (w io.Writer) error {
		return jsonPrint(w, contents)
	}, nil
}
