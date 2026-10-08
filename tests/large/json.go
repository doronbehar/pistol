//go:build ignore

// Generates a ~5 MiB, high entropy, valid JSON file. It's not larger, because
// libmagic detects only JSON files that fit in the first 7 MiB it reads as
// such.
package main

import (
	"bufio"
	"fmt"
	"math/rand/v2"
	"os"
)

const target = 5 << 20

// 90 printable ASCII chars, excluding space, `"` and `\`, which would have to
// be escaped in a JSON string.
const alpha = "!#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[]^_`abcdefghijklmnopqrstuvwxyz{|}~"

func randStr(r *rand.Rand, buf []byte) {
	for i := range buf {
		buf[i] = alpha[r.IntN(len(alpha))]
	}
}

func main() {
	f, err := os.Create(os.Args[1])
	if err != nil {
		panic(err)
	}
	w := bufio.NewWriterSize(f, 1<<20)
	r := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	id, sess, msg := make([]byte, 22), make([]byte, 22), make([]byte, 140)
	written := 0
	n, err := w.WriteString("[\n")
	if err != nil {
		panic(err)
	}
	written += n
	for first := true; written < target; first = false {
		if !first {
			n, err = w.WriteString(",\n")
			if err != nil {
				panic(err)
			}
			written += n
		}
		randStr(r, id)
		randStr(r, sess)
		randStr(r, msg)
		n, err = fmt.Fprintf(w, `  {"request_id": "%s", "session": "%s", "message": "%s"}`, id, sess, msg)
		if err != nil {
			panic(err)
		}
		written += n
	}
	if _, err := w.WriteString("\n]\n"); err != nil {
		panic(err)
	}
	if err := w.Flush(); err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
}
