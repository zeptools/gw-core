package rw

import "io"

type CountWriter struct {
	w io.Writer
	n int64
}

func NewCountWriter(w io.Writer) *CountWriter {
	return &CountWriter{w: w}
}

// Write implements io.Writer
func (cw *CountWriter) Write(p []byte) (int, error) {
	n, err := cw.w.Write(p)
	cw.n += int64(n) // cuz Write() can be called multiple times internally
	return n, err
}

// BytesWritten returns the total number of bytes written
func (cw *CountWriter) BytesWritten() int64 {
	return cw.n
}
