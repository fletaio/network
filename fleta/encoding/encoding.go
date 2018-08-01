package encoding

import (
	"io"
)

// include channel coding and compression

type BaseWriter struct {
	io.Writer
	w         io.Writer
	integrity Integrity
}

func NewWriter(w io.Writer, Integrity Integrity) io.Writer {
	return &BaseWriter{
		w:         w,
		integrity: Integrity,
	}
}

func (encoding *BaseWriter) Write(bs []byte) (int, error) {
	n, err := encoding.w.Write(bs)
	if err != nil {
		return n, err
	}
	encoding.integrity.Update(bs[:n])

	return n, nil
}

type BaseReader struct {
	io.Reader
	r         io.Reader
	integrity Integrity
}

func NewReader(r io.Reader, Integrity Integrity) (io.Reader, error) {
	return &BaseReader{
		r:         r,
		integrity: Integrity,
	}, nil
}

func (encoding *BaseReader) Read(bs []byte) (int, error) {
	n, err := encoding.r.Read(bs)
	if err != nil {
		return n, err
	}
	encoding.integrity.Update(bs[:n])

	return n, nil
}
