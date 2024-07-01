package pistol

import (
	"io"
	"os"
	"fmt"
	"encoding/json"

	log "github.com/sirupsen/logrus"
	jc "github.com/nwidger/jsoncolor"
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
	contents, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Encountered error reading file %s", filePath)
	}
	return func (w io.Writer) error {
		return jsonPrint(w, contents)
	}, nil
}
