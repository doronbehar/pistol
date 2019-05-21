package pistol

import (
	"io"
	"fmt"

	"github.com/rakyll/magicmime"
)

func NewFallbackWriter(mimeType, filePath string, verbose bool) (func(w io.Writer) error, error) {
	if err := magicmime.Open(magicmime.MAGIC_SYMLINK); err != nil {
		return emptyWriter, err
	}
	complete_filetype_description, err := magicmime.TypeByFile(filePath)
	if err != nil {
		return emptyWriter, err
	}
	defer magicmime.Close()
	return func (w io.Writer) error {
		fmt.Fprintln(w, complete_filetype_description)
		return nil
	}, nil
}
