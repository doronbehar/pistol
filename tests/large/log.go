//go:build ignore

// Generates a ~1 GiB, high entropy file that still looks like a log file, so
// that it barely shrinks when compressed.
package main

import (
	"bufio"
	"fmt"
	"math/rand/v2"
	"os"
)

const target = 1 << 30

// 93 printable ASCII chars, excluding space, `!` and `"`.
const alpha = "#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"

var levels = []string{"INFO ", "WARN ", "ERROR", "DEBUG"}

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
	for written < target {
		randStr(r, id)
		randStr(r, sess)
		randStr(r, msg)
		n, err := fmt.Fprintf(w, "2026-%02d-%02dT%02d:%02d:%02d.%03dZ %s request_id=%s session=%s message=\"%s\"\n",
			1+r.IntN(12), 1+r.IntN(28), r.IntN(24), r.IntN(60), r.IntN(60), r.IntN(1000),
			levels[r.IntN(len(levels))], id, sess, msg)
		if err != nil {
			panic(err)
		}
		written += n
	}
	if err := w.Flush(); err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
}
