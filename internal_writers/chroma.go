package pistol

import (
	"io"
	"os"

	"github.com/alecthomas/chroma/v2"
	log "github.com/sirupsen/logrus"
	cformatters "github.com/alecthomas/chroma/v2/formatters"
	clexers "github.com/alecthomas/chroma/v2/lexers"
	cstyles "github.com/alecthomas/chroma/v2/styles"
)

func chromaPrint(w io.Writer, contents string, lexer chroma.Lexer) error {
	iterator, err := lexer.Tokenise(nil, string(contents))
	if err != nil {
		panic(err)
	}
	env_formatter := os.Getenv("PISTOL_CHROMA_FORMATTER")
	var formatter chroma.Formatter
	if env_formatter != "" {
		log.Infof("Using style from environment: %s\n", env_formatter)
		formatter = cformatters.Get(env_formatter)
	} else {
		formatter = cformatters.TTY8
	}
	env_style := os.Getenv("PISTOL_CHROMA_STYLE")
	var style *chroma.Style
	if env_style != "" {
		log.Infof("Using style from environment: %s\n", env_style)
		style = cstyles.Get(env_style)
	} else {
		// I think this is the most impressive one on default usage with Lf
		style = cstyles.Get("pygments")
	}
	return formatter.Format(w, style, iterator)
}

func NewChromaWriter(magic_db, mimeType, filePath string) (func(w io.Writer) error, error) {
	lexer := clexers.Match(filePath)
	if lexer == nil {
		lexer = clexers.Fallback
	}
	log.Infof("using chroma to print %s with lexer %s\n", filePath, lexer)
	contents, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Encountered error reading file %s", filePath)
	}
	return func (w io.Writer) error {
		return chromaPrint(w, string(contents), lexer)
	}, nil
}

