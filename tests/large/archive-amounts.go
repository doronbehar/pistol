//go:build ignore

// Generates a .tar.gz archive of millions of empty files. Its listing is huge,
// while the archive itself is relatively small.
package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"os"
	"time"
)

const amount = 2_000_000

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
	for i := 0; i < amount; i++ {
		err := tw.WriteHeader(&tar.Header{
			Typeflag: tar.TypeReg,
			Name:     fmt.Sprintf("dir%03d/file%07d.txt", i%1000, i),
			Mode:     0o644,
			ModTime:  modTime,
		})
		if err != nil {
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
