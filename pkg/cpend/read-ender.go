package cpend

import "io"

//	Accept manually end coping without errors
//	Implement io.ReadWriter
//
type ReadWriteEnder struct {
	Reader io.Reader
	Writer io.Writer
	wr     *io.PipeWriter
}

func NewRW(arw io.ReadWriter) *ReadWriteEnder {
	rd, wr := io.Pipe()
	mrd := io.MultiReader(arw, rd)

	return &ReadWriteEnder{
		Writer: arw,
		Reader: mrd,
		wr:     wr,
	}
}

func (re *ReadWriteEnder) Read(p []byte) (int, error) {
	return re.Reader.Read(p)
}

func (re *ReadWriteEnder) Write(p []byte) (int, error) {
	return re.Writer.Write(p)
}

func (re *ReadWriteEnder) End() error {
	return re.wr.Close()
}

//	Accept manually end coping without errors
//	Implement io.Reader
//
type ReadEnder struct {
	Reader io.Reader
	wr     *io.PipeWriter
}

func NewR(ard io.Reader) *ReadEnder {
	rd, wr := io.Pipe()
	mrd := io.MultiReader(ard, rd)
	return &ReadEnder{
		Reader: mrd,
		wr:     wr,
	}
}

func (re *ReadEnder) Read(p []byte) (int, error) {
	return re.Reader.Read(p)
}

func (re *ReadEnder) End() error {
	return re.wr.Close()
}
