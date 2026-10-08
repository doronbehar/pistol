//go:build ignore

// Generates a .tar.gz archive of a few huge files, all zeros, which makes the
// archive itself small, although it has to be decompressed completely in order
// to be listed.
package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"time"
)

const (
	files    = 4
	fileSize = 4 << 30
)

type zeros struct{}

func (zeros) Read(p []byte) (int, error) {
	clear(p)
	return len(p), nil
}

func main() {
	f, err := os.Create(os.Args[1])
	if err != nil {
		panic(err)
	}
	gz, err := gzip.NewWriterLevel(f, gzip.BestSpeed)
	if err != nil {
		panic(err)
	}
	tw := tar.NewWriter(gz)
	modTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < files; i++ {
		err := tw.WriteHeader(&tar.Header{
			Typeflag: tar.TypeReg,
			Name:     fmt.Sprintf("file%d.bin", i),
			Mode:     0o644,
			Size:     fileSize,
			ModTime:  modTime,
		})
		if err != nil {
			panic(err)
		}
		if _, err := io.CopyN(tw, zeros{}, fileSize); err != nil {
			panic(err)
		}
	}
	if err := tw.Close(); err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
}
