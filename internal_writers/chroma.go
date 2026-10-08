package pistol

import (
	"io"
	"os"
	"strconv"

	"github.com/alecthomas/chroma/v2"
	log "github.com/sirupsen/logrus"
	cformatters "github.com/alecthomas/chroma/v2/formatters"
	clexers "github.com/alecthomas/chroma/v2/lexers"
	cstyles "github.com/alecthomas/chroma/v2/styles"
)

func chromaPrint(w io.Writer, contents string, lexer chroma.Lexer) error {
	iterator, err := lexer.Tokenise(nil, contents)
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

// Default amount of bytes read from a file to be highlighted by chroma, see
// PISTOL_CHROMA_SIZE in the README.
const defaultChromaSize = 100000

// chromaSize returns the maximal amount of bytes of a text file we feed to
// chroma. chroma needs the whole text in memory, and then even more for its own
// representation of it, so we avoid feeding it the whole of large files.
func chromaSize() int64 {
	if env_size := os.Getenv("PISTOL_CHROMA_SIZE"); env_size != "" {
		size, err := strconv.ParseInt(env_size, 10, 64)
		if err != nil || size < 1 {
			log.Fatalf("PISTOL_CHROMA_SIZE must be a positive number of bytes, got: %q", env_size)
		}
		return size
	}
	return defaultChromaSize
}

func NewChromaWriter(magic_db, mimeType, filePath string) (func(w io.Writer) error, error) {
	lexer := clexers.Match(filePath)
	if lexer == nil {
		lexer = clexers.Fallback
	}
	log.Infof("using chroma to print %s with lexer %s\n", filePath, lexer)
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Encountered error opening file %s: %v", filePath, err)
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, chromaSize()))
	if err != nil {
		log.Fatalf("Encountered error reading file %s: %v", filePath, err)
	}
	contents := string(raw)
	return func (w io.Writer) error {
		return chromaPrint(w, contents, lexer)
	}, nil
}

