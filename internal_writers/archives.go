package pistol

import (
	"archive/tar"
	"archive/zip"
	"log"
	"io"
	"fmt"

	"github.com/mholt/archiver"
	"github.com/nwaples/rardecode"
)

func NewArchiveLister(mimeType, filePath string, verbose bool) (func(w io.Writer) error, error) {
	if verbose {
		log.Printf("listing files in archive %s\n", filePath)
	}
	return func (w io.Writer) error {
		var count int
		err := archiver.Walk(filePath, func(f archiver.File) error {
			count++
			switch h := f.Header.(type) {
			case zip.FileHeader:
				fmt.Fprintf(w, "%s\t%d\t%d\t%s\t%s\n",
					f.Mode(),
					h.Method,
					f.Size(),
					f.ModTime(),
					h.Name,
				)
			case *tar.Header:
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n",
					f.Mode(),
					h.Uname,
					h.Gname,
					f.Size(),
					f.ModTime(),
					h.Name,
				)
			case *rardecode.FileHeader:
				fmt.Fprintf(w, "%s\t%d\t%d\t%s\t%s\n",
					f.Mode(),
					int(h.HostOS),
					f.Size(),
					f.ModTime(),
					h.Name,
				)
			default:
				fmt.Fprintf(w, "%s\t%d\t%s\t?/%s\n",
					f.Mode(),
					f.Size(),
					f.ModTime(),
					f.Name(), // we don't know full path from this
				)
			}
			return nil
		})
		fmt.Fprintf(w, "total %d\n", count)
		return err
	}, nil
}
