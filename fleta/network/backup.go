package network

// func (pr *Reader) Read(bs []byte) (int, error) {
// 	ofs := 0
// 	for {
// 		n, err := pr.reader.Read(bs[ofs:])
// 		if err != nil {
// 			return n, err
// 		}
// 		ofs += n
// 		if ofs >= len(bs) {
// 			break
// 		}
// 	}
// 	pr.read += len(bs)
// 	if !pr.bHeaderRead {
// 		if pr.read >= HeaderSize {
// 			pr.bHeaderRead = true
// 			if pr.Compression == COMPRESSED {
// 				pr.reader = gzip.NewReader(pr.reader)
// 			}
// 		}
// 	}
// }

// type Writer struct {
// 	writer io.Writer
// }

// func NewWriter(w io.Writer) *Writer {
// 	return &Writer{
// 		writer: w,
// 	}
// }

// func (pw *Writer) Write(bs []byte) (int, error) {
// 	n, err := pw.writer.Write(bs)
// 	if err != nil {
// 		return n, err
// 	}
// 	return n, nil
// }
