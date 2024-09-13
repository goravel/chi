package chi

import (
	"io"

	"github.com/go-rat/chix"
)

type StreamWriter struct {
	render *chix.Render
	w      io.Writer
}

func NewStreamWriter(render *chix.Render, w io.Writer) *StreamWriter {
	return &StreamWriter{render, w}
}

func (w *StreamWriter) Flush() error {
	w.render.Flush()
	return nil
}

func (w *StreamWriter) Write(data []byte) (int, error) {
	return w.w.Write(data)
}

func (w *StreamWriter) WriteString(s string) (int, error) {
	return w.w.Write([]byte(s))
}
