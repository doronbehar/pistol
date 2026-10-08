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


// The amount of files of an archive that are listed in a single table.
const archiveRowsBatch = 1000

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
			// Tables of archives with many files are rendered in batches, and
			// the header is rendered only in the first one. So the minimal
			// widths of the first columns are set here, to keep them aligned
			// between batches.
			t.SetColumnConfigs([]table.ColumnConfig{
				// The width of the "Permissions" header, as modes are always 10
				// characters long.
				{Number: 1, WidthMin: 11},
				// The width of the longest humanize.Bytes output, e.g. "1.0 GB".
				{Number: 2, WidthMin: 6},
				// The width of the "Modification Time" header, as the formatted
				// time is always 16 characters long.
				{Number: 3, WidthMin: 17},
			})
			if term.IsTerminal(0) {
				width, _, err := term.GetSize(0)
				if err == nil {
					t.SetAllowedRowLength(width)
				}
			}
			batched := false
			archiveHandler := func(ctx context.Context, f archives.FileInfo) error {
				// The table has to hold all of its rows in memory in order to be
				// rendered, so archives with many files are listed in batches of
				// rows.
				if t.Length() >= archiveRowsBatch {
					if !batched {
						// The width of the file names column isn't bounded, and
						// rows rendered later can't affect the columns above
						// them, so the border to the right of it is dropped.
						t.Style().Options.DrawBorder = false
						batched = true
					}
					t.Render()
					t.ResetHeaders()
					t.ResetRows()
				}
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
			if !batched || t.Length() > 0 {
				t.Render()
			}
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
			nBytes, readErr := fReader.Read(fBytes)
			if readErr != nil && readErr != io.EOF {
				panic(readErr)
			}
			if err := magicmime.OpenWithPath(magic_db, magicmime.MAGIC_MIME_TYPE | magicmime.MAGIC_SYMLINK); err != nil {
				log.Fatalf("Failed to open database again from some reason %v", err)
				return err
			}
			innerMimeType, err := magicmime.TypeByBuffer(fBytes[:nBytes])
			defer magicmime.Close()
			if err != nil {
				panic(err)
			}
			log.Infof("Detected inner mimetype of compressed file as %s", innerMimeType)
			isText, _ := regexp.MatchString("text/*", innerMimeType)
			isJson, _ := regexp.MatchString("application/json", innerMimeType)
			sizeLimit := chromaSize()
			var fContents []byte
			if readErr == io.EOF {
				fContents = fBytes[:nBytes]
			} else {
				var fRestReader io.Reader = fReader
				if isText {
					// Don't decompress more than chroma is going to get.
					fRestReader = io.LimitReader(fReader, sizeLimit - int64(nBytes))
				} else if isJson {
					// JSON can't be parsed when truncated. Read one byte more
					// than the limit, to know whether the file is too large.
					fRestReader = io.LimitReader(fReader, sizeLimit - int64(nBytes) + 1)
				}
				fRest,err := io.ReadAll(fRestReader)
				if err != nil {
					panic(err)
				}
				fContents = append(
					fBytes[:nBytes],
					fRest...
				)
			}
			if isText {
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
			} else if isJson {
				// In principle, this can never happen, and it is unfortunate. It seems
				// that libmagic doesn't detect JSON as a mimetype, just with 512 bytes.
				if int64(len(fContents)) > sizeLimit {
					fmt.Fprintf(
						w,
						"Compressed JSON file larger then %s\n",
						humanize.Bytes(uint64(sizeLimit)),
					)
				} else {
					jsonPrint(w, fContents)
				}
			} else {
				fmt.Fprintf(w, "%s file compressed in a %s archive\n", innerMimeType, mimeType)
			}
			defer fReader.Close()
			defer fCompressed.Close()
		}
		return nil
	}, nil
}
