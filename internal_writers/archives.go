package pistol

import (
	"os"
	"io"
	"fmt"
	"regexp"
	"context"
	"golang.org/x/term"

	"github.com/mholt/archiver/v4"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/dustin/go-humanize"
	"github.com/doronbehar/magicmime"

	clexers "github.com/alecthomas/chroma/v2/lexers"
)


func NewArchiveLister(magic_db, mimeType, filePath string) (func(w io.Writer) error, error) {
	return func (w io.Writer) error {
		isArchive := true
		var singleFileFormat archiver.Decompressor
		var format archiver.Archival
		switch mimeType {
		// zip
		case "application/zip":
			format = archiver.Zip{}
		case "application/x-rar-compressed":
			format = archiver.Rar{}
		case "application/x-tar":
			format = archiver.Tar{}
		case "application/x-xz":
			if res, _ := regexp.MatchString(`.*\.tar\.xz`, filePath); res {
				format = archiver.CompressedArchive{
					Compression: archiver.Xz{},
					Archival: archiver.Tar{},
				}
			} else {
				singleFileFormat = archiver.Xz{}
				isArchive = false
			}
		case "application/x-bzip2":
			if res, _ := regexp.MatchString(`.*\.tar\.bz2`, filePath); res {
				format = archiver.CompressedArchive{
					Compression: archiver.Bz2{},
					Archival: archiver.Tar{},
				}
			} else {
				singleFileFormat = archiver.Bz2{}
				isArchive = false
			}
		case "application/gzip":
			if res, _ := regexp.MatchString(`.*\.tar\.gz`, filePath); res {
				format = archiver.CompressedArchive{
					Compression: archiver.Gz{},
					Archival: archiver.Tar{},
				}
			} else {
				singleFileFormat = archiver.Gz{}
				isArchive = false
			}
		case "application/x-lz4":
			if res, _ := regexp.MatchString(`.*\.tar\.lz`, filePath); res {
				format = archiver.CompressedArchive{
					Compression: archiver.Lz4{},
					Archival: archiver.Tar{},
				}
			} else {
				singleFileFormat = archiver.Lz4{}
				isArchive = false
			}
		case "application/x-snappy-framed":
			if res, _ := regexp.MatchString(`.*\.tar\.sz`, filePath); res {
				format = archiver.CompressedArchive{
					Compression: archiver.Sz{},
					Archival: archiver.Tar{},
				}
			} else {
				singleFileFormat = archiver.Sz{}
				isArchive = false
			}
		case "application/x-zstd":
			if res, _ := regexp.MatchString(`.*\.tar\.zst`, filePath); res {
				format = archiver.CompressedArchive{
					Compression: archiver.Zstd{},
					Archival: archiver.Tar{},
				}
			} else {
				singleFileFormat = archiver.Zstd{}
				isArchive = false
			}
		case "application/x-7z-compressed":
			format = archiver.SevenZip{}
		// brotli - currently unsupported by libmagic
		// case "application/x-brotli":
			// // This may be a brotli compressed file / tar
			// if compressedTar(filePath) {
				// format := archiver.NewTarBrotli()
			// }
		}
		if isArchive {
			t := table.NewWriter()
			t.SetOutputMirror(w)
			t.AppendHeader(table.Row{
				"Permissions",
				"Size",
				"Modification Time",
				"File Name",
			})
			if term.IsTerminal(0) {
				width, _, err := term.GetSize(0)
				if err == nil {
					t.SetAllowedRowLength(width)
				}
			}
			archiveHandler := func(ctx context.Context, f archiver.File) error {
				fPerm := fmt.Sprintf("%v", f.Mode())
				fSize := humanize.Bytes(uint64(f.Size()))
				fModtS := f.ModTime()
				fModt := fmt.Sprintf(
					"%04d-%02d-%02d %02d:%02d",
					fModtS.Year(),
					fModtS.Month(),
					fModtS.Day(),
					fModtS.Hour(),
					fModtS.Minute(),
				)
				t.AppendRow([]interface{}{
					fPerm,
					fSize,
					fModt,
					f.NameInArchive,
				})
				return nil
			}
			reader, err := os.Open(filePath)
			if err != nil {
				log.Fatalf(
					"Encountered errors opening file %s: %v\n",
					filePath,
					err,
				)
				return err
			}
			err = format.Extract(context.TODO(), reader, nil, archiveHandler)
			if err != nil {
				log.Fatalf(
					"Encountered errors extracting file %s: %v\n",
					filePath,
					err,
				)
				return err
			}
			defer reader.Close()
			t.Render()
		} else {
			fCompressed, err := os.Open(filePath)
			if err != nil {
				log.Fatalf(
					"Encountered errors opening compressed file %s: %v\n",
					filePath,
					err,
				)
				return err
			}
			fReader, err := singleFileFormat.OpenReader(fCompressed)
			if err != nil {
				panic(err)
			}
			// Why 512? https://stackoverflow.com/a/17741765/4935114
			fBytes := make([]byte, 512)
			nBytes, err := fReader.Read(fBytes)
			var fContents []byte
			if err != nil {
				if err != io.EOF {
					panic(err)
				}
				fContents = fBytes[:nBytes]
			} else {
				// TODO: Perhaps put protections here against too large files
				fRest,err := io.ReadAll(fReader)
				if err != nil {
					panic(err)
				}
				fContents = append(
					fBytes[:nBytes],
					fRest...
				)
			}
			if err := magicmime.OpenWithPath(magic_db, magicmime.MAGIC_MIME_TYPE | magicmime.MAGIC_SYMLINK); err != nil {
				log.Fatalf("Failed to open database again from some reason %v", err)
				return err
			}
			innerMimeType, err := magicmime.TypeByBuffer(fBytes[:nBytes])
			defer magicmime.Close()
			log.Infof("Detected inner mimetype of compressed file as %s", innerMimeType)
			if isText, _ := regexp.MatchString("text/*", innerMimeType); isText {
				lexer := clexers.MatchMimeType(innerMimeType)
				if lexer == nil {
					lexer = clexers.Fallback
				}
				log.Infof(
					"Using chroma to print inner contents of %s with lexer %s\n",
					filePath,
					lexer,
				)
				chromaPrint(w,string(fContents), lexer)
			} else if isJson, _ := regexp.MatchString("application/json", innerMimeType); isJson {
				jsonPrint(w, fContents)
			} else {
				fmt.Fprintf(w, "%s file compressed in a %s archive\n", innerMimeType, mimeType)
			}
			defer fReader.Close()
			defer fCompressed.Close()
		}
		return nil
	}, nil
}
