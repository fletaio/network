package packet

import (
	"io"
)

type Payload interface {
	io.Reader
	io.WriterTo
	Len() int
}
