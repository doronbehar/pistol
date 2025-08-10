package pistol

import (
	"os"
	"io"
	"fmt"
	"regexp"
	"context"
	"golang.org/x/term"

	"github.com/mholt/archives"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/dustin/go-humanize"
	"github.com/doronbehar/magicmime"

	clexers "github.com/alecthomas/chroma/v2/lexers"
)


func NewArchiveLister(magic_db, mimeType, filePath string) (func(w io.Writer) error, error) {
	return func (w io.Writer) error {
		isArchive := true
		var singleFileFormat interface {
			OpenReader(r io.Reader) (io.ReadCloser, error)
		}
		var format archives.Format
		switch mimeType {
		// zip
		case "application/zip":
			format = &archives.Zip{}
		case "application/x-rar-compressed":
			format = &archives.Rar{}
		case "application/x-tar":
			format = &archives.Tar{}
		case "application/x-xz":
			if res, _ := regexp.MatchString(`.*\.tar\.xz$`, filePath); res {
				format = archives.CompressedArchive{
					Compression: &archives.Xz{},
					Archival:    &archives.Tar{},
					Extraction:  &archives.Tar{},
				}
			} else {
				singleFileFormat = &archives.Xz{}
				isArchive = false
			}
		case "application/x-bzip2":
			if res, _ := regexp.MatchString(`.*\.tar\.bz2$`, filePath); res {
				format = archives.CompressedArchive{
					Compression: &archives.Bz2{},
					Archival:    &archives.Tar{},
					Extraction:  &archives.Tar{},
				}
			} else {
				singleFileFormat = &archives.Bz2{}
				isArchive = false
			}
		case "application/gzip":
			if res, _ := regexp.MatchString(`.*\.tar\.gz$`, filePath); res {
				format = archives.CompressedArchive{
					Compression: &archives.Gz{},
					Archival:    &archives.Tar{},
					Extraction:  &archives.Tar{},
				}
			} else {
				singleFileFormat = &archives.Gz{}
				isArchive = false
			}
		case "application/x-lz4":
			if res, _ := regexp.MatchString(`.*\.tar\.lz4$`, filePath); res {
				format = archives.CompressedArchive{
					Compression: &archives.Lz4{},
					Archival:    &archives.Tar{},
					Extraction:  &archives.Tar{},
				}
			} else {
				singleFileFormat = &archives.Lz4{}
				isArchive = false
			}
		case "application/x-snappy-framed":
			if res, _ := regexp.MatchString(`.*\.tar\.sz$`, filePath); res {
				format = archives.CompressedArchive{
					Compression: &archives.Sz{},
					Archival:    &archives.Tar{},
					Extraction:  &archives.Tar{},
				}
			} else {
				singleFileFormat = &archives.Sz{}
				isArchive = false
			}
		case "application/x-zstd":
			if res, _ := regexp.MatchString(`.*\.tar\.zst$`, filePath); res {
				format = archives.CompressedArchive{
					Compression: &archives.Zstd{},
					Archival:    &archives.Tar{},
					Extraction:  &archives.Tar{},
				}
			} else {
				singleFileFormat = &archives.Zstd{}
				isArchive = false
			}
		case "application/x-7z-compressed":
			format = &archives.SevenZip{}
		// brotli - currently unsupported by libmagic, but we don't mind putting it
		// here anyway.
		case "application/x-brotli":
			if res, _ := regexp.MatchString(`.*\.tar\.br$`, filePath); res {
				format = archives.CompressedArchive{
					Compression: &archives.Brotli{},
					Archival:    &archives.Tar{},
					Extraction:  &archives.Tar{},
				}
			} else {
				singleFileFormat = &archives.Brotli{}
				isArchive = false
			}
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
			archiveHandler := func(ctx context.Context, f archives.FileInfo) error {
				fPerm := fmt.Sprintf("%v", f.FileInfo.Mode())
				fSize := humanize.Bytes(uint64(f.FileInfo.Size()))
				fModtS := f.FileInfo.ModTime()
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
			err = format.(archives.Extractor).Extract(context.Background(), reader, archiveHandler)
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
