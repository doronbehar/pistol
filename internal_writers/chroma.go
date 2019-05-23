package pistol

import (
	"log"
	"io/ioutil"
	"io"
	"os"

	"github.com/alecthomas/chroma"
	cformatters "github.com/alecthomas/chroma/formatters"
	clexers "github.com/alecthomas/chroma/lexers"
	cstyles "github.com/alecthomas/chroma/styles"
)

func NewChromaWriter(mimeType, filePath string, verbose bool) (func(w io.Writer) error, error) {
	if verbose {
		log.Printf("using chroma to print %s with syntax highlighting\n", filePath)
	}
	lexer := clexers.Match(filePath)
	if lexer == nil {
		lexer = clexers.Fallback
	}
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return emptyWriter, err
	}
	iterator, err := lexer.Tokenise(nil, string(contents))
	if err != nil {
		return emptyWriter, err
	}
	env_formatter := os.Getenv("PISTOL_CHROMA_FORMATTER")
	var formatter chroma.Formatter
	if env_formatter != "" {
		if verbose {
			log.Printf("Using style from environment: %s\n", env_formatter)
		}
		formatter = cformatters.Get(env_formatter)
	} else {
		formatter = cformatters.Get("terminal")
	}
	env_style := os.Getenv("PISTOL_CHROMA_STYLE")
	var style *chroma.Style
	if env_style != "" {
		if verbose {
			log.Printf("Using style from environment: %s\n", env_style)
		}
		style = cstyles.Get(env_style)
	} else {
		style = cstyles.Get("vim")
	}
	return func (w io.Writer) error {
		return formatter.Format(w, style, iterator)
	}, nil
}

