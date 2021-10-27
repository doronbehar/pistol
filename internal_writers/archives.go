package pistol

import (
	"io"
	"fmt"
	"math"
	"regexp"
	"archive/tar"
	"archive/zip"
	"github.com/nwaples/rardecode"

	"github.com/mholt/archiver/v3"
	log "github.com/sirupsen/logrus"
	"github.com/dustin/go-humanize"
)

type archiveFileInfo struct {
	Permissions string
	Size string
	ModifiedTime string
	FileName string
}

func NewArchiveLister(magic_db, mimeType, filePath string) (func(w io.Writer) error, error) {
	log.Infof("listing files in archive %s\n", filePath)
	return func (w io.Writer) error {
		var wIface interface{}
		// We can count upon libmagic to give the right mime type and choose the appropriate uncompresser accordingly
		switch mimeType {
		// zip
		case "application/zip":
			log.Infoln("Creating a new zip archiver walker interface")
			wIface = archiver.NewZip()
		case "application/x-rar-compressed":
			log.Infoln("Creating a new rar archiver walker interface")
			wIface = archiver.NewRar()
		case "application/x-tar":
			log.Infoln("Creating a new tar (no compression) archiver walker interface")
			wIface = archiver.NewTar()
		case "application/x-xz":
			// Test file name for maybe it's a tar.xz file
			if compressedTar(filePath) {
				log.Infoln("Creating a new tar xz archiver walker interface")
				wIface = archiver.NewTarXz()
			}
		case "application/x-bzip2":
			// Test file name for maybe it's a tar.bz2 file
			if compressedTar(filePath) {
				log.Infoln("Creating a new tar bz archiver walker interface")
				wIface = archiver.NewTarBz2()
			}
		case "application/gzip":
			// Test file name for maybe it's a tar.gz file
			if compressedTar(filePath) {
				log.Infoln("Creating a new tar gz archiver walker interface")
				wIface = archiver.NewTarGz()
			}
		case "application/x-lz4":
			// Test file name for maybe it's a tar.lz file
			if compressedTar(filePath) {
				log.Infoln("Creating a new tar lz4 archiver walker interface")
				wIface = archiver.NewTarLz4()
			}
		case "application/x-snappy-framed":
			// Test file name for maybe it's a tar.sz file
			if compressedTar(filePath) {
				log.Infoln("Creating a new tar snappy archiver walker interface")
				wIface = archiver.NewTarSz()
			}
		case "application/x-zstd":
			if compressedTar(filePath) {
				log.Infoln("Creating a new tar zstd archiver walker interface")
				wIface = archiver.NewTarZstd()
			}
		// brotli - currently unsupported by libmagic
		// case "application/x-brotli":
			// // This may be a brotli compressed file / tar
			// if compressedTar(filePath) {
				// wIface = archiver.NewTarBrotli()
			// }
		// 7z - currently unsupported by archiver, see https://github.com/mholt/archiver/issues/53
		// case "application/x-7z-compressed":
			// wIface = archiver.New
		}
		walker, ok := wIface.(archiver.Walker)
		if !ok {
			log.Infof("format specified by archive filename (%s) is not a walker format: (%T)", filePath, wIface)
			fmt.Fprintf(w, "%s compressed file\n", mimeType)
			return nil
		} else {
			log.Infof("format specified by archive filename (%s) is: (%T)", filePath, wIface)
			var count int
			header := archiveFileInfo{
				"Permissions",
				"Size",
				"Modification Time",
				"File Name",
			}
			var filesInfo []archiveFileInfo
			var fPermMaxWidth int = 11
			var fSizeMaxWidth int = 4
			var fModtMaxWidth int = 13

			err := walker.Walk(filePath, func(f archiver.File) error {
				fPerm := fmt.Sprintf("%v", f.Mode())
				fSize := humanize.Bytes(uint64(f.Size()))
				fModtS := f.ModTime()
				fModt := fmt.Sprintf("%04d-%02d-%02d %02d:%02d", fModtS.Year(), fModtS.Month(),
				fModtS.Day(), fModtS.Hour(), fModtS.Minute())
				var fName string
				switch h := f.Header.(type) {
				case zip.FileHeader:
					fName = h.Name
				case *tar.Header:
					fName = h.Name
				case *rardecode.FileHeader:
					fName = h.Name
				default:
					// We don't know the full path when another type of archive
					// file is read but we don't need it, as other archive
					// types are not a collection of files but rather a single
					// file compressed.
					fName = f.Name()
				}
				fPermMaxWidth = int(math.Max(float64(fPermMaxWidth), float64(len(fPerm))))
				fSizeMaxWidth = int(math.Max(float64(fSizeMaxWidth), float64(len(fSize))))
				fModtMaxWidth = int(math.Max(float64(fModtMaxWidth), float64(len(fModt))))
				filesInfo = append(filesInfo, archiveFileInfo{
					fPerm,
					fSize,
					fModt,
					fName,
				})
				count++
				return nil
			})
			fPermMaxWidthFmt := fmt.Sprintf("%%-%ds", fPermMaxWidth + 3)
			fSizeMaxWidthFmt := fmt.Sprintf("%%-%ds", fSizeMaxWidth + 3)
			fModtMaxWidthFmt := fmt.Sprintf("%%-%ds", fModtMaxWidth + 3)
			lineFmt := fmt.Sprintf("%s%s%s%%s\n", fPermMaxWidthFmt, fSizeMaxWidthFmt, fModtMaxWidthFmt)
			fmt.Fprintf(w, lineFmt, header.Permissions, header.Size, header.ModifiedTime, header.FileName)
			for _, fMaxWidth := range []int{
				fPermMaxWidth,
				fSizeMaxWidth,
				fModtMaxWidth,
				// File name will always have a 9 = size header line
				9,
			} {
				for i := 0; i <= fMaxWidth; i++ {
					fmt.Fprintf(w, "=")
				}
				fmt.Fprintf(w, "  ")
			}
			fmt.Fprintf(w, "\n")

			for _, fileInfo := range filesInfo {
				fmt.Fprintf(w, lineFmt, fileInfo.Permissions, fileInfo.Size, fileInfo.ModifiedTime, fileInfo.FileName)
			}
			fmt.Fprintf(w, "total %d\n", count)
			return err
		}
	}, nil
}

// utility function to check if a given filepath ends with .tar.*
func compressedTar(filePath string) bool {
	res, _ := regexp.MatchString(`.*\.tar\.`, filePath)
	return res
}
