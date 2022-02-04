package main

import (
	"io"
	"log"
	"time"
)

type prefixWriter struct {
	f func() string
	w io.Writer
}

func (p prefixWriter) Write(b []byte) (n int, err error) {
	if n, err = p.w.Write([]byte(p.f())); err != nil {
		return
	}
	nn, err := p.w.Write(b)
	return n + nn, err
}

func init() {
	log.SetFlags(0)
	log.SetOutput(prefixWriter{
		f: func() string { return time.Now().Format("2006-01-02 15:04:05.00") + " " },
		w: log.Writer(),
	})
}
